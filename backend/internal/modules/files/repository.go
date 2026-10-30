package files

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"university-api/internal/platform/apperr"
	"university-api/internal/platform/database"
)

type Repository struct{ db *pgxpool.Pool }

func newRepository(db *pgxpool.Pool) *Repository { return &Repository{db: db} }

func (r *Repository) Insert(ctx context.Context, key string) (int64, error) {
	var id int64
	err := r.db.QueryRow(ctx, `INSERT INTO file (storage_key) VALUES ($1) RETURNING id`, key).Scan(&id)
	return id, database.Translate(err)
}

func (r *Repository) Key(ctx context.Context, id int64) (string, error) {
	var key string
	err := r.db.QueryRow(ctx, `SELECT storage_key FROM file WHERE id = $1`, id).Scan(&key)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", apperr.NewNotFound("file_not_found")
	}
	return key, database.Translate(err)
}

// Delete: если на файл кто-то ссылается (материал, задание, ответ) — 409 in_use.
func (r *Repository) Delete(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM file WHERE id = $1`, id)
	return database.AffectedDelete(tag, err, "file_not_found")
}

// courseMember — SQL-условие «пользователь $3 ведёт курс или учится в его группе».
func courseMember(courseExpr string) string {
	return `(EXISTS (SELECT 1 FROM workload w WHERE w.course_id = ` + courseExpr + ` AND w.teacher_id = $3)
		OR EXISTS (SELECT 1 FROM course_offering c JOIN student st ON st.group_id = c.group_id
		           WHERE c.id = ` + courseExpr + ` AND st.user_id = $3))`
}

// CanRead — может ли пользователь скачать файл: админ, владелец (его папка), любой — аватар,
// участники курса — материалы и вложения заданий, автор и преподаватели курса — ответы студентов.
// Для несуществующего файла вернёт NotFound.
func (r *Repository) CanRead(ctx context.Context, fileID, userID int64, isAdmin bool, ownerPrefix string) (bool, error) {
	q := `
		SELECT (
			$2::bool
			OR starts_with(f.storage_key, $4::text)
			OR EXISTS (SELECT 1 FROM app_user u WHERE u.avatar_file_id = f.id)
			OR EXISTS (
				SELECT 1 FROM material m JOIN topic t ON t.id = m.topic_id
				WHERE m.file_id = f.id AND ` + courseMember("t.course_id") + `)
			OR EXISTS (
				SELECT 1 FROM assignment_file af JOIN assignment a ON a.id = af.assignment_id
				WHERE af.file_id = f.id AND ` + courseMember("a.course_id") + `)
			OR EXISTS (
				SELECT 1 FROM submission_file sf
				JOIN submission sb ON sb.id = sf.submission_id
				JOIN assignment a  ON a.id = sb.assignment_id
				WHERE sf.file_id = f.id AND (
					sb.student_id = $3
					OR EXISTS (SELECT 1 FROM workload w WHERE w.course_id = a.course_id AND w.teacher_id = $3)))
		)
		FROM file f WHERE f.id = $1`
	var ok bool
	err := r.db.QueryRow(ctx, q, fileID, isAdmin, userID, ownerPrefix).Scan(&ok)
	return ok, database.Translate(err)
}
