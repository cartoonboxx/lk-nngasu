// Package access — сквозные проверки доступа к курсам и файлам.
// Это инфраструктура, а не бизнес-модуль: ей пользуются content, assignments, users.
//
// Студенты курса — это студенты его группы (отдельной записи на курс нет).
package access

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"university-api/internal/platform/apperr"
	"university-api/internal/platform/filekey"
	"university-api/internal/platform/middleware"
)

var (
	errForbidden   = apperr.NewForbidden("forbidden", "no access to this resource")
	errInvalidFile = apperr.NewInvalid("invalid_file", "file does not exist or is not yours")
)

type Checker struct{ db *pgxpool.Pool }

func NewChecker(db *pgxpool.Pool) *Checker { return &Checker{db: db} }

func (k *Checker) exists(ctx context.Context, q string, args ...any) (bool, error) {
	var ok bool
	err := k.db.QueryRow(ctx, q, args...).Scan(&ok)
	return ok, err
}

func (k *Checker) IsCourseTeacher(ctx context.Context, userID, courseID int64) (bool, error) {
	return k.exists(ctx,
		`SELECT EXISTS (SELECT 1 FROM workload WHERE course_id = $1 AND teacher_id = $2)`,
		courseID, userID)
}

// StudentOfCourse — студент учится в группе, у которой идёт этот курс.
func (k *Checker) StudentOfCourse(ctx context.Context, studentID, courseID int64) (bool, error) {
	return k.exists(ctx,
		`SELECT EXISTS (SELECT 1 FROM course_offering c
		                JOIN student s ON s.group_id = c.group_id
		                WHERE c.id = $1 AND s.user_id = $2)`,
		courseID, studentID)
}

// CourseView: админ, преподаватель курса или студент группы курса.
func (k *Checker) CourseView(ctx context.Context, p middleware.Principal, courseID int64) error {
	if p.IsAdmin() {
		return nil
	}
	var ok bool
	var err error
	if p.IsTeacher() {
		ok, err = k.IsCourseTeacher(ctx, p.UserID, courseID)
	} else {
		ok, err = k.StudentOfCourse(ctx, p.UserID, courseID)
	}
	if err != nil {
		return err
	}
	if !ok {
		return errForbidden
	}
	return nil
}

// CourseTeacher: только преподаватель этого курса. Нужен там, где в БД стоит FK на teacher
// (created_by, reviewed_by) — администратор такие записи создавать не может.
func (k *Checker) CourseTeacher(ctx context.Context, p middleware.Principal, courseID int64) error {
	if !p.IsTeacher() {
		return errForbidden
	}
	ok, err := k.IsCourseTeacher(ctx, p.UserID, courseID)
	if err != nil {
		return err
	}
	if !ok {
		return errForbidden
	}
	return nil
}

// CourseManage: администратор или преподаватель этого курса.
func (k *Checker) CourseManage(ctx context.Context, p middleware.Principal, courseID int64) error {
	if p.IsAdmin() {
		return nil
	}
	return k.CourseTeacher(ctx, p, courseID)
}

// OwnsFile: файл существует и лежит в папке этого пользователя
// (защита от привязки чужих файлов по id).
func (k *Checker) OwnsFile(ctx context.Context, userID, fileID int64) error {
	ok, err := k.exists(ctx,
		`SELECT EXISTS (SELECT 1 FROM file WHERE id = $1 AND starts_with(storage_key, $2::text))`,
		fileID, filekey.Prefix(userID))
	if err != nil {
		return err
	}
	if !ok {
		return errInvalidFile
	}
	return nil
}
