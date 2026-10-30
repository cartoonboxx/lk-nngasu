package schedule

import (
	"github.com/gin-gonic/gin"

	"university-api/internal/platform/httpx"
	"university-api/internal/platform/middleware"
)

type Handler struct{ svc *Service }

type lessonRequest struct {
	WeekNumber int    `json:"week_number" binding:"required,oneof=1 2"`
	DayOfWeek  int    `json:"day_of_week" binding:"required,min=1,max=6"`
	StartsAt   string `json:"starts_at" binding:"required"` // "08:30"
	RoomID     int64  `json:"room_id" binding:"required"`
}

type createLessonRequest struct {
	WorkloadID int64 `json:"workload_id" binding:"required"`
	lessonRequest
}

func (r lessonRequest) input() LessonInput {
	return LessonInput{WeekNumber: r.WeekNumber, DayOfWeek: r.DayOfWeek, StartsAt: r.StartsAt, RoomID: r.RoomID}
}

// parseQuery читает временной срез: ?week=1|2 &day=1..6 либо ?from=YYYY-MM-DD&to=YYYY-MM-DD.
func parseQuery(c *gin.Context) (Query, bool) {
	week, ok := httpx.OptInt64(c, "week")
	if !ok {
		return Query{}, false
	}
	day, ok := httpx.OptInt64(c, "day")
	if !ok {
		return Query{}, false
	}
	from, ok := httpx.OptDate(c, "from")
	if !ok {
		return Query{}, false
	}
	to, ok := httpx.OptDate(c, "to")
	if !ok {
		return Query{}, false
	}
	q := Query{From: from, To: to}
	if week != nil {
		w := int(*week)
		q.Week = &w
	}
	if day != nil {
		d := int(*day)
		q.Day = &d
	}
	return q, true
}

// GET /schedule/week?date=YYYY-MM-DD — верхняя или нижняя неделя (по умолчанию сегодня).
func (h *Handler) Week(c *gin.Context) {
	date, ok := httpx.OptDate(c, "date")
	if !ok {
		return
	}
	httpx.OK(c, h.svc.Week(date))
}

// GET /schedule/groups, /schedule/teachers — списки для выбора (публичные).
func (h *Handler) Groups(c *gin.Context) {
	list, err := h.svc.Groups(c.Request.Context())
	httpx.Reply(c, list, err)
}

func (h *Handler) Teachers(c *gin.Context) {
	list, err := h.svc.Teachers(c.Request.Context())
	httpx.Reply(c, list, err)
}

// GET /schedule/groups/:id?semester_id=&week=&day=&from=&to=
func (h *Handler) ByGroup(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	h.list(c, LessonFilter{GroupID: &id})
}

// GET /schedule/teachers/:id?...
func (h *Handler) ByTeacher(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	h.list(c, LessonFilter{TeacherID: &id})
}

// GET /schedule/me?... — студент: своей группы; преподаватель: свои занятия.
func (h *Handler) Mine(c *gin.Context) {
	f, q, ok := h.filterAndQuery(c, LessonFilter{})
	if !ok {
		return
	}
	list, err := h.svc.Mine(c.Request.Context(), middleware.Me(c), f, q)
	httpx.Reply(c, list, err)
}

func (h *Handler) list(c *gin.Context, base LessonFilter) {
	f, q, ok := h.filterAndQuery(c, base)
	if !ok {
		return
	}
	list, err := h.svc.Lessons(c.Request.Context(), f, q)
	httpx.Reply(c, list, err)
}

func (h *Handler) filterAndQuery(c *gin.Context, f LessonFilter) (LessonFilter, Query, bool) {
	sem, ok := httpx.OptInt64(c, "semester_id")
	if !ok {
		return f, Query{}, false
	}
	f.SemesterID = sem
	q, ok := parseQuery(c)
	return f, q, ok
}

// ---- управление расписанием (администратор) ----

func (h *Handler) CreateLesson(c *gin.Context) {
	var req createLessonRequest
	if !httpx.Bind(c, &req) {
		return
	}
	l, err := h.svc.Create(c.Request.Context(), req.WorkloadID, req.input())
	httpx.ReplyCreated(c, l, err)
}

func (h *Handler) UpdateLesson(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	var req lessonRequest
	if !httpx.Bind(c, &req) {
		return
	}
	l, err := h.svc.Update(c.Request.Context(), id, req.input())
	httpx.Reply(c, l, err)
}

func (h *Handler) DeleteLesson(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	httpx.ReplyEmpty(c, h.svc.Delete(c.Request.Context(), id))
}
