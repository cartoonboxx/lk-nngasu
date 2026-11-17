package auth

import (
	"errors"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"university-api/internal/platform/httpx"
)

// Refresh-cookie по документации: HttpOnly; Secure; SameSite=Strict; Path=/api/v1/auth.
// Path ограничивает cookie маршрутами авторизации: к остальным запросам она не прикладывается.
const (
	refreshCookie = "refresh_token"
	cookiePath    = "/api/v1/auth"
	csrfHeader    = "X-CSRF-Token"
)

// Handler — HTTP-слой авторизации: cookie, заголовки CSRF и Origin; логика — в Service.
type Handler struct {
	svc  *Service
	opts Options
}

// loginRequest — LoginInput: вход в активный аккаунт.
type loginRequest struct {
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required"`
	RememberMe bool   `json:"remember_me"` // false — сессия на 1 сутки, true — на 30
}

// activationRequest — ActivationInput: токен из письма и новый пароль.
type activationRequest struct {
	Token    string `json:"token" binding:"required,max=200"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}

// Login — вход. Ответ 200 с Token и Set-Cookie с refresh-токеном.
//
// POST /auth/login
func (h *Handler) Login(c *gin.Context) {
	var req loginRequest
	if !httpx.Bind(c, &req) {
		return
	}
	tok, sess, err := h.svc.Login(c.Request.Context(), req.Email, req.Password, req.RememberMe)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	h.setRefreshCookie(c, sess)
	httpx.OK(c, tok)
}

// CSRF — CSRF-токен текущей refresh-сессии (нужен после перезагрузки страницы).
//
// GET /auth/csrf
func (h *Handler) CSRF(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	tok, err := h.svc.CSRF(c.Request.Context(), refreshFrom(c))
	httpx.Reply(c, tok, err)
}

// Refresh — новый access-токен по refresh-cookie. Нужны X-CSRF-Token и допустимый Origin.
// Ответ 200 с Token и Set-Cookie с новым refresh-токеном.
//
// POST /auth/refresh
func (h *Handler) Refresh(c *gin.Context) {
	if !h.originAllowed(c) {
		httpx.Fail(c, errCSRFInvalid)
		return
	}
	tok, sess, err := h.svc.Refresh(c.Request.Context(), refreshFrom(c), c.GetHeader(csrfHeader))
	if err != nil {
		// Сессии больше нет — убираем и cookie, чтобы браузер не присылал её снова.
		if errors.Is(err, errSessionExpired) || errors.Is(err, errRefreshReused) {
			h.clearRefreshCookie(c)
		}
		httpx.Fail(c, err)
		return
	}
	h.setRefreshCookie(c, sess)
	httpx.OK(c, tok)
}

// Logout — завершить сессию на сервере (её access-токены тоже перестают действовать)
// и удалить cookie. Ответ 204, в том числе если сессия уже закончилась.
//
// POST /auth/logout
func (h *Handler) Logout(c *gin.Context) {
	if !h.originAllowed(c) {
		httpx.Fail(c, errCSRFInvalid)
		return
	}
	if err := h.svc.Logout(c.Request.Context(), refreshFrom(c), c.GetHeader(csrfHeader)); err != nil {
		httpx.Fail(c, err)
		return
	}
	h.clearRefreshCookie(c)
	httpx.NoContent(c)
}

// Activate — задать пароль по приглашению администратора. Ответ 204.
//
// POST /auth/activate
func (h *Handler) Activate(c *gin.Context) {
	var req activationRequest
	if !httpx.Bind(c, &req) {
		return
	}
	httpx.ReplyEmpty(c, h.svc.Activate(c.Request.Context(), req.Token, req.Password))
}

// refreshFrom — значение refresh-cookie или "".
func refreshFrom(c *gin.Context) string {
	v, err := c.Cookie(refreshCookie)
	if err != nil {
		return ""
	}
	return v
}

func (h *Handler) setRefreshCookie(c *gin.Context, s session) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     refreshCookie,
		Value:    s.refreshToken,
		Path:     cookiePath,
		Expires:  s.expiresAt,
		MaxAge:   max(int(time.Until(s.expiresAt).Seconds()), 1),
		HttpOnly: true,
		Secure:   h.opts.CookieSecure,
		SameSite: http.SameSiteStrictMode,
	})
}

func (h *Handler) clearRefreshCookie(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name: refreshCookie, Value: "", Path: cookiePath, MaxAge: -1,
		HttpOnly: true, Secure: h.opts.CookieSecure, SameSite: http.SameSiteStrictMode,
	})
}

// originAllowed — защита от CSRF по заголовку Origin. Браузер всегда присылает Origin
// в POST; он должен совпадать с адресом самого API (фронт и API на одном домене)
// или быть в ALLOWED_ORIGINS. Без Origin (Postman, curl) запрос не блокируется —
// его всё равно защищает X-CSRF-Token.
func (h *Handler) originAllowed(c *gin.Context) bool {
	origin := strings.TrimRight(c.GetHeader("Origin"), "/")
	if origin == "" {
		return true
	}
	if slices.Contains(h.opts.AllowedOrigins, origin) {
		return true
	}
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	host := c.Request.Host
	if fh := c.GetHeader("X-Forwarded-Host"); fh != "" {
		host = fh // за nginx исходный адрес приходит в X-Forwarded-Host
	}
	return u.Host == host
}
