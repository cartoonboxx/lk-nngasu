package content

import (
	"context"

	"university-api/internal/platform/apperr"
	"university-api/internal/platform/middleware"
	"university-api/internal/platform/validate"
)

// CourseGuard — проверки доступа к курсу и файлам (реализует access.Checker).
type CourseGuard interface {
	CourseView(ctx context.Context, p middleware.Principal, courseID int64) error
	CourseManage(ctx context.Context, p middleware.Principal, courseID int64) error
	CourseTeacher(ctx context.Context, p middleware.Principal, courseID int64) error
	OwnsFile(ctx context.Context, userID, fileID int64) error
}

// Blobs стирает файлы с диска (реализует files.Storage).
type Blobs interface {
	Remove(key string) error
}

type Service struct {
	repo  *Repository
	guard CourseGuard
	blobs Blobs
}

func NewService(repo *Repository, guard CourseGuard, blobs Blobs) *Service {
	return &Service{repo: repo, guard: guard, blobs: blobs}
}

// removeBlobs — best effort: запись в БД уже удалена, сбой диска не должен ронять запрос.
func (s *Service) removeBlobs(keys []string) {
	for _, k := range keys {
		_ = s.blobs.Remove(k)
	}
}

func (s *Service) ListTopics(ctx context.Context, p middleware.Principal, courseID int64) ([]Topic, error) {
	if err := s.guard.CourseView(ctx, p, courseID); err != nil {
		return nil, err
	}
	return s.repo.ListTopics(ctx, courseID)
}

// topicForManage загружает тему и проверяет, что её курс пользователь вправе менять.
func (s *Service) topicForManage(ctx context.Context, p middleware.Principal, id int64) (Topic, error) {
	t, err := s.repo.GetTopic(ctx, id)
	if err != nil {
		return Topic{}, err
	}
	return t, s.guard.CourseManage(ctx, p, t.CourseID)
}

func (s *Service) CreateTopic(ctx context.Context, p middleware.Principal, courseID int64, in TopicInput) (Topic, error) {
	if err := s.guard.CourseManage(ctx, p, courseID); err != nil {
		return Topic{}, err
	}
	id, err := s.repo.CreateTopic(ctx, courseID, in)
	if err != nil {
		return Topic{}, err
	}
	return s.repo.GetTopic(ctx, id)
}

func (s *Service) UpdateTopic(ctx context.Context, p middleware.Principal, id int64, in TopicInput) (Topic, error) {
	if _, err := s.topicForManage(ctx, p, id); err != nil {
		return Topic{}, err
	}
	if err := s.repo.UpdateTopic(ctx, id, in); err != nil {
		return Topic{}, err
	}
	return s.repo.GetTopic(ctx, id)
}

func (s *Service) DeleteTopic(ctx context.Context, p middleware.Principal, id int64) error {
	if _, err := s.topicForManage(ctx, p, id); err != nil {
		return err
	}
	keys, err := s.repo.DeleteTopic(ctx, id)
	s.removeBlobs(keys)
	return err
}

// checkMaterial: нужен файл или ссылка (можно оба), ссылка — http(s), файл — свой
// (или уже привязанный к этому материалу).
func (s *Service) checkMaterial(ctx context.Context, p middleware.Principal, in MaterialInput, current *int64) error {
	if in.FileID == nil && in.ExternalURL == nil {
		return apperr.NewInvalid("material_source", "provide file_id or external_url")
	}
	if in.ExternalURL != nil && !validate.HTTPURL(*in.ExternalURL) {
		return apperr.NewInvalid("bad_url", "external_url must be an http(s) link")
	}
	if in.FileID != nil && (current == nil || *current != *in.FileID) {
		return s.guard.OwnsFile(ctx, p.UserID, *in.FileID)
	}
	return nil
}

// CreateMaterial: created_by ссылается на teacher, поэтому материалы добавляет преподаватель курса.
func (s *Service) CreateMaterial(ctx context.Context, p middleware.Principal, topicID int64, in MaterialInput) (Material, error) {
	t, err := s.repo.GetTopic(ctx, topicID)
	if err != nil {
		return Material{}, err
	}
	if err := s.guard.CourseTeacher(ctx, p, t.CourseID); err != nil {
		return Material{}, err
	}
	if err := s.checkMaterial(ctx, p, in, nil); err != nil {
		return Material{}, err
	}
	id, err := s.repo.CreateMaterial(ctx, topicID, p.UserID, in)
	if err != nil {
		return Material{}, err
	}
	return s.repo.GetMaterial(ctx, id)
}

func (s *Service) materialForManage(ctx context.Context, p middleware.Principal, id int64) (Material, error) {
	m, err := s.repo.GetMaterial(ctx, id)
	if err != nil {
		return Material{}, err
	}
	return m, s.guard.CourseManage(ctx, p, m.CourseID)
}

func (s *Service) UpdateMaterial(ctx context.Context, p middleware.Principal, id int64, in MaterialInput) (Material, error) {
	m, err := s.materialForManage(ctx, p, id)
	if err != nil {
		return Material{}, err
	}
	if err := s.checkMaterial(ctx, p, in, m.FileID); err != nil {
		return Material{}, err
	}
	keys, err := s.repo.UpdateMaterial(ctx, id, in)
	s.removeBlobs(keys)
	if err != nil {
		return Material{}, err
	}
	return s.repo.GetMaterial(ctx, id)
}

func (s *Service) DeleteMaterial(ctx context.Context, p middleware.Principal, id int64) error {
	if _, err := s.materialForManage(ctx, p, id); err != nil {
		return err
	}
	keys, err := s.repo.DeleteMaterial(ctx, id)
	s.removeBlobs(keys)
	return err
}
