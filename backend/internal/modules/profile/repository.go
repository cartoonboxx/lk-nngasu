package profile

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// errUserNotFound — пользователя с таким id нет (например, удалён после выдачи токена).
// Репозиторий возвращает её вместо pgx.ErrNoRows, чтобы сервис не зависел от драйвера БД.
var errUserNotFound = errors.New("user not found")

// profileRow — «плоская» строка результата запроса, по одному полю на колонку.
// Колонки из LEFT JOIN могут быть NULL, поэтому они указатели.
// Вложенный Profile из этой строки собирает сервис.
type profileRow struct {
	// app_user — есть всегда.
	ID           int64
	FirstName    string
	LastName     string
	MiddleName   *string
	Email        string
	AvatarFileID *int64 // id файла в таблице file; nil — аватара нет
	IsActive     bool
	IsAdmin      bool

	// student → study_group → specialization — NULL, если пользователь не студент.
	IsStudent        bool // есть ли запись в student (вычисляется в запросе)
	RecordBookNumber *string
	IsHeadman        *bool
	GroupID          *int64
	GroupName        *string
	AdmissionYear    *int
	SpecID           *int64
	SpecCode         *string
	SpecName         *string

	// teacher → department — NULL, если пользователь не преподаватель.
	IsTeacher bool // есть ли запись в teacher (вычисляется в запросе)
	DeptID    *int64
	DeptName  *string
}

// Repository — чтение профиля из БД.
type Repository struct{ db *pgxpool.Pool }

// newRepository создаёт репозиторий поверх общего пула подключений (пул создаётся в main).
func newRepository(db *pgxpool.Pool) *Repository { return &Repository{db: db} }

// Get возвращает профиль пользователя одним запросом.
// Если пользователя нет — errUserNotFound.
func (r *Repository) Get(ctx context.Context, userID int64) (profileRow, error) {
	// Везде LEFT JOIN, а не JOIN: у админа нет записей ни в student, ни в teacher,
	// у студента нет записи в teacher и т.д. — обычный JOIN выбросил бы такого пользователя.
	//
	// s.user_id IS NOT NULL / t.user_id IS NOT NULL — признаки «студент» / «преподаватель».
	// По остальным колонкам этого не понять: студент без группы даст NULL в g.id.
	const q = `
		SELECT u.id, u.first_name, u.last_name, u.middle_names, u.email, u.avatar_file_id,
		       u.is_active, u.is_admin,
		       s.user_id IS NOT NULL, s.record_book_number, s.is_headman,
		       g.id, g.name, g.admission_year,
		       sp.id, sp.code, sp.name,
		       t.user_id IS NOT NULL,
		       d.id, d.name
		FROM app_user u
		LEFT JOIN student        s  ON s.user_id = u.id
		LEFT JOIN study_group    g  ON g.id = s.group_id
		LEFT JOIN specialization sp ON sp.id = g.specialization_id
		LEFT JOIN teacher        t  ON t.user_id = u.id
		LEFT JOIN department     d  ON d.id = t.department_id
		WHERE u.id = $1`

	// Порядок полей в Scan должен совпадать с порядком колонок в SELECT.
	var p profileRow
	err := r.db.QueryRow(ctx, q, userID).Scan(
		&p.ID, &p.FirstName, &p.LastName, &p.MiddleName, &p.Email, &p.AvatarFileID,
		&p.IsActive, &p.IsAdmin,
		&p.IsStudent, &p.RecordBookNumber, &p.IsHeadman,
		&p.GroupID, &p.GroupName, &p.AdmissionYear,
		&p.SpecID, &p.SpecCode, &p.SpecName,
		&p.IsTeacher,
		&p.DeptID, &p.DeptName,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return profileRow{}, errUserNotFound
	}
	if err != nil {
		// При ошибке не отдаём частично заполненную структуру.
		return profileRow{}, err
	}
	return p, nil
}
