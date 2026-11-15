package users

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"university-api/internal/platform/apperr"
	"university-api/internal/platform/database"
)

var (
	errRequestNotFound = apperr.NewNotFound("registration_request_not_found")
	errRequestPending  = apperr.NewConflict("registration_exists", "a registration request with this email is already waiting for review")
	errBadReference    = apperr.NewInvalid("invalid_reference", "group or department does not exist")
)

// errAlreadyReviewed — 409 при повторном решении; в details — актуальное состояние заявки.
func errAlreadyReviewed(id int64, status string) error {
	return apperr.NewConflict("request_already_reviewed", "registration request has already been reviewed").
		WithDetails(map[string]any{"request_id": id, "status": status})
}

// requestSelect — заявка с администратором, принявшим решение (PersonRef).
// Порядок колонок — как в scanRequest.
const requestSelect = `
	SELECT r.id, r.email, r.passport, r.snils, r.last_name, r.first_name, r.middle_names,
	       r.comment, r.status::text, r.created_at, r.reviewed_at,
	       rv.id, concat_ws(' ', rv.last_name, rv.first_name, rv.middle_names),
	       r.user_id
	FROM registration_request r
	LEFT JOIN app_user rv ON rv.id = r.reviewed_by`

func scanRequest(row pgx.Row) (RegistrationRequest, error) {
	var (
		r          RegistrationRequest
		reviewerID *int64
		reviewer   *string
	)
	err := row.Scan(&r.ID, &r.Email, &r.Passport, &r.Snils, &r.LastName, &r.FirstName, &r.MiddleNames,
		&r.Comment, &r.Status, &r.CreatedAt, &r.ReviewedAt, &reviewerID, &reviewer, &r.UserID)
	if reviewerID != nil {
		r.ReviewedBy = &PersonRef{ID: *reviewerID, Name: deref(reviewer)}
	}
	return r, err
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// CreateRequest сохраняет заявку и возвращает квитанцию.
// Email уже занятый учёткой — 409 email_exists; уже есть необработанная заявка
// с этим email — 409 registration_exists. После отклонения подать заявку можно снова.
func (r *Repository) CreateRequest(ctx context.Context, in SubmitRequest, passwordHash string) (RegistrationReceipt, error) {
	var taken bool
	if err := r.db.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM app_user WHERE email = $1)`, in.Email).Scan(&taken); err != nil {
		return RegistrationReceipt{}, database.Translate(err)
	}
	if taken {
		return RegistrationReceipt{}, apperr.NewConflict("email_exists", "email is already taken")
	}

	var rc RegistrationReceipt
	err := r.db.QueryRow(ctx, `
		INSERT INTO registration_request (email, password_hash, passport, snils, comment, terms_version, accepted_terms_at)
		VALUES ($1, $2, $3, $4, $5, $6, now())
		RETURNING id, status::text, created_at`,
		in.Email, passwordHash, in.Passport, in.Snils, in.Comment, in.TermsVersion).
		Scan(&rc.ID, &rc.Status, &rc.CreatedAt)
	if database.PgCode(err) == database.CodeUnique {
		return RegistrationReceipt{}, errRequestPending // частичный уникальный индекс по email среди pending
	}
	return rc, database.Translate(err)
}

// GetRequest — заявка по id.
func (r *Repository) GetRequest(ctx context.Context, id int64) (RegistrationRequest, error) {
	req, err := scanRequest(r.db.QueryRow(ctx, requestSelect+` WHERE r.id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return RegistrationRequest{}, errRequestNotFound
	}
	return req, database.Translate(err)
}

// ListRequests — страница заявок и общее число под фильтрами (total учитывает q и status).
func (r *Repository) ListRequests(ctx context.Context, f RequestFilter) ([]RegistrationRequest, int, error) {
	var w database.Filter
	if f.Status != "" {
		w.Add("r.status = ?::registration_status", f.Status)
	}
	if f.Q != "" {
		w.Add(`(r.email ILIKE ? OR concat_ws(' ', r.last_name, r.first_name, r.middle_names) ILIKE ?)`, "%"+f.Q+"%")
	}

	var total int
	countArgs := append([]any(nil), w.Args()...)
	if err := r.db.QueryRow(ctx, `SELECT count(*) FROM registration_request r`+w.Where(), countArgs...).Scan(&total); err != nil {
		return nil, 0, database.Translate(err)
	}

	order := "ASC"
	if f.Newest {
		order = "DESC"
	}
	q := requestSelect + w.Where() +
		fmt.Sprintf(" ORDER BY r.created_at %s, r.id %s LIMIT %s OFFSET %s", order, order, w.Arg(f.Limit), w.Arg(f.Offset))
	list, err := database.Collect(ctx, r.db, scanRequest, q, w.Args()...)
	return list, total, err
}

// lockPending блокирует заявку до конца транзакции и проверяет, что она ещё не рассмотрена.
// FOR UPDATE не даёт двум администраторам одновременно принять решение по одной заявке.
func lockPending(ctx context.Context, tx pgx.Tx, id int64) (email, passwordHash string, err error) {
	var status string
	err = tx.QueryRow(ctx,
		`SELECT status::text, email, password_hash FROM registration_request WHERE id = $1 FOR UPDATE`, id).
		Scan(&status, &email, &passwordHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", "", errRequestNotFound
	}
	if err != nil {
		return "", "", database.Translate(err)
	}
	if status != "pending" {
		return "", "", errAlreadyReviewed(id, status)
	}
	return email, passwordHash, nil
}

// ApproveRequest в одной транзакции создаёт пользователя с email и хешем пароля из заявки,
// записывает в заявку ФИО и уточнённые документы и закрывает её.
func (r *Repository) ApproveRequest(ctx context.Context, id, adminID int64, u ApprovalUser) error {
	return database.InTx(ctx, r.db, func(tx pgx.Tx) error {
		email, hash, err := lockPending(ctx, tx, id)
		if err != nil {
			return err
		}
		// Пользователь получает email и хеш пароля из заявки, документы — уточнённые администратором.
		in := CreateUser{LastName: u.LastName, FirstName: u.FirstName, MiddleNames: u.MiddleNames,
			Email: email, Passport: u.Passport, Snils: u.Snils}
		in.Student, in.Teacher = u.Student, u.Teacher
		userID, err := createUserTx(ctx, tx, in, hash)
		if err != nil {
			return err // invalid_reference, email_exists, record_book_exists
		}

		_, err = tx.Exec(ctx, `
			UPDATE registration_request
			SET status = 'approved', reviewed_by = $2, reviewed_at = now(), user_id = $3,
			    passport = $4, snils = $5, last_name = $6, first_name = $7, middle_names = $8
			WHERE id = $1`, id, adminID, userID, u.Passport, u.Snils, u.LastName, u.FirstName, u.MiddleNames)
		return database.Translate(err)
	})
}

// RejectRequest отклоняет необработанную заявку. Причина в текущем интерфейсе не запрашивается.
func (r *Repository) RejectRequest(ctx context.Context, id, adminID int64) error {
	return database.InTx(ctx, r.db, func(tx pgx.Tx) error {
		if _, _, err := lockPending(ctx, tx, id); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, `
			UPDATE registration_request SET status = 'rejected', reviewed_by = $2, reviewed_at = now()
			WHERE id = $1`, id, adminID)
		return database.Translate(err)
	})
}
