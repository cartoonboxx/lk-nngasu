package assignments

import (
	"time"

	"github.com/gin-gonic/gin"

	"university-api/internal/platform/httpx"
	"university-api/internal/platform/middleware"
)

type Handler struct{ svc *Service }

type assignmentRequest struct {
	TopicID     int64     `json:"topic_id" binding:"required,gt=0"`
	Title       string    `json:"title" binding:"required,max=255"`
	Description *string   `json:"description"`
	DeadlineAt  time.Time `json:"deadline_at" binding:"required"`
	FileIDs     []int64   `json:"file_ids" binding:"max=5,dive,gt=0"`
}

func (r assignmentRequest) input() AssignmentInput {
	return AssignmentInput{TopicID: r.TopicID, Title: r.Title, Description: r.Description,
		DeadlineAt: r.DeadlineAt, FileIDs: r.FileIDs}
}

// GET /courses/:id/assignments
func (h *Handler) List(c *gin.Context) {
	courseID, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	list, err := h.svc.ListAssignments(c.Request.Context(), middleware.Me(c), courseID)
	httpx.Reply(c, list, err)
}

// POST /courses/:id/assignments
func (h *Handler) Create(c *gin.Context) {
	courseID, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	var req assignmentRequest
	if !httpx.Bind(c, &req) {
		return
	}
	a, err := h.svc.CreateAssignment(c.Request.Context(), middleware.Me(c), courseID, req.input())
	httpx.ReplyCreated(c, a, err)
}

// GET /assignments/:id
func (h *Handler) Get(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	a, err := h.svc.GetAssignment(c.Request.Context(), middleware.Me(c), id)
	httpx.Reply(c, a, err)
}

// PUT /assignments/:id
func (h *Handler) Update(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	var req assignmentRequest
	if !httpx.Bind(c, &req) {
		return
	}
	a, err := h.svc.UpdateAssignment(c.Request.Context(), middleware.Me(c), id, req.input())
	httpx.Reply(c, a, err)
}

// DELETE /assignments/:id
func (h *Handler) Delete(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	httpx.ReplyEmpty(c, h.svc.DeleteAssignment(c.Request.Context(), middleware.Me(c), id))
}

// GET /me/grades?course_id=
func (h *Handler) MyGrades(c *gin.Context) {
	courseID, ok := httpx.OptInt64(c, "course_id")
	if !ok {
		return
	}
	list, err := h.svc.MyGrades(c.Request.Context(), middleware.Me(c), courseID)
	httpx.Reply(c, list, err)
}

// GET /courses/:id/grades
func (h *Handler) CourseGrades(c *gin.Context) {
	courseID, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	list, err := h.svc.CourseGrades(c.Request.Context(), middleware.Me(c), courseID)
	httpx.Reply(c, list, err)
}
