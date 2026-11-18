package news

import (
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"

	"university-api/internal/platform/apperr"
	"university-api/internal/platform/httpx"
	"university-api/internal/platform/middleware"
)

type Handler struct{ repo *Repository }

// newsInput — NewsInput: текст и статус (архивирование — отдельной командой PATCH).
type newsInput struct {
	Body   string `json:"body" binding:"required"`
	Status string `json:"status" binding:"required,oneof=draft published"`
}

// stateInput — NewsStateInput.
type stateInput struct {
	Status string `json:"status" binding:"required,oneof=draft published archived"`
}

func fieldErr(kind, field, problem string) error {
	return apperr.NewInvalid(kind, "invalid "+strings.TrimPrefix(kind, "invalid_")).
		WithDetails(map[string]any{"fields": map[string][]string{field: {problem}}})
}

// cleanBody — текст новости: без пробелов по краям, 1–50000 символов.
func cleanBody(s string) (string, error) {
	s = strings.TrimSpace(s)
	if n := utf8.RuneCountInString(s); n == 0 || n > 50000 {
		return "", fieldErr("invalid_request", "body", "must be 1-50000 characters after trim")
	}
	return s, nil
}

// listQuery читает q, limit (1–200, по умолчанию 50) и offset (>= 0).
func listQuery(c *gin.Context) (ListQuery, bool) {
	q := ListQuery{Q: strings.TrimSpace(c.Query("q")), Limit: 50}
	if utf8.RuneCountInString(q.Q) > 200 {
		httpx.Fail(c, fieldErr("invalid_query", "q", "must be at most 200 characters"))
		return q, false
	}
	if raw := c.Query("limit"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 1 || v > 200 {
			httpx.Fail(c, fieldErr("invalid_query", "limit", "must be an integer from 1 to 200"))
			return q, false
		}
		q.Limit = v
	}
	if raw := c.Query("offset"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 0 {
			httpx.Fail(c, fieldErr("invalid_query", "offset", "must be an integer >= 0"))
			return q, false
		}
		q.Offset = v
	}
	return q, true
}

func summaries(list []AdminNews) []NewsSummary {
	out := make([]NewsSummary, len(list))
	for i, n := range list {
		out[i] = n.NewsSummary
	}
	return out
}

// ---- публичная лента ----

// GET /news?q=&limit=&offset= — опубликованные новости, свежие первыми.
func (h *Handler) Feed(c *gin.Context) {
	q, ok := listQuery(c)
	if !ok {
		return
	}
	q.Public = true
	list, total, err := h.repo.List(c.Request.Context(), q)
	httpx.ReplyPage(c, summaries(list), httpx.Meta{Limit: q.Limit, Offset: q.Offset, Total: total}, err)
}

// GET /news/:id — опубликованная новость; черновик, архив и неизвестный id — 404.
func (h *Handler) Show(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	n, err := h.repo.Get(c.Request.Context(), id, true)
	httpx.Reply(c, n.NewsDetail, err)
}

// ---- редактор (администратор) ----

var sorts = map[string]bool{"updated_at": true, "-updated_at": true, "published_at": true, "-published_at": true}
var statuses = map[string]bool{StatusDraft: true, StatusPublished: true, StatusArchived: true}

// GET /admin/news?q=&status=&sort=&limit=&offset=
func (h *Handler) AdminList(c *gin.Context) {
	q, ok := listQuery(c)
	if !ok {
		return
	}
	if st := c.Query("status"); st != "" {
		if !statuses[st] {
			httpx.Fail(c, fieldErr("invalid_query", "status", "must be draft, published or archived"))
			return
		}
		q.Status = &st
	}
	q.Sort = c.DefaultQuery("sort", "-updated_at")
	if !sorts[q.Sort] {
		httpx.Fail(c, fieldErr("invalid_query", "sort", "must be updated_at, -updated_at, published_at or -published_at"))
		return
	}
	list, total, err := h.repo.List(c.Request.Context(), q)
	httpx.ReplyPage(c, list, httpx.Meta{Limit: q.Limit, Offset: q.Offset, Total: total}, err)
}

// GET /admin/news/:id — с черновиками и архивом.
func (h *Handler) AdminGet(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	n, err := h.repo.Get(c.Request.Context(), id, false)
	httpx.Reply(c, n, err)
}

// POST /admin/news — черновик или сразу публикация; автор — текущий администратор.
func (h *Handler) Create(c *gin.Context) {
	var in newsInput
	if !httpx.Bind(c, &in) {
		return
	}
	body, err := cleanBody(in.Body)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	ctx := c.Request.Context()
	id, err := h.repo.Create(ctx, middleware.Me(c).UserID, body, in.Status)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	n, err := h.repo.Get(ctx, id, false)
	httpx.ReplyCreated(c, n, err)
}

// PUT /admin/news/:id — новый текст и статус (draft | published). Правка опубликованной
// новости без смены статуса сохраняет дату публикации. Архивная — 409 news_archived.
func (h *Handler) Replace(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	var in newsInput
	if !httpx.Bind(c, &in) {
		return
	}
	body, err := cleanBody(in.Body)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	ctx := c.Request.Context()
	if err := h.repo.Replace(ctx, id, body, in.Status); err != nil {
		httpx.Fail(c, err)
		return
	}
	n, err := h.repo.Get(ctx, id, false)
	httpx.Reply(c, n, err)
}

// PATCH /admin/news/:id — опубликовать, снять с публикации или архивировать.
func (h *Handler) SetStatus(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	var in stateInput
	if !httpx.Bind(c, &in) {
		return
	}
	ctx := c.Request.Context()
	if err := h.repo.SetStatus(ctx, id, in.Status); err != nil {
		httpx.Fail(c, err)
		return
	}
	n, err := h.repo.Get(ctx, id, false)
	httpx.Reply(c, n, err)
}

// DELETE /admin/news/:id — физически удалить черновик или архивную новость.
func (h *Handler) Delete(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	httpx.ReplyEmpty(c, h.repo.Delete(c.Request.Context(), id))
}
