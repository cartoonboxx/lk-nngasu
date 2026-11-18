package assignments

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"university-api/internal/platform/apperr"
	"university-api/internal/platform/database"
	"university-api/internal/platform/filegc"
)

type Repository struct{ db *pgxpool.Pool }

func newRepository(db *pgxpool.Pool) *Repository { return &Repository{db: db} }

// SQL-выражения для вычисляемых полей. Живут в одном месте, чтобы список для студента,
// список для преподавателя и карточка ответа считали статус одинаково.
const (
	// derivedStatus: sb — submission (может быть NULL после LEFT JOIN), a — assignment.
	derivedStatus = `(CASE
		WHEN sb.status = 'reviewed'  THEN 'reviewed'
		WHEN sb.status = 'submitted' THEN 'submitted'
		WHEN a.deadline_at IS NOT NULL AND a.deadline_at < now() THEN 'missing'
		ELSE 'not_submitted' END)`

	// isLate: ответ отправлен после дедлайна.
	isLate = `(sb.submitted_at IS NOT NULL AND a.deadline_at IS NOT NULL AND sb.submitted_at > a.deadline_at)`
)

// filesJSON — подзапрос «файлы владельца» в виде json-массива FileInfo.
// size = -1 — размер не записан (файлы до миграции 010): сервис возьмёт его с диска.
func filesJSON(linkTable, ownerCol, ownerExpr string) string {
	return fmt.Sprintf(`(SELECT COALESCE(json_agg(json_build_object(
			'id', f.id, 'name', %s, 'mime_type', COALESCE(f.mime_type, 'application/octet-stream'),
			'size_bytes', COALESCE(f.size_bytes, -1), 'key', f.storage_key) ORDER BY f.id), '[]'::json)
		FROM %s l JOIN file f ON f.id = l.file_id WHERE l.%s = %s)`,
		database.FileName("f"), linkTable, ownerCol, ownerExpr)
}

func decodeFiles(raw []byte) ([]FileInfo, error) {
	files := []FileInfo{}
	if len(raw) == 0 {
		return files, nil
	}
	var rows []struct {
		ID        int64  `json:"id"`
		Name      string `json:"name"`
		MimeType  string `json:"mime_type"`
		SizeBytes int64  `json:"size_bytes"`
		Key       string `json:"key"`
	}
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, err
	}
	for _, r := range rows {
		files = append(files, FileInfo{ID: r.ID, Name: r.Name, MimeType: r.MimeType, SizeBytes: r.SizeBytes, key: r.Key})
	}
	return files, nil
}

var (
	errTopicMismatch  = apperr.NewInvalid("topic_course_mismatch", "topic does not exist or belongs to another course")
	errTopicTaken     = apperr.NewConflict("topic_assignment_exists", "this topic already has an assignment")
	errFileNotFound   = apperr.NewNotFound("file_not_found")
	errHasSubmissions = apperr.NewConflict("assignment_has_submissions", "the assignment has student work and cannot be deleted")
)

// assignmentErr уточняет ошибки ограничений по их имени.
func assignmentErr(err error) error {
	switch database.ConstraintName(err) {
	case "assignment_topic_course_fk":
		return errTopicMismatch
	case "assignment_topic_unique":
		return errTopicTaken
	case "assignment_course_id_fkey":
		return apperr.NewNotFound("course_not_found")
	case "assignment_file_file_id_fkey", "submission_file_file_id_fkey":
		return errFileNotFound
	}
	return database.Translate(err)
}

// ---- assignments ----

// $1 — смотрит ли преподаватель/админ (тогда считается статистика), $2 — id студента-зрителя
// (0, если зритель не студент: тогда LEFT JOIN ничего не находит).
var assignmentSelect = `
	SELECT a.id, a.course_id, a.topic_id, a.created_by, a.title, a.description, a.deadline_at, a.published_at,
	       ` + filesJSON("assignment_file", "assignment_id", "a.id") + `,
	       ` + derivedStatus + `, sb.submitted_at, ` + isLate + `, sb.grade::float8,
	       CASE WHEN $1::bool THEN (SELECT count(*) FROM course_offering c JOIN student st ON st.group_id = c.group_id
	                                WHERE c.id = a.course_id) END,
	       CASE WHEN $1::bool THEN (SELECT count(*) FROM submission x
	                                WHERE x.assignment_id = a.id AND x.status IN ('submitted', 'reviewed')) END,
	       CASE WHEN $1::bool THEN (SELECT count(*) FROM submission x
	                                WHERE x.assignment_id = a.id AND x.status = 'reviewed') END
	FROM assignment a
	JOIN topic t ON t.id = a.topic_id
	LEFT JOIN submission sb ON sb.assignment_id = a.id AND sb.student_id = $2`

// scanAssignment: staff — заполнить Stats, student — заполнить Mine.
func scanAssignment(staff, student bool) func(pgx.Row) (Assignment, error) {
	return func(row pgx.Row) (Assignment, error) {
		var (
			a                    Assignment
			filesRaw             []byte
			mine                 MySubmission
			total, sent, checked *int
		)
		err := row.Scan(&a.ID, &a.CourseID, &a.TopicID, &a.CreatedBy, &a.Title, &a.Description,
			&a.DeadlineAt, &a.PublishedAt, &filesRaw,
			&mine.Status, &mine.SubmittedAt, &mine.IsLate, &mine.Grade, &total, &sent, &checked)
		if err != nil {
			return Assignment{}, err
		}
		if a.Files, err = decodeFiles(filesRaw); err != nil {
			return Assignment{}, err
		}
		if student {
			a.Mine = &mine
		}
		if staff && total != nil && sent != nil && checked != nil {
			a.Stats = &Stats{StudentsTotal: *total, Submitted: *sent, Reviewed: *checked}
		}
		return a, nil
	}
}

func (r *Repository) ListAssignments(ctx context.Context, courseID int64, staff bool, viewerID int64) ([]Assignment, error) {
	return database.Collect(ctx, r.db, scanAssignment(staff, viewerID > 0),
		assignmentSelect+` WHERE a.course_id = $3 ORDER BY a.published_at DESC, a.id DESC`,
		staff, viewerID, courseID)
}

func (r *Repository) GetAssignment(ctx context.Context, id int64, staff bool, viewerID int64) (Assignment, error) {
	a, err := scanAssignment(staff, viewerID > 0)(r.db.QueryRow(ctx, assignmentSelect+` WHERE a.id = $3`, staff, viewerID, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Assignment{}, apperr.NewNotFound("assignment_not_found")
	}
	return a, database.Translate(err)
}

func (r *Repository) CreateAssignment(ctx context.Context, courseID, createdBy int64, in AssignmentInput) (int64, error) {
	var id int64
	err := database.InTx(ctx, r.db, func(tx pgx.Tx) error {
		err := tx.QueryRow(ctx, `
			INSERT INTO assignment (course_id, topic_id, created_by, title, description, deadline_at)
			VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
			courseID, in.TopicID, createdBy, in.Title, in.Description, in.DeadlineAt).Scan(&id)
		if err != nil {
			return assignmentErr(err)
		}
		_, err = replaceFiles(ctx, tx, "assignment_file", "assignment_id", id, in.FileIDs)
		return err
	})
	return id, err
}

// UpdateAssignment меняет тему, название, описание, дедлайн и набор файлов. Курс (группа) не меняется.
// Возвращает ключи файлов, оставшихся без ссылок, — их надо стереть с диска.
func (r *Repository) UpdateAssignment(ctx context.Context, id int64, in AssignmentInput) ([]string, error) {
	var keys []string
	err := database.InTx(ctx, r.db, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, `
			UPDATE assignment SET topic_id = $2, title = $3, description = $4, deadline_at = $5 WHERE id = $1`,
			id, in.TopicID, in.Title, in.Description, in.DeadlineAt)
		if err != nil {
			return assignmentErr(err)
		}
		if err := database.Affected(tag, nil, "assignment_not_found"); err != nil {
			return err
		}
		keys, err = replaceFiles(ctx, tx, "assignment_file", "assignment_id", id, in.FileIDs)
		return err
	})
	return keys, err
}

// DeleteAssignment удаляет задание без ответов и освободившиеся файлы задания.
// Если есть хоть один сохранённый ответ (даже черновик) — 409 assignment_has_submissions:
// удаление не должно уничтожать работы и оценки. Строка задания блокируется, чтобы
// студент не успел сохранить ответ между проверкой и удалением.
func (r *Repository) DeleteAssignment(ctx context.Context, id int64) ([]string, error) {
	var keys []string
	err := database.InTx(ctx, r.db, func(tx pgx.Tx) error {
		var hasWork bool
		err := tx.QueryRow(ctx, `
			SELECT EXISTS (SELECT 1 FROM submission WHERE assignment_id = a.id)
			FROM assignment a WHERE a.id = $1 FOR UPDATE`, id).Scan(&hasWork)
		if errors.Is(err, pgx.ErrNoRows) {
			return apperr.NewNotFound("assignment_not_found")
		}
		if err != nil {
			return database.Translate(err)
		}
		if hasWork {
			return errHasSubmissions
		}
		fileIDs, err := database.Collect(ctx, tx, scanID,
			`SELECT file_id FROM assignment_file WHERE assignment_id = $1`, id)
		if err != nil {
			return err
		}
		tag, err := tx.Exec(ctx, `DELETE FROM assignment WHERE id = $1`, id)
		if err := database.AffectedDelete(tag, err, "assignment_not_found"); err != nil {
			return err
		}
		keys, err = filegc.Purge(ctx, tx, fileIDs)
		return err
	})
	return keys, err
}

func scanID(row pgx.Row) (int64, error) {
	var id int64
	err := row.Scan(&id)
	return id, err
}

// replaceFiles заменяет набор файлов владельца (задания или ответа) и стирает записи о файлах,
// которые ни к чему больше не привязаны. linkTable и ownerCol — константы из кода.
func replaceFiles(ctx context.Context, tx pgx.Tx, linkTable, ownerCol string, ownerID int64, fileIDs []int64) ([]string, error) {
	old, err := database.Collect(ctx, tx, scanID,
		fmt.Sprintf(`DELETE FROM %s WHERE %s = $1 RETURNING file_id`, linkTable, ownerCol), ownerID)
	if err != nil {
		return nil, err
	}
	if len(fileIDs) > 0 {
		if _, err := tx.Exec(ctx,
			fmt.Sprintf(`INSERT INTO %s (%s, file_id) SELECT $1::int, unnest($2::bigint[])::int`, linkTable, ownerCol),
			ownerID, fileIDs); err != nil {
			return nil, assignmentErr(err)
		}
	}
	return filegc.Purge(ctx, tx, old)
}
