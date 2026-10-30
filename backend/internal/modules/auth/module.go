// Package auth — вход по email и паролю, выпуск JWT.
package auth

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Register(r *gin.RouterGroup, pool *pgxpool.Pool, secret []byte, ttl time.Duration) {
	h := &Handler{svc: NewService(newPgRepository(pool), secret, ttl)}
	r.POST("/auth/login", h.Login)
}
