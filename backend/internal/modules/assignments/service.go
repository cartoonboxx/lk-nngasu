package assignments

import (
	"context"
	"errors"
	"io"
	"strings"
	"unicode/utf8"

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
	Size(key string) (int64, error) // размер на диске — для старых файлов без size_bytes
}

var errCourseForbidden = apperr.NewForbidden("course_forbidden", "you have no access to this course")

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

// withSizes дополняет FileInfo размером с диска там, где он не записан в БД (файлы до миграции 010).
func (s *Service) withSizes(files []FileInfo) {
	for i := range files {
		if files[i].SizeBytes < 0 {
			files[i].SizeBytes = 0
			if n, err := s.blobs.Size(files[i].key); err == nil {
				files[i].SizeBytes = n
			}
		}
	}
}

func (s *Service) assignmentOut(a Assignment) Assignment {
	s.withSizes(a.Files)
	return a
}

func (s *Service) submissionOut(sub Submission) Submission {
	s.withSizes(sub.Files)
	return sub
}

// checkFiles: не больше MaxFiles, без повторов, все файлы — свои
// (или уже привязанные к этому заданию/ответу — keep).
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
			var ae *apperr.Error
			if errors.As(err, &ae) && ae.Code == "invalid_file" {
				return errFileNotFound // чужой и несуществующий файл неотличимы
			}
			return err
		}
	}
	return nil
}

func normalizeInput(in *AssignmentInput) error {
	in.Title = strings.TrimSpace(in.Title)
	if in.Title == "" || utf8.RuneCountInString(in.Title) > 200 {
		return fieldErr("title", "must be 1-200 characters after trim")
	}
	if in.Description != nil {
		d := strings.TrimSpace(*in.Description)
		if d == "" {
			in.Description = nil
		} else {
			in.Description = &d
		}
	}
	return nil
}

func (s *Service) ListAssignments(ctx context.Context, p middleware.Principal, courseID int64) ([]Assignment, error) {
	if err := s.guard.CourseView(ctx, p, courseID); err != nil {
		return nil, err
	}
	list, err := s.repo.ListAssignments(ctx, courseID, s.staff(p), viewer(p))
	for i := range list {
		s.withSizes(list[i].Files)
	}
	return list, err
}

// GetAssignment — карточка задания (AssignmentDetail). Студенту дополнительно отдаётся его
// ответ целиком или null, чтобы модальному окну задания хватало одного запроса;
// преподавателю и админу — stats, а список работ — отдельным запросом.
func (s *Service) GetAssignment(ctx context.Context, p middleware.Principal, id int64) (AssignmentDetail, error) {
	a, err := s.repo.GetAssignment(ctx, id, s.staff(p), viewer(p))
	if err != nil {
		return AssignmentDetail{}, err
	}
	if err := s.guard.CourseView(ctx, p, a.CourseID); err != nil {
		return AssignmentDetail{}, err
	}
	out := AssignmentDetail{Assignment: s.assignmentOut(a)}
	if p.IsStudent() {
		sub, err := s.repo.GetSubmissionOf(ctx, id, p.UserID)
		switch {
		case err == nil:
			sub = s.submissionOut(sub)
			out.Submission = &sub
		case !apperr.IsNotFound(err): // «ответа ещё нет» — не ошибка: submission = null
			return AssignmentDetail{}, err
		}
	}
	return out, nil
}

// CreateAssignment: задание создаёт преподаватель курса или администратор
// (контракт: администратор управляет заданиями); created_by — тот, кто создал.
// Оценивание при этом остаётся только за преподавателем курса.
func (s *Service) CreateAssignment(ctx context.Context, p middleware.Principal, courseID int64, in AssignmentInput) (Assignment, error) {
	if err := s.guard.CourseManage(ctx, p, courseID); err != nil {
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
	a, err := s.repo.GetAssignment(ctx, id, true, 0)
	return s.assignmentOut(a), err
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
	a, err := s.repo.GetAssignment(ctx, id, true, 0)
	return s.assignmentOut(a), err
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

func checkGradeFilter(f GradeFilter) error {
	if f.Status != nil && !validStatuses[*f.Status] {
		return queryErr("status", "must be one of not_submitted, missing, submitted, reviewed")
	}
	if utf8.RuneCountInString(f.Q) > 200 {
		return queryErr("q", "must be at most 200 characters")
	}
	return nil
}

// StudentGrades — GET /grades: все задания своих курсов семестра (по умолчанию текущего)
// со статусами и оценками, включая несданные и непроверенные. student_id — только из сессии.
func (s *Service) StudentGrades(ctx context.Context, p middleware.Principal, f GradeFilter) ([]GradeRow, int, error) {
	if err := checkGradeFilter(f); err != nil {
		return nil, 0, err
	}
	if f.SemesterID == nil {
		id, ok, err := s.repo.CurrentSemester(ctx)
		if err != nil {
			return nil, 0, err
		}
		if !ok {
			return nil, 0, apperr.NewConflict("current_semester_not_set", "current semester is not set")
		}
		f.SemesterID = &id
	} else if err := s.repo.SemesterExists(ctx, *f.SemesterID); err != nil {
		return nil, 0, err
	}
	if f.CourseID != nil {
		ok, err := s.guard.StudentOfCourse(ctx, p.UserID, *f.CourseID)
		if err != nil {
			return nil, 0, err
		}
		if !ok {
			return nil, 0, errCourseForbidden
		}
	}
	f.StudentID = &p.UserID
	return s.repo.ListGrades(ctx, f)
}

// CourseGrades — GET /courses/:id/grades: рабочая таблица результатов курса для его
// преподавателя или администратора — все задания всех студентов группы, включая несданные.
func (s *Service) CourseGrades(ctx context.Context, p middleware.Principal, courseID int64, f GradeFilter) ([]GradeRow, int, error) {
	if err := checkGradeFilter(f); err != nil {
		return nil, 0, err
	}
	if err := s.guard.CourseManage(ctx, p, courseID); err != nil {
		return nil, 0, err
	}
	f.CourseID, f.SemesterID = &courseID, nil
	return s.repo.ListGrades(ctx, f)
}
