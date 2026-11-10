package users

import (
	"context"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"university-api/internal/platform/apperr"
	"university-api/internal/platform/middleware"
)

// FileOwnership проверяет, что файл лежит в папке пользователя (реализует access.Checker).
// Возвращает nil, если файл принадлежит userID, иначе — готовую apperr-ошибку
// (не найден / нет доступа), которая отдаётся клиенту как есть.
type FileOwnership interface {
	OwnsFile(ctx context.Context, userID, fileID int64) error
}

// Service — бизнес-правила пакета users: проверка ролей при создании,
// хеширование паролей, нормализация email, подтверждение паролем,
// защита администратора от блокировки самого себя.
// Права доступа к маршрутам проверяются раньше, в роутере.
type Service struct {
	repo  *Repository
	files FileOwnership
}

func NewService(repo *Repository, files FileOwnership) *Service {
	return &Service{repo: repo, files: files}
}

// normEmail приводит email к единому виду (без пробелов, в нижнем регистре),
// чтобы «Ivan@Mail.ru» и «ivan@mail.ru» считались одним адресом
// для ограничения уникальности. При входе email нужно нормализовать так же.
func normEmail(s string) string { return strings.ToLower(strings.TrimSpace(s)) }

// hashPassword хеширует пароль bcrypt со стандартной стоимостью.
// Соль генерируется внутри bcrypt и хранится в самом хеше.
func hashPassword(plain string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	return string(h), err
}

// Get и List — сквозные вызовы репозитория, своих правил у них нет.

func (s *Service) Get(ctx context.Context, id int64) (User, error) { return s.repo.Get(ctx, id) }

func (s *Service) List(ctx context.Context, f ListFilter) ([]User, error) { return s.repo.List(ctx, f) }

// Create создаёт пользователя и возвращает его в полном виде.
// Здесь проверяется то, что не выразить тегами binding: обязательные поля
// зависят от роли. Без этой проверки репозиторий разыменовал бы nil-указатель.
// Лишние поля чужой роли (например, DepartmentID у студента) молча игнорируются.
func (s *Service) Create(ctx context.Context, in CreateUser) (User, error) {
	switch in.Role {
	case "student":
		if in.GroupID == nil {
			return User{}, apperr.NewInvalid("group_required", "group_id is required for a student")
		}
	case "teacher":
		if in.DepartmentID == nil {
			return User{}, apperr.NewInvalid("department_required", "department_id is required for a teacher")
		}
	case "admin":
		// Профиля нет, дополнительных полей не требуется.
	default:
		// Хендлер уже проверяет oneof, но сервис не должен на это полагаться.
		return User{}, apperr.NewInvalid("invalid_role", "role must be student, teacher or admin")
	}

	hash, err := hashPassword(in.Password)
	if err != nil {
		return User{}, err
	}
	in.Email = normEmail(in.Email)
	id, err := s.repo.Create(ctx, in, hash)
	if err != nil {
		return User{}, err
	}
	// Перечитываем, чтобы вернуть вычисляемые поля: роль, ФИО, названия группы и кафедры.
	return s.repo.Get(ctx, id)
}

// teacherNeedsApproval — преподаватель после самостоятельной регистрации неактивен,
// пока его не включит администратор. Роль преподавателя даёт доступ к карточкам
// пользователей и составам групп, поэтому выдавать её без проверки нельзя.
// false — преподаватель входит сразу, как студент.
const teacherNeedsApproval = true

// SelfRegister — самостоятельная регистрация по email и паролю (POST /auth/register).
// Доступна только для ролей student и teacher: администратора так создать нельзя.
// Студент активен сразу, преподаватель — после одобрения (см. teacherNeedsApproval).
// Остальные правила (обязательная группа или кафедра, уникальные email и зачётка)
// те же, что при создании администратором.
func (s *Service) SelfRegister(ctx context.Context, in CreateUser) (User, error) {
	switch in.Role {
	case "student":
		in.PendingApproval = false
	case "teacher":
		in.PendingApproval = teacherNeedsApproval
	default:
		return User{}, apperr.NewInvalid("invalid_role", "role must be student or teacher")
	}
	in.IsHeadman = false // старосту назначает администратор, не сам студент
	return s.Create(ctx, in)
}

// UpdateProfile обновляет данные, которые пользователь меняет о себе,
// и возвращает обновлённый профиль.
// Аватаром можно сделать только свой файл, иначе через этот метод
// можно было бы получить ссылку на чужой файл.
func (s *Service) UpdateProfile(ctx context.Context, id int64, in UpdateProfile) (User, error) {
	if in.AvatarFileID != nil {
		if err := s.files.OwnsFile(ctx, id, *in.AvatarFileID); err != nil {
			return User{}, err
		}
	}
	if err := s.repo.UpdateProfile(ctx, id, in); err != nil {
		return User{}, err
	}
	return s.repo.Get(ctx, id)
}

// AdminUpdate — правка пользователя администратором actor.
// Администратор не может отключить себя или снять с себя права администратора:
// иначе можно случайно остаться без единого активного админа.
// Пароль хешируется здесь, в репозиторий уходит только хеш.
func (s *Service) AdminUpdate(ctx context.Context, actor middleware.Principal, id int64, in AdminUpdate) (User, error) {
	if id == actor.UserID && ((in.IsActive != nil && !*in.IsActive) || (in.IsAdmin != nil && !*in.IsAdmin)) {
		return User{}, apperr.NewInvalid("self_lockout", "you cannot deactivate yourself or remove your own admin rights")
	}
	var hash *string
	if in.Password != nil {
		h, err := hashPassword(*in.Password)
		if err != nil {
			return User{}, err
		}
		hash = &h
	}
	if in.Email != nil {
		e := normEmail(*in.Email)
		in.Email = &e
	}
	if err := s.repo.AdminUpdate(ctx, id, in, hash); err != nil {
		return User{}, err
	}
	return s.repo.Get(ctx, id)
}

// verifyPassword сверяет пароль пользователя; неверный — 401 wrong_password.
// Используется как подтверждение перед чувствительными действиями
// (смена пароля, смена email), даже при активной сессии.
func (s *Service) verifyPassword(ctx context.Context, id int64, plain string) error {
	hash, err := s.repo.PasswordHash(ctx, id)
	if err != nil {
		return err
	}
	// CompareHashAndPassword сравнивает за постоянное время,
	// поэтому по времени ответа пароль не подобрать.
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) != nil {
		return apperr.NewUnauthorized("wrong_password", "current password is incorrect")
	}
	return nil
}

// ChangePassword меняет пароль пользователя после проверки текущего.
func (s *Service) ChangePassword(ctx context.Context, id int64, oldPass, newPass string) error {
	if err := s.verifyPassword(ctx, id, oldPass); err != nil {
		return err
	}
	hash, err := hashPassword(newPass)
	if err != nil {
		return err
	}
	return s.repo.SetPassword(ctx, id, hash)
}

// ChangeEmail меняет почту после подтверждения паролем. Подтверждения самой почты нет:
// в схеме нет поля email_verified_at. Поэтому опечатка в новом адресе
// не обнаружится, пока пользователь не попробует восстановить пароль.
func (s *Service) ChangeEmail(ctx context.Context, id int64, password, newEmail string) (User, error) {
	if err := s.verifyPassword(ctx, id, password); err != nil {
		return User{}, err
	}
	if err := s.repo.SetEmail(ctx, id, normEmail(newEmail)); err != nil {
		return User{}, err
	}
	return s.repo.Get(ctx, id)
}
