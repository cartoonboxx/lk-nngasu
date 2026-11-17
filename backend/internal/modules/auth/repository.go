package auth

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"university-api/internal/platform/database"
)

// errNotFound — запись не найдена; сервис превращает её в ошибку API по контексту.
var errNotFound = errors.New("not found")

// repository — доступ к пользователям, заявкам, сессиям и приглашениям.
type repository struct{ db *pgxpool.Pool }

func newRepository(db *pgxpool.Pool) *repository { return &repository{db: db} }

// userRecord — пользователь для входа и профиля в ответе.
type userRecord struct {
	UserInfo
	PasswordHash string // пустой — пароль ещё не задан (ждёт активации по приглашению)
	IsActive     bool   // дублирует UserInfo.IsActive для проверок в сервисе
}

// userSelect — пользователь с группой и кафедрой; порядок колонок — как в scanUser.
const userSelect = `
	SELECT u.id, u.email, u.password_hash, u.is_active, u.is_admin,
	       u.last_name, u.first_name, u.middle_names, u.avatar_file_id,
	       s.group_id, g.name, s.record_book_number, s.is_headman,
	       t.department_id, d.name
	FROM app_user u
	LEFT JOIN student     s ON s.user_id = u.id
	LEFT JOIN study_group g ON g.id = s.group_id
	LEFT JOIN teacher     t ON t.user_id = u.id
	LEFT JOIN department  d ON d.id = t.department_id`

func scanUser(row pgx.Row) (userRecord, error) {
	var (
		u              userRecord
		avatarID       *int64
		groupID        *int64
		groupName      *string
		recordBook     *string
		isHeadman      *bool
		departmentID   *int64
		departmentName *string
	)
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.IsActive, &u.IsAdmin,
		&u.LastName, &u.FirstName, &u.MiddleNames, &avatarID,
		&groupID, &groupName, &recordBook, &isHeadman, &departmentID, &departmentName)
	if errors.Is(err, pgx.ErrNoRows) {
		return userRecord{}, errNotFound
	}
	if err != nil {
		return userRecord{}, err
	}
	u.UserInfo.IsActive = u.IsActive
	u.FullName = strings.Join(nonEmpty(u.LastName, u.FirstName, deref(u.MiddleNames)), " ")
	if avatarID != nil {
		u.Avatar = &FileRef{ID: *avatarID, URL: fileURLBase + strconv.FormatInt(*avatarID, 10)}
	}
	if groupID != nil {
		u.Student = &StudentInfo{GroupID: *groupID, GroupName: deref(groupName), RecordBookNumber: recordBook,
			IsHeadman: isHeadman != nil && *isHeadman}
	}
	if departmentID != nil {
		u.Teacher = &TeacherInfo{DepartmentID: *departmentID, DepartmentName: deref(departmentName)}
	}
	return u, nil
}

// fileURLBase — адрес скачивания файла через модуль files (GET /api/v1/files/:id).
const fileURLBase = "/api/v1/files/"

// nonEmpty — непустые части ФИО, чтобы без отчества не было двойного пробела.
func nonEmpty(parts ...string) []string {
	out := parts[:0]
	for _, p := range parts {
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// UserByEmail — пользователь по email (email уже нормализован сервисом).
func (r *repository) UserByEmail(ctx context.Context, email string) (userRecord, error) {
	return scanUser(r.db.QueryRow(ctx, userSelect+` WHERE u.email = $1`, email))
}

// userByID — пользователь по id внутри транзакции (refresh перечитывает роли из БД).
func userByID(ctx context.Context, q pgx.Tx, id int64) (userRecord, error) {
	return scanUser(q.QueryRow(ctx, userSelect+` WHERE u.id = $1`, id))
}

// LatestRequest — последняя заявка на регистрацию с этим email: статус и хеш пароля.
// Нужна, чтобы на вход до одобрения ответить account_pending / account_rejected.
func (r *repository) LatestRequest(ctx context.Context, email string) (status, passwordHash string, err error) {
	err = r.db.QueryRow(ctx, `
		SELECT status::text, password_hash FROM registration_request
		WHERE email = $1 ORDER BY created_at DESC, id DESC LIMIT 1`, email).Scan(&status, &passwordHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", "", errNotFound
	}
	return status, passwordHash, err
}

// CreateSession создаёт сессию и её первый refresh-токен. Возвращает id сессии.
func (r *repository) CreateSession(ctx context.Context, userID int64, remember bool, expiresAt time.Time, tokenHash []byte) (string, error) {
	var sid string
	err := database.InTx(ctx, r.db, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, `
			INSERT INTO auth_session (user_id, remember_me, expires_at) VALUES ($1, $2, $3)
			RETURNING id::text`, userID, remember, expiresAt).Scan(&sid); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, `INSERT INTO auth_refresh_token (token_hash, session_id) VALUES ($1, $2)`, tokenHash, sid)
		return err
	})
	return sid, err
}

// refreshRow — refresh-токен и его сессия.
type refreshRow struct {
	SessionID string
	UserID    int64
	ExpiresAt time.Time
	Revoked   bool
	Used      bool // токен уже обменян на новый
}

// active — сессия ещё действует (не отозвана и не истекла).
func (r refreshRow) active() bool { return !r.Revoked && time.Now().Before(r.ExpiresAt) }

const refreshSelect = `
	SELECT s.id::text, s.user_id, s.expires_at, s.revoked_at IS NOT NULL, rt.used_at IS NOT NULL
	FROM auth_refresh_token rt
	JOIN auth_session s ON s.id = rt.session_id
	WHERE rt.token_hash = $1`

func scanRefresh(row pgx.Row) (refreshRow, error) {
	var r refreshRow
	err := row.Scan(&r.SessionID, &r.UserID, &r.ExpiresAt, &r.Revoked, &r.Used)
	if errors.Is(err, pgx.ErrNoRows) {
		return refreshRow{}, errNotFound
	}
	return r, err
}

// RefreshByHash — сессия по refresh-токену (для CSRF и logout, без блокировки).
func (r *repository) RefreshByHash(ctx context.Context, tokenHash []byte) (refreshRow, error) {
	return scanRefresh(r.db.QueryRow(ctx, refreshSelect, tokenHash))
}

// lockRefresh — то же внутри транзакции с блокировкой: два параллельных refresh
// с одним токеном выполнятся по очереди, второй увидит used = true.
func lockRefresh(ctx context.Context, tx pgx.Tx, tokenHash []byte) (refreshRow, error) {
	return scanRefresh(tx.QueryRow(ctx, refreshSelect+` FOR UPDATE OF rt, s`, tokenHash))
}

// rotate помечает старый refresh-токен использованным и сохраняет новый.
func rotate(ctx context.Context, tx pgx.Tx, oldHash, newHash []byte, sid string) error {
	if _, err := tx.Exec(ctx, `UPDATE auth_refresh_token SET used_at = now() WHERE token_hash = $1`, oldHash); err != nil {
		return err
	}
	_, err := tx.Exec(ctx, `INSERT INTO auth_refresh_token (token_hash, session_id) VALUES ($1, $2)`, newHash, sid)
	return err
}

// revokeSession отзывает сессию внутри транзакции; с ней перестают приниматься
// и её access-токены (middleware проверяет сессию на каждом запросе).
func revokeSession(ctx context.Context, tx pgx.Tx, sid, reason string) error {
	_, err := tx.Exec(ctx, `UPDATE auth_session SET revoked_at = now(), revoke_reason = $2
		WHERE id = $1::uuid AND revoked_at IS NULL`, sid, reason)
	return err
}

// RevokeSession — отзыв вне транзакции (logout).
func (r *repository) RevokeSession(ctx context.Context, sid, reason string) error {
	_, err := r.db.Exec(ctx, `UPDATE auth_session SET revoked_at = now(), revoke_reason = $2
		WHERE id = $1::uuid AND revoked_at IS NULL`, sid, reason)
	return err
}

// SessionActive — сессия не отозвана, не истекла, пользователь не заблокирован.
func (r *repository) SessionActive(ctx context.Context, sid string, userID int64) (bool, error) {
	var ok bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM auth_session s JOIN app_user u ON u.id = s.user_id
			WHERE s.id = $1::uuid AND s.user_id = $2
			  AND s.revoked_at IS NULL AND s.expires_at > now() AND u.is_active)`, sid, userID).Scan(&ok)
	return ok, err
}
