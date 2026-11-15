package users

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
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

// Inviter отправляет пользователю приглашение задать пароль (реализует auth.Service):
// создаёт одноразовый токен и отправляет письмо со ссылкой на /activation.
type Inviter interface {
	Invite(ctx context.Context, userID, invitedBy int64) error
}

// Options — настройки пакета из конфигурации приложения.
type Options struct {
	// TermsVersion — версия условий использования, которую поддерживает сервер.
	// Заявка с другой версией (фронт показал устаревшие условия) отклоняется.
	TermsVersion string

	// Inviter — отправка приглашений. Без него администратор не может создавать
	// пользователей: пароль при создании не передаётся.
	Inviter Inviter

	// RegisterLimit — ограничение частоты подачи заявок (429 rate_limited); nil — без ограничения.
	RegisterLimit gin.HandlerFunc
}

// Service — бизнес-правила пакета users: нормализация ФИО, email и документов,
// проверка владения аватаром, хеширование паролей, защита от потери
// последнего администратора. Права доступа к маршрутам проверяются раньше, в роутере.
type Service struct {
	repo  *Repository
	files FileOwnership
	opts  Options
}

func NewService(repo *Repository, files FileOwnership, opts Options) *Service {
	return &Service{repo: repo, files: files, opts: opts}
}

// --- Нормализация и проверки, которые не выразить тегами binding ---

// normEmail приводит email к единому виду (без пробелов, в нижнем регистре):
// в БД стоит CHECK email = lower(email), вход ищет email в том же виде.
func normEmail(s string) string { return strings.ToLower(strings.TrimSpace(s)) }

// maxNameLen — предел длины фамилии, имени и отчества (varchar(50) в БД).
const maxNameLen = 50

// normName — фамилия или имя: 1–50 символов после trim.
func normName(field, s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", fieldErr(field, "must not be empty")
	}
	if utf8.RuneCountInString(s) > maxNameLen {
		return "", fieldErr(field, "must be at most 50 characters")
	}
	return s, nil
}

// normMiddle — отчество: до 50 символов; пустая строка равна отсутствию отчества.
func normMiddle(field string, s *string) (*string, error) {
	if s == nil {
		return nil, nil
	}
	v := strings.TrimSpace(*s)
	if v == "" {
		return nil, nil
	}
	if utf8.RuneCountInString(v) > maxNameLen {
		return nil, fieldErr(field, "must be at most 50 characters")
	}
	return &v, nil
}

// normRecordBook — номер зачётной книжки: 1–50 символов после trim.
func normRecordBook(field string, r *StudentRole) error {
	if r == nil || r.RecordBookNumber == nil {
		return nil
	}
	v, err := normName(field, *r.RecordBookNumber)
	if err != nil {
		return err
	}
	r.RecordBookNumber = &v
	return nil
}

// checkEmail — тот же валидатор, что у тега binding:"email", для полей PATCH.
func checkEmail(field, s string) error {
	if len(s) > 255 {
		return fieldErr(field, "must be at most 255 characters")
	}
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok && v.Var(s, "required,email") != nil {
		return fieldErr(field, "must be a valid email")
	}
	return nil
}

// notNull — ключ передан со значением null там, где очистка не предусмотрена.
func notNull[T any](field string, f Field[T]) error {
	if f.Null() {
		return fieldErr(field, "must not be null")
	}
	return nil
}

// errEmptyPatch — PATCH без единого известного ключа.
var errEmptyPatch = apperr.NewInvalid("invalid_request", "at least one field is required")

// hashPassword хеширует пароль bcrypt со стандартной стоимостью.
// Соль генерируется внутри bcrypt и хранится в самом хеше.
//
// bcrypt учитывает не больше 72 байт. Ограничение binding max=72 считает символы,
// а кириллическая буква занимает 2 байта, поэтому длину проверяем ещё и в байтах.
func hashPassword(plain string) (string, error) {
	if len(plain) > maxPasswordBytes {
		return "", errPasswordTooLong
	}
	h, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	return string(h), err
}

// maxPasswordBytes — предел длины пароля у bcrypt.
const maxPasswordBytes = 72

var errPasswordTooLong = fieldErr("new_password", "must be at most 72 bytes in UTF-8")

// --- Чтение ---

// Get возвращает полную карточку (AdminUser). Хендлер сам решает,
// отдать её целиком (администратору) или только профиль.
func (s *Service) Get(ctx context.Context, id int64) (AdminUser, error) { return s.repo.Get(ctx, id) }

// List возвращает страницу UserSummary и общее число записей под фильтром.
func (s *Service) List(ctx context.Context, f ListFilter) ([]User, int, error) {
	return s.repo.List(ctx, f)
}

// --- Администратор ---

// Create создаёт пользователя без пароля и отправляет приглашение на email.
// Учебные роли независимы: можно задать обе, одну или ни одной.
// Аватар — только файл, загруженный самим администратором.
func (s *Service) Create(ctx context.Context, in CreateUser) (AdminUser, error) {
	if s.opts.Inviter == nil {
		return AdminUser{}, apperr.New(apperr.Unavailable, "invitations_disabled", "invitations are not configured")
	}
	var err error
	if in.LastName, err = normName("last_name", in.LastName); err != nil {
		return AdminUser{}, err
	}
	if in.FirstName, err = normName("first_name", in.FirstName); err != nil {
		return AdminUser{}, err
	}
	if in.MiddleNames, err = normMiddle("middle_names", in.MiddleNames); err != nil {
		return AdminUser{}, err
	}
	if in.Passport, err = normPassport("passport", in.Passport); err != nil {
		return AdminUser{}, err
	}
	if in.Snils, err = normSnils("snils", in.Snils); err != nil {
		return AdminUser{}, err
	}
	if err := normRecordBook("student.record_book_number", in.Student); err != nil {
		return AdminUser{}, err
	}
	if in.AvatarFileID != nil {
		if err := s.files.OwnsFile(ctx, in.CreatedBy, *in.AvatarFileID); err != nil {
			return AdminUser{}, err
		}
	}
	in.Email = normEmail(in.Email)

	// Пароль пустой: войти нельзя, пока пользователь не активирует приглашение.
	id, err := s.repo.Create(ctx, in, "")
	if err != nil {
		return AdminUser{}, err
	}
	// Учётка уже создана. Если письмо не ушло — ошибка 503 invitation_not_sent,
	// администратор может повторить отправку через POST /users/:id/invitation.
	if err := s.opts.Inviter.Invite(ctx, id, in.CreatedBy); err != nil {
		return AdminUser{}, err
	}
	return s.repo.Get(ctx, id)
}

// ResendInvitation отправляет приглашение заново (прежние ссылки перестают работать).
// Пользователь, уже задавший пароль, — 409 already_activated.
func (s *Service) ResendInvitation(ctx context.Context, userID, adminID int64) (InvitationReceipt, error) {
	if s.opts.Inviter == nil {
		return InvitationReceipt{}, apperr.New(apperr.Unavailable, "invitations_disabled", "invitations are not configured")
	}
	if err := s.opts.Inviter.Invite(ctx, userID, adminID); err != nil {
		return InvitationReceipt{}, err
	}
	return InvitationReceipt{UserID: userID, DeliveryStatus: "queued"}, nil
}

// AdminUpdate — частичная правка администратором actor (UserPatchInput).
//
// Правила документации: нужен хотя бы один ключ; null очищает отчество, снимает
// роль или аватар; роль с учебными данными или нагрузкой не снимается (409);
// последнего активного администратора заблокировать нельзя (409).
// Дополнительно администратор не может заблокировать себя или снять с себя права.
func (s *Service) AdminUpdate(ctx context.Context, actor middleware.Principal, id int64, in AdminUpdate) (AdminUser, error) {
	if !(in.LastName.Set || in.FirstName.Set || in.MiddleNames.Set || in.Email.Set || in.Passport.Set ||
		in.Snils.Set || in.Student.Set || in.Teacher.Set || in.AvatarFileID.Set || in.IsActive.Set || in.IsAdmin.Set) {
		return AdminUser{}, errEmptyPatch
	}
	for _, e := range []error{
		notNull("last_name", in.LastName), notNull("first_name", in.FirstName), notNull("email", in.Email),
		notNull("passport", in.Passport), notNull("snils", in.Snils),
		notNull("is_active", in.IsActive), notNull("is_admin", in.IsAdmin),
	} {
		if e != nil {
			return AdminUser{}, e
		}
	}

	var err error
	if in.LastName.Value != nil {
		if *in.LastName.Value, err = normName("last_name", *in.LastName.Value); err != nil {
			return AdminUser{}, err
		}
	}
	if in.FirstName.Value != nil {
		if *in.FirstName.Value, err = normName("first_name", *in.FirstName.Value); err != nil {
			return AdminUser{}, err
		}
	}
	if in.MiddleNames.Value, err = normMiddle("middle_names", in.MiddleNames.Value); err != nil {
		return AdminUser{}, err
	}
	if in.Email.Value != nil {
		*in.Email.Value = normEmail(*in.Email.Value)
		if err := checkEmail("email", *in.Email.Value); err != nil {
			return AdminUser{}, err
		}
	}
	if in.Passport.Value != nil {
		if *in.Passport.Value, err = normPassport("passport", *in.Passport.Value); err != nil {
			return AdminUser{}, err
		}
	}
	if in.Snils.Value != nil {
		if *in.Snils.Value, err = normSnils("snils", *in.Snils.Value); err != nil {
			return AdminUser{}, err
		}
	}
	if err := normRecordBook("student.record_book_number", in.Student.Value); err != nil {
		return AdminUser{}, err
	}
	if in.AvatarFileID.Value != nil {
		if err := s.files.OwnsFile(ctx, actor.UserID, *in.AvatarFileID.Value); err != nil {
			return AdminUser{}, err
		}
	}
	if id == actor.UserID && ((in.IsActive.Value != nil && !*in.IsActive.Value) || (in.IsAdmin.Value != nil && !*in.IsAdmin.Value)) {
		return AdminUser{}, apperr.NewConflict("self_lockout", "you cannot deactivate yourself or remove your own admin rights")
	}

	if err := s.repo.AdminUpdate(ctx, id, in); err != nil {
		return AdminUser{}, err
	}
	return s.repo.Get(ctx, id)
}

// --- Свой профиль ---

// UpdateProfile обновляет ФИО и аватар (ProfilePatchInput) и возвращает профиль.
// Аватаром можно сделать только свой файл, иначе через этот метод
// можно было бы получить доступ к чужому файлу.
func (s *Service) UpdateProfile(ctx context.Context, id int64, in UpdateProfile) (User, error) {
	if !(in.LastName.Set || in.FirstName.Set || in.MiddleNames.Set || in.AvatarFileID.Set) {
		return User{}, errEmptyPatch
	}
	if err := notNull("last_name", in.LastName); err != nil {
		return User{}, err
	}
	if err := notNull("first_name", in.FirstName); err != nil {
		return User{}, err
	}
	var err error
	if in.LastName.Value != nil {
		if *in.LastName.Value, err = normName("last_name", *in.LastName.Value); err != nil {
			return User{}, err
		}
	}
	if in.FirstName.Value != nil {
		if *in.FirstName.Value, err = normName("first_name", *in.FirstName.Value); err != nil {
			return User{}, err
		}
	}
	if in.MiddleNames.Value, err = normMiddle("middle_names", in.MiddleNames.Value); err != nil {
		return User{}, err
	}
	if in.AvatarFileID.Value != nil {
		if err := s.files.OwnsFile(ctx, id, *in.AvatarFileID.Value); err != nil {
			return User{}, err
		}
	}
	if err := s.repo.UpdateProfile(ctx, id, in); err != nil {
		return User{}, err
	}
	u, err := s.repo.Get(ctx, id)
	return u.User, err
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

// ChangePassword меняет пароль после проверки текущего (PasswordInput).
// Новый пароль должен отличаться от старого. Все сессии, кроме текущей (sid), отзываются.
func (s *Service) ChangePassword(ctx context.Context, id int64, sid, oldPass, newPass string) error {
	if oldPass == newPass {
		return fieldErr("new_password", "must differ from the current password")
	}
	if err := s.verifyPassword(ctx, id, oldPass); err != nil {
		return err
	}
	hash, err := hashPassword(newPass)
	if err != nil {
		return err
	}
	return s.repo.SetPassword(ctx, id, hash, sid) // остальные сессии закрываются
}

// ChangeEmail меняет почту после подтверждения паролем. Подтверждения самой почты нет:
// в схеме нет поля email_verified_at.
func (s *Service) ChangeEmail(ctx context.Context, id int64, password, newEmail string) (User, error) {
	if err := s.verifyPassword(ctx, id, password); err != nil {
		return User{}, err
	}
	if err := s.repo.SetEmail(ctx, id, normEmail(newEmail)); err != nil {
		return User{}, err
	}
	u, err := s.repo.Get(ctx, id)
	return u.User, err
}
