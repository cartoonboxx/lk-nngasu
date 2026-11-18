package assignments

import (
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"

	"university-api/internal/platform/httpx"
	"university-api/internal/platform/middleware"
)

type Handler struct{ svc *Service }

// assignmentRequest — AssignmentInput. Все ключи обязательны; description и deadline_at
// допускают null (без описания / без срока). Лимит файлов (5) проверяет сервис —
// с кодом too_many_files.
type assignmentRequest struct {
	TopicID     int64               `json:"topic_id" binding:"required,gt=0"`
	Title       string              `json:"title" binding:"required"`
	Description nullable[string]    `json:"description"`
	DeadlineAt  nullable[time.Time] `json:"deadline_at"`
	FileIDs     []int64             `json:"file_ids" binding:"required,max=100,dive,gt=0"`
}

func (r assignmentRequest) input() (AssignmentInput, error) {
	if err := requireKeys(map[string]bool{"description": r.Description.Set, "deadline_at": r.DeadlineAt.Set}); err != nil {
		return AssignmentInput{}, err
	}
	if r.Description.Value != nil && utf8.RuneCountInString(*r.Description.Value) > 20000 {
		return AssignmentInput{}, fieldErr("description", "must be at most 20000 characters")
	}
	return AssignmentInput{TopicID: r.TopicID, Title: r.Title, Description: r.Description.Value,
		DeadlineAt: r.DeadlineAt.Value, FileIDs: r.FileIDs}, nil
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
	in, err := req.input()
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	a, err := h.svc.CreateAssignment(c.Request.Context(), middleware.Me(c), courseID, in)
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
	in, err := req.input()
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	a, err := h.svc.UpdateAssignment(c.Request.Context(), middleware.Me(c), id, in)
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

// gradeFilter читает общие параметры таблиц оценок: status, limit, offset.
func gradeFilter(c *gin.Context) (GradeFilter, bool) {
	var f GradeFilter
	if st := c.Query("status"); st != "" {
		f.Status = &st
	}
	var ok bool
	f.Limit, f.Offset, ok = parsePage(c)
	return f, ok
}

// GET /grades?semester_id=&course_id=&status=&q=&limit=&offset= — оценки студента
func (h *Handler) StudentGrades(c *gin.Context) {
	f, ok := gradeFilter(c)
	if !ok {
		return
	}
	if f.SemesterID, ok = optID(c, "semester_id"); !ok {
		return
	}
	if f.CourseID, ok = optID(c, "course_id"); !ok {
		return
	}
	f.Q = strings.TrimSpace(c.Query("q"))
	list, total, err := h.svc.StudentGrades(c.Request.Context(), middleware.Me(c), f)
	httpx.ReplyPage(c, list, httpx.Meta{Limit: f.Limit, Offset: f.Offset, Total: total}, err)
}

// GET /courses/:id/grades?student_id=&status=&limit=&offset= — результаты курса
func (h *Handler) CourseGrades(c *gin.Context) {
	courseID, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	f, ok := gradeFilter(c)
	if !ok {
		return
	}
	if f.StudentID, ok = optID(c, "student_id"); !ok {
		return
	}
	list, total, err := h.svc.CourseGrades(c.Request.Context(), middleware.Me(c), courseID, f)
	httpx.ReplyPage(c, list, httpx.Meta{Limit: f.Limit, Offset: f.Offset, Total: total}, err)
}
