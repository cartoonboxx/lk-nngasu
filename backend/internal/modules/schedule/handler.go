package schedule

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"university-api/internal/platform/apperr"
	"university-api/internal/platform/httpx"
	"university-api/internal/platform/middleware"
)

// Handler — HTTP-слой расписания: разбирает параметры запроса и тела,
// логику (календарь, права в Mine, проверки занятий) выполняет Service.
type Handler struct{ svc *Service }

// lessonRequest — где и когда проходит занятие. Общая часть для создания и изменения.
//
// Расписание двухнедельное: занятие привязано не к дате, а к паре
// «неделя (верхняя/нижняя) + день недели»; конкретные даты получаются
// из учебного календаря.
type lessonRequest struct {
	WeekNumber int    `json:"week_number" binding:"required,oneof=1 2"`   // 1 — верхняя, 2 — нижняя неделя
	DayOfWeek  int    `json:"day_of_week" binding:"required,min=1,max=6"` // 1 — понедельник … 6 — суббота
	StartsAt   string `json:"starts_at" binding:"required"`               // начало пары по сетке звонков, "08:00"
	RoomID     int64  `json:"room_id" binding:"required,gt=0"`            // действующая аудитория
}

// createLessonRequest — тело POST /schedule/lessons.
// WorkloadID — строка учебной нагрузки (кто, что и какой группе ведёт в семестре);
// из неё занятие берёт преподавателя, дисциплину, группу и семестр.
// Встроенная структура раскрывается в JSON на верхний уровень,
// её теги binding тоже проверяются.
type createLessonRequest struct {
	WorkloadID int64 `json:"workload_id" binding:"required,gt=0"`
	lessonRequest
}

// conflictRequest — LessonConflictInput: проверка без записи.
// lesson_id — редактируемое занятие, оно исключается из сравнения.
type conflictRequest struct {
	LessonID *int64 `json:"lesson_id" binding:"omitempty,gt=0"`
	createLessonRequest
}

// input переводит тело запроса в тип сервиса.
func (r lessonRequest) input() LessonInput {
	return LessonInput{WeekNumber: r.WeekNumber, DayOfWeek: r.DayOfWeek, StartsAt: r.StartsAt, RoomID: r.RoomID}
}

// parseQuery читает временной срез: ?week=1|2 &day=1..6 либо ?from=YYYY-MM-DD&to=YYYY-MM-DD.
//   - week/day — срез по сетке расписания (вся верхняя неделя, все понедельники);
//   - from/to  — конкретные даты, сервис раскладывает сетку на календарь.
//
// Все параметры необязательны. Некорректный формат даёт 400:
// ответ отправляют сами httpx.Opt*, поэтому при false нужно просто выйти.
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
	// OptInt64 отдаёт *int64, а Query хранит *int: переупаковываем.
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

// failQuery — 400 invalid_query с указанием параметра.
func failQuery(c *gin.Context, name, problem string) {
	httpx.Fail(c, apperr.NewInvalid("invalid_query", "invalid query parameters").
		WithDetails(map[string]any{"fields": map[string][]string{name: {problem}}}))
}

// optID — необязательный id в query: положительное безопасное целое.
func optID(c *gin.Context, name string) (*int64, bool) {
	v, ok := httpx.OptInt64(c, name)
	if !ok {
		return nil, false
	}
	if v != nil && (*v <= 0 || *v > 1<<53-1) {
		failQuery(c, name, "must be a positive integer")
		return nil, false
	}
	return v, true
}

// parseRefQuery читает q, limit (1–200, по умолчанию 50) и offset (>= 0).
func parseRefQuery(c *gin.Context) (RefQuery, bool) {
	q := RefQuery{Q: strings.TrimSpace(c.Query("q")), Limit: 50}
	if len([]rune(q.Q)) > 200 {
		failQuery(c, "q", "must be at most 200 characters")
		return q, false
	}
	if raw := c.Query("limit"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 1 || v > 200 {
			failQuery(c, "limit", "must be an integer from 1 to 200")
			return q, false
		}
		q.Limit = v
	}
	if raw := c.Query("offset"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 0 {
			failQuery(c, "offset", "must be an integer >= 0")
			return q, false
		}
		q.Offset = v
	}
	return q, true
}

// ---- публичный просмотр ----

// Week сообщает, верхняя или нижняя неделя на заданную дату (по умолчанию сегодня).
// Без WEEK_ONE_DATE — 409 academic_calendar_not_configured.
//
// GET /schedule/week?date=YYYY-MM-DD
func (h *Handler) Week(c *gin.Context) {
	date, ok := httpx.OptDate(c, "date")
	if !ok {
		return
	}
	w, err := h.svc.Week(date)
	httpx.Reply(c, w, err)
}

// Groups — страница действующих групп для выбора расписания. Публичный.
//
// GET /schedule/groups?q=&limit=&offset=
func (h *Handler) Groups(c *gin.Context) {
	q, ok := parseRefQuery(c)
	if !ok {
		return
	}
	list, total, err := h.svc.Groups(c.Request.Context(), q)
	httpx.ReplyPage(c, list, httpx.Meta{Limit: q.Limit, Offset: q.Offset, Total: total}, err)
}

// Teachers — страница преподавателей для выбора расписания. Публичный,
// поэтому в ответе только id, ФИО и кафедра.
//
// GET /schedule/teachers?q=&limit=&offset=
func (h *Handler) Teachers(c *gin.Context) {
	q, ok := parseRefQuery(c)
	if !ok {
		return
	}
	list, total, err := h.svc.Teachers(c.Request.Context(), q)
	httpx.ReplyPage(c, list, httpx.Meta{Limit: q.Limit, Offset: q.Offset, Total: total}, err)
}

// ByGroup — расписание группы.
//
// GET /schedule/groups/:id?semester_id=&week=&day=&from=&to=
func (h *Handler) ByGroup(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	f, q, ok := h.filterAndQuery(c)
	if !ok {
		return
	}
	list, err := h.svc.GroupLessons(c.Request.Context(), id, f, q)
	httpx.Reply(c, list, err)
}

// ByTeacher — расписание преподавателя. Параметры те же, что у ByGroup.
//
// GET /schedule/teachers/:id?semester_id=&week=&day=&from=&to=
func (h *Handler) ByTeacher(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	f, q, ok := h.filterAndQuery(c)
	if !ok {
		return
	}
	list, err := h.svc.TeacherLessons(c.Request.Context(), id, f, q)
	httpx.Reply(c, list, err)
}

// ---- личное расписание ----

// Mine — расписание текущего пользователя: студенту — его группы,
// преподавателю — его занятия (по JWT). Параметры те же, что у ByGroup.
//
// GET /schedule/me?semester_id=&week=&day=&from=&to=
func (h *Handler) Mine(c *gin.Context) {
	f, q, ok := h.filterAndQuery(c)
	if !ok {
		return
	}
	list, err := h.svc.Mine(c.Request.Context(), middleware.Me(c), f, q)
	httpx.Reply(c, list, err)
}

// Overview — календарная страница расписания ЛК.
// Без group_id — своё расписание, с group_id — любой группы.
// Всегда по датам: date_from и date_to вместе (до 62 дней), без них — текущая неделя Пн–Вс.
//
// GET /schedule?group_id=&semester_id=&date_from=YYYY-MM-DD&date_to=YYYY-MM-DD
func (h *Handler) Overview(c *gin.Context) {
	groupID, ok := optID(c, "group_id")
	if !ok {
		return
	}
	semID, ok := optID(c, "semester_id")
	if !ok {
		return
	}
	from, ok := httpx.OptDate(c, "date_from")
	if !ok {
		return
	}
	to, ok := httpx.OptDate(c, "date_to")
	if !ok {
		return
	}
	list, err := h.svc.Overview(c.Request.Context(), middleware.Me(c), groupID, semID, from, to)
	httpx.Reply(c, list, err)
}

// filterAndQuery читает semester_id (без него — текущий семестр) и временной срез.
func (h *Handler) filterAndQuery(c *gin.Context) (LessonFilter, Query, bool) {
	sem, ok := optID(c, "semester_id")
	if !ok {
		return LessonFilter{}, Query{}, false
	}
	q, ok := parseQuery(c)
	return LessonFilter{SemesterID: sem}, q, ok
}

// ---- управление расписанием (администратор) ----

// Lessons — шаблон семестра для административного редактора: массив без пагинации.
// semester_id обязателен; week, group_id, teacher_id, room_id — фильтры через AND.
//
// GET /schedule/lessons?semester_id=&week=&group_id=&teacher_id=&room_id=
func (h *Handler) Lessons(c *gin.Context) {
	var f LessonFilter
	ids := []struct {
		name string
		dst  **int64
	}{{"semester_id", &f.SemesterID}, {"group_id", &f.GroupID}, {"teacher_id", &f.TeacherID}, {"room_id", &f.RoomID}}
	for _, p := range ids {
		v, ok := optID(c, p.name)
		if !ok {
			return
		}
		*p.dst = v
	}
	if f.SemesterID == nil {
		failQuery(c, "semester_id", "is required")
		return
	}
	week, ok := httpx.OptInt64(c, "week")
	if !ok {
		return
	}
	var w *int
	if week != nil {
		v := int(*week)
		w = &v
	}
	list, err := h.svc.Editor(c.Request.Context(), f, w)
	httpx.Reply(c, list, err)
}

// CreateLesson ставит занятие из строки нагрузки в сетку расписания.
//
// POST /schedule/lessons
func (h *Handler) CreateLesson(c *gin.Context) {
	var req createLessonRequest
	if !httpx.Bind(c, &req) {
		return
	}
	l, err := h.svc.Create(c.Request.Context(), req.WorkloadID, req.input())
	httpx.ReplyCreated(c, l, err)
}

// UpdateLesson переносит занятие: меняет неделю, день, время и аудиторию.
// Нагрузку (кто, что и кому ведёт) поменять нельзя: для этого занятие
// удаляют и создают заново из другой строки нагрузки.
//
// PUT /schedule/lessons/:id
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

// CheckConflicts — предварительная проверка пересечений без записи.
// Ответ — {data: ScheduleConflict[]}, пустой массив — пересечений нет.
// Окончательная проверка всё равно выполняется при сохранении (409 schedule_conflict).
//
// POST /schedule/lessons/conflicts
func (h *Handler) CheckConflicts(c *gin.Context) {
	var req conflictRequest
	if !httpx.Bind(c, &req) {
		return
	}
	list, err := h.svc.Conflicts(c.Request.Context(), req.WorkloadID, req.input(), req.LessonID)
	httpx.Reply(c, list, err)
}

// DeleteLesson удаляет занятие из расписания.
//
// DELETE /schedule/lessons/:id
func (h *Handler) DeleteLesson(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	httpx.ReplyEmpty(c, h.svc.Delete(c.Request.Context(), id))
}
