package courses

import (
	"math"
	"strconv"

	"github.com/gin-gonic/gin"

	"university-api/internal/platform/apperr"
	"university-api/internal/platform/httpx"
	"university-api/internal/platform/middleware"
)

// Handler: чтение с проверкой доступа идёт через сервис, административные записи — прямо в репозиторий
// (правило там одно — RequireRole(admin) на маршруте).
type Handler struct {
	svc  *Service
	repo *Repository
}

// createCourseRequest — CourseInput: действующие дисциплина, группа и семестр.
type createCourseRequest struct {
	SubjectID  int64 `json:"subject_id" binding:"required,gt=0"`
	GroupID    int64 `json:"group_id" binding:"required,gt=0"`
	SemesterID int64 `json:"semester_id" binding:"required,gt=0"`
}

// workloadRequest — WorkloadInput. planned_hours — указатель, чтобы 0 был допустим,
// а отсутствие поля — нет (required у числа отверг бы 0).
type workloadRequest struct {
	TeacherID    int64    `json:"teacher_id" binding:"required,gt=0"`
	LessonType   string   `json:"lesson_type" binding:"required,oneof=lecture practice laboratory consultation exam"`
	PlannedHours *float64 `json:"planned_hours" binding:"required,min=0"`
}

// validHours — шаг 0.25 академического часа и разумный верхний предел (numeric(7,2)).
func validHours(h float64) bool {
	return h >= 0 && h < 100000 && math.Abs(h*4-math.Round(h*4)) < 1e-9
}

// failQuery — 400 invalid_query с указанием параметра.
func failQuery(c *gin.Context, name, problem string) {
	httpx.Fail(c, apperr.NewInvalid("invalid_query", "invalid query parameters").
		WithDetails(map[string]any{"fields": map[string][]string{name: {problem}}}))
}

// ListCourses — страница доступных пользователю курсов (Page<Course>).
// current=true — только текущий семестр (страница «Предметы»); с semester_id не сочетается.
//
// GET /courses?group_id=&semester_id=&current=&limit=&offset=
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
	if semID != nil && current != nil && *current {
		failQuery(c, "current", "cannot be combined with semester_id")
		return
	}
	f := CourseFilter{GroupID: groupID, SemesterID: semID, CurrentSemester: current != nil && *current, Limit: 50}
	if raw := c.Query("limit"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 1 || v > 200 {
			failQuery(c, "limit", "must be an integer from 1 to 200")
			return
		}
		f.Limit = v
	}
	if raw := c.Query("offset"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 0 {
			failQuery(c, "offset", "must be an integer >= 0")
			return
		}
		f.Offset = v
	}
	list, total, err := h.svc.List(c.Request.Context(), middleware.Me(c), f)
	httpx.ReplyPage(c, list, httpx.Meta{Limit: f.Limit, Offset: f.Offset, Total: total}, err)
}

// Subjects — предметы текущего семестра для ЛК: тот же список, что /courses?current=true,
// но массивом (маршрут из прежнего контракта фронта; лимит 200).
//
// GET /subjects
func (h *Handler) Subjects(c *gin.Context) {
	list, _, err := h.svc.List(c.Request.Context(), middleware.Me(c), CourseFilter{CurrentSemester: true, Limit: 200})
	httpx.Reply(c, list, err)
}

// GetCourse — страница предмета: курс, нагрузка и темы с материалами и заданиями.
//
// GET /courses/:id
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

// replaceSubjectRequest — CourseReplacement: новая действующая дисциплина.
type replaceSubjectRequest struct {
	SubjectID int64 `json:"subject_id" binding:"required,gt=0"`
}

// ReplaceSubject меняет дисциплину у курса без пересоздания нагрузки, занятий,
// тем, материалов и заданий. Группа и семестр не меняются. Ответ — CourseDetail.
//
// PATCH /courses/:id
func (h *Handler) ReplaceSubject(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	var req replaceSubjectRequest
	if !httpx.Bind(c, &req) {
		return
	}
	if err := h.repo.ReplaceSubject(c.Request.Context(), id, req.SubjectID); err != nil {
		httpx.Fail(c, err)
		return
	}
	d, err := h.svc.Get(c.Request.Context(), middleware.Me(c), id)
	httpx.Reply(c, d, err)
}

// DeleteCourse удаляет пустое назначение: при нагрузке, темах, материалах,
// заданиях или работах — 409 in_use (каскадного удаления нет).
//
// DELETE /courses/:id
func (h *Handler) DeleteCourse(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	httpx.ReplyEmpty(c, h.repo.DeleteCourse(c.Request.Context(), id))
}

// ListWorkloads — нагрузка курса массивом, без пагинации.
//
// GET /courses/:id/workloads
func (h *Handler) ListWorkloads(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	list, err := h.svc.Workloads(c.Request.Context(), middleware.Me(c), id)
	httpx.Reply(c, list, err)
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
	if !validHours(*req.PlannedHours) {
		httpx.Fail(c, apperr.NewInvalid("invalid_request", "request body is invalid").
			WithDetails(map[string]any{"fields": map[string][]string{"planned_hours": {"must be >= 0 with a step of 0.25"}}}))
		return
	}
	wid, err := h.repo.AddWorkload(c.Request.Context(), id, req.TeacherID, req.LessonType, *req.PlannedHours)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	w, err := h.repo.GetWorkload(c.Request.Context(), wid)
	httpx.ReplyCreated(c, w, err)
}

// UpdateWorkload — полная замена нагрузки: преподаватель, вид занятия, часы.
// Занятия сохраняются и переходят к новому преподавателю; при его занятости — 409 schedule_conflict.
//
// PUT /workloads/:id
func (h *Handler) UpdateWorkload(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	var req workloadRequest
	if !httpx.Bind(c, &req) {
		return
	}
	if !validHours(*req.PlannedHours) {
		httpx.Fail(c, apperr.NewInvalid("invalid_request", "request body is invalid").
			WithDetails(map[string]any{"fields": map[string][]string{"planned_hours": {"must be >= 0 with a step of 0.25"}}}))
		return
	}
	if err := h.repo.UpdateWorkload(c.Request.Context(), id, req.TeacherID, req.LessonType, *req.PlannedHours); err != nil {
		httpx.Fail(c, err)
		return
	}
	w, err := h.repo.GetWorkload(c.Request.Context(), id)
	httpx.Reply(c, w, err)
}

// DeleteWorkload удаляет нагрузку без занятий; есть занятия — 409 in_use.
//
// DELETE /workloads/:id
func (h *Handler) DeleteWorkload(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	httpx.ReplyEmpty(c, h.repo.DeleteWorkload(c.Request.Context(), id))
}
