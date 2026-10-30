package assignments

import (
	"mime"

	"github.com/gin-gonic/gin"

	"university-api/internal/platform/httpx"
	"university-api/internal/platform/middleware"
)

type submissionRequest struct {
	AnswerText *string `json:"answer_text"`
	FileIDs    []int64 `json:"file_ids" binding:"max=5,dive,gt=0"`
	Submit     *bool   `json:"submit"` // по умолчанию true
}

type reviewRequest struct {
	Grade   *float64 `json:"grade" binding:"required"`
	Comment *string  `json:"comment"`
}

// GET /assignments/:id/submissions?status=&late=&has_work=
func (h *Handler) ListSubmissions(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	var f SubmissionFilter
	if st := c.Query("status"); st != "" {
		f.Status = &st
	}
	if f.Late, ok = httpx.OptBool(c, "late"); !ok {
		return
	}
	if f.HasWork, ok = httpx.OptBool(c, "has_work"); !ok {
		return
	}
	list, err := h.svc.ListSubmissions(c.Request.Context(), middleware.Me(c), id, f)
	httpx.Reply(c, list, err)
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
	in := SubmissionInput{AnswerText: req.AnswerText, FileIDs: req.FileIDs, Submit: true}
	if req.Submit != nil {
		in.Submit = *req.Submit
	}
	s, err := h.svc.SaveSubmission(c.Request.Context(), middleware.Me(c), id, in)
	httpx.Reply(c, s, err)
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
	s, err := h.svc.Review(c.Request.Context(), middleware.Me(c), id, ReviewInput{Grade: *req.Grade, Comment: req.Comment})
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
	c.Header("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": src.Name}))
	c.Status(200)
	// Заголовки уже отправлены: сбой посреди потока можно только оборвать соединение.
	if err := h.svc.WriteArchive(c.Writer, src); err != nil {
		_ = c.Error(err)
	}
}
