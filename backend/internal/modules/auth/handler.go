package auth

import (
	"github.com/gin-gonic/gin"

	"university-api/internal/platform/httpx"
)

// loginRequest — тело запроса на вход: почта и пароль.
// Теги binding проверяются при Bind: оба поля обязательны, email должен быть валидным.
type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Handler — HTTP-слой авторизации: принимает запрос, разбирает JSON,
// вызывает Service и отправляет ответ. Бизнес-логики здесь нет.
type Handler struct{ svc *Service }

// Login обрабатывает POST /auth/login.
func (h *Handler) Login(c *gin.Context) {
	// Пустая структура, которую Bind заполнит из JSON-тела.
	var req loginRequest

	// Парсим JSON и проверяем теги binding.
	// При ошибке Bind сам отвечает 400, нам остаётся только выйти.
	if !httpx.Bind(c, &req) {
		return
	}

	// Передаём в сервис стандартный context.Context запроса (не gin.Context).
	// Сервис проверяет пользователя и пароль и возвращает токен или ошибку.
	token, err := h.svc.Login(c.Request.Context(), req.Email, req.Password)

	// Если ошибки нет — 200 с токеном, иначе ошибка переводится в HTTP-статус.
	httpx.Reply(c, token, err)
}
