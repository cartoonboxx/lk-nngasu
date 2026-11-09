package auth

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// errUserNotFound — пользователь с таким email не найден.
// Репозиторий возвращает её вместо pgx.ErrNoRows, чтобы сервис не зависел от драйвера БД.
var errUserNotFound = errors.New("user not found")

// userRecord — данные пользователя, необходимые для входа (не полная модель пользователя).
type userRecord struct {
	ID           int64
	PasswordHash string // bcrypt-хеш пароля
	IsActive     bool
	IsAdmin      bool
	IsTeacher    bool // вычисляется: есть ли запись в таблице teacher
}

// Repository — доступ к данным пользователей, нужный сервису авторизации.
// Сервис зависит от интерфейса, поэтому в тестах его можно подменить фейком.
type Repository interface {
	FindByEmail(ctx context.Context, email string) (userRecord, error)
}

// pgRepository — реализация Repository поверх Postgres.
type pgRepository struct{ db *pgxpool.Pool }

// newPgRepository создаёт репозиторий поверх общего пула подключений (пул создаётся в main).
func newPgRepository(db *pgxpool.Pool) *pgRepository { return &pgRepository{db: db} }

// FindByEmail ищет пользователя по email.
// Сравнение регистрозависимое: email должен приходить уже нормализованным
// (нижний регистр, без пробелов) — это делает сервис.
// Если пользователя нет, возвращает errUserNotFound.
func (r *pgRepository) FindByEmail(ctx context.Context, email string) (userRecord, error) {
	// is_teacher не хранится в app_user — определяем по наличию записи в teacher.
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
	if err != nil {
		return userRecord{}, err
	}
	return u, nil
}
