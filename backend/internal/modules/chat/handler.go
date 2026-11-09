package chat

import (
	"github.com/gin-gonic/gin"

	"university-api/internal/platform/httpx"
	"university-api/internal/platform/middleware"
)

type Handler struct{ svc *Service }

type directRequest struct {
	UserID int64 `json:"user_id" binding:"required,gt=0"`
}

type groupRequest struct {
	GroupID int64 `json:"group_id" binding:"required,gt=0"`
}

type sendRequest struct {
	Body    string  `json:"body" binding:"max=4000"`
	FileIDs []int64 `json:"file_ids" binding:"max=5,dive,gt=0"`
}

type editRequest struct {
	Body string `json:"body" binding:"max=4000"`
}

type readRequest struct {
	MessageID int64 `json:"message_id" binding:"required,gt=0"`
}

// GET /chats
func (h *Handler) List(c *gin.Context) {
	list, err := h.svc.List(c.Request.Context(), middleware.Me(c))
	httpx.Reply(c, list, err)
}

// POST /chats/direct
func (h *Handler) OpenDirect(c *gin.Context) {
	var req directRequest
	if !httpx.Bind(c, &req) {
		return
	}
	ch, err := h.svc.OpenDirect(c.Request.Context(), middleware.Me(c), req.UserID)
	httpx.Reply(c, ch, err)
}

// POST /chats/group
func (h *Handler) OpenGroup(c *gin.Context) {
	var req groupRequest
	if !httpx.Bind(c, &req) {
		return
	}
	ch, err := h.svc.OpenGroup(c.Request.Context(), middleware.Me(c), req.GroupID)
	httpx.Reply(c, ch, err)
}

// GET /chats/:id
func (h *Handler) Get(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	ch, err := h.svc.Get(c.Request.Context(), middleware.Me(c), id)
	httpx.Reply(c, ch, err)
}

// GET /chats/:id/messages?before=<id>|after=<id>&limit=50
func (h *Handler) Messages(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	before, ok := httpx.OptInt64(c, "before")
	if !ok {
		return
	}
	after, ok := httpx.OptInt64(c, "after")
	if !ok {
		return
	}
	limit, _ := httpx.Page(c)
	list, err := h.svc.Messages(c.Request.Context(), middleware.Me(c), id, before, after, limit)
	httpx.Reply(c, list, err)
}

// POST /chats/:id/messages
func (h *Handler) Send(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	var req sendRequest
	if !httpx.Bind(c, &req) {
		return
	}
	m, err := h.svc.Send(c.Request.Context(), middleware.Me(c), id, req.Body, req.FileIDs)
	httpx.ReplyCreated(c, m, err)
}

// PATCH /chats/:id/messages/:mid
func (h *Handler) Edit(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	mid, ok := httpx.ID(c, "mid")
	if !ok {
		return
	}
	var req editRequest
	if !httpx.Bind(c, &req) {
		return
	}
	m, err := h.svc.Edit(c.Request.Context(), middleware.Me(c), id, mid, req.Body)
	httpx.Reply(c, m, err)
}

// DELETE /chats/:id/messages/:mid
func (h *Handler) DeleteMessage(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	mid, ok := httpx.ID(c, "mid")
	if !ok {
		return
	}
	httpx.ReplyEmpty(c, h.svc.Delete(c.Request.Context(), middleware.Me(c), id, mid))
}

// POST /chats/:id/read
func (h *Handler) Read(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	var req readRequest
	if !httpx.Bind(c, &req) {
		return
	}
	httpx.ReplyEmpty(c, h.svc.MarkRead(c.Request.Context(), middleware.Me(c), id, req.MessageID))
}
