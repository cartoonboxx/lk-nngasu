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

	// Кафедры и дисциплины устроены одинаково (id, name, archived) и
	// обслуживаются одним кодом catalog; отличаются таблицей и кодом ошибки.
	departments catalog
	subjects    catalog
}

func newRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
		// Кафедру нельзя заархивировать, пока к ней привязан хоть один преподаватель,
		// в том числе заблокированный: блокировка аккаунта учебные связи не снимает.
		departments: catalog{db: db, table: "department", notFound: "department_not_found",
			inUse: func(ctx context.Context, tx pgx.Tx, id int64) error {
				used, err := exists(ctx, tx, `SELECT 1 FROM teacher WHERE department_id = $1`, id)
				if err == nil && used {
					err = errInUse("teachers of this department")
				}
				return err
			}},
		// Дисциплину нельзя заархивировать, пока есть хоть один курс (course_offering) с ней.
		subjects: catalog{db: db, table: "subject", notFound: "subject_not_found",
			inUse: func(ctx context.Context, tx pgx.Tx, id int64) error {
				used, err := exists(ctx, tx, `SELECT 1 FROM course_offering WHERE subject_id = $1`, id)
				if err == nil && used {
					err = errInUse("courses with this subject")
				}
				return err
			}},
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

// errInUse — запись нельзя заархивировать: на неё опираются действующие данные.
func errInUse(what string) error {
	return apperr.NewConflict("catalog_in_use", "the record is still used by "+what)
}

// pageQuery — общий список справочника: условие по архиву, поиск q по searchCols
// без учёта регистра, total с учётом всех фильтров и страница LIMIT/OFFSET.
// Колонки и порядок — константы из кода, значения идут параметрами.
func pageQuery[T any](ctx context.Context, db *pgxpool.Pool, scan func(pgx.Row) (T, error),
	selectSQL string, w *database.Filter, archivedCol string, searchCols []string, lq ListQuery, order string) ([]T, int, error) {
	if lq.Archived != nil {
		w.Add(archivedCol+" = ?", *lq.Archived)
	}
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

// archiveTx — общая схема архивации: блокировка строки, проверка зависимостей
// (check, может быть nil) и пометка — одной транзакцией. Повтор над уже архивной
// записью ничего не проверяет и не меняет (ответ 200 с той же записью).
func archiveTx(ctx context.Context, db *pgxpool.Pool, table string, id int64, notFound string,
	check func(tx pgx.Tx) error) error {
	return database.InTx(ctx, db, func(tx pgx.Tx) error {
		var archived bool
		err := tx.QueryRow(ctx, `SELECT archived FROM `+table+` WHERE id = $1 FOR UPDATE`, id).Scan(&archived)
		if errors.Is(err, pgx.ErrNoRows) {
			return apperr.NewNotFound(notFound)
		}
		if err != nil {
			return database.Translate(err)
		}
		if archived {
			return nil
		}
		if check != nil {
			if err := check(tx); err != nil {
				return err
			}
		}
		_, err = tx.Exec(ctx, `UPDATE `+table+` SET archived = true WHERE id = $1`, id)
		return database.Translate(err)
	})
}

// exists — вспомогательный запрос «есть ли хоть одна строка».
func exists(ctx context.Context, tx pgx.Tx, sql string, args ...any) (bool, error) {
	var ok bool
	err := tx.QueryRow(ctx, `SELECT EXISTS (`+sql+`)`, args...).Scan(&ok)
	return ok, database.Translate(err)
}

// errBadReference — ссылка на несуществующую или архивную запись.
func errBadReference(field, problem string) error {
	return apperr.NewInvalid("invalid_reference", "referenced record does not exist or is archived").
		WithDetails(map[string]any{"fields": map[string][]string{field: {problem}}})
}

// ---- catalog: кафедры и дисциплины ----

// catalog — CRUD для одной таблицы вида (id, name, archived).
// Имя таблицы подставляется в SQL строкой (параметром его не передать),
// поэтому оно задаётся только константой в newRepository и никогда не берётся из запроса.
type catalog struct {
	db       *pgxpool.Pool
	table    string // константа из кода, не пользовательский ввод
	notFound string // код ошибки, когда запись с таким id не найдена
	// inUse — проверка зависимостей перед архивацией (nil — архивировать можно всегда).
	inUse func(ctx context.Context, tx pgx.Tx, id int64) error
}

func scanNamed(row pgx.Row) (Named, error) {
	var n Named
	err := row.Scan(&n.ID, &n.Name, &n.Archived)
	return n, err
}

// list — страница записей по ListQuery: name ASC, id ASC.
func (c catalog) list(ctx context.Context, lq ListQuery) ([]Named, int, error) {
	var w database.Filter
	return pageQuery(ctx, c.db, scanNamed, `SELECT id, name, archived FROM `+c.table, &w,
		"archived", []string{"name"}, lq, "name, id")
}

// create добавляет запись (archived = false задаёт БД). Дубликат названия без учёта
// регистра (уникальный индекс по lower(name)) — 409 catalog_entry_exists.
func (c catalog) create(ctx context.Context, name string) (Named, error) {
	n, err := scanNamed(c.db.QueryRow(ctx,
		`INSERT INTO `+c.table+` (name) VALUES ($1) RETURNING id, name, archived`, name))
	return n, writeErr(err)
}

// update переименовывает запись (полная замена NamedInput); ссылки на неё сохраняются.
func (c catalog) update(ctx context.Context, id int64, name string) (Named, error) {
	n, err := scanNamed(c.db.QueryRow(ctx,
		`UPDATE `+c.table+` SET name = $2 WHERE id = $1 RETURNING id, name, archived`, id, name))
	if errors.Is(err, pgx.ErrNoRows) {
		return Named{}, apperr.NewNotFound(c.notFound)
	}
	return n, writeErr(err)
}

// get возвращает запись по id.
func (c catalog) get(ctx context.Context, id int64) (Named, error) {
	n, err := scanNamed(c.db.QueryRow(ctx, `SELECT id, name, archived FROM `+c.table+` WHERE id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Named{}, apperr.NewNotFound(c.notFound)
	}
	return n, database.Translate(err)
}

// archive переводит запись в архив и возвращает её.
func (c catalog) archive(ctx context.Context, id int64) (Named, error) {
	var check func(tx pgx.Tx) error
	if c.inUse != nil {
		check = func(tx pgx.Tx) error { return c.inUse(ctx, tx, id) }
	}
	if err := archiveTx(ctx, c.db, c.table, id, c.notFound, check); err != nil {
		return Named{}, err
	}
	return c.get(ctx, id)
}

// remove удаляет запись. Если на неё ссылаются (кафедра — преподаватели,
// дисциплина — курсы), AffectedDelete вернёт ошибку «запись используется».
func (c catalog) remove(ctx context.Context, id int64) error {
	tag, err := c.db.Exec(ctx, `DELETE FROM `+c.table+` WHERE id = $1`, id)
	return database.AffectedDelete(tag, err, c.notFound)
}

// ---- specializations ----

const specializationSelect = `SELECT id, name, code, archived FROM specialization`

func scanSpecialization(row pgx.Row) (Specialization, error) {
	var s Specialization
	err := row.Scan(&s.ID, &s.Name, &s.Code, &s.Archived)
	return s, err
}

// ListSpecializations — страница направлений: name ASC, id ASC; q ищет и по коду.
func (r *Repository) ListSpecializations(ctx context.Context, lq ListQuery) ([]Specialization, int, error) {
	var w database.Filter
	return pageQuery(ctx, r.db, scanSpecialization, specializationSelect, &w,
		"archived", []string{"name", "code"}, lq, "name, id")
}

func (r *Repository) GetSpecialization(ctx context.Context, id int64) (Specialization, error) {
	s, err := scanSpecialization(r.db.QueryRow(ctx, specializationSelect+` WHERE id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Specialization{}, apperr.NewNotFound("specialization_not_found")
	}
	return s, database.Translate(err)
}

func (r *Repository) CreateSpecialization(ctx context.Context, code, name string) (Specialization, error) {
	s, err := scanSpecialization(r.db.QueryRow(ctx,
		`INSERT INTO specialization (code, name) VALUES ($1, $2) RETURNING id, name, code, archived`, code, name))
	return s, writeErr(err)
}

func (r *Repository) UpdateSpecialization(ctx context.Context, id int64, code, name string) (Specialization, error) {
	s, err := scanSpecialization(r.db.QueryRow(ctx,
		`UPDATE specialization SET code = $2, name = $3 WHERE id = $1 RETURNING id, name, code, archived`, id, code, name))
	if errors.Is(err, pgx.ErrNoRows) {
		return Specialization{}, apperr.NewNotFound("specialization_not_found")
	}
	return s, writeErr(err)
}

func (r *Repository) ArchiveSpecialization(ctx context.Context, id int64) (Specialization, error) {
	// Направление нельзя заархивировать, пока у него есть неархивная учебная группа.
	err := archiveTx(ctx, r.db, "specialization", id, "specialization_not_found", func(tx pgx.Tx) error {
		used, err := exists(ctx, tx, `SELECT 1 FROM study_group WHERE specialization_id = $1 AND NOT archived`, id)
		if err == nil && used {
			err = errInUse("active groups of this specialization")
		}
		return err
	})
	if err != nil {
		return Specialization{}, err
	}
	return r.GetSpecialization(ctx, id)
}

// DeleteSpecialization удаляет направление; если на него ссылаются группы — ошибка внешнего ключа.
func (r *Repository) DeleteSpecialization(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM specialization WHERE id = $1`, id)
	return database.AffectedDelete(tag, err, "specialization_not_found")
}

// ---- groups ----

// groupSelect — группа вместе с названием и кодом направления. Порядок колонок — как в scanGroup.
const groupSelect = `
	SELECT g.id, g.name, g.archived, g.specialization_id, sp.name, sp.code, g.admission_year
	FROM study_group g
	JOIN specialization sp ON sp.id = g.specialization_id`

func scanGroup(row pgx.Row) (Group, error) {
	var g Group
	err := row.Scan(&g.ID, &g.Name, &g.Archived, &g.SpecializationID, &g.SpecializationName,
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
	return pageQuery(ctx, r.db, scanGroup, groupSelect, &w, "g.archived", []string{"g.name"}, f.ListQuery, "g.name, g.id")
}

// checkSpecialization: направление существует и действует (не в архиве).
// keep — направление, которое уже стоит у группы: его оставить можно, даже если оно в архиве.
func checkSpecialization(ctx context.Context, tx pgx.Tx, id int64, keep *int64) error {
	var archived bool
	err := tx.QueryRow(ctx, `SELECT archived FROM specialization WHERE id = $1 FOR SHARE`, id).Scan(&archived)
	if errors.Is(err, pgx.ErrNoRows) {
		return errBadReference("specialization_id", "specialization does not exist")
	}
	if err != nil {
		return database.Translate(err)
	}
	if archived && (keep == nil || *keep != id) {
		return errBadReference("specialization_id", "specialization is archived")
	}
	return nil
}

// CreateGroup создаёт группу и возвращает её id. Направление должно быть действующим;
// FOR SHARE не даёт заархивировать его параллельно, пока идёт вставка.
func (r *Repository) CreateGroup(ctx context.Context, in GroupInput) (int64, error) {
	var id int64
	err := database.InTx(ctx, r.db, func(tx pgx.Tx) error {
		if err := checkSpecialization(ctx, tx, in.SpecializationID, nil); err != nil {
			return err
		}
		err := tx.QueryRow(ctx, `
			INSERT INTO study_group (name, specialization_id, admission_year) VALUES ($1, $2, $3) RETURNING id`,
			in.Name, in.SpecializationID, in.AdmissionYear).Scan(&id)
		return writeErr(err)
	})
	return id, err
}

// UpdateGroup заменяет все поля группы.
func (r *Repository) UpdateGroup(ctx context.Context, id int64, in GroupInput) error {
	return database.InTx(ctx, r.db, func(tx pgx.Tx) error {
		var current int64
		err := tx.QueryRow(ctx, `SELECT specialization_id FROM study_group WHERE id = $1 FOR UPDATE`, id).Scan(&current)
		if errors.Is(err, pgx.ErrNoRows) {
			return apperr.NewNotFound("group_not_found")
		}
		if err != nil {
			return database.Translate(err)
		}
		if err := checkSpecialization(ctx, tx, in.SpecializationID, &current); err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `
			UPDATE study_group SET name = $2, specialization_id = $3, admission_year = $4 WHERE id = $1`,
			id, in.Name, in.SpecializationID, in.AdmissionYear)
		return writeErr(err)
	})
}

func (r *Repository) ArchiveGroup(ctx context.Context, id int64) (Group, error) {
	// Группу нельзя заархивировать, пока в ней есть студенты (в том числе заблокированные)
	// или на неё назначен хоть один курс.
	err := archiveTx(ctx, r.db, "study_group", id, "group_not_found", func(tx pgx.Tx) error {
		used, err := exists(ctx, tx, `SELECT 1 FROM student WHERE group_id = $1
			UNION ALL SELECT 1 FROM course_offering WHERE group_id = $1`, id)
		if err == nil && used {
			err = errInUse("students or courses of this group")
		}
		return err
	})
	if err != nil {
		return Group{}, err
	}
	return r.GetGroup(ctx, id)
}

// DeleteGroup удаляет группу; если в ней есть студенты или курсы — ошибка внешнего ключа.
func (r *Repository) DeleteGroup(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM study_group WHERE id = $1`, id)
	return database.AffectedDelete(tag, err, "group_not_found")
}

// ---- semesters ----

const semesterSelect = `SELECT id, name, archived, academic_year_start, number, is_current FROM semester`

func scanSemester(row pgx.Row) (Semester, error) {
	var s Semester
	err := row.Scan(&s.ID, &s.Name, &s.Archived, &s.AcademicYearStart, &s.Number, &s.IsCurrent)
	return s, err
}

// ListSemesters — страница семестров от новых к старым: academic_year_start DESC, number DESC, id ASC.
func (r *Repository) ListSemesters(ctx context.Context, lq ListQuery) ([]Semester, int, error) {
	var w database.Filter
	return pageQuery(ctx, r.db, scanSemester, semesterSelect, &w,
		"archived", []string{"name"}, lq, "academic_year_start DESC, number DESC, id")
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
		`INSERT INTO semester (name, academic_year_start, number) VALUES ($1, $2, $3) RETURNING id`,
		in.Name, in.AcademicYearStart, in.Number).Scan(&id)
	return id, writeErr(err)
}

// UpdateSemester заменяет название, год и номер; флаг «текущий» не трогает.
func (r *Repository) UpdateSemester(ctx context.Context, id int64, in SemesterInput) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE semester SET name = $2, academic_year_start = $3, number = $4 WHERE id = $1`,
		id, in.Name, in.AcademicYearStart, in.Number)
	if err != nil {
		return writeErr(err)
	}
	return database.Affected(tag, nil, "semester_not_found")
}

// ArchiveSemester архивирует семестр. Нельзя, если он текущий (409 current_semester_in_use:
// сначала выберите другой) или на него назначены курсы (409 catalog_in_use).
func (r *Repository) ArchiveSemester(ctx context.Context, id int64) (Semester, error) {
	err := archiveTx(ctx, r.db, "semester", id, "semester_not_found", func(tx pgx.Tx) error {
		current, err := exists(ctx, tx, `SELECT 1 FROM semester WHERE id = $1 AND is_current`, id)
		if err != nil {
			return err
		}
		if current {
			return apperr.NewConflict("current_semester_in_use", "the current semester cannot be archived")
		}
		// И пока на семестр назначен хоть один курс.
		used, err := exists(ctx, tx, `SELECT 1 FROM course_offering WHERE semester_id = $1`, id)
		if err == nil && used {
			err = errInUse("courses of this semester")
		}
		return err
	})
	if err != nil {
		return Semester{}, err
	}
	return r.GetSemester(ctx, id)
}

// DeleteSemester удаляет семестр; если на него ссылаются курсы — ошибка внешнего ключа.
func (r *Repository) DeleteSemester(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM semester WHERE id = $1`, id)
	return database.AffectedDelete(tag, err, "semester_not_found")
}

// SetCurrentSemester делает семестр текущим, снимая флаг с прежнего (одной транзакцией).
// Архивный семестр текущим сделать нельзя.
func (r *Repository) SetCurrentSemester(ctx context.Context, id int64) error {
	return database.InTx(ctx, r.db, func(tx pgx.Tx) error {
		var archived bool
		err := tx.QueryRow(ctx, `SELECT archived FROM semester WHERE id = $1 FOR UPDATE`, id).Scan(&archived)
		if errors.Is(err, pgx.ErrNoRows) {
			return apperr.NewNotFound("semester_not_found")
		}
		if err != nil {
			return database.Translate(err)
		}
		if archived {
			return apperr.NewConflict("semester_archived", "an archived semester cannot be made current")
		}
		if _, err := tx.Exec(ctx, `UPDATE semester SET is_current = false WHERE is_current AND id <> $1`, id); err != nil {
			return database.Translate(err)
		}
		_, err = tx.Exec(ctx, `UPDATE semester SET is_current = true WHERE id = $1`, id)
		return database.Translate(err)
	})
}

// ---- rooms ----
// В таблице room обозначение хранится в колонке number; наружу оно отдаётся как name.

func scanRoom(row pgx.Row) (Room, error) {
	var rm Room
	err := row.Scan(&rm.ID, &rm.Name, &rm.Archived)
	return rm, err
}

// ListRooms — страница аудиторий: name ASC, id ASC.
func (r *Repository) ListRooms(ctx context.Context, lq ListQuery) ([]Room, int, error) {
	var w database.Filter
	return pageQuery(ctx, r.db, scanRoom, `SELECT id, number, archived FROM room`, &w,
		"archived", []string{"number"}, lq, "number, id")
}

func (r *Repository) GetRoom(ctx context.Context, id int64) (Room, error) {
	rm, err := scanRoom(r.db.QueryRow(ctx, `SELECT id, number, archived FROM room WHERE id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Room{}, apperr.NewNotFound("room_not_found")
	}
	return rm, database.Translate(err)
}

func (r *Repository) CreateRoom(ctx context.Context, name string) (Room, error) {
	rm, err := scanRoom(r.db.QueryRow(ctx, `INSERT INTO room (number) VALUES ($1) RETURNING id, number, archived`, name))
	return rm, writeErr(err)
}

func (r *Repository) UpdateRoom(ctx context.Context, id int64, name string) (Room, error) {
	rm, err := scanRoom(r.db.QueryRow(ctx,
		`UPDATE room SET number = $2 WHERE id = $1 RETURNING id, number, archived`, id, name))
	if errors.Is(err, pgx.ErrNoRows) {
		return Room{}, apperr.NewNotFound("room_not_found")
	}
	return rm, writeErr(err)
}

func (r *Repository) ArchiveRoom(ctx context.Context, id int64) (Room, error) {
	// Аудиторию нельзя заархивировать, пока в ней стоит хоть одно занятие.
	err := archiveTx(ctx, r.db, "room", id, "room_not_found", func(tx pgx.Tx) error {
		used, err := exists(ctx, tx, `SELECT 1 FROM lesson WHERE room_id = $1`, id)
		if err == nil && used {
			err = errInUse("lessons in this room")
		}
		return err
	})
	if err != nil {
		return Room{}, err
	}
	return r.GetRoom(ctx, id)
}

// DeleteRoom удаляет аудиторию; если она занята в расписании — ошибка внешнего ключа.
func (r *Repository) DeleteRoom(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM room WHERE id = $1`, id)
	return database.AffectedDelete(tag, err, "room_not_found")
}
