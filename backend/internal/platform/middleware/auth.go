package middleware

import (
	"context"
	"errors"
	"slices"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"university-api/internal/platform/apperr"
	"university-api/internal/platform/httpx"
)

const (
	RoleStudent = "student"
	RoleTeacher = "teacher"
	RoleAdmin   = "admin"
)

// Claims — содержимое access-токена. Subject = id пользователя.
// Sid — id сессии входа: по нему сервер отзывает access-токены при logout,
// повторном использовании refresh-токена или блокировке пользователя.
type Claims struct {
	Role string `json:"role"`
	Sid  string `json:"sid"`
	jwt.RegisteredClaims
}

// SessionChecker проверяет, что сессия токена ещё действует: не отозвана,
// не истекла и пользователь не заблокирован (реализует auth.Service).
// Возвращает false, если токен больше не должен приниматься.
type SessionChecker interface {
	SessionActive(ctx context.Context, sid string, userID int64) (bool, error)
}

// Коды 401 по документации API.
var (
	errUnauthorized   = apperr.NewUnauthorized("unauthorized", "authentication required")
	errTokenExpired   = apperr.NewUnauthorized("token_expired", "access token has expired, call /auth/refresh")
	errSessionExpired = apperr.NewUnauthorized("session_expired", "session has ended, sign in again")
)

// Principal — кто делает запрос. Кладётся в gin.Context.
type Principal struct {
	UserID int64
	Role   string
	Sid    string // id сессии из access-токена: нужен, чтобы не закрыть текущую сессию при смене пароля
}

const principalKey = "principal"

func PrincipalFrom(c *gin.Context) (Principal, bool) {
	v, ok := c.Get(principalKey)
	if !ok {
		return Principal{}, false
	}
	p, ok := v.(Principal)
	return p, ok
}

// Auth проверяет access-токен из заголовка Authorization: Bearer <token>.
//
//	нет токена или он подделан — 401 unauthorized;
//	срок токена истёк          — 401 token_expired (клиент вызывает /auth/refresh);
//	сессия отозвана/истекла
//	или пользователь заблокирован — 401 session_expired (клиент открывает вход).
//
// Сессия проверяется по БД на каждом запросе: так logout и блокировка действуют
// сразу, а не когда истечёт уже выданный access-токен.
func Auth(secret []byte, sessions SessionChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, ok := strings.CutPrefix(c.GetHeader("Authorization"), "Bearer ")
		if !ok || raw == "" {
			httpx.Fail(c, errUnauthorized)
			return
		}

		claims := &Claims{}
		tok, err := jwt.ParseWithClaims(raw, claims,
			func(*jwt.Token) (any, error) { return secret, nil },
			jwt.WithValidMethods([]string{"HS256"}),
		)
		if errors.Is(err, jwt.ErrTokenExpired) {
			httpx.Fail(c, errTokenExpired)
			return
		}
		if err != nil || !tok.Valid || claims.Sid == "" {
			httpx.Fail(c, errUnauthorized)
			return
		}

		id, err := strconv.ParseInt(claims.Subject, 10, 64)
		if err != nil {
			httpx.Fail(c, errUnauthorized)
			return
		}

		active, err := sessions.SessionActive(c.Request.Context(), claims.Sid, id)
		if err != nil {
			httpx.Fail(c, err)
			return
		}
		if !active {
			httpx.Fail(c, errSessionExpired)
			return
		}

		// Ответы защищённых маршрутов — приватные данные: не кешировать (контракт, раздел «Заголовки»).
		c.Header("Cache-Control", "no-store")
		c.Set(principalKey, Principal{UserID: id, Role: claims.Role, Sid: claims.Sid})
		c.Next()
	}
}

// RequireRole пропускает только перечисленные роли. Ставится после Auth.
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		p, ok := PrincipalFrom(c)
		if !ok {
			httpx.Fail(c, errUnauthorized)
			return
		}
		if !slices.Contains(roles, p.Role) {
			httpx.Fail(c, apperr.New(apperr.Forbidden, "forbidden", "insufficient permissions"))
			return
		}
		c.Next()
	}
}
