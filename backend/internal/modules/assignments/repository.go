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

// filesJSON — подзапрос «файлы владельца» в виде json-массива [{id, name}].
func filesJSON(linkTable, ownerCol, ownerExpr string) string {
	return fmt.Sprintf(`(SELECT COALESCE(json_agg(json_build_object('id', f.id, 'name', %s) ORDER BY f.id), '[]'::json)
		FROM %s l JOIN file f ON f.id = l.file_id WHERE l.%s = %s)`,
		database.FileName("f"), linkTable, ownerCol, ownerExpr)
}

func decodeFiles(raw []byte) ([]FileRef, error) {
	files := []FileRef{}
	if len(raw) == 0 {
		return files, nil
	}
	if err := json.Unmarshal(raw, &files); err != nil {
		return nil, err
	}
	return files, nil
}

// assignmentErr уточняет FK-ошибки по имени ограничения.
func assignmentErr(err error) error {
	switch database.ConstraintName(err) {
	case "assignment_topic_course_fk":
		return apperr.NewInvalid("topic_not_in_course", "topic does not exist or belongs to another course")
	case "assignment_course_id_fkey":
		return apperr.NewNotFound("course_not_found")
	case "assignment_file_file_id_fkey", "submission_file_file_id_fkey":
		return apperr.NewInvalid("file_not_found", "attached file does not exist")
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
		assignmentSelect+` WHERE a.course_id = $3 ORDER BY t.position, a.deadline_at NULLS LAST, a.id`,
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

// DeleteAssignment удаляет задание со всеми ответами (ON DELETE CASCADE) и освободившимися файлами.
func (r *Repository) DeleteAssignment(ctx context.Context, id int64) ([]string, error) {
	var keys []string
	err := database.InTx(ctx, r.db, func(tx pgx.Tx) error {
		fileIDs, err := database.Collect(ctx, tx, scanID, `
			SELECT file_id FROM assignment_file WHERE assignment_id = $1
			UNION
			SELECT sf.file_id FROM submission_file sf JOIN submission s ON s.id = sf.submission_id
			WHERE s.assignment_id = $1`, id)
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
