package courses

import (
	"context"

	"university-api/internal/platform/access"
	"university-api/internal/platform/middleware"
)

type Service struct {
	repo   *Repository
	access *access.Checker
}

func NewService(repo *Repository, checker *access.Checker) *Service {
	return &Service{repo: repo, access: checker}
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
	return CourseDetail{Course: c, Workloads: ws}, nil
}

// Students — состав курса; студентам список не отдаётся.
func (s *Service) Students(ctx context.Context, p middleware.Principal, courseID int64) ([]CourseStudent, error) {
	if err := s.access.CourseManage(ctx, p, courseID); err != nil {
		return nil, err
	}
	return s.repo.ListCourseStudents(ctx, courseID)
}
