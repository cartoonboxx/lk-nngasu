package courses

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"university-api/internal/platform/apperr"
	"university-api/internal/platform/database"
)

type Repository struct {
	db *pgxpool.Pool
	// lessonMinutes — длительность одного занятия; в БД хранится только время начала,
	// поэтому часы шаблона считаются как число занятий × lessonMinutes / 45.
	lessonMinutes int
}

func newRepository(db *pgxpool.Pool, lessonMinutes int) *Repository {
	return &Repository{db: db, lessonMinutes: lessonMinutes}
}

const courseSelect = `
	SELECT c.id, c.subject_id, s.name, c.group_id, g.name,
	       c.semester_id, sem.name, sem.academic_year_start, sem.number,
	       (SELECT count(*) FROM workload w WHERE w.course_id = c.id), c.created_at
	FROM course_offering c
	JOIN subject s      ON s.id = c.subject_id
	JOIN study_group g  ON g.id = c.group_id
	JOIN semester sem   ON sem.id = c.semester_id`

func scanCourse(row pgx.Row) (Course, error) {
	var c Course
	err := row.Scan(&c.ID, &c.SubjectID, &c.SubjectName, &c.GroupID, &c.GroupName,
		&c.SemesterID, &c.SemesterName, &c.AcademicYearStart, &c.SemesterNumber, &c.WorkloadCount, &c.CreatedAt)
	return c, err
}

// ListCourses — страница курсов по фильтру; total учитывает все фильтры и права.
func (r *Repository) ListCourses(ctx context.Context, f CourseFilter) ([]Course, int, error) {
	var w database.Filter
	if f.GroupID != nil {
		w.Add("c.group_id = ?", *f.GroupID)
	}
	if f.SemesterID != nil {
		w.Add("c.semester_id = ?", *f.SemesterID)
	}
	if f.CurrentSemester {
		w.Raw("sem.is_current")
	}
	if f.Viewer != nil {
		// Фильтр group_id не расширяет доступ: он лишь сужает видимое пользователю.
		w.Add(`(c.group_id IN (SELECT st.group_id FROM student st WHERE st.user_id = ?)
		        OR EXISTS (SELECT 1 FROM workload wl WHERE wl.course_id = c.id AND wl.teacher_id = ?))`, *f.Viewer)
	}
	var total int
	countArgs := append([]any(nil), w.Args()...)
	if err := r.db.QueryRow(ctx, `SELECT count(*) FROM (`+courseSelect+w.Where()+`) x`, countArgs...).Scan(&total); err != nil {
		return nil, 0, database.Translate(err)
	}
	q := courseSelect + w.Where() + ` ORDER BY sem.academic_year_start DESC, sem.number DESC, s.name, g.name, c.id` +
		` LIMIT ` + w.Arg(f.Limit) + ` OFFSET ` + w.Arg(f.Offset)
	list, err := database.Collect(ctx, r.db, scanCourse, q, w.Args()...)
	return list, total, err
}

// errCourseExists — тройка дисциплина/группа/семестр уже занята.
var errCourseExists = apperr.NewConflict("course_exists", "this subject is already assigned to the group in this semester")

// errWorkloadExists — у курса уже есть нагрузка с тем же преподавателем и видом занятия.
var errWorkloadExists = apperr.NewConflict("workload_exists", "this teacher already has this lesson type in the course")

func (r *Repository) GetCourse(ctx context.Context, id int64) (Course, error) {
	c, err := scanCourse(r.db.QueryRow(ctx, courseSelect+` WHERE c.id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Course{}, apperr.NewNotFound("course_not_found")
	}
	return c, database.Translate(err)
}

// errBadReference — ссылка на несуществующую или архивную запись (400 invalid_reference).
func errBadReference(field, problem string) error {
	return apperr.NewInvalid("invalid_reference", "referenced record does not exist or is archived").
		WithDetails(map[string]any{"fields": map[string][]string{field: {problem}}})
}

// activeRef проверяет, что запись table с этим id есть и не в архиве.
// FOR SHARE не даёт заархивировать её, пока транзакция не завершится.
// table и field — константы из кода.
func activeRef(ctx context.Context, tx pgx.Tx, table, field string, id int64) error {
	var archived bool
	err := tx.QueryRow(ctx, `SELECT archived FROM `+table+` WHERE id = $1 FOR SHARE`, id).Scan(&archived)
	if errors.Is(err, pgx.ErrNoRows) {
		return errBadReference(field, "does not exist")
	}
	if err != nil {
		return database.Translate(err)
	}
	if archived {
		return errBadReference(field, "is archived")
	}
	return nil
}

// CreateCourse назначает дисциплину группе на семестр (CourseInput).
// Дисциплина, группа и семестр должны быть действующими; тройка уникальна (409).
func (r *Repository) CreateCourse(ctx context.Context, subjectID, groupID, semesterID int64) (int64, error) {
	var id int64
	err := database.InTx(ctx, r.db, func(tx pgx.Tx) error {
		for _, ref := range []struct {
			table, field string
			id           int64
		}{{"subject", "subject_id", subjectID}, {"study_group", "group_id", groupID}, {"semester", "semester_id", semesterID}} {
			if err := activeRef(ctx, tx, ref.table, ref.field, ref.id); err != nil {
				return err
			}
		}
		err := tx.QueryRow(ctx, `
			INSERT INTO course_offering (subject_id, group_id, semester_id) VALUES ($1, $2, $3) RETURNING id`,
			subjectID, groupID, semesterID).Scan(&id)
		if database.PgCode(err) == database.CodeUnique {
			return errCourseExists
		}
		return database.Translate(err)
	})
	return id, err
}

// ReplaceSubject меняет дисциплину у курса (CourseReplacement). Сохраняются id курса,
// нагрузка, занятия, темы, материалы и задания — меняется только ссылка на справочник.
// Новая дисциплина должна быть действующей; тройка остаётся уникальной (409 course_exists).
func (r *Repository) ReplaceSubject(ctx context.Context, courseID, subjectID int64) error {
	return database.InTx(ctx, r.db, func(tx pgx.Tx) error {
		var cur int64
		err := tx.QueryRow(ctx, `SELECT subject_id FROM course_offering WHERE id = $1 FOR UPDATE`, courseID).Scan(&cur)
		if errors.Is(err, pgx.ErrNoRows) {
			return apperr.NewNotFound("course_not_found")
		}
		if err != nil {
			return database.Translate(err)
		}
		if cur == subjectID {
			return nil
		}
		if err := activeRef(ctx, tx, "subject", "subject_id", subjectID); err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `UPDATE course_offering SET subject_id = $2 WHERE id = $1`, courseID, subjectID)
		if database.PgCode(err) == database.CodeUnique {
			return errCourseExists
		}
		return database.Translate(err)
	})
}

// DeleteCourse удаляет только пустой курс: с нагрузкой, темами или заданиями — 409 in_use.
func (r *Repository) DeleteCourse(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM course_offering WHERE id = $1`, id)
	return database.AffectedDelete(tag, err, "course_not_found")
}

// ---- students: студенты группы курса ----

func (r *Repository) ListCourseStudents(ctx context.Context, courseID int64) ([]CourseStudent, error) {
	return database.Collect(ctx, r.db, func(row pgx.Row) (CourseStudent, error) {
		var s CourseStudent
		err := row.Scan(&s.StudentID, &s.FullName, &s.RecordBookNumber, &s.IsHeadman)
		return s, err
	}, `
		SELECT st.user_id, `+database.FullName("u")+`, st.record_book_number, st.is_headman
		FROM course_offering c
		JOIN student st ON st.group_id = c.group_id
		JOIN app_user u ON u.id = st.user_id
		WHERE c.id = $1
		ORDER BY u.last_name, u.first_name, u.middle_names NULLS FIRST, u.id`, courseID)
}

// ---- workloads ----

func scanWorkload(row pgx.Row) (Workload, error) {
	var w Workload
	err := row.Scan(&w.ID, &w.CourseID, &w.TeacherID, &w.TeacherName, &w.LessonType, &w.PlannedHours, &w.ScheduledHours)
	return w, err
}

// workloadSelect — нагрузка с ФИО преподавателя и часами двухнедельного шаблона:
// число занятий × длительность занятия ($1, минуты) / 45, округление до 2 знаков.
const workloadSelect = `
	SELECT w.id, w.course_id, w.teacher_id, ` + "%s" + `, w.lesson_type::text, w.planned_hours::float8,
	       round((SELECT count(*) FROM lesson l WHERE l.workload_id = w.id) * $1::numeric / 45, 2)::float8
	FROM workload w
	JOIN app_user u ON u.id = w.teacher_id`

func (r *Repository) workloadQuery() string {
	return fmt.Sprintf(workloadSelect, database.FullName("u"))
}

func (r *Repository) ListWorkloads(ctx context.Context, courseID int64) ([]Workload, error) {
	return database.Collect(ctx, r.db, scanWorkload,
		r.workloadQuery()+` WHERE w.course_id = $2 ORDER BY w.lesson_type, u.last_name, w.id`, r.lessonMinutes, courseID)
}

func (r *Repository) GetWorkload(ctx context.Context, id int64) (Workload, error) {
	w, err := scanWorkload(r.db.QueryRow(ctx, r.workloadQuery()+` WHERE w.id = $2`, r.lessonMinutes, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Workload{}, apperr.NewNotFound("workload_not_found")
	}
	return w, database.Translate(err)
}

// checkTeacher: пользователь — преподаватель (есть роль teacher) и активен.
// Нет роли — 400 invalid_reference, заблокирован — 400 teacher_inactive.
// FOR SHARE: преподавателя не заблокируют и не снимут роль, пока идёт запись.
func checkTeacher(ctx context.Context, tx pgx.Tx, teacherID int64) error {
	var active bool
	err := tx.QueryRow(ctx, `
		SELECT u.is_active FROM teacher t JOIN app_user u ON u.id = t.user_id
		WHERE t.user_id = $1 FOR SHARE`, teacherID).Scan(&active)
	if errors.Is(err, pgx.ErrNoRows) {
		return errBadReference("teacher_id", "must be a user with the teacher role")
	}
	if err != nil {
		return database.Translate(err)
	}
	if !active {
		return apperr.NewInvalid("teacher_inactive", "the teacher's account is blocked")
	}
	return nil
}

// workloadErr — нарушение уникальности курс/преподаватель/вид занятия → 409 workload_exists.
func workloadErr(err error) error {
	if database.PgCode(err) == database.CodeUnique {
		return errWorkloadExists
	}
	return database.Translate(err)
}

// AddWorkload назначает преподавателя на курс (WorkloadInput).
func (r *Repository) AddWorkload(ctx context.Context, courseID, teacherID int64, lessonType string, hours float64) (int64, error) {
	var id int64
	err := database.InTx(ctx, r.db, func(tx pgx.Tx) error {
		var exists bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM course_offering WHERE id = $1)`, courseID).Scan(&exists); err != nil {
			return database.Translate(err)
		}
		if !exists {
			return apperr.NewNotFound("course_not_found")
		}
		if err := checkTeacher(ctx, tx, teacherID); err != nil {
			return err
		}
		err := tx.QueryRow(ctx, `
			INSERT INTO workload (course_id, teacher_id, lesson_type, planned_hours)
			VALUES ($1, $2, $3::lesson_kind, $4) RETURNING id`,
			courseID, teacherID, lessonType, hours).Scan(&id)
		return workloadErr(err)
	})
	return id, err
}

// teacherLockKey — тот же ключ advisory-блокировки преподавателя, что в модуле schedule
// (класс 2 в старших битах): смена преподавателя нагрузки и сохранение занятия
// с ним же выполняются по очереди и не пропустят пересечение.
func teacherLockKey(teacherID int64) int64 { return 2<<40 | teacherID }

// UpdateWorkload заменяет преподавателя, вид занятия и часы (WorkloadInput).
// Все занятия нагрузки сохраняются и переходят к новому преподавателю — для всего
// двухнедельного шаблона. Если новый преподаватель занят в каком-то из этих интервалов
// (в том же семестре), не сохраняется ничего: 409 schedule_conflict со списком пересечений.
func (r *Repository) UpdateWorkload(ctx context.Context, id, teacherID int64, lessonType string, hours float64) error {
	return database.InTx(ctx, r.db, func(tx pgx.Tx) error {
		var current int64
		err := tx.QueryRow(ctx, `SELECT teacher_id FROM workload WHERE id = $1 FOR UPDATE`, id).Scan(&current)
		if errors.Is(err, pgx.ErrNoRows) {
			return apperr.NewNotFound("workload_not_found")
		}
		if err != nil {
			return database.Translate(err)
		}
		if err := checkTeacher(ctx, tx, teacherID); err != nil {
			return err
		}
		if teacherID != current {
			if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock($1)`, teacherLockKey(teacherID)); err != nil {
				return database.Translate(err)
			}
			list, err := teacherConflicts(ctx, tx, id, teacherID)
			if err != nil {
				return err
			}
			if len(list) > 0 {
				return apperr.NewConflict("schedule_conflict", "the new teacher is busy during lessons of this workload").
					WithDetails(map[string]any{"conflicts": list})
			}
		}
		_, err = tx.Exec(ctx, `
			UPDATE workload SET teacher_id = $2, lesson_type = $3::lesson_kind, planned_hours = $4 WHERE id = $1`,
			id, teacherID, lessonType, hours)
		return workloadErr(err)
	})
}

// teacherConflicts — занятия преподавателя teacherID (из других нагрузок того же семестра),
// пересекающиеся по [starts_at, ends_at) с занятиями нагрузки workloadID.
func teacherConflicts(ctx context.Context, tx pgx.Tx, workloadID, teacherID int64) ([]ScheduleConflict, error) {
	rows, err := tx.Query(ctx, `
		SELECT DISTINCT o.id, o.week_number, o.day_of_week, to_char(o.starts_at, 'HH24:MI'), to_char(o.ends_at, 'HH24:MI'), o.starts_at
		FROM lesson l
		JOIN workload w        ON w.id = l.workload_id
		JOIN course_offering c ON c.id = w.course_id
		JOIN lesson o          ON o.week_number = l.week_number AND o.day_of_week = l.day_of_week
		                      AND o.starts_at < l.ends_at AND o.ends_at > l.starts_at
		JOIN workload ow       ON ow.id = o.workload_id
		JOIN course_offering oc ON oc.id = ow.course_id
		WHERE l.workload_id = $1 AND ow.teacher_id = $2 AND oc.semester_id = c.semester_id
		  AND o.workload_id <> $1
		ORDER BY o.week_number, o.day_of_week, o.starts_at, o.id`, workloadID, teacherID)
	if err != nil {
		return nil, database.Translate(err)
	}
	defer rows.Close()
	out := []ScheduleConflict{}
	for rows.Next() {
		var c ScheduleConflict
		var skip any
		if err := rows.Scan(&c.LessonID, &c.WeekNumber, &c.DayOfWeek, &c.StartsAt, &c.EndsAt, &skip); err != nil {
			return nil, database.Translate(err)
		}
		c.Resources = []string{"teacher"}
		out = append(out, c)
	}
	return out, database.Translate(rows.Err())
}

// DeleteWorkload: если по нагрузке уже есть занятия в расписании — 409 in_use.
func (r *Repository) DeleteWorkload(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM workload WHERE id = $1`, id)
	return database.AffectedDelete(tag, err, "workload_not_found")
}
