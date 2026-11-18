// Package realtime — WebSocket-канал личного кабинета (протокол lk-chat.v1).
//
// Подключение в два шага, потому что браузерный WebSocket не умеет слать Authorization:
//  1. POST /api/v1/realtime/tickets с Bearer — одноразовый билет на 30 секунд,
//     привязанный к пользователю, сессии и Origin;
//  2. GET /api/v1/realtime?ticket=<билет> с подпротоколом lk-chat.v1 — Upgrade.
//
// Один сокет на вкладку получает события всех доступных пользователю чатов.
// После session.ready ({user_id, expires_at}) клиент может слать команды
// {type, request_id, data}; ответ — {type:"ack", request_id, data} или
// {type:"error", request_id, error:{code, message, details}}. Ошибка команды соединение
// не закрывает. События — {type, occurred_at, data}, без replay после обрыва.
//
// Коды закрытия: 4401 — истёк допуск (expires_at = срок access-токена) или сессия отозвана
// (проверяется перед каждой командой и раз в PingEvery); 1008 — клиент не успевает читать.
package realtime

import (
	"context"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"

	"university-api/internal/platform/apperr"
	"university-api/internal/platform/httpx"
	"university-api/internal/platform/middleware"
)

const (
	Protocol = "lk-chat.v1"
	Endpoint = "/api/v1/realtime"

	// PingEvery — интервал ping и перепроверки сессии.
	PingEvery = 30 * time.Second
	// CommandTimeout — сколько может выполняться одна команда.
	CommandTimeout = 10 * time.Second
	// RateLimit команд на соединение: не больше rateBurst за rateWindow.
	rateBurst  = 30
	rateWindow = 10 * time.Second
)

var (
	errTicketExpired  = apperr.NewUnauthorized("ticket_expired", "the websocket ticket is unknown or expired")
	errTicketUsed     = apperr.NewUnauthorized("ticket_used", "the websocket ticket has already been used")
	errSessionRevoked = apperr.NewUnauthorized("session_revoked", "the session of this ticket is no longer active")
	errOrigin         = apperr.NewForbidden("origin_forbidden", "this origin is not allowed")
	errNeedWebSocket  = apperr.New(apperr.UpgradeRequired, "websocket_required", "use a WebSocket upgrade with subprotocol "+Protocol)
)

// Sessions — проверка, что сессия не отозвана (реализует auth.Service).
type Sessions interface {
	SessionActive(ctx context.Context, sid string, userID int64) (bool, error)
}

type Options struct {
	AllowedOrigins []string // кроме адреса самого API
}

type Handler struct {
	tickets  *tickets
	sessions Sessions
	opts     Options
	hub      *Hub
}

// Register: билет — за JWT (protected), само подключение — по билету (public).
// Команды и рассылка событий — через hub (его же получают модули, см. chat.RegisterCommands).
func Register(public, protected *gin.RouterGroup, hub *Hub, sessions Sessions, opts Options) {
	h := &Handler{tickets: newTickets(), sessions: sessions, opts: opts, hub: hub}
	protected.POST("/realtime/tickets", h.Ticket)
	public.GET("/realtime", h.Connect)
}

// WebSocketTicket — ответ POST /realtime/tickets.
type WebSocketTicket struct {
	Ticket    string    `json:"ticket"`
	ExpiresAt time.Time `json:"expires_at"`
	Endpoint  string    `json:"endpoint"`
	Protocol  string    `json:"protocol"`
}

// POST /realtime/tickets → 201
func (h *Handler) Ticket(c *gin.Context) {
	origin := normOrigin(c.GetHeader("Origin"))
	if origin != "" && !h.originAllowed(c.Request, origin) {
		httpx.Fail(c, errOrigin)
		return
	}
	id, exp, err := h.tickets.issue(middleware.Me(c), origin)
	c.Header("Cache-Control", "no-store")
	httpx.ReplyCreated(c, WebSocketTicket{Ticket: id, ExpiresAt: exp.UTC(), Endpoint: Endpoint, Protocol: Protocol}, err)
}

// GET /realtime?ticket=… → 101
func (h *Handler) Connect(c *gin.Context) {
	r := c.Request
	if !isUpgrade(r) || !slices.Contains(subprotocols(r), Protocol) {
		c.Header("Upgrade", "websocket")
		httpx.Fail(c, errNeedWebSocket)
		return
	}
	origin := normOrigin(r.Header.Get("Origin"))
	if origin == "" || !h.originAllowed(r, origin) {
		httpx.Fail(c, errOrigin)
		return
	}
	t, err := h.tickets.redeem(r.URL.Query().Get("ticket"))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	if t.origin != "" && t.origin != origin {
		httpx.Fail(c, errOrigin)
		return
	}
	p := t.principal
	active, err := h.sessions.SessionActive(r.Context(), p.Sid, p.UserID)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	if !active {
		httpx.Fail(c, errSessionRevoked)
		return
	}

	// Таймауты http.Server (ReadTimeout 60s) иначе оборвут долгоживущий сокет.
	rc := http.NewResponseController(c.Writer)
	_ = rc.SetReadDeadline(time.Time{})
	_ = rc.SetWriteDeadline(time.Time{})

	// Origin уже проверен выше, поэтому встроенную проверку библиотеки отключаем.
	conn, err := websocket.Accept(c.Writer, r, &websocket.AcceptOptions{
		Subprotocols: []string{Protocol}, InsecureSkipVerify: true,
	})
	if err != nil {
		return // библиотека уже ответила клиенту
	}
	conn.SetReadLimit(64 << 10)
	h.serve(context.WithoutCancel(r.Context()), conn, p)
}

func (h *Handler) serve(ctx context.Context, conn *websocket.Conn, p middleware.Principal) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	defer conn.CloseNow()

	c := &client{conn: conn, p: p, send: make(chan []byte, sendBuffer), cancel: cancel}
	h.hub.add(c)
	defer h.hub.remove(c)

	// Единственный писатель в сокет: кадры уходят в порядке постановки (ack раньше события).
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case frame := <-c.send:
				wctx, done := context.WithTimeout(ctx, 10*time.Second)
				err := conn.Write(wctx, websocket.MessageText, frame)
				done()
				if err != nil {
					cancel()
					return
				}
			}
		}
	}()

	expires := p.Expires
	if expires.IsZero() {
		expires = time.Now().Add(15 * time.Minute)
	}
	c.enqueueJSON(readyFrame{Type: "session.ready", Data: map[string]any{
		"user_id": p.UserID, "expires_at": expires.UTC(),
	}})

	// Срок допуска: по его истечении — 4401, клиент берёт новый билет.
	expiry := time.AfterFunc(time.Until(expires), func() { c.close(closeUnauthorized, "ticket_expired") })
	defer expiry.Stop()

	// Пульс: ping и перепроверка сессии.
	go func() {
		tick := time.NewTicker(PingEvery)
		defer tick.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-tick.C:
			}
			if !h.alive(ctx, c) {
				return
			}
			pctx, done := context.WithTimeout(ctx, 10*time.Second)
			err := conn.Ping(pctx)
			done()
			if err != nil {
				cancel()
				return
			}
		}
	}()

	var windowStart time.Time // окно ограничения частоты команд
	count := 0
	for {
		typ, raw, err := conn.Read(ctx)
		if err != nil {
			return
		}
		if typ != websocket.MessageText {
			c.enqueueJSON(errorOf(nil, errInvalidPayload))
			continue
		}
		if now := time.Now(); now.Sub(windowStart) > rateWindow {
			windowStart, count = now, 0
		}
		count++
		var cmd commandFrame
		if err := Decode(raw, &cmd); err != nil {
			c.enqueueJSON(errorOf(nil, errInvalidPayload))
			continue
		}
		if count > rateBurst {
			wait := rateWindow - time.Since(windowStart)
			c.enqueueJSON(errorOf(cmd.RequestID, errRateLimited.WithDetails(map[string]any{
				"retry_after_ms": max(wait.Milliseconds(), 1)})))
			continue
		}
		h.dispatch(ctx, c, cmd)
	}
}

// alive перепроверяет сессию; отозванная — закрыть 4401.
func (h *Handler) alive(ctx context.Context, c *client) bool {
	active, err := h.sessions.SessionActive(ctx, c.p.Sid, c.p.UserID)
	if err != nil {
		return true // временная ошибка БД не рвёт соединение
	}
	if !active {
		c.close(closeUnauthorized, "session_revoked")
	}
	return active
}

func (h *Handler) dispatch(ctx context.Context, c *client, cmd commandFrame) {
	if cmd.RequestID == nil || !uuidRe.MatchString(*cmd.RequestID) {
		c.enqueueJSON(errorOf(cmd.RequestID, FieldError("request_id", "must be a UUID")))
		return
	}
	fn, ok := h.hub.command(cmd.Type)
	if !ok {
		c.enqueueJSON(errorOf(cmd.RequestID, errUnknownCommand))
		return
	}
	if len(cmd.Data) == 0 || cmd.Data[0] != '{' {
		c.enqueueJSON(errorOf(cmd.RequestID, FieldError("data", "must be an object")))
		return
	}
	if !h.alive(ctx, c) {
		return
	}
	cctx, done := context.WithTimeout(ctx, CommandTimeout)
	ack, after, err := fn(cctx, c.p, cmd.Data)
	done()
	if err != nil {
		c.enqueueJSON(errorOf(cmd.RequestID, err))
		return
	}
	c.enqueueJSON(ackFrame{Type: "ack", RequestID: *cmd.RequestID, Data: ack})
	if after != nil {
		after()
	}
}

// uuidRe — request_id по контракту: UUID (8-4-4-4-12 шестнадцатеричных цифр).
var uuidRe = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

func errorOf(requestID *string, err error) errorFrame {
	return errorFrame{Type: "error", RequestID: requestID, Error: toWSError(err)}
}

func isUpgrade(r *http.Request) bool {
	return r.Method == http.MethodGet &&
		strings.EqualFold(r.Header.Get("Upgrade"), "websocket") &&
		headerHasToken(r.Header.Values("Connection"), "upgrade")
}

func subprotocols(r *http.Request) []string {
	var out []string
	for _, v := range r.Header.Values("Sec-WebSocket-Protocol") {
		for _, p := range strings.Split(v, ",") {
			out = append(out, strings.TrimSpace(p))
		}
	}
	return out
}

func headerHasToken(values []string, token string) bool {
	for _, v := range values {
		for _, t := range strings.Split(v, ",") {
			if strings.EqualFold(strings.TrimSpace(t), token) {
				return true
			}
		}
	}
	return false
}

func normOrigin(o string) string { return strings.TrimRight(strings.TrimSpace(o), "/") }

// originAllowed — как у /auth/refresh: адрес самого API (за nginx — X-Forwarded-Host)
// или один из ALLOWED_ORIGINS.
func (h *Handler) originAllowed(r *http.Request, origin string) bool {
	if slices.Contains(h.opts.AllowedOrigins, origin) {
		return true
	}
	u, err := url.Parse(origin)
	if err != nil || u.Host == "" {
		return false
	}
	host := r.Host
	if fh := r.Header.Get("X-Forwarded-Host"); fh != "" {
		host = fh
	}
	return strings.EqualFold(u.Host, host)
}
