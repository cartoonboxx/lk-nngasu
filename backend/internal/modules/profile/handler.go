package profile

import (
	"github.com/gin-gonic/gin"

	"university-api/internal/platform/httpx"
	"university-api/internal/platform/middleware"
)

// Handler — HTTP-слой профиля. Бизнес-логики здесь нет: только вызов сервиса и ответ.
type Handler struct{ svc *Service }

// GET /auth/me — профиль текущего пользователя для личного кабинета.
// Ответ 200 с Profile; пользователь удалён или неактивен — 401.
func (h *Handler) Me(c *gin.Context) {
	// id пользователя берём из JWT (middleware.Me), а не из запроса,
	// поэтому получить так чужой профиль нельзя.
	// В сервис передаём стандартный context.Context запроса, а не gin.Context.
	p, err := h.svc.Me(c.Request.Context(), middleware.Me(c).UserID)

	// Если ошибки нет — 200 с профилем, иначе ошибка переводится в HTTP-статус.
	httpx.Reply(c, p, err)
}
