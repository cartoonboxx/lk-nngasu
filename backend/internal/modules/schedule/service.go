package schedule

import (
	"context"
	"time"

	"university-api/internal/platform/apperr"
	"university-api/internal/platform/middleware"
)

const maxRangeDays = 62

var (
	errBadPeriod  = apperr.NewInvalid("invalid_period", "from and to must be given together, and to must not be before from")
	errTooLong    = apperr.NewInvalid("range_too_big", "requested period must not exceed 62 days")
	errMixed      = apperr.NewInvalid("conflicting_filters", "use either week/day or from/to, not both")
	errBadWeek    = apperr.NewInvalid("invalid_week", "week must be 1 or 2")
	errBadDay     = apperr.NewInvalid("invalid_day", "day must be between 1 (Monday) and 6 (Saturday)")
	errBadTime    = apperr.NewInvalid("invalid_time", "starts_at must look like 08:30")
	errNoSchedule = apperr.NewForbidden("no_personal_schedule", "administrators have no personal schedule")
)

type Service struct {
	repo *Repository
	cal  Calendar
}

func NewService(repo *Repository, cal Calendar) *Service { return &Service{repo: repo, cal: cal} }

func (s *Service) Week(date *time.Time) WeekInfo {
	if date == nil {
		return s.cal.Week(s.cal.Today())
	}
	return s.cal.Week(*date)
}

func (s *Service) withEnds(ls []Lesson) []Lesson {
	for i := range ls {
		ls[i].EndsAt = s.cal.EndTime(ls[i].StartsAt)
	}
	return ls
}

func validWeek(w *int) bool { return w == nil || *w == 1 || *w == 2 }
func validDay(d *int) bool  { return d == nil || (*d >= 1 && *d <= 6) }

// Lessons отдаёт расписание по фильтру и временному срезу: шаблон (week/day) либо даты (from/to).
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
func (s *Service) Mine(ctx context.Context, p middleware.Principal, f LessonFilter, q Query) ([]Lesson, error) {
	switch {
	case p.IsStudent():
		f.StudentID = &p.UserID
	case p.IsTeacher():
		f.TeacherID = &p.UserID
	default:
		return nil, errNoSchedule
	}
	return s.Lessons(ctx, f, q)
}

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

func (s *Service) Delete(ctx context.Context, id int64) error { return s.repo.Delete(ctx, id) }

func (s *Service) get(ctx context.Context, id int64) (Lesson, error) {
	l, err := s.repo.Get(ctx, id)
	l.EndsAt = s.cal.EndTime(l.StartsAt)
	return l, err
}

func (s *Service) Groups(ctx context.Context) ([]GroupRef, error)     { return s.repo.Groups(ctx) }
func (s *Service) Teachers(ctx context.Context) ([]TeacherRef, error) { return s.repo.Teachers(ctx) }
