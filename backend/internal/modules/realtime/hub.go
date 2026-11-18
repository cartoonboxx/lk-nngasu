package realtime

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/coder/websocket"

	"university-api/internal/platform/apperr"
	"university-api/internal/platform/middleware"
)

// CommandFunc обрабатывает команду клиента. ack — data подтверждения; after (может быть nil)
// вызывается после отправки ack — там модуль рассылает события (уже после COMMIT).
type CommandFunc func(ctx context.Context, p middleware.Principal, data json.RawMessage) (ack any, after func(), err error)

// Hub — подключённые сокеты по пользователям и зарегистрированные команды.
// Модули (chat) регистрируют команды через Handle и рассылают события через Publish.
type Hub struct {
	mu    sync.RWMutex
	conns map[int64]map[*client]struct{}
	cmds  map[string]CommandFunc
	now   func() time.Time
}

func NewHub() *Hub {
	return &Hub{conns: map[int64]map[*client]struct{}{}, cmds: map[string]CommandFunc{}, now: time.Now}
}

// Handle регистрирует команду (при старте, до приёма соединений).
func (h *Hub) Handle(typ string, fn CommandFunc) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.cmds[typ] = fn
}

func (h *Hub) command(typ string) (CommandFunc, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	fn, ok := h.cmds[typ]
	return fn, ok
}

// Publish отправляет событие {type, occurred_at, data} во все сокеты перечисленных пользователей.
// Получателей вычисляет модуль по текущему членству; событий во время отключения не копим.
func (h *Hub) Publish(userIDs []int64, typ string, data any) {
	frame, err := json.Marshal(eventFrame{Type: typ, OccurredAt: h.now().UTC(), Data: data})
	if err != nil {
		slog.Error("realtime: marshal event", "type", typ, "err", err)
		return
	}
	h.mu.RLock()
	var targets []*client
	for _, id := range userIDs {
		for c := range h.conns[id] {
			targets = append(targets, c)
		}
	}
	h.mu.RUnlock()
	for _, c := range targets {
		c.enqueue(frame)
	}
}

func (h *Hub) add(c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	set := h.conns[c.p.UserID]
	if set == nil {
		set = map[*client]struct{}{}
		h.conns[c.p.UserID] = set
	}
	set[c] = struct{}{}
}

func (h *Hub) remove(c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if set := h.conns[c.p.UserID]; set != nil {
		delete(set, c)
		if len(set) == 0 {
			delete(h.conns, c.p.UserID)
		}
	}
}

// ---- кадры ----

type eventFrame struct {
	Type       string    `json:"type"`
	OccurredAt time.Time `json:"occurred_at"`
	Data       any       `json:"data"`
}

type readyFrame struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

type ackFrame struct {
	Type      string `json:"type"`
	RequestID string `json:"request_id"`
	Data      any    `json:"data"`
}

type errorFrame struct {
	Type      string  `json:"type"`
	RequestID *string `json:"request_id"`
	Error     wsError `json:"error"`
}

type wsError struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details"`
}

// commandFrame — WsCommandEnvelope.
type commandFrame struct {
	Type      string          `json:"type"`
	RequestID *string         `json:"request_id"`
	Data      json.RawMessage `json:"data"`
}

var (
	errInvalidPayload = apperr.NewInvalid("invalid_payload", "Некорректный кадр.")
	errUnknownCommand = apperr.NewInvalid("unknown_command", "Неизвестная команда.")
	errRateLimited    = apperr.New(apperr.TooManyRequests, "rate_limited", "Слишком много команд, повторите позже.")
	errInternal       = apperr.New(apperr.Internal, "internal_error", "Внутренняя ошибка сервера.")
)

// Decode разбирает data команды строго: неизвестные поля и неверные типы — invalid_payload.
func Decode(raw json.RawMessage, dst any) error {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		var te *json.UnmarshalTypeError
		if errors.As(err, &te) && te.Field != "" {
			return FieldError(te.Field, "has a wrong type")
		}
		return errInvalidPayload.WithDetails(map[string]any{"reason": err.Error()})
	}
	if dec.More() {
		return errInvalidPayload
	}
	return nil
}

// FieldError — invalid_payload с указанием поля: details.fields как в HTTP-ошибках валидации.
func FieldError(field, problem string) error {
	return errInvalidPayload.WithDetails(map[string]any{"fields": map[string][]string{field: {problem}}})
}

func toWSError(err error) wsError {
	var ae *apperr.Error
	if errors.As(err, &ae) && ae.Kind != apperr.Internal && ae.Kind != apperr.Unavailable {
		return wsError{Code: ae.Code, Message: ae.Message, Details: ae.Details}
	}
	slog.Error("realtime: command failed", "err", err)
	return wsError{Code: errInternal.Code, Message: errInternal.Message}
}

// ---- соединение ----

// Коды закрытия из контракта.
const (
	closeUnauthorized websocket.StatusCode = 4401 // допуск истёк, сессия отозвана или аккаунт заблокирован
	closeForbidden    websocket.StatusCode = 4403 // доступ запрещён
)

// sendBuffer — сколько кадров может ждать отправки; медленный клиент сверх этого отключается.
const sendBuffer = 64

type client struct {
	conn   *websocket.Conn
	p      middleware.Principal
	send   chan []byte
	cancel context.CancelFunc
	once   sync.Once
}

func (c *client) enqueue(frame []byte) {
	select {
	case c.send <- frame:
	default:
		c.close(websocket.StatusPolicyViolation, "slow_consumer")
	}
}

func (c *client) enqueueJSON(v any) {
	b, err := json.Marshal(v)
	if err != nil {
		slog.Error("realtime: marshal frame", "err", err)
		return
	}
	c.enqueue(b)
}

// close закрывает сокет с кодом; повторные вызовы ничего не делают.
func (c *client) close(code websocket.StatusCode, reason string) {
	c.once.Do(func() {
		go func() {
			_ = c.conn.Close(code, reason)
			c.cancel()
		}()
	})
}
