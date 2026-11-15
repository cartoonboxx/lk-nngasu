package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"

	"university-api/internal/platform/apperr"
	"university-api/internal/platform/database"
)

// invitationTTL — срок ссылки из письма.
const invitationTTL = 24 * time.Hour

var (
	errInvitationInvalid  = apperr.NewInvalid("invitation_invalid", "invitation link is invalid")
	errInvitationExpired  = apperr.NewGone("invitation_expired", "invitation link has expired, ask the administrator for a new one")
	errAlreadyActivated   = apperr.NewConflict("already_activated", "account is already activated")
	errUserNotFound       = apperr.NewNotFound("user_not_found")
	errInvitationNotSent  = apperr.New(apperr.Unavailable, "invitation_not_sent", "invitation was created but the email could not be sent; resend it with POST /users/{id}/invitation")
	errPasswordTooLongAct = apperr.NewInvalid("invalid_request", "password: must be at most 72 bytes in UTF-8").
				WithDetails(map[string]any{"fields": map[string][]string{"password": {"must be at most 72 bytes in UTF-8"}}})
)

// Invite создаёт приглашение задать пароль и отправляет письмо со ссылкой
// <APP_URL>/activation?token=…  Прежние неиспользованные приглашения отменяются.
// Пользователь, у которого пароль уже задан, — 409 already_activated.
//
// Реализует users.Inviter: вызывается при создании пользователя без пароля
// и из POST /users/:id/invitation.
func (s *Service) Invite(ctx context.Context, userID, invitedBy int64) error {
	token := newToken()
	var email, firstName string

	err := database.InTx(ctx, s.repo.db, func(tx pgx.Tx) error {
		var hash string
		err := tx.QueryRow(ctx, `SELECT email, first_name, password_hash FROM app_user WHERE id = $1 FOR UPDATE`, userID).
			Scan(&email, &firstName, &hash)
		if errors.Is(err, pgx.ErrNoRows) {
			return errUserNotFound
		}
		if err != nil {
			return err
		}
		if hash != "" {
			return errAlreadyActivated
		}
		if _, err := tx.Exec(ctx, `UPDATE user_invitation SET revoked_at = now()
			WHERE user_id = $1 AND used_at IS NULL AND revoked_at IS NULL`, userID); err != nil {
			return err
		}
		var by *int64
		if invitedBy > 0 {
			by = &invitedBy
		}
		_, err = tx.Exec(ctx, `INSERT INTO user_invitation (user_id, token_hash, created_by, expires_at)
			VALUES ($1, $2, $3, $4)`, userID, hashToken(token), by, time.Now().Add(invitationTTL))
		return err
	})
	if err != nil {
		return err
	}

	link := s.opts.AppURL + "/activation?token=" + url.QueryEscape(token)
	body := fmt.Sprintf("Здравствуйте, %s!\n\n"+
		"Для вас создана учётная запись в личном кабинете. Чтобы задать пароль, перейдите по ссылке "+
		"(она действует 24 часа):\n\n%s\n\n"+
		"Если вы не ожидали это письмо, просто проигнорируйте его.\n", firstName, link)
	if err := s.opts.Mail.Send(email, "Приглашение в личный кабинет", body); err != nil {
		slog.Error("invitation email failed", "user_id", userID, "err", err)
		return errInvitationNotSent
	}
	return nil
}

// Activate задаёт пароль по токену из письма. После этого пользователь входит обычным способом.
func (s *Service) Activate(ctx context.Context, token, password string) error {
	if len(password) > 72 {
		return errPasswordTooLongAct
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("auth: hash password: %w", err)
	}

	return database.InTx(ctx, s.repo.db, func(tx pgx.Tx) error {
		var (
			id, userID int64
			expiresAt  time.Time
			used       bool
			revoked    bool
		)
		err := tx.QueryRow(ctx, `
			SELECT id, user_id, expires_at, used_at IS NOT NULL, revoked_at IS NOT NULL
			FROM user_invitation WHERE token_hash = $1 FOR UPDATE`, hashToken(token)).
			Scan(&id, &userID, &expiresAt, &used, &revoked)
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return errInvitationInvalid
		case err != nil:
			return err
		case used:
			return errAlreadyActivated
		case revoked:
			return errInvitationInvalid // заменено более новым приглашением
		case time.Now().After(expiresAt):
			return errInvitationExpired
		}

		// Пароль задаётся, только если его ещё нет: администратор мог успеть задать его сам.
		tag, err := tx.Exec(ctx, `UPDATE app_user SET password_hash = $2 WHERE id = $1 AND password_hash = ''`, userID, string(hash))
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return errAlreadyActivated
		}
		_, err = tx.Exec(ctx, `UPDATE user_invitation SET used_at = now() WHERE id = $1`, id)
		return err
	})
}
