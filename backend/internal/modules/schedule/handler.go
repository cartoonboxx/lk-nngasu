package schedule

import (
	"github.com/gin-gonic/gin"

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
	StartsAt   string `json:"starts_at" binding:"required"`               // время начала пары, "08:30"
	RoomID     int64  `json:"room_id" binding:"required"`
}

// createLessonRequest — тело POST /schedule/lessons.
// WorkloadID — строка учебной нагрузки (кто, что и какой группе ведёт в семестре);
// из неё занятие берёт преподавателя, дисциплину, группу и семестр.
// Встроенная структура раскрывается в JSON на верхний уровень,
// её теги binding тоже проверяются.
type createLessonRequest struct {
	WorkloadID int64 `json:"workload_id" binding:"required"`
	lessonRequest
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

// ---- публичный просмотр ----

// Week сообщает, верхняя или нижняя неделя на заданную дату (по умолчанию сегодня).
// Ошибки нет: ответ вычисляется календарём без обращения к БД.
//
// GET /schedule/week?date=YYYY-MM-DD
func (h *Handler) Week(c *gin.Context) {
	date, ok := httpx.OptDate(c, "date")
	if !ok {
		return
	}
	httpx.OK(c, h.svc.Week(date))
}

// Groups — список групп для выбора расписания на сайте. Публичный.
//
// GET /schedule/groups
func (h *Handler) Groups(c *gin.Context) {
	list, err := h.svc.Groups(c.Request.Context())
	httpx.Reply(c, list, err)
}

// Teachers — список преподавателей для выбора расписания на сайте. Публичный,
// поэтому в ответе должны быть только данные, которые не страшно показать гостю.
//
// GET /schedule/teachers
func (h *Handler) Teachers(c *gin.Context) {
	list, err := h.svc.Teachers(c.Request.Context())
	httpx.Reply(c, list, err)
}

// ByGroup — расписание группы.
//
// GET /schedule/groups/:id?semester_id=&week=&day=&from=&to=
func (h *Handler) ByGroup(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	h.list(c, LessonFilter{GroupID: &id})
}

// ByTeacher — расписание преподавателя. Параметры те же, что у ByGroup.
//
// GET /schedule/teachers/:id?semester_id=&week=&day=&from=&to=
func (h *Handler) ByTeacher(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	h.list(c, LessonFilter{TeacherID: &id})
}

// ---- личное расписание ----

// Mine — расписание текущего пользователя: студенту — его группы,
// преподавателю — его занятия. Кто есть кто, сервис решает по middleware.Me,
// поэтому фильтр стартует пустым и группа/преподаватель в нём не задаются.
//
// GET /schedule/me?semester_id=&week=&day=&from=&to=
func (h *Handler) Mine(c *gin.Context) {
	f, q, ok := h.filterAndQuery(c, LessonFilter{})
	if !ok {
		return
	}
	list, err := h.svc.Mine(c.Request.Context(), middleware.Me(c), f, q)
	httpx.Reply(c, list, err)
}

// Overview — страница расписания ЛК.
// Без group_id — своё расписание (студенту — группы, преподавателю — его занятия),
// с group_id — расписание любой группы. Всегда по датам: без date_from/date_to —
// текущая неделя, с одной границей — неделя от неё.
//
// GET /schedule?group_id=&date_from=YYYY-MM-DD&date_to=YYYY-MM-DD
func (h *Handler) Overview(c *gin.Context) {
	groupID, ok := httpx.OptInt64(c, "group_id")
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
	from, to = h.svc.WeekRange(from, to)
	q := Query{From: from, To: to}

	// Чужую группу — тем же путём, что публичный ByGroup; своё — через Mine (по JWT).
	if groupID != nil {
		list, err := h.svc.Lessons(c.Request.Context(), LessonFilter{GroupID: groupID}, q)
		httpx.Reply(c, list, err)
		return
	}
	list, err := h.svc.Mine(c.Request.Context(), middleware.Me(c), LessonFilter{}, q)
	httpx.Reply(c, list, err)
}

// list — общий хвост ByGroup и ByTeacher: дочитать параметры и отдать занятия.
func (h *Handler) list(c *gin.Context, base LessonFilter) {
	f, q, ok := h.filterAndQuery(c, base)
	if !ok {
		return
	}
	list, err := h.svc.Lessons(c.Request.Context(), f, q)
	httpx.Reply(c, list, err)
}

// filterAndQuery дополняет базовый фильтр семестром (semester_id; если не задан,
// выбор семестра остаётся за сервисом) и читает временной срез.
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
