package auth

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"university-api/internal/platform/ratelimit"
)

// Register собирает модуль авторизации и регистрирует его маршруты в публичной группе r
// (все они работают без access-токена). Возвращает сервис: app передаёт его
// в middleware.Auth (проверка сессий) и в users (отправка приглашений).
//
// limiter ограничивает подбор паролей и токенов: 429 rate_limited с Retry-After.
func Register(r *gin.RouterGroup, pool *pgxpool.Pool, opts Options, limiter *ratelimit.Limiter) *Service {
	svc := NewService(newRepository(pool), opts)
	h := &Handler{svc: svc, opts: opts}

	r.GET("/auth/setup", h.SetupStatus)
	r.POST("/auth/setup", limiter.Middleware("setup", 5, 15*time.Minute, (*gin.Context).RemoteIP), h.Setup)
	r.POST("/auth/login", limiter.Middleware("login", 10, time.Minute), h.Login) // вход
	r.GET("/auth/csrf", h.CSRF)                                                  // CSRF-токен сессии
	r.POST("/auth/refresh", h.Refresh)                                           // новый access-токен
	r.POST("/auth/logout", h.Logout)                                             // завершить сессию
	r.POST("/auth/activate", limiter.Middleware("activate", 10, time.Minute), h.Activate)

	return svc
}
