package schedule

import (
	"context"
	"errors"
	"time"

	"university-api/internal/platform/apperr"
	"university-api/internal/platform/middleware"
)

// maxRangeDays — максимальная длина периода в днях, обе границы включительно.
// Ограничивает работу на публичных маршрутах: без него гость мог бы попросить
// развернуть расписание на годы вперёд.
const maxRangeDays = 62

var (
	errCalendar   = apperr.NewConflict("academic_calendar_not_configured", "the upper-week start date (WEEK_ONE_DATE) is not configured")
	errNoSemester = apperr.NewConflict("current_semester_not_set", "current semester is not set")
	errNoSchedule = apperr.NewForbidden("no_personal_schedule", "administrators have no personal schedule")
)

// queryErr — 400 invalid_query с указанием параметра (тот же формат, что у справочников).
func queryErr(name, problem string) error {
	return apperr.NewInvalid("invalid_query", "invalid query parameters").
		WithDetails(map[string]any{"fields": map[string][]string{name: {problem}}})
}

// Service — правила расписания: проверка временного среза и занятий,
// выбор «своего» расписания по роли, перевод двухнедельной сетки
// в конкретные даты через учебный календарь.
type Service struct {
	repo *Repository
	cal  Calendar // учебный календарь: сегодня, верхняя/нижняя неделя, звонки, развёртка по датам
}

func NewService(repo *Repository, cal Calendar) *Service { return &Service{repo: repo, cal: cal} }

// Week возвращает, верхняя или нижняя неделя на дату; nil — на сегодня (Europe/Moscow).
// Без опорной даты верхней недели — 409 academic_calendar_not_configured.
func (s *Service) Week(date *time.Time) (WeekInfo, error) {
	if !s.cal.Configured() {
		return WeekInfo{}, errCalendar
	}
	if date == nil {
		return s.cal.Week(s.cal.Today()), nil
	}
	return s.cal.Week(*date), nil
}

// checkQuery проверяет временной срез: шаблон (week/day) или даты (from/to), не вместе.
func checkQuery(q Query, fromName, toName string) error {
	if q.Week != nil && *q.Week != 1 && *q.Week != 2 {
		return queryErr("week", "must be 1 or 2")
	}
	if q.Day != nil && (*q.Day < 1 || *q.Day > 6) {
		return queryErr("day", "must be from 1 (Monday) to 6 (Saturday)")
	}
	if q.From == nil && q.To == nil {
		return nil
	}
	if q.Week != nil || q.Day != nil {
		return queryErr(fromName, "cannot be combined with week/day")
	}
	if q.From == nil {
		return queryErr(fromName, "must be given together with "+toName)
	}
	if q.To == nil {
		return queryErr(toName, "must be given together with "+fromName)
	}
	if q.To.Before(*q.From) {
		return queryErr(toName, "must not be before "+fromName)
	}
	if days := int(q.To.Sub(*q.From).Hours()/24) + 1; days > maxRangeDays {
		return queryErr(toName, "period must not exceed 62 days")
	}
	return nil
}

// semester подставляет текущий семестр, если он не задан, или проверяет заданный.
func (s *Service) semester(ctx context.Context, f *LessonFilter) error {
	if f.SemesterID == nil {
		id, ok, err := s.repo.CurrentSemester(ctx)
		if err != nil {
			return err
		}
		if !ok {
			return errNoSemester
		}
		f.SemesterID = &id
		return nil
	}
	return s.repo.SemesterExists(ctx, *f.SemesterID)
}

// lessons — общий хвост всех выборок: семестр, сетка из БД и, в режиме дат, развёртка по календарю.
// Срез q уже проверен checkQuery.
//
//   - Шаблон: занятия сетки как есть, date = null; week/day сужают выборку,
//     без них — обе недели.
//   - Даты: сетка разворачивается в занятия на конкретные даты периода
//     (порядок date, starts_at, id). Нужен настроенный календарь.
func (s *Service) lessons(ctx context.Context, f LessonFilter, q Query) ([]Lesson, error) {
	byDates := q.From != nil
	if byDates && !s.cal.Configured() {
		return nil, errCalendar
	}
	if err := s.semester(ctx, &f); err != nil {
		return nil, err
	}
	f.Week, f.Day = q.Week, q.Day
	list, err := s.repo.List(ctx, f)
	if err != nil {
		return nil, err
	}
	if byDates {
		return s.cal.expand(list, *q.From, *q.To), nil
	}
	return list, nil
}

// GroupLessons — публичное расписание группы; нет группы — 404 group_not_found.
func (s *Service) GroupLessons(ctx context.Context, groupID int64, f LessonFilter, q Query) ([]Lesson, error) {
	if err := checkQuery(q, "from", "to"); err != nil {
		return nil, err
	}
	if err := s.repo.GroupExists(ctx, groupID); err != nil {
		return nil, err
	}
	f.GroupID = &groupID
	return s.lessons(ctx, f, q)
}

// TeacherLessons — публичное расписание преподавателя; нет преподавателя — 404 teacher_not_found.
// Нет занятий — пустой список.
func (s *Service) TeacherLessons(ctx context.Context, teacherID int64, f LessonFilter, q Query) ([]Lesson, error) {
	if err := checkQuery(q, "from", "to"); err != nil {
		return nil, err
	}
	if err := s.repo.TeacherExists(ctx, teacherID); err != nil {
		return nil, err
	}
	f.TeacherID = &teacherID
	return s.lessons(ctx, f, q)
}

// Mine — личное расписание: студент видит занятия своей группы, преподаватель — свои.
// Привязка берётся из JWT (p), а не из запроса, поэтому подменить группу
// или преподавателя через параметры нельзя. У администратора своего
// расписания нет — 403.
func (s *Service) Mine(ctx context.Context, p middleware.Principal, f LessonFilter, q Query) ([]Lesson, error) {
	if err := checkQuery(q, "from", "to"); err != nil {
		return nil, err
	}
	return s.mine(ctx, p, f, q)
}

func (s *Service) mine(ctx context.Context, p middleware.Principal, f LessonFilter, q Query) ([]Lesson, error) {
	switch {
	case p.IsStudent():
		f.StudentID = &p.UserID // группу студента репозиторий находит сам по user_id
	case p.IsTeacher():
		f.TeacherID = &p.UserID
	default:
		return nil, errNoSchedule
	}
	return s.lessons(ctx, f, q)
}

// Overview — календарная страница расписания ЛК (GET /schedule).
// groupID задан — расписание этой группы (смотреть можно любую), иначе личное.
// Период: date_from и date_to только вместе; без них — текущая неделя Пн–Вс.
func (s *Service) Overview(ctx context.Context, p middleware.Principal, groupID, semesterID *int64, from, to *time.Time) ([]Lesson, error) {
	if from == nil && to == nil {
		if !s.cal.Configured() {
			return nil, errCalendar
		}
		mon := mondayOf(s.cal.Today())
		sun := mon.AddDate(0, 0, 6)
		from, to = &mon, &sun
	}
	q := Query{From: from, To: to}
	if err := checkQuery(q, "date_from", "date_to"); err != nil {
		return nil, err
	}
	f := LessonFilter{SemesterID: semesterID}
	if groupID != nil {
		if err := s.repo.GroupExists(ctx, *groupID); err != nil {
			return nil, err
		}
		f.GroupID = groupID
		return s.lessons(ctx, f, q)
	}
	return s.mine(ctx, p, f, q)
}

// Editor — шаблон семестра для административного редактора (GET /schedule/lessons):
// фильтры week, group_id, teacher_id, room_id через AND, без пагинации.
func (s *Service) Editor(ctx context.Context, f LessonFilter, week *int) ([]Lesson, error) {
	q := Query{Week: week}
	if err := checkQuery(q, "from", "to"); err != nil {
		return nil, err
	}
	return s.lessons(ctx, f, q)
}

// check проверяет занятие перед записью и приводит время начала
// к каноническому виду (например, "8:30" → "08:30"), чтобы в БД
// одинаковое время хранилось одинаково.
// Повторяет часть проверок binding: сервис не полагается на то,
// что его вызывают только из HTTP.
//
// Время: конец позже начала и пара совпадает с одной из пар сетки звонков.
// stored — сохранённое занятие при редактировании: его прежний интервал
// допустим, даже если он не из сетки (старые записи).
func (s *Service) check(in LessonInput, stored *Stored) (LessonInput, error) {
	if in.WeekNumber != 1 && in.WeekNumber != 2 {
		return in, fieldErr("week_number", "must be 1 or 2")
	}
	if in.DayOfWeek < 1 || in.DayOfWeek > 6 {
		return in, fieldErr("day_of_week", "must be from 1 to 6")
	}
	start, ok := parseClock(in.StartsAt)
	if !ok {
		return in, fieldErr("starts_at", "must look like 08:00")
	}
	// ends_at не принимается (LessonInput): конец пары берётся из сетки звонков по starts_at.
	// У сохранённого занятия с нестандартным временем прежний интервал остаётся при правке
	// аудитории, недели или дня, если начало не менялось.
	end := s.cal.EndTime(start)
	if stored != nil && stored.StartsAt == start {
		end = stored.EndsAt
	}
	if end <= start {
		return in, periodErr("ends_at", "must be later than starts_at")
	}
	in.StartsAt, in.EndsAt = start, end
	unchanged := stored != nil && stored.StartsAt == start && stored.EndsAt == end
	if !unchanged && !s.cal.IsBell(start, end) {
		return in, periodErr("starts_at", "starts_at and ends_at must be the start and end of one bell schedule pair")
	}
	return in, nil
}

// fieldErr — ошибка формата поля тела (400 invalid_request), как у binding.
func fieldErr(field, problem string) error {
	return apperr.NewInvalid("invalid_request", "request body is invalid").
		WithDetails(map[string]any{"fields": map[string][]string{field: {problem}}})
}

// periodErr — интервал занятия недопустим (400 invalid_period): конец не позже начала
// или пара не из сетки звонков.
func periodErr(field, problem string) error {
	return apperr.NewInvalid("invalid_period", "lesson time must be one bell schedule pair").
		WithDetails(map[string]any{"fields": map[string][]string{field: {problem}}})
}

// Conflicts — предварительная проверка пересечений без записи (LessonConflictInput).
func (s *Service) Conflicts(ctx context.Context, workloadID int64, in LessonInput, lessonID *int64) ([]ScheduleConflict, error) {
	var stored *Stored
	if lessonID != nil {
		st, err := s.repo.Stored(ctx, *lessonID)
		if err != nil {
			return nil, err
		}
		stored = &st
	}
	in, err := s.check(in, stored)
	if err != nil {
		return nil, err
	}
	var keepRoom *int64
	if stored != nil {
		keepRoom = &stored.RoomID
	}
	return s.repo.CheckConflicts(ctx, workloadID, in, lessonID, keepRoom)
}

// Create ставит занятие из строки нагрузки workloadID в сетку
// и возвращает его в полном виде.
func (s *Service) Create(ctx context.Context, workloadID int64, in LessonInput) (Lesson, error) {
	in, err := s.check(in, nil)
	if err != nil {
		return Lesson{}, err
	}
	id, err := s.repo.Create(ctx, workloadID, in)
	if err != nil {
		return Lesson{}, err
	}
	return s.get(ctx, id)
}

// Update переносит занятие (неделя, день, время, аудитория) и возвращает его в полном виде.
func (s *Service) Update(ctx context.Context, id int64, in LessonInput) (Lesson, error) {
	stored, err := s.repo.Stored(ctx, id)
	if err != nil {
		return Lesson{}, err
	}
	in, err = s.check(in, &stored)
	if err != nil {
		return Lesson{}, err
	}
	if err := s.repo.Update(ctx, id, in); err != nil {
		return Lesson{}, err
	}
	return s.get(ctx, id)
}

// Delete удаляет занятие из сетки.
func (s *Service) Delete(ctx context.Context, id int64) error { return s.repo.Delete(ctx, id) }

// get читает одно занятие и дополняет его временем окончания.
func (s *Service) get(ctx context.Context, id int64) (Lesson, error) {
	return s.repo.Get(ctx, id)
}

// Groups и Teachers — списки для выбора расписания на публичной странице, без своей логики.

func (s *Service) Groups(ctx context.Context, q RefQuery) ([]ScheduleRef, int, error) {
	return s.repo.Groups(ctx, q)
}
func (s *Service) Teachers(ctx context.Context, q RefQuery) ([]ScheduleRef, int, error) {
	return s.repo.Teachers(ctx, q)
}

// Today — главная страница: учебная неделя на сегодня и пары пользователя на сегодня
// (студент — пары группы, преподаватель — пары нагрузки). Ничего не роняет:
// у администратора, без текущего семестра или без настроенного календаря пар нет — [],
// а без календаря и недели нет — week = nil.
func (s *Service) Today(ctx context.Context, p middleware.Principal) (*WeekInfo, []Lesson, error) {
	if !s.cal.Configured() {
		return nil, []Lesson{}, nil
	}
	today := s.cal.Today()
	week := s.cal.Week(today)
	if !p.IsStudent() && !p.IsTeacher() {
		return &week, []Lesson{}, nil
	}
	list, err := s.mine(ctx, p, LessonFilter{}, Query{From: &today, To: &today})
	if errors.Is(err, errNoSemester) {
		return &week, []Lesson{}, nil
	}
	if err != nil {
		return nil, nil, err
	}
	if list == nil {
		list = []Lesson{}
	}
	return &week, list, nil
}
