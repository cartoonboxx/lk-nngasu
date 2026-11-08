package courses

import (
	"context"

	"university-api/internal/platform/access"
	"university-api/internal/platform/middleware"
)

// TopicsLoader загружает темы курса для страницы предмета.
// Реализация — в app: она объединяет темы с материалами (content) и задания (assignments).
// Права на курс проверяют сами эти модули.
type TopicsLoader func(ctx context.Context, p middleware.Principal, courseID int64) (any, error)

type Service struct {
	repo   *Repository
	access *access.Checker
	topics TopicsLoader // nil — страница предмета без тем
}

func NewService(repo *Repository, checker *access.Checker, topics TopicsLoader) *Service {
	return &Service{repo: repo, access: checker, topics: topics}
}

// List возвращает курсы, видимые пользователю: студент — своей группы, преподаватель — где он ведёт,
// админ — все.
func (s *Service) List(ctx context.Context, p middleware.Principal, f CourseFilter) ([]Course, error) {
	switch {
	case p.IsStudent():
		f.ForStudent = &p.UserID
	case p.IsTeacher():
		f.ForTeacher = &p.UserID
	}
	return s.repo.ListCourses(ctx, f)
}

func (s *Service) Get(ctx context.Context, p middleware.Principal, id int64) (CourseDetail, error) {
	if err := s.access.CourseView(ctx, p, id); err != nil {
		return CourseDetail{}, err
	}
	c, err := s.repo.GetCourse(ctx, id)
	if err != nil {
		return CourseDetail{}, err
	}
	ws, err := s.repo.ListWorkloads(ctx, id)
	if err != nil {
		return CourseDetail{}, err
	}
	d := CourseDetail{Course: c, Workloads: ws, Topics: []any{}}
	if s.topics != nil {
		if d.Topics, err = s.topics(ctx, p, id); err != nil {
			return CourseDetail{}, err
		}
	}
	return d, nil
}

// Students — состав курса; студентам список не отдаётся.
func (s *Service) Students(ctx context.Context, p middleware.Principal, courseID int64) ([]CourseStudent, error) {
	if err := s.access.CourseManage(ctx, p, courseID); err != nil {
		return nil, err
	}
	return s.repo.ListCourseStudents(ctx, courseID)
}
