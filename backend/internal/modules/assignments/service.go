package assignments

import (
	"context"
	"io"
	"strings"

	"university-api/internal/platform/apperr"
	"university-api/internal/platform/middleware"
)

// Guard — проверки доступа (реализует access.Checker).
type Guard interface {
	CourseView(ctx context.Context, p middleware.Principal, courseID int64) error
	CourseManage(ctx context.Context, p middleware.Principal, courseID int64) error
	CourseTeacher(ctx context.Context, p middleware.Principal, courseID int64) error
	StudentOfCourse(ctx context.Context, studentID, courseID int64) (bool, error)
	OwnsFile(ctx context.Context, userID, fileID int64) error
}

// Blobs — файловое хранилище (реализует files.Storage).
type Blobs interface {
	Open(key string) (io.ReadCloser, error)
	Remove(key string) error
}

type Service struct {
	repo  *Repository
	guard Guard
	blobs Blobs
}

func NewService(repo *Repository, guard Guard, blobs Blobs) *Service {
	return &Service{repo: repo, guard: guard, blobs: blobs}
}

func (s *Service) removeBlobs(keys []string) {
	for _, k := range keys {
		_ = s.blobs.Remove(k)
	}
}

func (s *Service) staff(p middleware.Principal) bool { return !p.IsStudent() }

// viewer — id для LEFT JOIN со своим ответом: только у студента.
func viewer(p middleware.Principal) int64 {
	if p.IsStudent() {
		return p.UserID
	}
	return 0
}

// checkFiles: не больше MaxFiles, без повторов, все файлы — свои.
func (s *Service) checkFiles(ctx context.Context, p middleware.Principal, ids []int64, keep map[int64]bool) error {
	if len(ids) > MaxFiles {
		return apperr.NewInvalid("too_many_files", "at most 5 files can be attached")
	}
	seen := make(map[int64]bool, len(ids))
	for _, id := range ids {
		if seen[id] {
			return apperr.NewInvalid("duplicate_file", "the same file is attached twice")
		}
		seen[id] = true
		if keep[id] {
			continue // уже привязан к этому объекту — владельца повторно не проверяем
		}
		if err := s.guard.OwnsFile(ctx, p.UserID, id); err != nil {
			return err
		}
	}
	return nil
}

func normalizeInput(in *AssignmentInput) error {
	in.Title = strings.TrimSpace(in.Title)
	if in.Title == "" {
		return apperr.NewInvalid("title_required", "title is required")
	}
	if in.Description != nil {
		d := strings.TrimSpace(*in.Description)
		if d == "" {
			in.Description = nil
		} else {
			in.Description = &d
		}
	}
	if in.DeadlineAt.IsZero() {
		return apperr.NewInvalid("deadline_required", "deadline_at is required")
	}
	return nil
}

func (s *Service) ListAssignments(ctx context.Context, p middleware.Principal, courseID int64) ([]Assignment, error) {
	if err := s.guard.CourseView(ctx, p, courseID); err != nil {
		return nil, err
	}
	return s.repo.ListAssignments(ctx, courseID, s.staff(p), viewer(p))
}

func (s *Service) GetAssignment(ctx context.Context, p middleware.Principal, id int64) (Assignment, error) {
	a, err := s.repo.GetAssignment(ctx, id, s.staff(p), viewer(p))
	if err != nil {
		return Assignment{}, err
	}
	return a, s.guard.CourseView(ctx, p, a.CourseID)
}

// CreateAssignment: created_by ссылается на teacher — создаёт преподаватель курса.
func (s *Service) CreateAssignment(ctx context.Context, p middleware.Principal, courseID int64, in AssignmentInput) (Assignment, error) {
	if err := s.guard.CourseTeacher(ctx, p, courseID); err != nil {
		return Assignment{}, err
	}
	if err := normalizeInput(&in); err != nil {
		return Assignment{}, err
	}
	if err := s.checkFiles(ctx, p, in.FileIDs, nil); err != nil {
		return Assignment{}, err
	}
	id, err := s.repo.CreateAssignment(ctx, courseID, p.UserID, in)
	if err != nil {
		return Assignment{}, err
	}
	return s.repo.GetAssignment(ctx, id, true, 0)
}

func (s *Service) UpdateAssignment(ctx context.Context, p middleware.Principal, id int64, in AssignmentInput) (Assignment, error) {
	cur, err := s.repo.GetAssignment(ctx, id, true, 0)
	if err != nil {
		return Assignment{}, err
	}
	if err := s.guard.CourseManage(ctx, p, cur.CourseID); err != nil {
		return Assignment{}, err
	}
	if err := normalizeInput(&in); err != nil {
		return Assignment{}, err
	}
	keep := make(map[int64]bool, len(cur.Files))
	for _, f := range cur.Files {
		keep[f.ID] = true
	}
	if err := s.checkFiles(ctx, p, in.FileIDs, keep); err != nil {
		return Assignment{}, err
	}
	keys, err := s.repo.UpdateAssignment(ctx, id, in)
	s.removeBlobs(keys)
	if err != nil {
		return Assignment{}, err
	}
	return s.repo.GetAssignment(ctx, id, true, 0)
}

func (s *Service) DeleteAssignment(ctx context.Context, p middleware.Principal, id int64) error {
	cur, err := s.repo.GetAssignment(ctx, id, true, 0)
	if err != nil {
		return err
	}
	if err := s.guard.CourseManage(ctx, p, cur.CourseID); err != nil {
		return err
	}
	keys, err := s.repo.DeleteAssignment(ctx, id)
	s.removeBlobs(keys)
	return err
}

// ---- оценки ----

func (s *Service) MyGrades(ctx context.Context, p middleware.Principal, courseID *int64) ([]GradeRow, error) {
	if !p.IsStudent() {
		return nil, apperr.NewForbidden("students_only", "only students have personal grades")
	}
	return s.repo.ListGrades(ctx, GradeFilter{StudentID: &p.UserID, CourseID: courseID})
}

func (s *Service) CourseGrades(ctx context.Context, p middleware.Principal, courseID int64) ([]GradeRow, error) {
	if err := s.guard.CourseManage(ctx, p, courseID); err != nil {
		return nil, err
	}
	return s.repo.ListGrades(ctx, GradeFilter{CourseID: &courseID})
}
