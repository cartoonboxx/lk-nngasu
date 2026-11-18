package structure

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"university-api/internal/platform/apperr"
	"university-api/internal/platform/database"
)

// Repository — доступ к таблицам справочников. Ошибки БД переводятся в apperr:
// «нет строки» — в not found с кодом конкретной сущности, остальное —
// через database.Translate (уникальность, внешние ключи).
type Repository struct {
	db *pgxpool.Pool

	// Кафедры и дисциплины устроены одинаково (id, name) и обслуживаются одним
	// кодом catalog; отличаются таблицей, длиной названия и зависимостями.
	departments catalog
	subjects    catalog
}

func newRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
		// Кафедру нельзя удалить, пока к ней привязан хоть один преподаватель,
		// в том числе заблокированный: блокировка аккаунта учебные связи не снимает.
		departments: catalog{db: db, table: "department", notFound: "department_not_found", maxName: 150,
			usedBy: `SELECT 1 FROM teacher WHERE department_id = $1`, what: "teachers of this department"},
		// Дисциплину нельзя удалить, пока есть хоть один курс (course_offering) с ней.
		subjects: catalog{db: db, table: "subject", notFound: "subject_not_found", maxName: 200,
			usedBy: `SELECT 1 FROM course_offering WHERE subject_id = $1`, what: "courses with this subject"},
	}
}

// errEntryExists — дубликат названия (или кода направления, пары год/номер семестра).
var errEntryExists = apperr.NewConflict("catalog_entry_exists", "a record with the same unique value (name, code or year and number) already exists")

// writeErr переводит ошибки записи справочника: нарушение уникальности — 409 catalog_entry_exists.
func writeErr(err error) error {
	if database.PgCode(err) == database.CodeUnique {
		return errEntryExists
	}
	return database.Translate(err)
}

// errInUse — запись нельзя удалить: на неё ссылаются другие данные.
func errInUse(what string) error {
	return apperr.NewConflict("catalog_in_use", "the record is still used by "+what)
}

// pageQuery — общий список справочника: поиск q по searchCols без учёта регистра,
// total с учётом всех фильтров и страница LIMIT/OFFSET.
// Колонки и порядок — константы из кода, значения идут параметрами.
func pageQuery[T any](ctx context.Context, db *pgxpool.Pool, scan func(pgx.Row) (T, error),
	selectSQL string, w *database.Filter, searchCols []string, lq ListQuery, order string) ([]T, int, error) {
	if lq.Q != "" {
		conds := make([]string, len(searchCols))
		for i, c := range searchCols {
			conds[i] = c + " ILIKE ?"
		}
		w.Add("("+strings.Join(conds, " OR ")+")", "%"+likeEscape(lq.Q)+"%")
	}
	var total int
	countArgs := append([]any(nil), w.Args()...)
	if err := db.QueryRow(ctx, `SELECT count(*) FROM (`+selectSQL+w.Where()+`) x`, countArgs...).Scan(&total); err != nil {
		return nil, 0, database.Translate(err)
	}
	q := selectSQL + w.Where() + ` ORDER BY ` + order + ` LIMIT ` + w.Arg(lq.Limit) + ` OFFSET ` + w.Arg(lq.Offset)
	list, err := database.Collect(ctx, db, scan, q, w.Args()...)
	return list, total, err
}

// likeEscape экранирует спецсимволы шаблона LIKE, чтобы «_» и «%» искались буквально.
func likeEscape(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}

// deleteTx — физическое удаление: блокировка строки, проверка зависимостей и DELETE
// одной транзакцией. Ссылки, которые check не учёл, всё равно не дадут удалить
// (внешний ключ) — это тоже 409 catalog_in_use.
func deleteTx(ctx context.Context, db *pgxpool.Pool, table string, id int64, notFound string,
	check func(tx pgx.Tx) error) error {
	return database.InTx(ctx, db, func(tx pgx.Tx) error {
		var found bool
		err := tx.QueryRow(ctx, `SELECT true FROM `+table+` WHERE id = $1 FOR UPDATE`, id).Scan(&found)
		if errors.Is(err, pgx.ErrNoRows) {
			return apperr.NewNotFound(notFound)
		}
		if err != nil {
			return database.Translate(err)
		}
		if check != nil {
			if err := check(tx); err != nil {
				return err
			}
		}
		if _, err := tx.Exec(ctx, `DELETE FROM `+table+` WHERE id = $1`, id); err != nil {
			if database.PgCode(err) == database.CodeFK {
				return errInUse("other records")
			}
			return database.Translate(err)
		}
		return nil
	})
}

// usedBy — проверка зависимостей: если запрос находит строку, запись используется.
func usedBy(ctx context.Context, sql, what string, id int64) func(tx pgx.Tx) error {
	return func(tx pgx.Tx) error {
		var used bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS (`+sql+`)`, id).Scan(&used); err != nil {
			return database.Translate(err)
		}
		if used {
			return errInUse(what)
		}
		return nil
	}
}

// errBadReference — ссылка на несуществующую запись.
func errBadReference(field, problem string) error {
	return apperr.NewInvalid("invalid_reference", "referenced record does not exist").
		WithDetails(map[string]any{"fields": map[string][]string{field: {problem}}})
}

// ---- catalog: кафедры и дисциплины ----

// catalog — CRUD для одной таблицы вида (id, name).
// Имя таблицы подставляется в SQL строкой (параметром его не передать),
// поэтому оно задаётся только константой в newRepository и никогда не берётся из запроса.
type catalog struct {
	db       *pgxpool.Pool
	table    string // константа из кода, не пользовательский ввод
	notFound string // код ошибки, когда запись с таким id не найдена
	maxName  int    // длина названия: кафедра до 150, дисциплина до 200 символов
	usedBy   string // запрос зависимостей ($1 — id записи)
	what     string // что мешает удалению — для текста ошибки
}

func scanNamed(row pgx.Row) (Named, error) {
	var n Named
	err := row.Scan(&n.ID, &n.Name)
	return n, err
}

// list — страница записей по ListQuery: name ASC, id ASC.
func (c catalog) list(ctx context.Context, lq ListQuery) ([]Named, int, error) {
	var w database.Filter
	return pageQuery(ctx, c.db, scanNamed, `SELECT id, name FROM `+c.table, &w, []string{"name"}, lq, "name, id")
}

// create добавляет запись. Дубликат названия без учёта регистра — 409 catalog_entry_exists.
func (c catalog) create(ctx context.Context, name string) (Named, error) {
	n, err := scanNamed(c.db.QueryRow(ctx, `INSERT INTO `+c.table+` (name) VALUES ($1) RETURNING id, name`, name))
	return n, writeErr(err)
}

// update переименовывает запись (полная замена NamedInput); ссылки на неё сохраняются.
func (c catalog) update(ctx context.Context, id int64, name string) (Named, error) {
	n, err := scanNamed(c.db.QueryRow(ctx,
		`UPDATE `+c.table+` SET name = $2 WHERE id = $1 RETURNING id, name`, id, name))
	if errors.Is(err, pgx.ErrNoRows) {
		return Named{}, apperr.NewNotFound(c.notFound)
	}
	return n, writeErr(err)
}

// remove удаляет запись без зависимостей; иначе 409 catalog_in_use.
func (c catalog) remove(ctx context.Context, id int64) error {
	return deleteTx(ctx, c.db, c.table, id, c.notFound, usedBy(ctx, c.usedBy, c.what, id))
}

// ---- specializations ----

const specializationSelect = `SELECT id, name, code FROM specialization`

func scanSpecialization(row pgx.Row) (Specialization, error) {
	var s Specialization
	err := row.Scan(&s.ID, &s.Name, &s.Code)
	return s, err
}

// ListSpecializations — страница направлений: name ASC, id ASC; q ищет и по коду.
func (r *Repository) ListSpecializations(ctx context.Context, lq ListQuery) ([]Specialization, int, error) {
	var w database.Filter
	return pageQuery(ctx, r.db, scanSpecialization, specializationSelect, &w, []string{"name", "code"}, lq, "name, id")
}

func (r *Repository) CreateSpecialization(ctx context.Context, code, name string) (Specialization, error) {
	s, err := scanSpecialization(r.db.QueryRow(ctx,
		`INSERT INTO specialization (code, name) VALUES ($1, $2) RETURNING id, name, code`, code, name))
	return s, writeErr(err)
}

func (r *Repository) UpdateSpecialization(ctx context.Context, id int64, code, name string) (Specialization, error) {
	s, err := scanSpecialization(r.db.QueryRow(ctx,
		`UPDATE specialization SET code = $2, name = $3 WHERE id = $1 RETURNING id, name, code`, id, code, name))
	if errors.Is(err, pgx.ErrNoRows) {
		return Specialization{}, apperr.NewNotFound("specialization_not_found")
	}
	return s, writeErr(err)
}

// DeleteSpecialization — направление удаляется, только если у него нет учебных групп.
func (r *Repository) DeleteSpecialization(ctx context.Context, id int64) error {
	return deleteTx(ctx, r.db, "specialization", id, "specialization_not_found",
		usedBy(ctx, `SELECT 1 FROM study_group WHERE specialization_id = $1`, "groups of this specialization", id))
}

// ---- groups ----

// groupSelect — группа вместе с названием и кодом направления. Порядок колонок — как в scanGroup.
const groupSelect = `
	SELECT g.id, g.name, g.specialization_id, sp.name, sp.code, g.admission_year
	FROM study_group g
	JOIN specialization sp ON sp.id = g.specialization_id`

func scanGroup(row pgx.Row) (Group, error) {
	var g Group
	err := row.Scan(&g.ID, &g.Name, &g.SpecializationID, &g.SpecializationName,
		&g.SpecializationCode, &g.AdmissionYear)
	return g, err
}

func (r *Repository) GetGroup(ctx context.Context, id int64) (Group, error) {
	g, err := scanGroup(r.db.QueryRow(ctx, groupSelect+` WHERE g.id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Group{}, apperr.NewNotFound("group_not_found")
	}
	return g, database.Translate(err)
}

// ListGroups — страница групп по фильтру: name ASC, id ASC.
func (r *Repository) ListGroups(ctx context.Context, f GroupFilter) ([]Group, int, error) {
	var w database.Filter
	if f.SpecializationID != nil {
		w.Add("g.specialization_id = ?", *f.SpecializationID)
	}
	if f.AdmissionYear != nil {
		w.Add("g.admission_year = ?", *f.AdmissionYear)
	}
	return pageQuery(ctx, r.db, scanGroup, groupSelect, &w, []string{"g.name"}, f.ListQuery, "g.name, g.id")
}

// checkSpecialization: направление существует. FOR SHARE не даёт удалить его,
// пока идёт запись группы.
func checkSpecialization(ctx context.Context, tx pgx.Tx, id int64) error {
	var found bool
	err := tx.QueryRow(ctx, `SELECT true FROM specialization WHERE id = $1 FOR SHARE`, id).Scan(&found)
	if errors.Is(err, pgx.ErrNoRows) {
		return errBadReference("specialization_id", "specialization does not exist")
	}
	return database.Translate(err)
}

// CreateGroup создаёт группу и возвращает её id.
func (r *Repository) CreateGroup(ctx context.Context, in GroupInput) (int64, error) {
	var id int64
	err := database.InTx(ctx, r.db, func(tx pgx.Tx) error {
		if err := checkSpecialization(ctx, tx, in.SpecializationID); err != nil {
			return err
		}
		err := tx.QueryRow(ctx, `
			INSERT INTO study_group (name, specialization_id, admission_year) VALUES ($1, $2, $3) RETURNING id`,
			in.Name, in.SpecializationID, in.AdmissionYear).Scan(&id)
		return writeErr(err)
	})
	return id, err
}

// UpdateGroup заменяет все поля группы; студенты и курсы остаются у неё.
func (r *Repository) UpdateGroup(ctx context.Context, id int64, in GroupInput) error {
	return database.InTx(ctx, r.db, func(tx pgx.Tx) error {
		var found bool
		err := tx.QueryRow(ctx, `SELECT true FROM study_group WHERE id = $1 FOR UPDATE`, id).Scan(&found)
		if errors.Is(err, pgx.ErrNoRows) {
			return apperr.NewNotFound("group_not_found")
		}
		if err != nil {
			return database.Translate(err)
		}
		if err := checkSpecialization(ctx, tx, in.SpecializationID); err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `
			UPDATE study_group SET name = $2, specialization_id = $3, admission_year = $4 WHERE id = $1`,
			id, in.Name, in.SpecializationID, in.AdmissionYear)
		return writeErr(err)
	})
}

// DeleteGroup — группа удаляется, только если в ней нет студентов (в том числе
// заблокированных) и на неё не назначен ни один курс.
func (r *Repository) DeleteGroup(ctx context.Context, id int64) error {
	return deleteTx(ctx, r.db, "study_group", id, "group_not_found",
		usedBy(ctx, `SELECT 1 FROM student WHERE group_id = $1
			UNION ALL SELECT 1 FROM course_offering WHERE group_id = $1`, "students or courses of this group", id))
}

// ---- semesters ----

// Название семестра — вычисляемая колонка БД («2026/27 · Осенний», миграция 016).
const semesterSelect = `SELECT id, name, academic_year_start, number, is_current FROM semester`

func scanSemester(row pgx.Row) (Semester, error) {
	var s Semester
	err := row.Scan(&s.ID, &s.Name, &s.AcademicYearStart, &s.Number, &s.IsCurrent)
	return s, err
}

// ListSemesters — страница семестров от новых к старым: academic_year_start DESC, number DESC, id ASC.
func (r *Repository) ListSemesters(ctx context.Context, lq ListQuery) ([]Semester, int, error) {
	var w database.Filter
	return pageQuery(ctx, r.db, scanSemester, semesterSelect, &w, []string{"name"}, lq,
		"academic_year_start DESC, number DESC, id")
}

func (r *Repository) GetSemester(ctx context.Context, id int64) (Semester, error) {
	s, err := scanSemester(r.db.QueryRow(ctx, semesterSelect+` WHERE id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Semester{}, apperr.NewNotFound("semester_not_found")
	}
	return s, database.Translate(err)
}

// CurrentSemester — текущий семестр; если он не выбран — 404 current_semester_not_set.
func (r *Repository) CurrentSemester(ctx context.Context) (Semester, error) {
	s, err := scanSemester(r.db.QueryRow(ctx, semesterSelect+` WHERE is_current`))
	if errors.Is(err, pgx.ErrNoRows) {
		return Semester{}, apperr.NewNotFound("current_semester_not_set")
	}
	return s, database.Translate(err)
}

// CreateSemester создаёт семестр; текущим он не становится (is_current по умолчанию false).
// Пара год/номер уникальна — дубликат вернётся как конфликт.
func (r *Repository) CreateSemester(ctx context.Context, in SemesterInput) (int64, error) {
	var id int64
	err := r.db.QueryRow(ctx,
		`INSERT INTO semester (academic_year_start, number) VALUES ($1, $2) RETURNING id`,
		in.AcademicYearStart, in.Number).Scan(&id)
	return id, writeErr(err)
}

// UpdateSemester заменяет год и номер; флаг «текущий» не трогает.
func (r *Repository) UpdateSemester(ctx context.Context, id int64, in SemesterInput) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE semester SET academic_year_start = $2, number = $3 WHERE id = $1`,
		id, in.AcademicYearStart, in.Number)
	if err != nil {
		return writeErr(err)
	}
	return database.Affected(tag, nil, "semester_not_found")
}

// DeleteSemester — нельзя удалить текущий семестр (409 current_semester_in_use)
// и семестр, на который назначены курсы (409 catalog_in_use).
func (r *Repository) DeleteSemester(ctx context.Context, id int64) error {
	return deleteTx(ctx, r.db, "semester", id, "semester_not_found", func(tx pgx.Tx) error {
		var current bool
		if err := tx.QueryRow(ctx, `SELECT is_current FROM semester WHERE id = $1`, id).Scan(&current); err != nil {
			return database.Translate(err)
		}
		if current {
			return apperr.NewConflict("current_semester_in_use", "the current semester cannot be deleted")
		}
		return usedBy(ctx, `SELECT 1 FROM course_offering WHERE semester_id = $1`, "courses of this semester", id)(tx)
	})
}

// SetCurrentSemester делает семестр текущим, снимая флаг с прежнего (одной транзакцией).
func (r *Repository) SetCurrentSemester(ctx context.Context, id int64) error {
	return database.InTx(ctx, r.db, func(tx pgx.Tx) error {
		var found bool
		err := tx.QueryRow(ctx, `SELECT true FROM semester WHERE id = $1 FOR UPDATE`, id).Scan(&found)
		if errors.Is(err, pgx.ErrNoRows) {
			return apperr.NewNotFound("semester_not_found")
		}
		if err != nil {
			return database.Translate(err)
		}
		if _, err := tx.Exec(ctx, `UPDATE semester SET is_current = false WHERE is_current AND id <> $1`, id); err != nil {
			return database.Translate(err)
		}
		_, err = tx.Exec(ctx, `UPDATE semester SET is_current = true WHERE id = $1`, id)
		return database.Translate(err)
	})
}

// ---- rooms ----

func scanRoom(row pgx.Row) (Room, error) {
	var rm Room
	err := row.Scan(&rm.ID, &rm.Number)
	return rm, err
}

// ListRooms — страница аудиторий: number ASC, id ASC.
func (r *Repository) ListRooms(ctx context.Context, lq ListQuery) ([]Room, int, error) {
	var w database.Filter
	return pageQuery(ctx, r.db, scanRoom, `SELECT id, number FROM room`, &w, []string{"number"}, lq, "number, id")
}

func (r *Repository) CreateRoom(ctx context.Context, number string) (Room, error) {
	rm, err := scanRoom(r.db.QueryRow(ctx, `INSERT INTO room (number) VALUES ($1) RETURNING id, number`, number))
	return rm, writeErr(err)
}

func (r *Repository) UpdateRoom(ctx context.Context, id int64, number string) (Room, error) {
	rm, err := scanRoom(r.db.QueryRow(ctx, `UPDATE room SET number = $2 WHERE id = $1 RETURNING id, number`, id, number))
	if errors.Is(err, pgx.ErrNoRows) {
		return Room{}, apperr.NewNotFound("room_not_found")
	}
	return rm, writeErr(err)
}

// DeleteRoom — аудитория удаляется, только если в ней нет ни одного занятия.
func (r *Repository) DeleteRoom(ctx context.Context, id int64) error {
	return deleteTx(ctx, r.db, "room", id, "room_not_found",
		usedBy(ctx, `SELECT 1 FROM lesson WHERE room_id = $1`, "lessons in this room", id))
}
