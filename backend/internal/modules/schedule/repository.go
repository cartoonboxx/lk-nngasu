package schedule

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"university-api/internal/platform/apperr"
	"university-api/internal/platform/database"
)

// Repository — доступ к таблице lesson и чтение связанных данных для показа расписания.
type Repository struct{ db *pgxpool.Pool }

func newRepository(db *pgxpool.Pool) *Repository { return &Repository{db: db} }

// lessonErr уточняет ошибки внешних ключей и уникальности по имени ограничения.
// Несуществующие нагрузка или аудитория — 400 с понятным кодом вместо общей
// ошибки внешнего ключа; повтор занятия той же нагрузки в том же слоте — 409.
// Остальное (и nil) уходит в database.Translate.
// Имена ограничений должны совпадать с миграциями.
func lessonErr(err error) error {
	switch database.ConstraintName(err) {
	case "lesson_workload_id_fkey":
		return apperr.NewInvalid("workload_not_found", "workload does not exist")
	case "lesson_room_id_fkey":
		return apperr.NewInvalid("room_not_found", "room does not exist")
	case "lesson_workload_id_week_number_day_of_week_starts_at_key":
		return apperr.NewConflict("lesson_exists", "this workload already has a lesson in this slot")
	}
	return database.Translate(err)
}

// Расписание — read-модель: осознанно читает таблицы соседних модулей одним запросом.
//
// Цепочка связей: занятие (lesson) → строка нагрузки (workload: преподаватель
// и вид занятия) → дисциплина группы в семестре (course_offering) → дисциплина,
// группа. Плюс преподаватель и аудитория для отображения.
// Все JOIN-ы внутренние: каждая связь обязательна (NOT NULL + внешний ключ).
// Время отдаётся строкой "HH:MM", без секунд.
// Порядок колонок должен совпадать со scanLesson.
var lessonSelect = `
	SELECT l.id, l.workload_id, c.id, s.name, w.lesson_type::text, w.teacher_id, ` + database.FullName("u") + `,
	       c.group_id, g.name, l.week_number, l.day_of_week, to_char(l.starts_at, 'HH24:MI'), l.room_id, rm.number
	FROM lesson l
	JOIN workload w        ON w.id = l.workload_id
	JOIN course_offering c ON c.id = w.course_id
	JOIN subject s         ON s.id = c.subject_id
	JOIN study_group g     ON g.id = c.group_id
	JOIN app_user u        ON u.id = w.teacher_id
	JOIN room rm           ON rm.id = l.room_id`

// scanLesson читает строку lessonSelect. EndsAt здесь не заполняется:
// его вычисляет сервис по сетке звонков.
func scanLesson(row pgx.Row) (Lesson, error) {
	var l Lesson
	err := row.Scan(&l.ID, &l.WorkloadID, &l.CourseID, &l.Subject, &l.LessonType, &l.TeacherID, &l.TeacherName,
		&l.GroupID, &l.GroupName, &l.WeekNumber, &l.DayOfWeek, &l.StartsAt, &l.RoomID, &l.Room)
	return l, err
}

// Get возвращает занятие по id.
func (r *Repository) Get(ctx context.Context, id int64) (Lesson, error) {
	l, err := scanLesson(r.db.QueryRow(ctx, lessonSelect+` WHERE l.id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Lesson{}, apperr.NewNotFound("lesson_not_found")
	}
	return l, database.Translate(err)
}

// List возвращает занятия сетки по фильтру, упорядоченные как в расписании:
// неделя → день → время, внутри слота — по дисциплине.
//
// Семестр: если SemesterID не задан, берётся текущий. Если текущий не выбран,
// результат пустой, а не ошибка.
// StudentID — «группа этого студента»: подзапрос находит её по user_id.
// Если такого студента нет, подзапрос даёт NULL и список будет пустым.
func (r *Repository) List(ctx context.Context, f LessonFilter) ([]Lesson, error) {
	var w database.Filter
	if f.SemesterID != nil {
		w.Add("c.semester_id = ?", *f.SemesterID)
	} else {
		w.Raw("c.semester_id IN (SELECT id FROM semester WHERE is_current)")
	}
	if f.GroupID != nil {
		w.Add("c.group_id = ?", *f.GroupID)
	}
	if f.TeacherID != nil {
		w.Add("w.teacher_id = ?", *f.TeacherID)
	}
	if f.StudentID != nil {
		w.Add("c.group_id = (SELECT st.group_id FROM student st WHERE st.user_id = ?)", *f.StudentID)
	}
	if f.Week != nil {
		w.Add("l.week_number = ?", *f.Week)
	}
	if f.Day != nil {
		w.Add("l.day_of_week = ?", *f.Day)
	}
	return database.Collect(ctx, r.db, scanLesson,
		lessonSelect+w.Where()+` ORDER BY l.week_number, l.day_of_week, l.starts_at, s.name, l.id`, w.Args()...)
}

// Create добавляет занятие в сетку и возвращает его id.
// StartsAt уже приведён сервисом к "HH:MM", здесь он приводится к типу time.
func (r *Repository) Create(ctx context.Context, workloadID int64, in LessonInput) (int64, error) {
	var id int64
	err := r.db.QueryRow(ctx, `
		INSERT INTO lesson (workload_id, week_number, day_of_week, starts_at, room_id)
		VALUES ($1, $2, $3, $4::time, $5) RETURNING id`,
		workloadID, in.WeekNumber, in.DayOfWeek, in.StartsAt, in.RoomID).Scan(&id)
	return id, lessonErr(err)
}

// Update переносит занятие: слот и аудиторию. workload_id не меняется.
func (r *Repository) Update(ctx context.Context, id int64, in LessonInput) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE lesson SET week_number = $2, day_of_week = $3, starts_at = $4::time, room_id = $5
		WHERE id = $1`, id, in.WeekNumber, in.DayOfWeek, in.StartsAt, in.RoomID)
	if err != nil {
		return lessonErr(err)
	}
	return database.Affected(tag, nil, "lesson_not_found")
}

// Delete удаляет занятие. Если на него ссылаются другие данные
// (например, посещаемость), AffectedDelete вернёт ошибку «используется».
func (r *Repository) Delete(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM lesson WHERE id = $1`, id)
	return database.AffectedDelete(tag, err, "lesson_not_found")
}

// ---- справочники для выбора в расписании (доступны гостям) ----
// Возвращают только то, что безопасно показать без входа: id и названия.

// Groups — все группы: сначала младшие наборы, внутри набора — по названию.
func (r *Repository) Groups(ctx context.Context) ([]GroupRef, error) {
	return database.Collect(ctx, r.db, func(row pgx.Row) (GroupRef, error) {
		var g GroupRef
		err := row.Scan(&g.ID, &g.Name)
		return g, err
	}, `SELECT id, name FROM study_group ORDER BY admission_year DESC, name`)
}

// Teachers — активные преподаватели: id, ФИО и кафедра. Email и прочие
// личные данные намеренно не выбираются: список публичный.
// Отключённые учётки скрыты из списка, но их занятия в расписании остаются.
func (r *Repository) Teachers(ctx context.Context) ([]TeacherRef, error) {
	return database.Collect(ctx, r.db, func(row pgx.Row) (TeacherRef, error) {
		var t TeacherRef
		err := row.Scan(&t.ID, &t.FullName, &t.DepartmentName)
		return t, err
	}, `
		SELECT t.user_id, `+database.FullName("u")+`, d.name
		FROM teacher t
		JOIN app_user u   ON u.id = t.user_id
		JOIN department d ON d.id = t.department_id
		WHERE u.is_active
		ORDER BY u.last_name, u.first_name, u.id`)
}
