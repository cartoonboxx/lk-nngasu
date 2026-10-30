package auth

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var errUserNotFound = errors.New("user not found")

type userRecord struct {
	ID           int64
	PasswordHash string
	IsActive     bool
	IsAdmin      bool
	IsTeacher    bool
}

type Repository interface {
	FindByEmail(ctx context.Context, email string) (userRecord, error)
}

type pgRepository struct{ db *pgxpool.Pool }

func newPgRepository(db *pgxpool.Pool) *pgRepository { return &pgRepository{db: db} }

func (r *pgRepository) FindByEmail(ctx context.Context, email string) (userRecord, error) {
	const q = `
		SELECT u.id, u.password_hash, u.is_active, u.is_admin,
		       EXISTS (SELECT 1 FROM teacher t WHERE t.user_id = u.id)
		FROM app_user u
		WHERE u.email = $1`

	var u userRecord
	err := r.db.QueryRow(ctx, q, email).Scan(&u.ID, &u.PasswordHash, &u.IsActive, &u.IsAdmin, &u.IsTeacher)
	if errors.Is(err, pgx.ErrNoRows) {
		return userRecord{}, errUserNotFound
	}
	return u, err
}
