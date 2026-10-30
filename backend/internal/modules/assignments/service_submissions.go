package assignments

import (
	"context"
	"strings"

	"university-api/internal/platform/apperr"
	"university-api/internal/platform/middleware"
)

var validStatuses = map[string]bool{
	StatusNotSubmitted: true, StatusMissing: true, StatusSubmitted: true, StatusReviewed: true,
}

// ListSubmissions — таблица «студент — статус» по заданию (для преподавателя курса и админа).
func (s *Service) ListSubmissions(ctx context.Context, p middleware.Principal, assignmentID int64, f SubmissionFilter) ([]SubmissionRow, error) {
	if f.Status != nil && !validStatuses[*f.Status] {
		return nil, apperr.NewInvalid("invalid_query", "invalid query parameter: status")
	}
	a, err := s.repo.GetAssignment(ctx, assignmentID, false, 0)
	if err != nil {
		return nil, err
	}
	if err := s.guard.CourseManage(ctx, p, a.CourseID); err != nil {
		return nil, err
	}
	return s.repo.ListSubmissionRows(ctx, assignmentID, f)
}

// GetSubmission: ответ видит его автор и преподаватели курса (и админ).
func (s *Service) GetSubmission(ctx context.Context, p middleware.Principal, id int64) (Submission, error) {
	sub, err := s.repo.GetSubmission(ctx, id)
	if err != nil {
		return Submission{}, err
	}
	return sub, s.canReadSubmission(ctx, p, sub)
}

func (s *Service) canReadSubmission(ctx context.Context, p middleware.Principal, sub Submission) error {
	if p.IsStudent() {
		if sub.StudentID != p.UserID {
			return apperr.NewNotFound("submission_not_found") // чужие ответы не раскрываем
		}
		return nil
	}
	return s.guard.CourseManage(ctx, p, sub.CourseID)
}

// MySubmission — собственный ответ студента на задание; пока его нет, возвращается пустой черновик.
func (s *Service) MySubmission(ctx context.Context, p middleware.Principal, assignmentID int64) (Submission, error) {
	a, err := s.studentAssignment(ctx, p, assignmentID)
	if err != nil {
		return Submission{}, err
	}
	sub, err := s.repo.GetSubmissionOf(ctx, assignmentID, p.UserID)
	if apperr.IsNotFound(err) {
		return Submission{AssignmentID: a.ID, CourseID: a.CourseID, StudentID: p.UserID,
			Status: StatusNotSubmitted, Files: []FileRef{}}, nil
	}
	return sub, err
}

// studentAssignment: задание существует и студент учится в группе его курса.
func (s *Service) studentAssignment(ctx context.Context, p middleware.Principal, assignmentID int64) (Assignment, error) {
	if !p.IsStudent() {
		return Assignment{}, apperr.NewForbidden("students_only", "only students submit answers")
	}
	a, err := s.repo.GetAssignment(ctx, assignmentID, false, 0)
	if err != nil {
		return Assignment{}, err
	}
	ok, err := s.guard.StudentOfCourse(ctx, p.UserID, a.CourseID)
	if err != nil {
		return Assignment{}, err
	}
	if !ok {
		return Assignment{}, apperr.NewNotFound("assignment_not_found")
	}
	return a, nil
}

// SaveSubmission — студент сохраняет черновик или отправляет ответ. Опоздание допускается
// (помечается is_late), проверенный ответ менять нельзя.
func (s *Service) SaveSubmission(ctx context.Context, p middleware.Principal, assignmentID int64, in SubmissionInput) (Submission, error) {
	if _, err := s.studentAssignment(ctx, p, assignmentID); err != nil {
		return Submission{}, err
	}
	if in.AnswerText != nil {
		t := strings.TrimSpace(*in.AnswerText)
		if t == "" {
			in.AnswerText = nil
		} else {
			in.AnswerText = &t
		}
	}
	if in.Submit && in.AnswerText == nil && len(in.FileIDs) == 0 {
		return Submission{}, apperr.NewInvalid("empty_answer", "attach a file or write an answer before submitting")
	}

	keep := map[int64]bool{}
	if cur, err := s.repo.GetSubmissionOf(ctx, assignmentID, p.UserID); err == nil {
		for _, f := range cur.Files {
			keep[f.ID] = true
		}
	} else if !apperr.IsNotFound(err) {
		return Submission{}, err
	}
	if err := s.checkFiles(ctx, p, in.FileIDs, keep); err != nil {
		return Submission{}, err
	}

	id, keys, err := s.repo.SaveSubmission(ctx, assignmentID, p.UserID, in)
	s.removeBlobs(keys)
	if err != nil {
		return Submission{}, err
	}
	return s.repo.GetSubmission(ctx, id)
}

// Review — оценка и комментарий; только преподаватель этого курса (reviewed_by ссылается на teacher).
func (s *Service) Review(ctx context.Context, p middleware.Principal, id int64, in ReviewInput) (Submission, error) {
	sub, err := s.repo.GetSubmission(ctx, id)
	if err != nil {
		return Submission{}, err
	}
	if err := s.guard.CourseTeacher(ctx, p, sub.CourseID); err != nil {
		return Submission{}, err
	}
	if err := validateReview(&in); err != nil {
		return Submission{}, err
	}
	if err := s.repo.Review(ctx, id, p.UserID, in); err != nil {
		return Submission{}, err
	}
	return s.repo.GetSubmission(ctx, id)
}
