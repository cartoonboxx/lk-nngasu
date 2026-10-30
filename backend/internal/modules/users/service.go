package users

import (
	"context"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"university-api/internal/platform/apperr"
	"university-api/internal/platform/middleware"
)

// FileOwnership проверяет, что файл лежит в папке пользователя (реализует access.Checker).
type FileOwnership interface {
	OwnsFile(ctx context.Context, userID, fileID int64) error
}

type Service struct {
	repo  *Repository
	files FileOwnership
}

func NewService(repo *Repository, files FileOwnership) *Service {
	return &Service{repo: repo, files: files}
}

func normEmail(s string) string { return strings.ToLower(strings.TrimSpace(s)) }

func hashPassword(plain string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	return string(h), err
}

func (s *Service) Get(ctx context.Context, id int64) (User, error) { return s.repo.Get(ctx, id) }

func (s *Service) List(ctx context.Context, f ListFilter) ([]User, error) { return s.repo.List(ctx, f) }

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
	default:
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
	return s.repo.Get(ctx, id)
}

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
func (s *Service) verifyPassword(ctx context.Context, id int64, plain string) error {
	hash, err := s.repo.PasswordHash(ctx, id)
	if err != nil {
		return err
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) != nil {
		return apperr.NewUnauthorized("wrong_password", "current password is incorrect")
	}
	return nil
}

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
// в схеме нет поля email_verified_at.
func (s *Service) ChangeEmail(ctx context.Context, id int64, password, newEmail string) (User, error) {
	if err := s.verifyPassword(ctx, id, password); err != nil {
		return User{}, err
	}
	if err := s.repo.SetEmail(ctx, id, normEmail(newEmail)); err != nil {
		return User{}, err
	}
	return s.repo.Get(ctx, id)
}
