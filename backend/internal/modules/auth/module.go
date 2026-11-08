// Package auth — вход по email и паролю, выпуск JWT.
package auth

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Register собирает зависимости модуля auth и регистрирует его маршруты в группе r.
//
// pool   — общий пул подключений к Postgres, создаётся один раз в main;
// secret — ключ для подписи JWT;
// ttl    — время жизни выдаваемого токена.
func Register(r *gin.RouterGroup, pool *pgxpool.Pool, secret []byte, ttl time.Duration) {
	// Сборка цепочки: репозиторий (работа с БД) → сервис (бизнес-логика) → handler (HTTP).
	h := &Handler{svc: NewService(newPgRepository(pool), secret, ttl)}

	// POST <префикс группы>/auth/login → Handler.Login
	r.POST("/auth/login", h.Login)
}
