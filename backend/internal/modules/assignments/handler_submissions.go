package assignments

import (
	"fmt"
	"net/url"
	"strconv"
	"unicode/utf8"

	"github.com/gin-gonic/gin"

	"university-api/internal/platform/httpx"
	"university-api/internal/platform/middleware"
)

// submissionRequest — SubmissionInput: все ключи обязательны, answer_text допускает null.
type submissionRequest struct {
	AnswerText nullable[string] `json:"answer_text"`
	FileIDs    []int64          `json:"file_ids" binding:"required,max=100,dive,gt=0"`
	Submit     *bool            `json:"submit" binding:"required"` // указатель: required у bool отверг бы false
}

// reviewRequest — ReviewInput: grade 0–5 с шагом 0.1, comment обязателен как ключ (может быть null).
type reviewRequest struct {
	Grade   *float64         `json:"grade" binding:"required"`
	Comment nullable[string] `json:"comment"`
}

// GET /assignments/:id/submissions?status=&late=&has_work=&limit=&offset=
func (h *Handler) ListSubmissions(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	f := SubmissionFilter{Limit: 50}
	if st := c.Query("status"); st != "" {
		f.Status = &st
	}
	if f.Late, ok = httpx.OptBool(c, "late"); !ok {
		return
	}
	if f.HasWork, ok = httpx.OptBool(c, "has_work"); !ok {
		return
	}
	if f.Limit, f.Offset, ok = parsePage(c); !ok {
		return
	}
	list, total, err := h.svc.ListSubmissions(c.Request.Context(), middleware.Me(c), id, f)
	httpx.ReplyPage(c, list, httpx.Meta{Limit: f.Limit, Offset: f.Offset, Total: total}, err)
}

// GET /assignments/:id/submission — свой ответ (студент)
func (h *Handler) MySubmission(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	s, err := h.svc.MySubmission(c.Request.Context(), middleware.Me(c), id)
	httpx.Reply(c, s, err)
}

// PUT /assignments/:id/submission — сохранить черновик или отправить ответ (студент)
func (h *Handler) SaveSubmission(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	var req submissionRequest
	if !httpx.Bind(c, &req) {
		return
	}
	if err := requireKeys(map[string]bool{"answer_text": req.AnswerText.Set}); err != nil {
		httpx.Fail(c, err)
		return
	}
	if req.AnswerText.Value != nil && utf8.RuneCountInString(*req.AnswerText.Value) > 2000 {
		httpx.Fail(c, fieldErr("answer_text", "must be at most 2000 characters"))
		return
	}
	in := SubmissionInput{AnswerText: req.AnswerText.Value, FileIDs: req.FileIDs, Submit: *req.Submit}
	s, err := h.svc.SaveSubmission(c.Request.Context(), middleware.Me(c), id, in)
	httpx.Reply(c, s, err)
}

// DELETE /assignments/:id/submission — отозвать свой непроверенный ответ (студент)
func (h *Handler) DeleteSubmission(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	httpx.ReplyEmpty(c, h.svc.DeleteSubmission(c.Request.Context(), middleware.Me(c), id))
}

// GET /submissions/:id
func (h *Handler) GetSubmission(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	s, err := h.svc.GetSubmission(c.Request.Context(), middleware.Me(c), id)
	httpx.Reply(c, s, err)
}

// PUT /submissions/:id/review
func (h *Handler) Review(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	var req reviewRequest
	if !httpx.Bind(c, &req) {
		return
	}
	if err := requireKeys(map[string]bool{"comment": req.Comment.Set}); err != nil {
		httpx.Fail(c, err)
		return
	}
	if req.Comment.Value != nil && utf8.RuneCountInString(*req.Comment.Value) > 4000 {
		httpx.Fail(c, fieldErr("comment", "must be at most 4000 characters"))
		return
	}
	s, err := h.svc.Review(c.Request.Context(), middleware.Me(c), id, ReviewInput{Grade: *req.Grade, Comment: req.Comment.Value})
	httpx.Reply(c, s, err)
}

// GET /submissions/:id/archive
func (h *Handler) SubmissionArchive(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	src, err := h.svc.ArchiveOfSubmission(c.Request.Context(), middleware.Me(c), id)
	h.sendArchive(c, src, err)
}

// GET /assignments/:id/submissions/archive
func (h *Handler) AssignmentArchive(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	src, err := h.svc.ArchiveOfAssignment(c.Request.Context(), middleware.Me(c), id)
	h.sendArchive(c, src, err)
}

func (h *Handler) sendArchive(c *gin.Context, src ArchiveSource, err error) {
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	if len(src.Entries) == 0 { // до первого байта ответа можно ещё вернуть нормальную ошибку
		httpx.Fail(c, h.svc.WriteArchive(nil, src))
		return
	}
	c.Header("Content-Type", "application/zip")
	// filename — для старых клиентов, filename*=UTF-8'' — по контракту (RFC 6266/5987).
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename=%q; filename*=UTF-8''%s`,
		src.Name, url.PathEscape(src.Name)))
	c.Status(200)
	// Заголовки уже отправлены: сбой посреди потока можно только оборвать соединение.
	if err := h.svc.WriteArchive(c.Writer, src); err != nil {
		_ = c.Error(err)
	}
}

// parsePage читает limit (1–200, по умолчанию 50) и offset (>= 0).
func parsePage(c *gin.Context) (limit, offset int, ok bool) {
	limit = 50
	if raw := c.Query("limit"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 1 || v > 200 {
			httpx.Fail(c, queryErr("limit", "must be an integer from 1 to 200"))
			return 0, 0, false
		}
		limit = v
	}
	if raw := c.Query("offset"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 0 {
			httpx.Fail(c, queryErr("offset", "must be an integer >= 0"))
			return 0, 0, false
		}
		offset = v
	}
	return limit, offset, true
}

// optID — необязательный id в query: положительное безопасное целое.
func optID(c *gin.Context, name string) (*int64, bool) {
	v, ok := httpx.OptInt64(c, name)
	if !ok {
		return nil, false
	}
	if v != nil && (*v <= 0 || *v > 1<<53-1) {
		httpx.Fail(c, queryErr(name, "must be a positive integer"))
		return nil, false
	}
	return v, true
}
