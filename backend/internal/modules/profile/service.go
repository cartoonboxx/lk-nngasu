package profile

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"university-api/internal/platform/apperr"
	"university-api/internal/platform/middleware"
)

// errNoUser — токен валиден, но пользователя уже нет или он деактивирован.
// Отдаётся как 401: получив его, фронт разлогинивает пользователя.
var errNoUser = apperr.NewUnauthorized("user_not_found", "user does not exist or is inactive")

// Service — бизнес-логика профиля: читает строку из БД и собирает из неё
// вложенную структуру Profile в формате, который ждёт фронт.
type Service struct {
	repo       *Repository
	avatarBase string // префикс ссылки на файл, например "/api/v1/files/"
}

// NewService создаёт сервис профиля.
func NewService(repo *Repository, avatarBase string) *Service {
	return &Service{repo: repo, avatarBase: avatarBase}
}

// Me возвращает профиль пользователя с данным id (id берётся из JWT).
// Пользователя нет или он неактивен — errNoUser (401), сбой БД — внутренняя ошибка (500).
func (s *Service) Me(ctx context.Context, userID int64) (Profile, error) {
	r, err := s.repo.Get(ctx, userID)

	// Пользователя удалили после выдачи токена — токен больше не годится.
	if errors.Is(err, errUserNotFound) {
		return Profile{}, errNoUser
	}
	// Любая другая ошибка — внутренняя; оборачиваем, чтобы в логах было видно источник.
	if err != nil {
		return Profile{}, fmt.Errorf("profile: get user: %w", err)
	}
	// Токен ещё действует, но пользователя деактивировали — не пускаем.
	if !r.IsActive {
		return Profile{}, errNoUser
	}

	// Общие поля — есть у любого пользователя.
	p := Profile{
		ID:         r.ID,
		Role:       roleOf(r),
		FirstName:  r.FirstName,
		LastName:   r.LastName,
		MiddleName: r.MiddleName,
		Email:      r.Email,
		IsActive:   r.IsActive,
	}

	// Аватар хранится в модуле files — отдаём ссылку на его скачивание.
	// Права на скачивание проверит сам модуль files.
	if r.AvatarFileID != nil {
		url := s.avatarBase + strconv.FormatInt(*r.AvatarFileID, 10)
		p.AvatarURL = &url
	}

	// Блок студента: группа и направление вкладываются, только если они есть,
	// иначе соответствующий объект остаётся null.
	if r.IsStudent {
		st := &Student{RecordBookNumber: deref(r.RecordBookNumber), IsHeadman: deref(r.IsHeadman)}
		if r.GroupID != nil {
			st.Group = &Group{ID: *r.GroupID, Name: deref(r.GroupName), AdmissionYear: deref(r.AdmissionYear)}
			if r.SpecID != nil {
				st.Group.Specialization = &Specialization{ID: *r.SpecID, Code: deref(r.SpecCode), Name: deref(r.SpecName)}
			}
		}
		p.Student = st
	}

	// Блок преподавателя: кафедра вкладывается, только если указана.
	if r.IsTeacher {
		t := &Teacher{}
		if r.DeptID != nil {
			t.Department = &Department{ID: *r.DeptID, Name: deref(r.DeptName)}
		}
		p.Teacher = t
	}

	return p, nil
}

// roleOf определяет роль так же, как при логине: admin > teacher > student.
// Должна совпадать с логикой auth.Service.Login, иначе роль в профиле и в токене разойдётся.
func roleOf(r profileRow) string {
	switch {
	case r.IsAdmin:
		return middleware.RoleAdmin
	case r.IsTeacher:
		return middleware.RoleTeacher
	default:
		return middleware.RoleStudent
	}
}

// deref возвращает значение по указателю или нулевое значение, если указатель nil.
// Нужна для колонок из LEFT JOIN, которые в БД могут быть NULL.
func deref[T any](p *T) T {
	var zero T
	if p == nil {
		return zero
	}
	return *p
}
