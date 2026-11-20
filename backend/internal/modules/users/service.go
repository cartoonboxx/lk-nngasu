package users

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
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

// Inviter создаёт приглашение задать пароль (реализует auth.Service):
// одноразовый токен и письмо со ссылкой на /activation в очереди отправки.
type Inviter interface {
	// Invite — в своей транзакции (повторная отправка).
	Invite(ctx context.Context, userID, invitedBy int64) error
	// InviteTx — в транзакции вызывающего: создание пользователя и приглашение атомарны.
	InviteTx(ctx context.Context, tx pgx.Tx, userID, invitedBy int64, grantsAdmin bool) error
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

	// PasswordLimit — ограничение попыток смены пароля (429 rate_limited): защита
	// от подбора текущего пароля через украденную сессию; nil — без ограничения.
	PasswordLimit gin.HandlerFunc
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

// errEmptyPatch — PATCH без единого известного ключа (400 empty_patch по документации).
var errEmptyPatch = apperr.NewInvalid("empty_patch", "at least one field is required")

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

// checkRoleProfiles — роль одна, профиль student/teacher только у своей роли.
func checkRoleProfiles(role string, st *StudentRole, te *TeacherRole) error {
	switch {
	case role != RoleStudent && role != RoleTeacher && role != RoleAdmin:
		return errInvalidRole("role", "must be student, teacher or admin")
	case role == RoleStudent && st == nil:
		return errInvalidRole("student", "is required for role student")
	case role == RoleTeacher && te == nil:
		return errInvalidRole("teacher", "is required for role teacher")
	case role != RoleStudent && st != nil:
		return errInvalidRole("student", "must be null unless role is student")
	case role != RoleTeacher && te != nil:
		return errInvalidRole("teacher", "must be null unless role is teacher")
	}
	return nil
}

// Create создаёт пользователя (UserCreateInput). С паролем учётка сразу активна
// (хранится только хеш); без пароля — приглашение на email, пароль пользователь
// задаёт сам (расширение /auth/activate). Административный доступ выдаётся в отдельном разделе.
// Аватар — только файл, загруженный самим администратором.
func (s *Service) Create(ctx context.Context, in CreateUser) (AdminUser, error) {
	if err := checkRoleProfiles(in.Role, in.Student, in.Teacher); err != nil {
		return AdminUser{}, err
	}
	if in.Role == RoleAdmin {
		return AdminUser{}, apperr.NewForbidden("admin_invitation_required", "administrator access must be granted through administrator management")
	}
	if in.Password == "" && s.opts.Inviter == nil {
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
		if err := s.repo.CheckAvatar(ctx, in.CreatedBy, *in.AvatarFileID); err != nil {
			return AdminUser{}, err
		}
	}
	in.Email = normEmail(in.Email)

	if in.Password != "" {
		hash, err := hashPassword(in.Password)
		if err != nil {
			return AdminUser{}, fieldErr("password", "must be at most 72 bytes in UTF-8")
		}
		id, err := s.repo.Create(ctx, in, hash, nil)
		if err != nil {
			return AdminUser{}, err
		}
		return s.repo.Get(ctx, id)
	}
	// Без пароля войти нельзя, пока пользователь не активирует приглашение.
	// Пользователь и письмо в очереди создаются одной транзакцией.
	id, err := s.repo.Create(ctx, in, "", func(tx pgx.Tx, id int64) error {
		return s.opts.Inviter.InviteTx(ctx, tx, id, in.CreatedBy, false)
	})
	if err != nil {
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
		in.Snils.Set || in.Student.Set || in.Teacher.Set || in.AvatarFileID.Set || in.IsActive.Set || in.Role.Set) {
		return AdminUser{}, errEmptyPatch
	}
	if r := in.Role.Value; r != nil && *r != RoleStudent && *r != RoleTeacher && *r != RoleAdmin {
		return AdminUser{}, errInvalidRole("role", "must be student, teacher or admin")
	}
	for _, e := range []error{
		notNull("last_name", in.LastName), notNull("first_name", in.FirstName), notNull("email", in.Email),
		notNull("passport", in.Passport), notNull("snils", in.Snils),
		notNull("is_active", in.IsActive), notNull("role", in.Role),
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
		if err := s.repo.CheckAvatar(ctx, actor.UserID, *in.AvatarFileID.Value); err != nil {
			return AdminUser{}, err
		}
	}
	if id == actor.UserID && in.IsActive.Value != nil && !*in.IsActive.Value {
		return AdminUser{}, apperr.NewConflict("self_lockout", "you cannot block yourself")
	}

	if err := s.repo.AdminUpdate(ctx, id, actor.UserID, in); err != nil {
		return AdminUser{}, err
	}
	return s.repo.Get(ctx, id)
}

// GroupStudents — состав группы. Администратору доступна любая группа,
// преподавателю — только та, у которой он ведёт хотя бы один курс.
// Иначе 404 group_not_found: по знанию id группу не открыть.
func (s *Service) GroupStudents(ctx context.Context, p middleware.Principal, groupID int64, limit, offset int) ([]CourseStudent, int, error) {
	ok, err := s.repo.GroupVisible(ctx, groupID, p.UserID, p.IsAdmin())
	if err != nil {
		return nil, 0, err
	}
	if !ok {
		return nil, 0, apperr.NewNotFound("group_not_found")
	}
	return s.repo.GroupStudents(ctx, groupID, limit, offset)
}

// --- Свой профиль ---

// UpdateProfile обновляет аватар и возвращает профиль; ФИО меняет только администратор.
// Аватаром можно сделать только свой файл, иначе через этот метод
// можно было бы получить доступ к чужому файлу.
func (s *Service) UpdateProfile(ctx context.Context, id int64, in UpdateProfile) (User, error) {
	if in.LastName.Set || in.FirstName.Set || in.MiddleNames.Set {
		u, err := s.repo.Get(ctx, id)
		if err != nil {
			return User{}, err
		}
		if !u.IsAdmin {
			return User{}, apperr.NewForbidden("profile_name_admin_only", "only administrators can change names")
		}
	}
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
		if err := s.repo.CheckAvatar(ctx, id, *in.AvatarFileID.Value); err != nil {
			return User{}, err
		}
	}
	if err := s.repo.UpdateProfile(ctx, id, in); err != nil {
		return User{}, err
	}
	u, err := s.repo.Get(ctx, id)
	return u.User, err
}

// verifyPassword сверяет пароль пользователя; неверный — 403 wrong_password.
// 403, а не 401: сессия действительна, клиент не должен разлогинивать пользователя.
func (s *Service) verifyPassword(ctx context.Context, id int64, plain string) error {
	hash, err := s.repo.PasswordHash(ctx, id)
	if err != nil {
		return err
	}
	// CompareHashAndPassword сравнивает за постоянное время,
	// поэтому по времени ответа пароль не подобрать.
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) != nil {
		return apperr.NewForbidden("wrong_password", "current password is incorrect")
	}
	return nil
}

// ChangePassword меняет пароль после проверки текущего (PasswordInput).
// Новый пароль должен отличаться от старого. После смены отзываются все сессии,
// включая текущую: клиент очищает токен и открывает вход (документация /me/password).
func (s *Service) ChangePassword(ctx context.Context, id int64, oldPass, newPass string) error {
	if oldPass == newPass {
		return apperr.NewInvalid("password_unchanged", "new password must differ from the current one")
	}
	if err := s.verifyPassword(ctx, id, oldPass); err != nil {
		return err
	}
	hash, err := hashPassword(newPass)
	if err != nil {
		return err
	}
	return s.repo.SetPassword(ctx, id, hash)
}
