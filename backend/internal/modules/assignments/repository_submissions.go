package assignments

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"

	"university-api/internal/platform/apperr"
	"university-api/internal/platform/database"
	"university-api/internal/platform/filegc"
)

var submissionSelect = `
	SELECT sb.id, sb.assignment_id, a.course_id, sb.student_id, ` + database.FullName("u") + `,
	       sb.status::text, sb.answer_text, sb.submitted_at, ` + isLate + `,
	       sb.review_comment, sb.reviewed_by, sb.reviewed_at, sb.grade::float8,
	       ` + filesJSON("submission_file", "submission_id", "sb.id") + `
	FROM submission sb
	JOIN assignment a ON a.id = sb.assignment_id
	JOIN app_user u   ON u.id = sb.student_id`

func scanSubmission(row pgx.Row) (Submission, error) {
	var (
		s        Submission
		filesRaw []byte
	)
	err := row.Scan(&s.ID, &s.AssignmentID, &s.CourseID, &s.StudentID, &s.StudentName,
		&s.Status, &s.AnswerText, &s.SubmittedAt, &s.IsLate,
		&s.ReviewComment, &s.ReviewedBy, &s.ReviewedAt, &s.Grade, &filesRaw)
	if err != nil {
		return Submission{}, err
	}
	s.Files, err = decodeFiles(filesRaw)
	return s, err
}

func (r *Repository) getSubmission(ctx context.Context, where string, args ...any) (Submission, error) {
	s, err := scanSubmission(r.db.QueryRow(ctx, submissionSelect+where, args...))
	if errors.Is(err, pgx.ErrNoRows) {
		return Submission{}, apperr.NewNotFound("submission_not_found")
	}
	return s, database.Translate(err)
}

func (r *Repository) GetSubmission(ctx context.Context, id int64) (Submission, error) {
	return r.getSubmission(ctx, ` WHERE sb.id = $1`, id)
}

func (r *Repository) GetSubmissionOf(ctx context.Context, assignmentID, studentID int64) (Submission, error) {
	return r.getSubmission(ctx, ` WHERE sb.assignment_id = $1 AND sb.student_id = $2`, assignmentID, studentID)
}

// ListSubmissionRows возвращает страницу «по строке на каждого студента группы курса: с ответом
// или без» и общее число строк с учётом фильтров. Черновик не считается работой: у студента
// с черновиком статус not_submitted/missing и нет submission_id — преподаватель черновиков не видит.
func (r *Repository) ListSubmissionRows(ctx context.Context, assignmentID int64, f SubmissionFilter) ([]SubmissionRow, int, error) {
	var w database.Filter
	w.Add("a.id = ?", assignmentID)
	w.Raw("(u.is_active OR sb.id IS NOT NULL)") // выбывших студентов показываем, только если они что-то сдавали
	if f.Status != nil {
		w.Add(derivedStatus+" = ?", *f.Status)
	}
	if f.Late != nil {
		if *f.Late {
			w.Raw(isLate)
		} else {
			w.Raw("(sb.submitted_at IS NOT NULL AND NOT " + isLate + ")")
		}
	}
	if f.HasWork != nil {
		if *f.HasWork {
			w.Raw("sb.id IS NOT NULL")
		} else {
			w.Raw("sb.id IS NULL")
		}
	}

	from := `
		FROM assignment a
		JOIN course_offering c ON c.id = a.course_id
		JOIN student st        ON st.group_id = c.group_id
		JOIN app_user u        ON u.id = st.user_id
		LEFT JOIN submission sb ON sb.assignment_id = a.id AND sb.student_id = st.user_id
		                       AND sb.status <> 'not_submitted'` + w.Where()

	var total int
	if err := r.db.QueryRow(ctx, `SELECT count(*)`+from, w.Args()...).Scan(&total); err != nil {
		return nil, 0, database.Translate(err)
	}
	args := append(w.Args(), f.Limit, f.Offset)
	n := len(w.Args())
	list, err := database.Collect(ctx, r.db, func(row pgx.Row) (SubmissionRow, error) {
		var s SubmissionRow
		err := row.Scan(&s.StudentID, &s.StudentName, &s.RecordBookNumber, &s.SubmissionID,
			&s.Status, &s.SubmittedAt, &s.IsLate, &s.Grade)
		return s, err
	}, `
		SELECT st.user_id, `+database.FullName("u")+`, st.record_book_number, sb.id,
		       `+derivedStatus+`, sb.submitted_at, `+isLate+`, sb.grade::float8`+from+
		` ORDER BY u.last_name, u.first_name, st.user_id`+
		fmt.Sprintf(` LIMIT $%d OFFSET $%d`, n+1, n+2), args...)
	return list, total, err
}

// SaveSubmission создаёт или обновляет ответ студента (одна строка на пару «задание — студент»,
// id сохраняется). Правила:
//   - submit=true — отправить: submitted_at = now(); повторная отправка до проверки разрешена;
//   - submit=false — черновик, только пока ответ не отправлен (иначе 409 already_submitted;
//     отозвать отправленный ответ можно через DELETE);
//   - проверенный ответ не меняется (409 already_reviewed).
//
// Текущий ответ читается с FOR UPDATE: параллельная проверка преподавателем не проскочит
// между чтением статуса и записью.
func (r *Repository) SaveSubmission(ctx context.Context, assignmentID, studentID int64, in SubmissionInput) (int64, []string, error) {
	var (
		id   int64
		keys []string
	)
	err := database.InTx(ctx, r.db, func(tx pgx.Tx) error {
		var status string
		err := tx.QueryRow(ctx, `
			SELECT id, status::text FROM submission WHERE assignment_id = $1 AND student_id = $2 FOR UPDATE`,
			assignmentID, studentID).Scan(&id, &status)
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			err = tx.QueryRow(ctx, `
				INSERT INTO submission (assignment_id, student_id, status, answer_text, submitted_at)
				VALUES ($1, $2, (CASE WHEN $3::bool THEN 'submitted' ELSE 'not_submitted' END)::submission_state, $4,
				        CASE WHEN $3::bool THEN now() END)
				RETURNING id`, assignmentID, studentID, in.Submit, in.AnswerText).Scan(&id)
			if err != nil {
				return assignmentErr(err) // в том числе гонка двух первых сохранений — уникальный ключ
			}
		case err != nil:
			return database.Translate(err)
		case status == "reviewed":
			return errReviewed
		case status == "submitted" && !in.Submit:
			return errAlreadySubmitted
		default:
			if _, err := tx.Exec(ctx, `
				UPDATE submission SET answer_text = $2,
				       status = (CASE WHEN $3::bool THEN 'submitted' ELSE 'not_submitted' END)::submission_state,
				       submitted_at = CASE WHEN $3::bool THEN now() END
				WHERE id = $1`, id, in.AnswerText, in.Submit); err != nil {
				return assignmentErr(err)
			}
		}
		keys, err = replaceFiles(ctx, tx, "submission_file", "submission_id", id, in.FileIDs)
		return err
	})
	return id, keys, err
}

// DeleteSubmission — студент отзывает свой ответ (черновик или отправленный) вместе с файлами.
// Проверенный ответ удалить нельзя (409 already_reviewed). Ответа нет — ничего не делает.
func (r *Repository) DeleteSubmission(ctx context.Context, assignmentID, studentID int64) ([]string, error) {
	var keys []string
	err := database.InTx(ctx, r.db, func(tx pgx.Tx) error {
		var (
			id     int64
			status string
		)
		err := tx.QueryRow(ctx, `
			SELECT id, status::text FROM submission WHERE assignment_id = $1 AND student_id = $2 FOR UPDATE`,
			assignmentID, studentID).Scan(&id, &status)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return database.Translate(err)
		}
		if status == "reviewed" {
			return errReviewed
		}
		fileIDs, err := database.Collect(ctx, tx, scanID, `DELETE FROM submission_file WHERE submission_id = $1 RETURNING file_id`, id)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `DELETE FROM submission WHERE id = $1`, id); err != nil {
			return database.Translate(err)
		}
		keys, err = filegc.Purge(ctx, tx, fileIDs)
		return err
	})
	return keys, err
}

var (
	errReviewed         = apperr.NewConflict("already_reviewed", "the answer has been reviewed and can no longer be changed")
	errAlreadySubmitted = apperr.NewConflict("already_submitted", "a submitted answer cannot become a draft again; withdraw it with DELETE")
)

// Review выставляет оценку и переводит ответ в reviewed. Повторная проверка перезаписывает оценку.
func (r *Repository) Review(ctx context.Context, id, teacherID int64, in ReviewInput) error {
	grade := strconv.FormatFloat(in.Grade, 'f', 1, 64) // текстом: numeric(2,1) разберёт сам PostgreSQL
	var got int64
	err := r.db.QueryRow(ctx, `
		UPDATE submission SET status = 'reviewed', grade = $2::numeric, review_comment = $3,
		       reviewed_by = $4, reviewed_at = now()
		WHERE id = $1 AND status IN ('submitted', 'reviewed')
		RETURNING id`, id, grade, in.Comment, teacherID).Scan(&got)
	if errors.Is(err, pgx.ErrNoRows) {
		// либо ответа нет, либо он ещё не отправлен
		if _, gerr := r.GetSubmission(ctx, id); gerr != nil {
			return gerr
		}
		return apperr.NewConflict("not_submitted", "the student has not submitted this answer yet")
	}
	return database.Translate(err)
}

// ---- оценки ----

// ListGrades — страница строк «задание × студент группы курса» с текущим статусом и оценкой
// и общее число строк с учётом фильтров. Ответы-черновики не учитываются (как будто ответа нет).
// Выбывших студентов показываем, только если они что-то сдавали.
func (r *Repository) ListGrades(ctx context.Context, f GradeFilter) ([]GradeRow, int, error) {
	var w database.Filter
	w.Raw("(u.is_active OR sb.id IS NOT NULL)")
	if f.StudentID != nil {
		w.Add("st.user_id = ?", *f.StudentID)
	}
	if f.CourseID != nil {
		w.Add("c.id = ?", *f.CourseID)
	}
	if f.SemesterID != nil {
		w.Add("c.semester_id = ?", *f.SemesterID)
	}
	if f.Status != nil {
		w.Add(derivedStatus+" = ?", *f.Status)
	}
	if f.Q != "" {
		ph := w.Arg("%" + likeEscape(f.Q) + "%")
		w.Raw(fmt.Sprintf("(a.title ILIKE %[1]s OR t.title ILIKE %[1]s OR subj.name ILIKE %[1]s)", ph))
	}
	from := `
		FROM assignment a
		JOIN topic t           ON t.id = a.topic_id
		JOIN course_offering c ON c.id = a.course_id
		JOIN subject subj      ON subj.id = c.subject_id
		JOIN semester sem      ON sem.id = c.semester_id
		JOIN student st        ON st.group_id = c.group_id
		JOIN app_user u        ON u.id = st.user_id
		LEFT JOIN submission sb ON sb.assignment_id = a.id AND sb.student_id = st.user_id
		                       AND sb.status <> 'not_submitted'` + w.Where()

	var total int
	if err := r.db.QueryRow(ctx, `SELECT count(*)`+from, w.Args()...).Scan(&total); err != nil {
		return nil, 0, database.Translate(err)
	}
	n := len(w.Args())
	args := append(w.Args(), f.Limit, f.Offset)
	list, err := database.Collect(ctx, r.db, func(row pgx.Row) (GradeRow, error) {
		var g GradeRow
		err := row.Scan(&g.AssignmentID, &g.AssignmentTitle, &g.TopicTitle, &g.CourseID, &g.SubjectName,
			&g.SemesterID, &g.SemesterName, &g.StudentID, &g.StudentName, &g.SubmissionID,
			&g.Status, &g.SubmittedAt, &g.Grade, &g.ReviewComment, &g.ReviewedAt)
		return g, err
	}, `
		SELECT a.id, a.title, t.title, c.id, subj.name, sem.id, sem.name,
		       st.user_id, `+database.FullName("u")+`, sb.id,
		       `+derivedStatus+`, sb.submitted_at, sb.grade::float8, sb.review_comment, sb.reviewed_at`+from+`
		ORDER BY subj.name, u.last_name, u.first_name, st.user_id, a.deadline_at NULLS LAST, a.id`+
		fmt.Sprintf(` LIMIT $%d OFFSET $%d`, n+1, n+2), args...)
	return list, total, err
}

// likeEscape экранирует спецсимволы LIKE, чтобы «_» и «%» в запросе искались буквально.
func likeEscape(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}

// CurrentSemester — id текущего семестра; ok = false, если он не выбран.
func (r *Repository) CurrentSemester(ctx context.Context) (int64, bool, error) {
	var id int64
	err := r.db.QueryRow(ctx, `SELECT id FROM semester WHERE is_current`).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, false, nil
	}
	return id, err == nil, database.Translate(err)
}

// SemesterExists — 404 semester_not_found для неизвестного id.
func (r *Repository) SemesterExists(ctx context.Context, id int64) error {
	var ok bool
	if err := r.db.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM semester WHERE id = $1)`, id).Scan(&ok); err != nil {
		return database.Translate(err)
	}
	if !ok {
		return apperr.NewNotFound("semester_not_found")
	}
	return nil
}

// ---- данные для zip-архивов ----

// ArchiveEntry — один отправленный ответ: текст и ключи его файлов.
type ArchiveEntry struct {
	SubmissionID int64
	StudentName  string
	AnswerText   *string
	Keys         []string
}

// ArchiveEntries собирает отправленные ответы задания или один ответ (onlySubmission).
// Черновики в архив задания не попадают; один ответ отдаётся как есть — права на него
// (в том числе на свой черновик) уже проверил сервис.
func (r *Repository) ArchiveEntries(ctx context.Context, assignmentID int64, onlySubmission *int64) ([]ArchiveEntry, error) {
	var w database.Filter
	w.Add("sb.assignment_id = ?", assignmentID)
	if onlySubmission != nil {
		w.Add("sb.id = ?", *onlySubmission)
	} else {
		w.Raw("sb.status IN ('submitted', 'reviewed')")
	}

	type row struct {
		id     int64
		name   string
		answer *string
		key    *string
	}
	rows, err := database.Collect(ctx, r.db, func(rw pgx.Row) (row, error) {
		var x row
		err := rw.Scan(&x.id, &x.name, &x.answer, &x.key)
		return x, err
	}, `
		SELECT sb.id, `+database.FullName("u")+`, sb.answer_text, f.storage_key
		FROM submission sb
		JOIN app_user u ON u.id = sb.student_id
		LEFT JOIN submission_file sf ON sf.submission_id = sb.id
		LEFT JOIN file f ON f.id = sf.file_id`+
		w.Where()+` ORDER BY u.last_name, u.first_name, sb.id, f.id`, w.Args()...)
	if err != nil {
		return nil, err
	}

	var out []ArchiveEntry
	for _, x := range rows {
		if len(out) == 0 || out[len(out)-1].SubmissionID != x.id {
			out = append(out, ArchiveEntry{SubmissionID: x.id, StudentName: x.name, AnswerText: x.answer})
		}
		if x.key != nil {
			last := &out[len(out)-1]
			last.Keys = append(last.Keys, *x.key)
		}
	}
	return out, nil
}
