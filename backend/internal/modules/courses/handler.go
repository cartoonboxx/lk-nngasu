package courses

import (
	"github.com/gin-gonic/gin"

	"university-api/internal/platform/httpx"
	"university-api/internal/platform/middleware"
)

// Handler: чтение с проверкой доступа идёт через сервис, административные записи — прямо в репозиторий
// (правило там одно — RequireRole(admin) на маршруте).
type Handler struct {
	svc  *Service
	repo *Repository
}

type createCourseRequest struct {
	SubjectID  int64 `json:"subject_id" binding:"required"`
	GroupID    int64 `json:"group_id" binding:"required"`
	SemesterID int64 `json:"semester_id" binding:"required"`
}

type workloadRequest struct {
	TeacherID    int64    `json:"teacher_id" binding:"required"`
	LessonType   string   `json:"lesson_type" binding:"required,oneof=lecture practice laboratory consultation exam"`
	PlannedHours *float64 `json:"planned_hours" binding:"omitempty,min=0"`
}

// GET /courses?group_id=&semester_id=&current=true
func (h *Handler) ListCourses(c *gin.Context) {
	groupID, ok := httpx.OptInt64(c, "group_id")
	if !ok {
		return
	}
	semID, ok := httpx.OptInt64(c, "semester_id")
	if !ok {
		return
	}
	current, ok := httpx.OptBool(c, "current")
	if !ok {
		return
	}
	list, err := h.svc.List(c.Request.Context(), middleware.Me(c), CourseFilter{
		GroupID: groupID, SemesterID: semID, CurrentSemester: current != nil && *current,
	})
	httpx.Reply(c, list, err)
}

func (h *Handler) GetCourse(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	d, err := h.svc.Get(c.Request.Context(), middleware.Me(c), id)
	httpx.Reply(c, d, err)
}

func (h *Handler) CreateCourse(c *gin.Context) {
	var req createCourseRequest
	if !httpx.Bind(c, &req) {
		return
	}
	id, err := h.repo.CreateCourse(c.Request.Context(), req.SubjectID, req.GroupID, req.SemesterID)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	d, err := h.svc.Get(c.Request.Context(), middleware.Me(c), id)
	httpx.ReplyCreated(c, d, err)
}

func (h *Handler) DeleteCourse(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	httpx.ReplyEmpty(c, h.repo.DeleteCourse(c.Request.Context(), id))
}

// GET /courses/:id/students
func (h *Handler) ListStudents(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	list, err := h.svc.Students(c.Request.Context(), middleware.Me(c), id)
	httpx.Reply(c, list, err)
}

func (h *Handler) AddWorkload(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	var req workloadRequest
	if !httpx.Bind(c, &req) {
		return
	}
	wid, err := h.repo.AddWorkload(c.Request.Context(), id, req.TeacherID, req.LessonType, req.PlannedHours)
	httpx.ReplyCreated(c, gin.H{"id": wid}, err)
}

func (h *Handler) DeleteWorkload(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	httpx.ReplyEmpty(c, h.repo.DeleteWorkload(c.Request.Context(), id))
}
