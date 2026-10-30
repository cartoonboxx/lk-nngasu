package middleware

import (
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
type Claims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

// Principal — кто делает запрос. Кладётся в gin.Context.
type Principal struct {
	UserID int64
	Role   string
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

func Auth(secret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, ok := strings.CutPrefix(c.GetHeader("Authorization"), "Bearer ")
		if !ok || raw == "" {
			httpx.Fail(c, apperr.New(apperr.Unauthorized, "missing_token", "authorization header required"))
			return
		}

		claims := &Claims{}
		tok, err := jwt.ParseWithClaims(raw, claims,
			func(*jwt.Token) (any, error) { return secret, nil },
			jwt.WithValidMethods([]string{"HS256"}),
		)
		if err != nil || !tok.Valid {
			httpx.Fail(c, apperr.New(apperr.Unauthorized, "invalid_token", "token is invalid or expired"))
			return
		}

		id, err := strconv.ParseInt(claims.Subject, 10, 64)
		if err != nil {
			httpx.Fail(c, apperr.New(apperr.Unauthorized, "invalid_token", "bad subject"))
			return
		}

		c.Set(principalKey, Principal{UserID: id, Role: claims.Role})
		c.Next()
	}
}

// RequireRole пропускает только перечисленные роли. Ставится после Auth.
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		p, ok := PrincipalFrom(c)
		if !ok {
			httpx.Fail(c, apperr.New(apperr.Unauthorized, "missing_token", "authentication required"))
			return
		}
		if !slices.Contains(roles, p.Role) {
			httpx.Fail(c, apperr.New(apperr.Forbidden, "forbidden", "insufficient permissions"))
			return
		}
		c.Next()
	}
}
