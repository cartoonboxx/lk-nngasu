package schedule

import (
	"context"
	"time"

	"university-api/internal/platform/apperr"
	"university-api/internal/platform/middleware"
)

// maxRangeDays — максимальная длина периода from..to. Ограничивает работу
// на публичных маршрутах: без него гость мог бы попросить развернуть
// расписание на годы вперёд. 62 дня — примерно два месяца.
const maxRangeDays = 62

// Ошибки валидации заданы заранее: они без параметров, и сервис
// возвращает одни и те же значения.
var (
	errBadPeriod  = apperr.NewInvalid("invalid_period", "from and to must be given together, and to must not be before from")
	errTooLong    = apperr.NewInvalid("range_too_big", "requested period must not exceed 62 days")
	errMixed      = apperr.NewInvalid("conflicting_filters", "use either week/day or from/to, not both")
	errBadWeek    = apperr.NewInvalid("invalid_week", "week must be 1 or 2")
	errBadDay     = apperr.NewInvalid("invalid_day", "day must be between 1 (Monday) and 6 (Saturday)")
	errBadTime    = apperr.NewInvalid("invalid_time", "starts_at must look like 08:30")
	errNoSchedule = apperr.NewForbidden("no_personal_schedule", "administrators have no personal schedule")
)

// Service — правила расписания: проверка временного среза и занятий,
// выбор «своего» расписания по роли, перевод двухнедельной сетки
// в конкретные даты через учебный календарь.
type Service struct {
	repo *Repository
	cal  Calendar // учебный календарь: сегодня, верхняя/нижняя неделя, звонки, развёртка по датам
}

func NewService(repo *Repository, cal Calendar) *Service { return &Service{repo: repo, cal: cal} }

// Week возвращает, верхняя или нижняя неделя на дату; nil — на сегодня.
// «Сегодня» берётся из календаря, а не из time.Now, чтобы в тестах дату можно было зафиксировать.
func (s *Service) Week(date *time.Time) WeekInfo {
	if date == nil {
		return s.cal.Week(s.cal.Today())
	}
	return s.cal.Week(*date)
}

// withEnds заполняет время окончания занятий. В БД хранится только начало,
// конец вычисляется по сетке звонков календаря. Срез меняется на месте.
func (s *Service) withEnds(ls []Lesson) []Lesson {
	for i := range ls {
		ls[i].EndsAt = s.cal.EndTime(ls[i].StartsAt)
	}
	return ls
}

// validWeek и validDay проверяют необязательные параметры среза; nil — «не задано», это допустимо.
func validWeek(w *int) bool { return w == nil || *w == 1 || *w == 2 }
func validDay(d *int) bool  { return d == nil || (*d >= 1 && *d <= 6) }

// Lessons отдаёт расписание по фильтру и временному срезу: шаблон (week/day) либо даты (from/to).
//
//   - Шаблон: занятия сетки как есть, без дат. week/day сужают выборку
//     (только верхняя неделя, только понедельники); без них — вся сетка.
//   - Даты: from и to обязательны вместе, не больше maxRangeDays,
//     с week/day не смешиваются. Сетка разворачивается календарём
//     в занятия на конкретные даты периода.
func (s *Service) Lessons(ctx context.Context, f LessonFilter, q Query) ([]Lesson, error) {
	if !validWeek(q.Week) {
		return nil, errBadWeek
	}
	if !validDay(q.Day) {
		return nil, errBadDay
	}
	byDates := q.From != nil || q.To != nil
	if byDates {
		if q.From == nil || q.To == nil || q.To.Before(*q.From) {
			return nil, errBadPeriod
		}
		if q.Week != nil || q.Day != nil {
			return nil, errMixed
		}
		if q.To.Sub(*q.From) > maxRangeDays*24*time.Hour {
			return nil, errTooLong
		}
	}
	// В режиме дат week/day гарантированно nil, так что из БД берётся вся сетка,
	// а нужные дни выбирает expand.
	f.Week, f.Day = q.Week, q.Day

	lessons, err := s.repo.List(ctx, f)
	if err != nil {
		return nil, err
	}
	lessons = s.withEnds(lessons)
	if byDates {
		return s.cal.expand(lessons, *q.From, *q.To), nil
	}
	return lessons, nil
}

// Mine — личное расписание: студент видит занятия своей группы, преподаватель — свои.
// Привязка берётся из JWT (p), а не из запроса, поэтому подменить группу
// или преподавателя через параметры нельзя. У администратора своего
// расписания нет — 403.
func (s *Service) Mine(ctx context.Context, p middleware.Principal, f LessonFilter, q Query) ([]Lesson, error) {
	switch {
	case p.IsStudent():
		f.StudentID = &p.UserID // группу студента репозиторий находит сам по user_id
	case p.IsTeacher():
		f.TeacherID = &p.UserID
	default:
		return nil, errNoSchedule
	}
	return s.Lessons(ctx, f, q)
}

// Today — личные занятия на сегодня (для главной страницы ЛК).
// У администратора личного расписания нет — для главной это не ошибка, а пустой список.
func (s *Service) Today(ctx context.Context, p middleware.Principal) ([]Lesson, error) {
	if p.IsAdmin() {
		return []Lesson{}, nil
	}
	today := s.cal.Today()
	return s.Mine(ctx, p, LessonFilter{}, Query{From: &today, To: &today})
}

// WeekRange дополняет период для страницы расписания:
//   - обе границы не заданы — текущая неделя (Пн–Вс);
//   - задана одна — вторая = та же ± 6 дней (неделя от неё).
//
// Заданные границы не меняются: их проверит Lessons.
func (s *Service) WeekRange(from, to *time.Time) (*time.Time, *time.Time) {
	switch {
	case from == nil && to == nil:
		mon := mondayOf(s.cal.Today())
		sun := mon.AddDate(0, 0, 6)
		return &mon, &sun
	case to == nil:
		t := from.AddDate(0, 0, 6)
		return from, &t
	case from == nil:
		f := to.AddDate(0, 0, -6)
		return &f, to
	}
	return from, to
}

// check проверяет занятие перед записью и приводит время начала
// к каноническому виду (например, "8:30" → "08:30"), чтобы в БД
// одинаковое время хранилось одинаково.
// Повторяет часть проверок binding: сервис не полагается на то,
// что его вызывают только из HTTP.
func (s *Service) check(in LessonInput) (LessonInput, error) {
	if in.WeekNumber != 1 && in.WeekNumber != 2 {
		return in, errBadWeek
	}
	if in.DayOfWeek < 1 || in.DayOfWeek > 6 {
		return in, errBadDay
	}
	clock, ok := parseClock(in.StartsAt)
	if !ok {
		return in, errBadTime
	}
	in.StartsAt = clock
	return in, nil
}

// Create ставит занятие из строки нагрузки workloadID в сетку
// и возвращает его в полном виде.
func (s *Service) Create(ctx context.Context, workloadID int64, in LessonInput) (Lesson, error) {
	in, err := s.check(in)
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
	in, err := s.check(in)
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
	l, err := s.repo.Get(ctx, id)
	if err != nil {
		return Lesson{}, err
	}
	l.EndsAt = s.cal.EndTime(l.StartsAt)
	return l, nil
}

// Groups и Teachers — списки для выбора расписания на публичной странице, без своей логики.

func (s *Service) Groups(ctx context.Context) ([]GroupRef, error)     { return s.repo.Groups(ctx) }
func (s *Service) Teachers(ctx context.Context) ([]TeacherRef, error) { return s.repo.Teachers(ctx) }
