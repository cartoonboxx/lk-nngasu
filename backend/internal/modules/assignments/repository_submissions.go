package assignments

import (
	"context"
	"errors"
	"strconv"

	"github.com/jackc/pgx/v5"

	"university-api/internal/platform/apperr"
	"university-api/internal/platform/database"
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

// ListSubmissionRows возвращает по строке на каждого студента группы курса: с ответом или без.
func (r *Repository) ListSubmissionRows(ctx context.Context, assignmentID int64, f SubmissionFilter) ([]SubmissionRow, error) {
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
			w.Raw("sb.status IN ('submitted', 'reviewed')")
		} else {
			w.Raw("(sb.id IS NULL OR sb.status = 'not_submitted')")
		}
	}

	return database.Collect(ctx, r.db, func(row pgx.Row) (SubmissionRow, error) {
		var s SubmissionRow
		err := row.Scan(&s.StudentID, &s.StudentName, &s.RecordBookNumber, &s.SubmissionID,
			&s.Status, &s.SubmittedAt, &s.IsLate, &s.Grade)
		return s, err
	}, `
		SELECT st.user_id, `+database.FullName("u")+`, st.record_book_number, sb.id,
		       `+derivedStatus+`, sb.submitted_at, `+isLate+`, sb.grade::float8
		FROM assignment a
		JOIN course_offering c ON c.id = a.course_id
		JOIN student st        ON st.group_id = c.group_id
		JOIN app_user u        ON u.id = st.user_id
		LEFT JOIN submission sb ON sb.assignment_id = a.id AND sb.student_id = st.user_id`+
		w.Where()+` ORDER BY u.last_name, u.first_name, st.user_id`, w.Args()...)
}

// SaveSubmission создаёт или обновляет ответ студента. Проверенный ответ менять нельзя (409).
// submit=true отправляет ответ на проверку; ответ, уже отправленный ранее, остаётся отправленным.
func (r *Repository) SaveSubmission(ctx context.Context, assignmentID, studentID int64, in SubmissionInput) (int64, []string, error) {
	var (
		id   int64
		keys []string
	)
	err := database.InTx(ctx, r.db, func(tx pgx.Tx) error {
		err := tx.QueryRow(ctx, `
			INSERT INTO submission (assignment_id, student_id, status, answer_text, submitted_at)
			VALUES ($1, $2, (CASE WHEN $3::bool THEN 'submitted' ELSE 'not_submitted' END)::submission_state, $4,
			        CASE WHEN $3::bool THEN now() END)
			ON CONFLICT (assignment_id, student_id) DO UPDATE SET
				answer_text  = EXCLUDED.answer_text,
				status       = (CASE WHEN $3::bool OR submission.status = 'submitted'
				                     THEN 'submitted' ELSE 'not_submitted' END)::submission_state,
				submitted_at = CASE WHEN $3::bool OR submission.status = 'submitted' THEN now() END
			WHERE submission.status <> 'reviewed'
			RETURNING id`, assignmentID, studentID, in.Submit, in.AnswerText).Scan(&id)
		if errors.Is(err, pgx.ErrNoRows) { // конфликт есть, но WHERE отсёк: ответ уже проверен
			return apperr.NewConflict("already_reviewed", "the answer has been reviewed and can no longer be changed")
		}
		if err != nil {
			return assignmentErr(err)
		}
		keys, err = replaceFiles(ctx, tx, "submission_file", "submission_id", id, in.FileIDs)
		return err
	})
	return id, keys, err
}

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

func (r *Repository) ListGrades(ctx context.Context, f GradeFilter) ([]GradeRow, error) {
	var w database.Filter
	w.Raw("sb.status = 'reviewed'")
	if f.StudentID != nil {
		w.Add("sb.student_id = ?", *f.StudentID)
	}
	if f.CourseID != nil {
		w.Add("a.course_id = ?", *f.CourseID)
	}
	return database.Collect(ctx, r.db, func(row pgx.Row) (GradeRow, error) {
		var g GradeRow
		err := row.Scan(&g.AssignmentID, &g.AssignmentTitle, &g.CourseID, &g.SubjectName,
			&g.StudentID, &g.StudentName, &g.Grade, &g.ReviewComment, &g.ReviewedAt)
		return g, err
	}, `
		SELECT a.id, a.title, a.course_id, subj.name, sb.student_id, `+database.FullName("u")+`,
		       sb.grade::float8, sb.review_comment, sb.reviewed_at
		FROM submission sb
		JOIN assignment a      ON a.id = sb.assignment_id
		JOIN course_offering c ON c.id = a.course_id
		JOIN subject subj      ON subj.id = c.subject_id
		JOIN app_user u        ON u.id = sb.student_id`+
		w.Where()+` ORDER BY sb.reviewed_at DESC, sb.id DESC`, w.Args()...)
}

// ---- данные для zip-архивов ----

// ArchiveEntry — один отправленный ответ: текст и ключи его файлов.
type ArchiveEntry struct {
	SubmissionID int64
	StudentName  string
	AnswerText   *string
	Keys         []string
}

// ArchiveEntries собирает отправленные ответы задания (или один ответ, если задан onlySubmission).
func (r *Repository) ArchiveEntries(ctx context.Context, assignmentID int64, onlySubmission *int64) ([]ArchiveEntry, error) {
	var w database.Filter
	w.Add("sb.assignment_id = ?", assignmentID)
	w.Raw("sb.status IN ('submitted', 'reviewed')")
	if onlySubmission != nil {
		w.Add("sb.id = ?", *onlySubmission)
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
