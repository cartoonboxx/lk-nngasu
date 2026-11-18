package auth

import (
	"context"
	"crypto/subtle"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"

	"university-api/internal/platform/apperr"
	"university-api/internal/platform/database"
)

var (
	errSetupClosed   = apperr.NewConflict("setup_closed", "initial setup is disabled or completed")
	errSetupKey      = apperr.NewForbidden("setup_key_invalid", "invalid setup key")
	errOwnerRequired = apperr.NewForbidden("owner_required", "only the system owner can manage administrators")
)

type AdminInput struct {
	Email     string `json:"email" binding:"required,email,max=255"`
	FirstName string `json:"first_name" binding:"required,max=50"`
	LastName  string `json:"last_name" binding:"required,max=50"`
}

type Administrator struct {
	ID                  int64      `json:"id"`
	Email               string     `json:"email"`
	FullName            string     `json:"full_name"`
	FirstName           string     `json:"first_name"`
	LastName            string     `json:"last_name"`
	CanManageAdmins     bool       `json:"can_manage_admins"`
	IsActive            bool       `json:"is_active"`
	CredentialState     string     `json:"credential_state"`
	InvitationExpiresAt *time.Time `json:"invitation_expires_at"`
}

func (s *Service) SetupAvailable(ctx context.Context) (bool, error) {
	if s.opts.AdminSetupKey == "" {
		return false, nil
	}
	var available bool
	err := s.repo.db.QueryRow(ctx, `SELECT completed_at IS NULL
		AND NOT EXISTS (SELECT 1 FROM app_user WHERE is_admin) FROM admin_setup WHERE id`).Scan(&available)
	return available, err
}

func (s *Service) Setup(ctx context.Context, key, password string, in AdminInput) error {
	if s.opts.AdminSetupKey == "" {
		return errSetupClosed
	}
	if subtle.ConstantTimeCompare(hashToken(key), hashToken(s.opts.AdminSetupKey)) != 1 {
		return errSetupKey
	}
	if len(password) > 72 {
		return errPasswordTooLongAct
	}
	in.FirstName, in.LastName = strings.TrimSpace(in.FirstName), strings.TrimSpace(in.LastName)
	if in.FirstName == "" || in.LastName == "" {
		return apperr.NewInvalid("invalid_request", "first and last name must not be blank")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return database.InTx(ctx, s.repo.db, func(tx pgx.Tx) error {
		var completed *time.Time
		if err := tx.QueryRow(ctx, `SELECT completed_at FROM admin_setup WHERE id FOR UPDATE`).Scan(&completed); err != nil {
			return err
		}
		if completed != nil {
			return errSetupClosed
		}
		var id int64
		err := tx.QueryRow(ctx, `INSERT INTO app_user
			(first_name, last_name, email, password_hash, is_admin, can_manage_admins)
			VALUES ($1, $2, $3, $4, true, true) RETURNING id`,
			in.FirstName, in.LastName, strings.ToLower(strings.TrimSpace(in.Email)), string(hash)).Scan(&id)
		if database.PgCode(err) == database.CodeUnique {
			return apperr.NewConflict("email_exists", "email is already taken")
		}
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `INSERT INTO admin_access_audit (target_id, target_email, action)
			SELECT id, email, 'setup' FROM app_user WHERE id = $1`, id)
		return err
	})
}

// Повторно проверяем право владельца в БД, независимо от старых claims JWT.
// Блокировка удерживается до завершения изменения прав и записи аудита.
func verifyOwner(ctx context.Context, tx pgx.Tx, actorID int64, password string) error {
	var hash string
	err := tx.QueryRow(ctx, `SELECT password_hash FROM app_user
		WHERE id = $1 AND is_active AND is_admin AND can_manage_admins FOR UPDATE`, actorID).Scan(&hash)
	if errors.Is(err, pgx.ErrNoRows) {
		return errOwnerRequired
	}
	if err != nil {
		return err
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		return apperr.NewForbidden("wrong_password", "current password is incorrect")
	}
	return nil
}

func (s *Service) Administrators(ctx context.Context, actorID int64) ([]Administrator, error) {
	var owner bool
	if err := s.repo.db.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM app_user
		WHERE id = $1 AND is_active AND is_admin AND can_manage_admins)`, actorID).Scan(&owner); err != nil {
		return nil, err
	}
	if !owner {
		return nil, errOwnerRequired
	}
	rows, err := s.repo.db.Query(ctx, `SELECT u.id, u.email, u.first_name, u.last_name, concat_ws(' ', u.last_name, u.first_name, u.middle_names),
		u.can_manage_admins, u.is_active, CASE WHEN u.password_hash = '' THEN 'invited' ELSE 'active' END,
		(SELECT max(expires_at) FROM user_invitation WHERE user_id = u.id AND grants_admin AND used_at IS NULL AND revoked_at IS NULL)
		FROM app_user u WHERE u.is_admin OR EXISTS
		(SELECT 1 FROM user_invitation WHERE user_id = u.id AND grants_admin AND used_at IS NULL AND revoked_at IS NULL)
		ORDER BY u.can_manage_admins DESC, u.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []Administrator{}
	for rows.Next() {
		var item Administrator
		if err := rows.Scan(&item.ID, &item.Email, &item.FirstName, &item.LastName, &item.FullName, &item.CanManageAdmins, &item.IsActive,
			&item.CredentialState, &item.InvitationExpiresAt); err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	return list, rows.Err()
}

// Повтор для той же почты переотправляет приглашение, отменяя предыдущую ссылку.
func (s *Service) InviteAdministrator(ctx context.Context, actorID int64, password string, in AdminInput) error {
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))
	in.FirstName, in.LastName = strings.TrimSpace(in.FirstName), strings.TrimSpace(in.LastName)
	if in.FirstName == "" || in.LastName == "" {
		return apperr.NewInvalid("invalid_request", "first and last name must not be blank")
	}
	return database.InTx(ctx, s.repo.db, func(tx pgx.Tx) error {
		if err := verifyOwner(ctx, tx, actorID, password); err != nil {
			return err
		}
		if !s.opts.AdminInvitationsEnabled {
			return apperr.New(apperr.Unavailable, "invitations_disabled", "SMTP must be configured for administrator invitations")
		}
		var id int64
		err := tx.QueryRow(ctx, `INSERT INTO app_user (first_name, last_name, email, password_hash)
			VALUES ($1, $2, $3, '') ON CONFLICT (email) DO NOTHING RETURNING id`, in.FirstName, in.LastName, in.Email).Scan(&id)
		if errors.Is(err, pgx.ErrNoRows) {
			var hash string
			var active bool
			if err := tx.QueryRow(ctx, `SELECT id, password_hash, is_active FROM app_user WHERE email = $1 FOR UPDATE`, in.Email).
				Scan(&id, &hash, &active); err != nil {
				return err
			}
			if hash != "" || !active {
				return apperr.NewConflict("email_exists", "use promotion for an existing active account")
			}
		} else if err != nil {
			return err
		}
		if err := s.InviteTx(ctx, tx, id, actorID, true); err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `INSERT INTO admin_access_audit (actor_id, target_id, target_email, action)
			VALUES ($1, $2, $3, 'invite')`, actorID, id, in.Email)
		return err
	})
}

func (s *Service) PromoteAdministrator(ctx context.Context, actorID int64, password, email string) error {
	return database.InTx(ctx, s.repo.db, func(tx pgx.Tx) error {
		if err := verifyOwner(ctx, tx, actorID, password); err != nil {
			return err
		}
		var id int64
		var isAdmin, active bool
		var hash string
		err := tx.QueryRow(ctx, `SELECT id, is_admin, is_active, password_hash FROM app_user WHERE email = $1 FOR UPDATE`,
			strings.ToLower(strings.TrimSpace(email))).Scan(&id, &isAdmin, &active, &hash)
		if errors.Is(err, pgx.ErrNoRows) {
			return errUserNotFound
		}
		if err != nil {
			return err
		}
		if isAdmin {
			return apperr.NewConflict("already_admin", "the user is already an administrator")
		}
		if !active || hash == "" {
			return apperr.NewConflict("account_not_ready", "an active account with a password is required")
		}
		if _, err := tx.Exec(ctx, `UPDATE app_user SET is_admin = true WHERE id = $1`, id); err != nil {
			return err
		}
		// После повторного входа пользователь получит JWT с актуальной ролью.
		if _, err := tx.Exec(ctx, `UPDATE auth_session SET revoked_at = now(), revoke_reason = 'admin_granted'
			WHERE user_id = $1 AND revoked_at IS NULL`, id); err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `INSERT INTO admin_access_audit (actor_id, target_id, target_email, action)
			SELECT $1, id, email, 'promote' FROM app_user WHERE id = $2`, actorID, id)
		return err
	})
}
