package structure

import (
	"context"
	"errors"

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

	// Справочники вида (id, name) обслуживаются одним кодом catalog,
	// отличаются только таблицей и кодом ошибки.
	departments catalog
	subjects    catalog
}

func newRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db:          db,
		departments: catalog{db: db, table: "department", notFound: "department_not_found"},
		subjects:    catalog{db: db, table: "subject", notFound: "subject_not_found"},
	}
}

// ---- catalog: таблицы вида (id, name) — кафедры и дисциплины ----

// catalog — CRUD для одной таблицы вида (id, name).
// Имя таблицы подставляется в SQL строкой (параметром его не передать),
// поэтому оно задаётся только константой в newRepository и никогда не берётся из запроса.
type catalog struct {
	db       *pgxpool.Pool
	table    string // константа из кода, не пользовательский ввод
	notFound string // код ошибки, когда запись с таким id не найдена
}

func scanNamed(row pgx.Row) (Named, error) {
	var n Named
	err := row.Scan(&n.ID, &n.Name)
	return n, err
}

// list возвращает записи по алфавиту; q — необязательный поиск подстроки
// в названии без учёта регистра.
func (c catalog) list(ctx context.Context, q string) ([]Named, error) {
	var w database.Filter
	if q != "" {
		w.Add("name ILIKE ?", "%"+q+"%")
	}
	return database.Collect(ctx, c.db, scanNamed,
		`SELECT id, name FROM `+c.table+w.Where()+` ORDER BY name`, w.Args()...)
}

// create добавляет запись. Дубликат названия (если на name есть
// уникальное ограничение) вернётся как конфликт через Translate.
func (c catalog) create(ctx context.Context, name string) (Named, error) {
	n := Named{Name: name}
	err := c.db.QueryRow(ctx, `INSERT INTO `+c.table+` (name) VALUES ($1) RETURNING id`, name).Scan(&n.ID)
	return n, database.Translate(err)
}

// update переименовывает запись. RETURNING даёт одновременно
// и проверку существования (ErrNoRows), и значение для ответа без второго запроса.
func (c catalog) update(ctx context.Context, id int64, name string) (Named, error) {
	n := Named{ID: id}
	err := c.db.QueryRow(ctx, `UPDATE `+c.table+` SET name = $2 WHERE id = $1 RETURNING name`, id, name).Scan(&n.Name)
	if errors.Is(err, pgx.ErrNoRows) {
		return Named{}, apperr.NewNotFound(c.notFound)
	}
	return n, database.Translate(err)
}

// remove удаляет запись. AffectedDelete превращает «0 строк» в not found,
// а нарушение внешнего ключа (на кафедру ссылаются преподаватели,
// на дисциплину — расписание) — в ошибку «запись используется».
func (c catalog) remove(ctx context.Context, id int64) error {
	tag, err := c.db.Exec(ctx, `DELETE FROM `+c.table+` WHERE id = $1`, id)
	return database.AffectedDelete(tag, err, c.notFound)
}

// ---- specializations ----

func scanSpecialization(row pgx.Row) (Specialization, error) {
	var s Specialization
	err := row.Scan(&s.ID, &s.Code, &s.Name)
	return s, err
}

// ListSpecializations возвращает все направления, отсортированные по шифру
// (шифры вида «09.03.04» при строковой сортировке идут в правильном порядке).
func (r *Repository) ListSpecializations(ctx context.Context) ([]Specialization, error) {
	return database.Collect(ctx, r.db, scanSpecialization, `SELECT id, code, name FROM specialization ORDER BY code`)
}

// CreateSpecialization добавляет направление и возвращает его целиком (RETURNING).
func (r *Repository) CreateSpecialization(ctx context.Context, code, name string) (Specialization, error) {
	s, err := scanSpecialization(r.db.QueryRow(ctx,
		`INSERT INTO specialization (code, name) VALUES ($1, $2) RETURNING id, code, name`, code, name))
	return s, database.Translate(err)
}

// UpdateSpecialization заменяет шифр и название направления.
func (r *Repository) UpdateSpecialization(ctx context.Context, id int64, code, name string) (Specialization, error) {
	s, err := scanSpecialization(r.db.QueryRow(ctx,
		`UPDATE specialization SET code = $2, name = $3 WHERE id = $1 RETURNING id, code, name`, id, code, name))
	if errors.Is(err, pgx.ErrNoRows) {
		return Specialization{}, apperr.NewNotFound("specialization_not_found")
	}
	return s, database.Translate(err)
}

// DeleteSpecialization удаляет направление; если на него ссылаются группы — ошибка внешнего ключа.
func (r *Repository) DeleteSpecialization(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM specialization WHERE id = $1`, id)
	return database.AffectedDelete(tag, err, "specialization_not_found")
}

// ---- groups ----

// groupSelect — общий SELECT группы вместе с названием направления.
// JOIN внутренний: specialization_id у группы обязателен (NOT NULL + внешний ключ).
// Порядок колонок должен совпадать со scanGroup.
const groupSelect = `
	SELECT g.id, g.name, g.specialization_id, sp.name, g.admission_year
	FROM study_group g
	JOIN specialization sp ON sp.id = g.specialization_id`

func scanGroup(row pgx.Row) (Group, error) {
	var g Group
	err := row.Scan(&g.ID, &g.Name, &g.SpecializationID, &g.SpecializationName, &g.AdmissionYear)
	return g, err
}

// GetGroup возвращает группу по id.
func (r *Repository) GetGroup(ctx context.Context, id int64) (Group, error) {
	g, err := scanGroup(r.db.QueryRow(ctx, groupSelect+` WHERE g.id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Group{}, apperr.NewNotFound("group_not_found")
	}
	return g, database.Translate(err)
}

// ListGroups возвращает группы по фильтру: сначала младшие наборы, внутри набора — по названию.
func (r *Repository) ListGroups(ctx context.Context, f GroupFilter) ([]Group, error) {
	var w database.Filter
	if f.SpecializationID != nil {
		w.Add("g.specialization_id = ?", *f.SpecializationID)
	}
	if f.AdmissionYear != nil {
		w.Add("g.admission_year = ?", *f.AdmissionYear)
	}
	return database.Collect(ctx, r.db, scanGroup,
		groupSelect+w.Where()+` ORDER BY g.admission_year DESC, g.name`, w.Args()...)
}

// CreateGroup создаёт группу и возвращает её id; полную запись хендлер перечитывает через GetGroup.
// Несуществующее направление вернётся как ошибка внешнего ключа.
func (r *Repository) CreateGroup(ctx context.Context, in GroupInput) (int64, error) {
	var id int64
	err := r.db.QueryRow(ctx, `
		INSERT INTO study_group (name, specialization_id, admission_year) VALUES ($1, $2, $3) RETURNING id`,
		in.Name, in.SpecializationID, in.AdmissionYear).Scan(&id)
	return id, database.Translate(err)
}

// UpdateGroup заменяет все поля группы.
func (r *Repository) UpdateGroup(ctx context.Context, id int64, in GroupInput) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE study_group SET name = $2, specialization_id = $3, admission_year = $4 WHERE id = $1`,
		id, in.Name, in.SpecializationID, in.AdmissionYear)
	return database.Affected(tag, err, "group_not_found")
}

// DeleteGroup удаляет группу; если в ней есть студенты — ошибка внешнего ключа.
func (r *Repository) DeleteGroup(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM study_group WHERE id = $1`, id)
	return database.AffectedDelete(tag, err, "group_not_found")
}

// ---- semesters ----

// semesterSelect — общий SELECT семестра; порядок колонок совпадает со scanSemester.
const semesterSelect = `SELECT id, academic_year_start, number, is_current FROM semester`

func scanSemester(row pgx.Row) (Semester, error) {
	var s Semester
	err := row.Scan(&s.ID, &s.AcademicYearStart, &s.Number, &s.IsCurrent)
	return s, err
}

// ListSemesters возвращает семестры от новых к старым.
func (r *Repository) ListSemesters(ctx context.Context) ([]Semester, error) {
	return database.Collect(ctx, r.db, scanSemester,
		semesterSelect+` ORDER BY academic_year_start DESC, number DESC`)
}

// GetSemester возвращает семестр по id.
func (r *Repository) GetSemester(ctx context.Context, id int64) (Semester, error) {
	s, err := scanSemester(r.db.QueryRow(ctx, semesterSelect+` WHERE id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Semester{}, apperr.NewNotFound("semester_not_found")
	}
	return s, database.Translate(err)
}

// CurrentSemester возвращает текущий семестр; если он не выбран — not found
// с кодом no_current_semester (фронт отличает его от «семестр с таким id не найден»).
func (r *Repository) CurrentSemester(ctx context.Context) (Semester, error) {
	s, err := scanSemester(r.db.QueryRow(ctx, semesterSelect+` WHERE is_current`))
	if errors.Is(err, pgx.ErrNoRows) {
		return Semester{}, apperr.NewNotFound("no_current_semester")
	}
	return s, database.Translate(err)
}

// CreateSemester создаёт семестр. is_current не задаётся и берётся из
// значения по умолчанию в схеме (false): новый семестр текущим не становится.
func (r *Repository) CreateSemester(ctx context.Context, year, number int) (int64, error) {
	var id int64
	err := r.db.QueryRow(ctx,
		`INSERT INTO semester (academic_year_start, number) VALUES ($1, $2) RETURNING id`, year, number).Scan(&id)
	return id, database.Translate(err)
}

// UpdateSemester меняет год и номер; флаг «текущий» не трогает.
func (r *Repository) UpdateSemester(ctx context.Context, id int64, year, number int) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE semester SET academic_year_start = $2, number = $3 WHERE id = $1`, id, year, number)
	return database.Affected(tag, err, "semester_not_found")
}

// DeleteSemester удаляет семестр. Удаление текущего не запрещено:
// после него текущего семестра не будет, пока его не выберут заново.
func (r *Repository) DeleteSemester(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM semester WHERE id = $1`, id)
	return database.AffectedDelete(tag, err, "semester_not_found")
}

// SetCurrentSemester делает семестр текущим, снимая флаг с прежнего (в одной транзакции).
// Если семестра с таким id нет, Affected вернёт ошибку и транзакция откатится,
// так что прежний текущий семестр останется на месте.
func (r *Repository) SetCurrentSemester(ctx context.Context, id int64) error {
	return database.InTx(ctx, r.db, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `UPDATE semester SET is_current = false WHERE is_current`); err != nil {
			return database.Translate(err)
		}
		tag, err := tx.Exec(ctx, `UPDATE semester SET is_current = true WHERE id = $1`, id)
		return database.Affected(tag, err, "semester_not_found")
	})
}

// ---- rooms ----

func scanRoom(row pgx.Row) (Room, error) {
	var rm Room
	err := row.Scan(&rm.ID, &rm.Number)
	return rm, err
}

// ListRooms возвращает все аудитории, отсортированные по номеру как строке.
func (r *Repository) ListRooms(ctx context.Context) ([]Room, error) {
	return database.Collect(ctx, r.db, scanRoom, `SELECT id, number FROM room ORDER BY number`)
}

// CreateRoom добавляет аудиторию.
func (r *Repository) CreateRoom(ctx context.Context, number string) (Room, error) {
	rm, err := scanRoom(r.db.QueryRow(ctx, `INSERT INTO room (number) VALUES ($1) RETURNING id, number`, number))
	return rm, database.Translate(err)
}

// UpdateRoom меняет номер аудитории.
func (r *Repository) UpdateRoom(ctx context.Context, id int64, number string) (Room, error) {
	rm, err := scanRoom(r.db.QueryRow(ctx, `UPDATE room SET number = $2 WHERE id = $1 RETURNING id, number`, id, number))
	if errors.Is(err, pgx.ErrNoRows) {
		return Room{}, apperr.NewNotFound("room_not_found")
	}
	return rm, database.Translate(err)
}

// DeleteRoom удаляет аудиторию; если она занята в расписании — ошибка внешнего ключа.
func (r *Repository) DeleteRoom(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM room WHERE id = $1`, id)
	return database.AffectedDelete(tag, err, "room_not_found")
}
