package content

import (
	"github.com/gin-gonic/gin"

	"university-api/internal/platform/httpx"
	"university-api/internal/platform/middleware"
)

type Handler struct{ svc *Service }

// Position не указан — тема/материал встаёт в конец списка (при обновлении — остаётся на месте).
type topicRequest struct {
	Title       string  `json:"title" binding:"required,max=200"`
	Description *string `json:"description"`
	Position    *int    `json:"position" binding:"omitempty,min=1"`
}

type materialRequest struct {
	Title       string  `json:"title" binding:"required,max=200"`
	FileID      *int64  `json:"file_id"`
	ExternalURL *string `json:"external_url"`
	Position    *int    `json:"position" binding:"omitempty,min=1"`
}

func (r topicRequest) input() TopicInput { return TopicInput(r) }

func (r materialRequest) input() MaterialInput { return MaterialInput(r) }

// GET /courses/:id/topics
func (h *Handler) ListTopics(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	list, err := h.svc.ListTopics(c.Request.Context(), middleware.Me(c), id)
	httpx.Reply(c, list, err)
}

// POST /courses/:id/topics
func (h *Handler) CreateTopic(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	var req topicRequest
	if !httpx.Bind(c, &req) {
		return
	}
	t, err := h.svc.CreateTopic(c.Request.Context(), middleware.Me(c), id, req.input())
	httpx.ReplyCreated(c, t, err)
}

// PUT /topics/:id
func (h *Handler) UpdateTopic(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	var req topicRequest
	if !httpx.Bind(c, &req) {
		return
	}
	t, err := h.svc.UpdateTopic(c.Request.Context(), middleware.Me(c), id, req.input())
	httpx.Reply(c, t, err)
}

// DELETE /topics/:id
func (h *Handler) DeleteTopic(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	httpx.ReplyEmpty(c, h.svc.DeleteTopic(c.Request.Context(), middleware.Me(c), id))
}

// POST /topics/:id/materials
func (h *Handler) CreateMaterial(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	var req materialRequest
	if !httpx.Bind(c, &req) {
		return
	}
	m, err := h.svc.CreateMaterial(c.Request.Context(), middleware.Me(c), id, req.input())
	httpx.ReplyCreated(c, m, err)
}

// PUT /materials/:id
func (h *Handler) UpdateMaterial(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	var req materialRequest
	if !httpx.Bind(c, &req) {
		return
	}
	m, err := h.svc.UpdateMaterial(c.Request.Context(), middleware.Me(c), id, req.input())
	httpx.Reply(c, m, err)
}

// DELETE /materials/:id
func (h *Handler) DeleteMaterial(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	httpx.ReplyEmpty(c, h.svc.DeleteMaterial(c.Request.Context(), middleware.Me(c), id))
}
