package courses

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

const courseSelect = `
	SELECT c.id, c.subject_id, s.name, c.group_id, g.name,
	       c.semester_id, sem.academic_year_start, sem.number, c.created_at
	FROM course_offering c
	JOIN subject s      ON s.id = c.subject_id
	JOIN study_group g  ON g.id = c.group_id
	JOIN semester sem   ON sem.id = c.semester_id`

func scanCourse(row pgx.Row) (Course, error) {
	var c Course
	err := row.Scan(&c.ID, &c.SubjectID, &c.SubjectName, &c.GroupID, &c.GroupName,
		&c.SemesterID, &c.AcademicYearStart, &c.SemesterNumber, &c.CreatedAt)
	return c, err
}

func (r *Repository) ListCourses(ctx context.Context, f CourseFilter) ([]Course, error) {
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
	if f.ForTeacher != nil {
		w.Add("EXISTS (SELECT 1 FROM workload w WHERE w.course_id = c.id AND w.teacher_id = ?)", *f.ForTeacher)
	}
	if f.ForStudent != nil {
		w.Add("c.group_id = (SELECT st.group_id FROM student st WHERE st.user_id = ?)", *f.ForStudent)
	}
	return database.Collect(ctx, r.db, scanCourse,
		courseSelect+w.Where()+` ORDER BY sem.academic_year_start DESC, sem.number DESC, s.name, g.name`, w.Args()...)
}

func (r *Repository) GetCourse(ctx context.Context, id int64) (Course, error) {
	c, err := scanCourse(r.db.QueryRow(ctx, courseSelect+` WHERE c.id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Course{}, apperr.NewNotFound("course_not_found")
	}
	return c, database.Translate(err)
}

func (r *Repository) CreateCourse(ctx context.Context, subjectID, groupID, semesterID int64) (int64, error) {
	var id int64
	err := r.db.QueryRow(ctx, `
		INSERT INTO course_offering (subject_id, group_id, semester_id) VALUES ($1, $2, $3) RETURNING id`,
		subjectID, groupID, semesterID).Scan(&id)
	return id, database.Translate(err)
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
		WHERE c.id = $1 AND u.is_active
		ORDER BY u.last_name, u.first_name, u.id`, courseID)
}

// ---- workloads ----

func scanWorkload(row pgx.Row) (Workload, error) {
	var w Workload
	err := row.Scan(&w.ID, &w.CourseID, &w.TeacherID, &w.TeacherName, &w.LessonType, &w.PlannedHours)
	return w, err
}

func (r *Repository) ListWorkloads(ctx context.Context, courseID int64) ([]Workload, error) {
	return database.Collect(ctx, r.db, scanWorkload, `
		SELECT w.id, w.course_id, w.teacher_id, `+database.FullName("u")+`, w.lesson_type::text, w.planned_hours::float8
		FROM workload w
		JOIN app_user u ON u.id = w.teacher_id
		WHERE w.course_id = $1
		ORDER BY w.lesson_type, u.last_name, w.id`, courseID)
}

func (r *Repository) AddWorkload(ctx context.Context, courseID, teacherID int64, lessonType string, hours *float64) (int64, error) {
	var id int64
	err := r.db.QueryRow(ctx, `
		INSERT INTO workload (course_id, teacher_id, lesson_type, planned_hours)
		VALUES ($1, $2, $3::lesson_kind, $4) RETURNING id`,
		courseID, teacherID, lessonType, hours).Scan(&id)
	return id, database.Translate(err)
}

// DeleteWorkload: если по нагрузке уже есть занятия в расписании — 409 in_use.
func (r *Repository) DeleteWorkload(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM workload WHERE id = $1`, id)
	return database.AffectedDelete(tag, err, "workload_not_found")
}
