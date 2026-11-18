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

// queryErr — 400 invalid_query с указанием параметра.
func queryErr(name, problem string) error {
	return apperr.NewInvalid("invalid_query", "invalid query parameters").
		WithDetails(map[string]any{"fields": map[string][]string{name: {problem}}})
}

// ListSubmissions — таблица «студент — статус» по заданию (для преподавателя курса и админа).
func (s *Service) ListSubmissions(ctx context.Context, p middleware.Principal, assignmentID int64, f SubmissionFilter) ([]SubmissionRow, int, error) {
	if f.Status != nil && !validStatuses[*f.Status] {
		return nil, 0, queryErr("status", "must be one of not_submitted, missing, submitted, reviewed")
	}
	a, err := s.repo.GetAssignment(ctx, assignmentID, false, 0)
	if err != nil {
		return nil, 0, err
	}
	if err := s.guard.CourseManage(ctx, p, a.CourseID); err != nil {
		return nil, 0, err
	}
	return s.repo.ListSubmissionRows(ctx, assignmentID, f)
}

// GetSubmission: ответ видит его автор, преподаватель курса и админ.
func (s *Service) GetSubmission(ctx context.Context, p middleware.Principal, id int64) (Submission, error) {
	sub, err := s.repo.GetSubmission(ctx, id)
	if err != nil {
		return Submission{}, err
	}
	if err := s.canReadSubmission(ctx, p, sub); err != nil {
		return Submission{}, err
	}
	return s.submissionOut(sub), nil
}

var errSubmissionNotFound = apperr.NewNotFound("submission_not_found")

// canReadSubmission: автор видит свой ответ (и черновик); чужой ответ для студента — 404,
// чтобы не раскрывать его существование. Преподавателю курса и админу черновики не
// показываются (404): черновик — личное дело студента, пока он его не отправил.
func (s *Service) canReadSubmission(ctx context.Context, p middleware.Principal, sub Submission) error {
	if sub.StudentID == p.UserID {
		return nil
	}
	if p.IsStudent() {
		return errSubmissionNotFound
	}
	if err := s.guard.CourseManage(ctx, p, sub.CourseID); err != nil {
		return err
	}
	if sub.Status == StatusNotSubmitted {
		return errSubmissionNotFound
	}
	return nil
}

// MySubmission — собственный ответ студента на задание; nil, если ответа ещё нет (data: null).
func (s *Service) MySubmission(ctx context.Context, p middleware.Principal, assignmentID int64) (*Submission, error) {
	if _, err := s.studentAssignment(ctx, p, assignmentID); err != nil {
		return nil, err
	}
	sub, err := s.repo.GetSubmissionOf(ctx, assignmentID, p.UserID)
	if apperr.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	sub = s.submissionOut(sub)
	return &sub, nil
}

// studentAssignment: задание существует и студент учится в группе его курса (иначе 403 course_forbidden).
func (s *Service) studentAssignment(ctx context.Context, p middleware.Principal, assignmentID int64) (Assignment, error) {
	a, err := s.repo.GetAssignment(ctx, assignmentID, false, 0)
	if err != nil {
		return Assignment{}, err
	}
	ok, err := s.guard.StudentOfCourse(ctx, p.UserID, a.CourseID)
	if err != nil {
		return Assignment{}, err
	}
	if !ok {
		return Assignment{}, errCourseForbidden
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
		return Submission{}, apperr.NewInvalid("empty_submission", "attach a file or write an answer before submitting")
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
	sub, err := s.repo.GetSubmission(ctx, id)
	return s.submissionOut(sub), err
}

// DeleteSubmission — студент отзывает свой непроверенный ответ: удаляются текст и связи
// с файлами, статус снова вычисляется как not_submitted/missing. Повтор без ответа — 204.
func (s *Service) DeleteSubmission(ctx context.Context, p middleware.Principal, assignmentID int64) error {
	if _, err := s.studentAssignment(ctx, p, assignmentID); err != nil {
		return err
	}
	keys, err := s.repo.DeleteSubmission(ctx, assignmentID, p.UserID)
	s.removeBlobs(keys)
	return err
}

// Review — оценка и комментарий; только преподаватель этого курса (по нагрузке): администрирование
// само по себе права оценивать не даёт. reviewed_by ссылается на teacher. Повторная оценка
// уже проверенной работы разрешена.
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
	sub, err = s.repo.GetSubmission(ctx, id)
	return s.submissionOut(sub), err
}
