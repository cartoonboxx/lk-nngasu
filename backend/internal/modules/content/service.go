package content

import (
	"context"

	"university-api/internal/platform/apperr"
	"university-api/internal/platform/middleware"
	"university-api/internal/platform/validate"
)

// CourseGuard — проверки доступа к курсу и файлам (реализует access.Checker).
// Каждая проверка возвращает nil, если доступ есть, иначе — готовую ошибку для клиента.
type CourseGuard interface {
	// CourseView — пользователь может просматривать курс (участник курса или админ).
	CourseView(ctx context.Context, p middleware.Principal, courseID int64) error
	// CourseManage — пользователь может управлять содержимым курса (его преподаватель или админ).
	CourseManage(ctx context.Context, p middleware.Principal, courseID int64) error
	// CourseTeacher — пользователь именно преподаватель этого курса (админ не проходит).
	CourseTeacher(ctx context.Context, p middleware.Principal, courseID int64) error
	// OwnsFile — файл загружен этим пользователем.
	OwnsFile(ctx context.Context, userID, fileID int64) error
}

// Blobs стирает файлы с диска (реализует files.Storage).
// Нужен, потому что при удалении материалов удаляются и их файлы.
type Blobs interface {
	Remove(key string) error
}

// Service — бизнес-логика тем и материалов: проверка прав через guard,
// валидация материалов, работа с БД через repo и уборка файлов с диска через blobs.
type Service struct {
	repo  *Repository
	guard CourseGuard
	blobs Blobs
}

// NewService создаёт сервис модуля content.
func NewService(repo *Repository, guard CourseGuard, blobs Blobs) *Service {
	return &Service{repo: repo, guard: guard, blobs: blobs}
}

// removeBlobs — best effort: запись в БД уже удалена, сбой диска не должен ронять запрос.
// В худшем случае на диске останется файл без записи.
// Вызывать только после успешной транзакции, иначе можно стереть файлы,
// записи которых после отката остались в БД.
func (s *Service) removeBlobs(keys []string) {
	for _, k := range keys {
		_ = s.blobs.Remove(k)
	}
}

// ListTopics возвращает темы курса. Доступно всем, кто может просматривать курс.
func (s *Service) ListTopics(ctx context.Context, p middleware.Principal, courseID int64) ([]Topic, error) {
	if err := s.guard.CourseView(ctx, p, courseID); err != nil {
		return nil, err
	}
	return s.repo.ListTopics(ctx, courseID)
}

// topicForManage загружает тему и проверяет, что её курс пользователь вправе менять.
// Тему грузим первой: в пути есть только id темы, а курс узнаём из неё.
func (s *Service) topicForManage(ctx context.Context, p middleware.Principal, id int64) (Topic, error) {
	t, err := s.repo.GetTopic(ctx, id)
	if err != nil {
		return Topic{}, err
	}
	return t, s.guard.CourseManage(ctx, p, t.CourseID)
}

// CreateTopic создаёт тему в курсе. Нужно право управлять курсом.
// После вставки тема перечитывается, чтобы вернуть её с вычисленными полями (позиция и т.п.).
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

// UpdateTopic изменяет тему и возвращает её актуальное состояние.
func (s *Service) UpdateTopic(ctx context.Context, p middleware.Principal, id int64, in TopicInput) (Topic, error) {
	if _, err := s.topicForManage(ctx, p, id); err != nil {
		return Topic{}, err
	}
	if err := s.repo.UpdateTopic(ctx, id, in); err != nil {
		return Topic{}, err
	}
	return s.repo.GetTopic(ctx, id)
}

// DeleteTopic удаляет тему вместе с её материалами.
// repo возвращает ключи файлов, освободившихся при удалении, — стираем их с диска.
func (s *Service) DeleteTopic(ctx context.Context, p middleware.Principal, id int64) error {
	if _, err := s.topicForManage(ctx, p, id); err != nil {
		return err
	}
	keys, err := s.repo.DeleteTopic(ctx, id)
	if err != nil {
		return err // транзакция откатилась — с диска ничего не удаляем
	}
	s.removeBlobs(keys)
	return nil
}

// checkMaterial: нужен файл или ссылка (можно оба), ссылка — http(s), файл — свой
// (или уже привязанный к этому материалу).
//
// current — id файла, который уже привязан к материалу (nil при создании).
// Если файл не меняется, владение не проверяем: так другой преподаватель курса
// или админ может править материал, не будучи автором файла.
func (s *Service) checkMaterial(ctx context.Context, p middleware.Principal, in MaterialInput, current *int64) error {
	if in.FileID == nil && in.ExternalURL == nil {
		return apperr.NewInvalid("material_source", "provide file_id or external_url")
	}
	// Только http(s): ссылки вида javascript: или data: не пропускаем.
	if in.ExternalURL != nil && !validate.HTTPURL(*in.ExternalURL) {
		return apperr.NewInvalid("bad_url", "external_url must be an http(s) link")
	}
	// Новый файл можно прикрепить только свой — чужой файл по id не привязать.
	if in.FileID != nil && (current == nil || *current != *in.FileID) {
		return s.guard.OwnsFile(ctx, p.UserID, *in.FileID)
	}
	return nil
}

// CreateMaterial: created_by ссылается на teacher, поэтому материалы добавляет преподаватель курса.
// Админ создать материал не может (он не преподаватель), но может изменять и удалять — см. materialForManage.
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

// materialForManage загружает материал и проверяет, что его курс пользователь вправе менять.
func (s *Service) materialForManage(ctx context.Context, p middleware.Principal, id int64) (Material, error) {
	m, err := s.repo.GetMaterial(ctx, id)
	if err != nil {
		return Material{}, err
	}
	return m, s.guard.CourseManage(ctx, p, m.CourseID)
}

// UpdateMaterial изменяет материал. Если файл заменили или убрали,
// repo возвращает ключ старого файла — стираем его с диска.
func (s *Service) UpdateMaterial(ctx context.Context, p middleware.Principal, id int64, in MaterialInput) (Material, error) {
	m, err := s.materialForManage(ctx, p, id)
	if err != nil {
		return Material{}, err
	}
	if err := s.checkMaterial(ctx, p, in, m.FileID); err != nil {
		return Material{}, err
	}
	keys, err := s.repo.UpdateMaterial(ctx, id, in)
	if err != nil {
		return Material{}, err // транзакция откатилась — с диска ничего не удаляем
	}
	s.removeBlobs(keys)
	return s.repo.GetMaterial(ctx, id)
}

// DeleteMaterial удаляет материал и его файл с диска (если файл больше нигде не используется).
func (s *Service) DeleteMaterial(ctx context.Context, p middleware.Principal, id int64) error {
	if _, err := s.materialForManage(ctx, p, id); err != nil {
		return err
	}
	keys, err := s.repo.DeleteMaterial(ctx, id)
	if err != nil {
		return err // транзакция откатилась — с диска ничего не удаляем
	}
	s.removeBlobs(keys)
	return nil
}
