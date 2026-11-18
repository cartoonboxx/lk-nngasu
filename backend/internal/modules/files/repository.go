package files

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"university-api/internal/platform/apperr"
	"university-api/internal/platform/database"
	"university-api/internal/platform/filegc"
)

type Repository struct{ db *pgxpool.Pool }

func newRepository(db *pgxpool.Pool) *Repository { return &Repository{db: db} }

// Insert создаёт запись о файле вместе с MIME-типом и размером (для FileInfo в ответах).
func (r *Repository) Insert(ctx context.Context, key, mime string, size int64) (int64, error) {
	var id int64
	err := r.db.QueryRow(ctx, `INSERT INTO file (storage_key, mime_type, size_bytes) VALUES ($1, $2, $3) RETURNING id`,
		key, mime, size).Scan(&id)
	return id, database.Translate(err)
}

// stored — запись о файле: ключ хранения и сохранённые при загрузке тип и размер
// (у файлов до миграции 010 размер может быть не записан — nil).
type stored struct {
	key  string
	mime *string
	size *int64
}

func (r *Repository) Get(ctx context.Context, id int64) (stored, error) {
	var f stored
	err := r.db.QueryRow(ctx, `SELECT storage_key, mime_type, size_bytes FROM file WHERE id = $1`, id).
		Scan(&f.key, &f.mime, &f.size)
	if errors.Is(err, pgx.ErrNoRows) {
		return stored{}, apperr.NewNotFound("file_not_found")
	}
	return f, database.Translate(err)
}

// Delete удаляет запись о файле, только если на него никто не ссылается
// (материал, задание, ответ, сообщение, аватар); иначе 409 file_in_use.
// Условие и удаление — одним запросом, без гонки с параллельной привязкой.
func (r *Repository) Delete(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM file f WHERE f.id = $1 AND `+filegc.Unreferenced, id)
	if err != nil {
		if database.PgCode(err) == database.CodeFK { // привязали между проверкой и удалением
			return errInUse
		}
		return database.Translate(err)
	}
	if tag.RowsAffected() == 0 {
		return errInUse // файл есть (владельца проверил handler), значит, он привязан
	}
	return nil
}

var errInUse = apperr.NewConflict("file_in_use", "the file is attached and cannot be deleted")

// courseMember — SQL-условие «пользователь $3 ведёт курс или учится в его группе».
func courseMember(courseExpr string) string {
	return `(EXISTS (SELECT 1 FROM workload w WHERE w.course_id = ` + courseExpr + ` AND w.teacher_id = $3)
		OR EXISTS (SELECT 1 FROM course_offering c JOIN student st ON st.group_id = c.group_id
		           WHERE c.id = ` + courseExpr + ` AND st.user_id = $3))`
}

// CanRead — может ли пользователь скачать файл: админ, владелец (его папка), любой — аватар,
// участники курса — материалы и вложения заданий, автор и преподаватели курса — ответы студентов,
// участники чата — вложения его сообщений.
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
					OR (sb.status <> 'not_submitted' -- черновик видит только автор
					    AND EXISTS (SELECT 1 FROM workload w WHERE w.course_id = a.course_id AND w.teacher_id = $3))))
			OR EXISTS (
				SELECT 1 FROM chat_message_file mf
				JOIN chat_message cm ON cm.id = mf.message_id
				JOIN chat ch ON ch.id = cm.chat_id
				WHERE mf.file_id = f.id AND cm.deleted_at IS NULL AND (
					ch.user_low = $3 OR ch.user_high = $3
					OR EXISTS (SELECT 1 FROM student st WHERE st.user_id = $3 AND st.group_id = ch.group_id)
					OR EXISTS (SELECT 1 FROM workload w JOIN course_offering c ON c.id = w.course_id
					           WHERE w.teacher_id = $3 AND c.group_id = ch.group_id)))
		)
		FROM file f WHERE f.id = $1`
	var ok bool
	err := r.db.QueryRow(ctx, q, fileID, isAdmin, userID, ownerPrefix).Scan(&ok)
	return ok, database.Translate(err)
}
