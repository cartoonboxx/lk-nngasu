package chat

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"

	"university-api/internal/platform/apperr"
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

func fieldErr(kind, field, problem string) error {
	return apperr.NewInvalid(kind, "invalid "+strings.TrimPrefix(kind, "invalid_")).
		WithDetails(map[string]any{"fields": map[string][]string{field: {problem}}})
}

// intParam читает необязательный целый параметр запроса в диапазоне [lo, hi].
func intParam(c *gin.Context, name string, def, lo, hi int) (int, bool) {
	raw := c.Query(name)
	if raw == "" {
		return def, true
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v < lo || v > hi {
		problem := fmt.Sprintf("must be an integer from %d to %d", lo, hi)
		if hi == math.MaxInt {
			problem = fmt.Sprintf("must be an integer >= %d", lo)
		}
		httpx.Fail(c, fieldErr("invalid_query", name, problem))
		return 0, false
	}
	return v, true
}

// listQuery — q (до 200 символов), limit 1–200 (по умолчанию 50), offset >= 0 — общие правила списков.
func listQuery(c *gin.Context) (ListQuery, bool) {
	q := ListQuery{Q: strings.TrimSpace(c.Query("q"))}
	if utf8.RuneCountInString(q.Q) > 200 {
		httpx.Fail(c, fieldErr("invalid_query", "q", "must be at most 200 characters"))
		return q, false
	}
	var ok bool
	if q.Limit, ok = intParam(c, "limit", 50, 1, 200); !ok {
		return q, false
	}
	if q.Offset, ok = intParam(c, "offset", 0, 0, math.MaxInt); !ok {
		return q, false
	}
	return q, true
}

// GET /contacts?q=&limit=&offset=
func (h *Handler) Contacts(c *gin.Context) {
	q, ok := listQuery(c)
	if !ok {
		return
	}
	list, total, err := h.svc.Contacts(c.Request.Context(), middleware.Me(c), q)
	httpx.ReplyPage(c, list, httpx.Meta{Limit: q.Limit, Offset: q.Offset, Total: total}, err)
}

// GET /chats?q=&limit=&offset=
func (h *Handler) List(c *gin.Context) {
	q, ok := listQuery(c)
	if !ok {
		return
	}
	list, total, err := h.svc.List(c.Request.Context(), middleware.Me(c), q)
	httpx.ReplyPage(c, list, httpx.Meta{Limit: q.Limit, Offset: q.Offset, Total: total}, err)
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

// GET /chats/:id/messages?before=<id>&limit=50
func (h *Handler) Messages(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	var before *int64
	if raw := c.Query("before"); raw != "" {
		v, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || v <= 0 {
			httpx.Fail(c, fieldErr("invalid_cursor", "before", "must be a positive message id"))
			return
		}
		before = &v
	}
	limit, ok := intParam(c, "limit", DefaultLimit, 1, MaxLimit)
	if !ok {
		return
	}
	page, err := h.svc.Messages(c.Request.Context(), middleware.Me(c), id, before, limit)
	httpx.Reply(c, page, err)
}

// GET /chats/:id/participants
func (h *Handler) Participants(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	list, err := h.svc.Participants(c.Request.Context(), middleware.Me(c), id)
	httpx.Reply(c, list, err)
}
