package schedule

import (
	"context"
	"errors"
	"slices"
	"strings"

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
		return errBadReference("workload_id", "workload does not exist")
	case "lesson_room_id_fkey":
		return errBadReference("room_id", "room does not exist")
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
	SELECT l.id, l.workload_id, c.id, c.semester_id, s.name, w.lesson_type::text, w.teacher_id, ` + database.FullName("u") + `,
	       c.group_id, g.name, l.week_number, l.day_of_week, to_char(l.starts_at, 'HH24:MI'), to_char(l.ends_at, 'HH24:MI'),
	       l.room_id, rm.number
	FROM lesson l
	JOIN workload w        ON w.id = l.workload_id
	JOIN course_offering c ON c.id = w.course_id
	JOIN subject s         ON s.id = c.subject_id
	JOIN study_group g     ON g.id = c.group_id
	JOIN app_user u        ON u.id = w.teacher_id
	JOIN room rm           ON rm.id = l.room_id`

// scanLesson читает строку lessonSelect.
func scanLesson(row pgx.Row) (Lesson, error) {
	var l Lesson
	err := row.Scan(&l.ID, &l.WorkloadID, &l.CourseID, &l.SemesterID, &l.Subject, &l.LessonType, &l.TeacherID, &l.TeacherName,
		&l.GroupID, &l.GroupName, &l.WeekNumber, &l.DayOfWeek, &l.StartsAt, &l.EndsAt, &l.RoomID, &l.Room)
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
// неделя → день → время → id. После развёртки по датам это даёт порядок date, starts_at, id.
//
// SemesterID обязателен: текущий семестр подставляет сервис.
// StudentID — «группа этого студента»: подзапрос находит её по user_id.
// Если такого студента нет, подзапрос даёт NULL и список будет пустым.
func (r *Repository) List(ctx context.Context, f LessonFilter) ([]Lesson, error) {
	var w database.Filter
	if f.SemesterID != nil {
		w.Add("c.semester_id = ?", *f.SemesterID)
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
	if f.RoomID != nil {
		w.Add("l.room_id = ?", *f.RoomID)
	}
	if f.Week != nil {
		w.Add("l.week_number = ?", *f.Week)
	}
	if f.Day != nil {
		w.Add("l.day_of_week = ?", *f.Day)
	}
	return database.Collect(ctx, r.db, scanLesson,
		lessonSelect+w.Where()+` ORDER BY l.week_number, l.day_of_week, l.starts_at, l.id`, w.Args()...)
}

// slot — всё, что нужно для проверки пересечений: ресурсы занятия и его интервал.
type slot struct {
	semesterID, groupID, teacherID, roomID int64
	in                                     LessonInput
}

// workloadSlot находит группу, преподавателя и семестр строки нагрузки.
func workloadSlot(ctx context.Context, q pgx.Tx, workloadID int64, in LessonInput) (slot, error) {
	sl := slot{roomID: in.RoomID, in: in}
	err := q.QueryRow(ctx, `
		SELECT c.semester_id, c.group_id, w.teacher_id
		FROM workload w JOIN course_offering c ON c.id = w.course_id
		WHERE w.id = $1`, workloadID).Scan(&sl.semesterID, &sl.groupID, &sl.teacherID)
	if errors.Is(err, pgx.ErrNoRows) {
		return slot{}, errBadReference("workload_id", "workload does not exist")
	}
	return sl, database.Translate(err)
}

// errBadReference — ссылка на несуществующую или архивную запись (400 invalid_reference).
func errBadReference(field, problem string) error {
	return apperr.NewInvalid("invalid_reference", "referenced record does not exist or is archived").
		WithDetails(map[string]any{"fields": map[string][]string{field: {problem}}})
}

// checkRoom: аудитория существует и действует. keep — аудитория, в которой занятие
// уже стоит: оставить её можно и после архивации.
func checkRoom(ctx context.Context, tx pgx.Tx, roomID int64, keep *int64) error {
	var archived bool
	err := tx.QueryRow(ctx, `SELECT archived FROM room WHERE id = $1 FOR SHARE`, roomID).Scan(&archived)
	if errors.Is(err, pgx.ErrNoRows) {
		return errBadReference("room_id", "room does not exist")
	}
	if err != nil {
		return database.Translate(err)
	}
	if archived && (keep == nil || *keep != roomID) {
		return errBadReference("room_id", "room is archived")
	}
	return nil
}

// lockResources берёт транзакционные advisory-блокировки на группу, преподавателя
// и аудиторию в стабильном порядке (по возрастанию ключа). Два параллельных сохранения,
// задевающих общий ресурс, выполняются по очереди, и проверка пересечений не пропустит
// коллизию; одинаковый порядок исключает взаимную блокировку.
func lockResources(ctx context.Context, tx pgx.Tx, sl slot) error {
	// Ключ: класс ресурса в старших битах, id — в младших.
	keys := []int64{1<<40 | sl.groupID, 2<<40 | sl.teacherID, 3<<40 | sl.roomID}
	slices.Sort(keys)
	for _, k := range keys {
		if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock($1)`, k); err != nil {
			return database.Translate(err)
		}
	}
	return nil
}

// conflicts ищет занятия того же семестра, недели и дня, чей интервал [starts_at, ends_at)
// пересекается с новым и которые занимают ту же группу, преподавателя или аудиторию.
// except — редактируемое занятие, оно в сравнении не участвует.
func conflicts(ctx context.Context, q pgx.Tx, sl slot, except *int64) ([]ScheduleConflict, error) {
	rows, err := q.Query(ctx, `
		SELECT l.id, l.week_number, l.day_of_week, to_char(l.starts_at, 'HH24:MI'), to_char(l.ends_at, 'HH24:MI'),
		       c.group_id = $5, w.teacher_id = $6, l.room_id = $7
		FROM lesson l
		JOIN workload w        ON w.id = l.workload_id
		JOIN course_offering c ON c.id = w.course_id
		WHERE c.semester_id = $1 AND l.week_number = $2 AND l.day_of_week = $3
		  AND l.starts_at < $4::time AND l.ends_at > $8::time
		  AND (c.group_id = $5 OR w.teacher_id = $6 OR l.room_id = $7)
		  AND ($9::int IS NULL OR l.id <> $9)
		ORDER BY l.starts_at, l.id`,
		sl.semesterID, sl.in.WeekNumber, sl.in.DayOfWeek, sl.in.EndsAt,
		sl.groupID, sl.teacherID, sl.roomID, sl.in.StartsAt, except)
	if err != nil {
		return nil, database.Translate(err)
	}
	defer rows.Close()
	out := []ScheduleConflict{}
	for rows.Next() {
		var c ScheduleConflict
		var group, teacher, room bool
		if err := rows.Scan(&c.LessonID, &c.WeekNumber, &c.DayOfWeek, &c.StartsAt, &c.EndsAt, &group, &teacher, &room); err != nil {
			return nil, database.Translate(err)
		}
		c.Resources = []string{}
		for _, r := range []struct {
			on   bool
			name string
		}{{group, "group"}, {teacher, "teacher"}, {room, "room"}} {
			if r.on {
				c.Resources = append(c.Resources, r.name)
			}
		}
		out = append(out, c)
	}
	return out, database.Translate(rows.Err())
}

// errConflict — 409 schedule_conflict со списком пересечений в details.conflicts.
func errConflict(list []ScheduleConflict) error {
	return apperr.NewConflict("schedule_conflict", "the lesson overlaps other lessons").
		WithDetails(map[string]any{"conflicts": list})
}

// CheckConflicts — предварительная проверка без записи (LessonConflictInput).
// Окончательная проверка всё равно выполняется в транзакции сохранения.
// keepRoom — аудитория редактируемого занятия: её можно оставить и после архивации.
func (r *Repository) CheckConflicts(ctx context.Context, workloadID int64, in LessonInput, except, keepRoom *int64) ([]ScheduleConflict, error) {
	var out []ScheduleConflict
	err := database.InTx(ctx, r.db, func(tx pgx.Tx) error {
		sl, err := workloadSlot(ctx, tx, workloadID, in)
		if err != nil {
			return err
		}
		if err := checkRoom(ctx, tx, in.RoomID, keepRoom); err != nil {
			return err
		}
		out, err = conflicts(ctx, tx, sl, except)
		return err
	})
	return out, err
}

// Create ставит занятие в сетку. Внутри одной транзакции: нагрузка и действующая
// аудитория, блокировки ресурсов, проверка пересечений, вставка.
func (r *Repository) Create(ctx context.Context, workloadID int64, in LessonInput) (int64, error) {
	var id int64
	err := database.InTx(ctx, r.db, func(tx pgx.Tx) error {
		sl, err := workloadSlot(ctx, tx, workloadID, in)
		if err != nil {
			return err
		}
		if err := checkRoom(ctx, tx, in.RoomID, nil); err != nil {
			return err
		}
		if err := lockResources(ctx, tx, sl); err != nil {
			return err
		}
		list, err := conflicts(ctx, tx, sl, nil)
		if err != nil {
			return err
		}
		if len(list) > 0 {
			return errConflict(list)
		}
		err = tx.QueryRow(ctx, `
			INSERT INTO lesson (workload_id, week_number, day_of_week, starts_at, ends_at, room_id)
			VALUES ($1, $2, $3, $4::time, $5::time, $6) RETURNING id`,
			workloadID, in.WeekNumber, in.DayOfWeek, in.StartsAt, in.EndsAt, in.RoomID).Scan(&id)
		return lessonErr(err)
	})
	return id, err
}

// Stored — сохранённые нагрузка, интервал и аудитория занятия (для правил редактирования).
type Stored struct {
	WorkloadID       int64
	StartsAt, EndsAt string
	RoomID           int64
}

// Stored возвращает сохранённые поля занятия; нет занятия — 404 lesson_not_found.
func (r *Repository) Stored(ctx context.Context, id int64) (Stored, error) {
	var s Stored
	err := r.db.QueryRow(ctx, `
		SELECT workload_id, to_char(starts_at, 'HH24:MI'), to_char(ends_at, 'HH24:MI'), room_id
		FROM lesson WHERE id = $1`, id).Scan(&s.WorkloadID, &s.StartsAt, &s.EndsAt, &s.RoomID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Stored{}, apperr.NewNotFound("lesson_not_found")
	}
	return s, database.Translate(err)
}

// Update переносит занятие (LessonUpdateInput): слот, интервал и аудиторию.
// Нагрузка не меняется. Проверки — как при создании, само занятие из сравнения исключается.
func (r *Repository) Update(ctx context.Context, id int64, in LessonInput) error {
	return database.InTx(ctx, r.db, func(tx pgx.Tx) error {
		var workloadID, roomID int64
		err := tx.QueryRow(ctx, `SELECT workload_id, room_id FROM lesson WHERE id = $1 FOR UPDATE`, id).Scan(&workloadID, &roomID)
		if errors.Is(err, pgx.ErrNoRows) {
			return apperr.NewNotFound("lesson_not_found")
		}
		if err != nil {
			return database.Translate(err)
		}
		sl, err := workloadSlot(ctx, tx, workloadID, in)
		if err != nil {
			return err
		}
		if err := checkRoom(ctx, tx, in.RoomID, &roomID); err != nil {
			return err
		}
		if err := lockResources(ctx, tx, sl); err != nil {
			return err
		}
		list, err := conflicts(ctx, tx, sl, &id)
		if err != nil {
			return err
		}
		if len(list) > 0 {
			return errConflict(list)
		}
		_, err = tx.Exec(ctx, `
			UPDATE lesson SET week_number = $2, day_of_week = $3, starts_at = $4::time, ends_at = $5::time, room_id = $6
			WHERE id = $1`, id, in.WeekNumber, in.DayOfWeek, in.StartsAt, in.EndsAt, in.RoomID)
		return lessonErr(err)
	})
}

// Delete удаляет занятие. Если на него ссылаются другие данные
// (например, посещаемость), AffectedDelete вернёт ошибку «используется».
func (r *Repository) Delete(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM lesson WHERE id = $1`, id)
	return database.AffectedDelete(tag, err, "lesson_not_found")
}

// ---- проверки существования для маршрутов расписания ----

// CurrentSemester — id текущего семестра; ok = false, если он не выбран.
func (r *Repository) CurrentSemester(ctx context.Context) (int64, bool, error) {
	var id int64
	err := r.db.QueryRow(ctx, `SELECT id FROM semester WHERE is_current`).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, false, nil
	}
	return id, err == nil, database.Translate(err)
}

// exists — 404 с кодом notFound, если запрос не вернул строк.
func (r *Repository) exists(ctx context.Context, notFound, sql string, id int64) error {
	var ok bool
	if err := r.db.QueryRow(ctx, `SELECT EXISTS (`+sql+`)`, id).Scan(&ok); err != nil {
		return database.Translate(err)
	}
	if !ok {
		return apperr.NewNotFound(notFound)
	}
	return nil
}

// SemesterExists, GroupExists, TeacherExists: архивные записи тоже существуют —
// их прошлое расписание остаётся доступным по прямой ссылке.
func (r *Repository) SemesterExists(ctx context.Context, id int64) error {
	return r.exists(ctx, "semester_not_found", `SELECT 1 FROM semester WHERE id = $1`, id)
}
func (r *Repository) GroupExists(ctx context.Context, id int64) error {
	return r.exists(ctx, "group_not_found", `SELECT 1 FROM study_group WHERE id = $1`, id)
}
func (r *Repository) TeacherExists(ctx context.Context, id int64) error {
	return r.exists(ctx, "teacher_not_found", `SELECT 1 FROM teacher WHERE user_id = $1`, id)
}

// ---- справочники для выбора в расписании (доступны гостям) ----
// Возвращают только то, что безопасно показать без входа: id, название или ФИО, кафедра.
// Поиск — серверный (q, ILIKE по названию/ФИО), страница — limit/offset; total учитывает поиск.

// likeEscape экранирует спецсимволы LIKE, чтобы «_» и «%» в запросе искались буквально.
func likeEscape(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}

// refPage выполняет запрос страницы справочника; последняя колонка — count(*) OVER ().
func (r *Repository) refPage(ctx context.Context, sql string, q RefQuery, withDept bool) ([]ScheduleRef, int, error) {
	rows, err := r.db.Query(ctx, sql, "%"+likeEscape(q.Q)+"%", q.Limit, q.Offset)
	if err != nil {
		return nil, 0, database.Translate(err)
	}
	defer rows.Close()
	out, total := []ScheduleRef{}, 0
	for rows.Next() {
		var ref ScheduleRef
		dest := []any{&ref.ID, &ref.Name}
		if withDept {
			dest = append(dest, &ref.DepartmentName)
		}
		if err := rows.Scan(append(dest, &total)...); err != nil {
			return nil, 0, database.Translate(err)
		}
		out = append(out, ref)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, database.Translate(err)
	}
	if len(out) == 0 && q.Offset > 0 {
		// Страница за концом списка: total всё равно нужен.
		err = r.db.QueryRow(ctx, `SELECT count(*) FROM (`+strings.Split(sql, "ORDER BY")[0]+`) t`,
			"%"+likeEscape(q.Q)+"%").Scan(&total)
	}
	return out, total, database.Translate(err)
}

// Groups — действующие группы: сначала младшие наборы, внутри набора — по названию.
// department_name у групп всегда null.
func (r *Repository) Groups(ctx context.Context, q RefQuery) ([]ScheduleRef, int, error) {
	return r.refPage(ctx, `
		SELECT id, name, count(*) OVER ()
		FROM study_group
		WHERE NOT archived AND name ILIKE $1
		ORDER BY admission_year DESC, name, id
		LIMIT $2 OFFSET $3`, q, false)
}

// Teachers — активные преподаватели: id, ФИО и кафедра. Email, документы и данные
// доступа намеренно не выбираются: список публичный. Отключённые учётки скрыты
// из списка, но их занятия в расписании остаются.
func (r *Repository) Teachers(ctx context.Context, q RefQuery) ([]ScheduleRef, int, error) {
	return r.refPage(ctx, `
		SELECT t.user_id, `+database.FullName("u")+`, d.name, count(*) OVER ()
		FROM teacher t
		JOIN app_user u        ON u.id = t.user_id
		LEFT JOIN department d ON d.id = t.department_id
		WHERE u.is_active AND (`+database.FullName("u")+`) ILIKE $1
		ORDER BY u.last_name, u.first_name, u.id
		LIMIT $2 OFFSET $3`, q, true)
}
