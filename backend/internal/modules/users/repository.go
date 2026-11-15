package users

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"university-api/internal/platform/apperr"
	"university-api/internal/platform/database"
)

// Repository — доступ к таблицам app_user, student, teacher.
// Знает только SQL: хеширование паролей, проверка ролей и прав живут в сервисе.
// Ошибки БД переводятся в apperr (через database.Translate или userErr),
// поэтому наружу не уходят ошибки pgx.
type Repository struct{ db *pgxpool.Pool }

func newRepository(db *pgxpool.Pool) *Repository { return &Repository{db: db} }

// userSelect — общий SELECT, который собирает AdminUser из app_user и ролей.
// К нему дописываются WHERE / ORDER BY / LIMIT.
//
// Роли независимы: LEFT JOIN к student и teacher, у пользователя может быть
// любая их комбинация; отсутствующая роль приходит NULL-ами.
// credential_state: пустой хеш — пароль ещё не задан, пользователь приглашён.
// Порядок колонок должен совпадать с scanUser.
var userSelect = `
	SELECT u.id, u.last_name, u.first_name, u.middle_names, ` + database.FullName("u") + `,
	       u.email, u.avatar_file_id, u.is_active, u.is_admin,
	       s.user_id IS NOT NULL, s.group_id, g.name, s.record_book_number, s.is_headman,
	       t.user_id IS NOT NULL, t.department_id, d.name,
	       u.passport, u.snils, u.password_hash = ''
	FROM app_user u
	LEFT JOIN teacher t        ON t.user_id = u.id
	LEFT JOIN department d     ON d.id = t.department_id
	LEFT JOIN student s        ON s.user_id = u.id
	LEFT JOIN study_group g    ON g.id = s.group_id`

// avatarURL — адрес скачивания аватара (GET /files/:id, с авторизацией).
const avatarURL = "/api/v1/files/"

// scanUser читает одну строку userSelect в AdminUser.
// Подходит и для QueryRow, и для строк из database.Collect.
func scanUser(row pgx.Row) (AdminUser, error) {
	var (
		u                    AdminUser
		avatar               *int64
		isStudent, isTeacher bool
		groupID, deptID      *int64
		groupName, deptName  *string
		isHeadman            *bool
		recordBook           *string
		invited              bool
	)
	err := row.Scan(&u.ID, &u.LastName, &u.FirstName, &u.MiddleNames, &u.FullName,
		&u.Email, &avatar, &u.IsActive, &u.IsAdmin,
		&isStudent, &groupID, &groupName, &recordBook, &isHeadman,
		&isTeacher, &deptID, &deptName,
		&u.Passport, &u.Snils, &invited)
	if err != nil {
		return AdminUser{}, err
	}
	if avatar != nil {
		u.Avatar = &FileRef{ID: *avatar, URL: avatarURL + strconv.FormatInt(*avatar, 10)}
	}
	if isStudent {
		u.Student = &StudentProfile{GroupID: *groupID, GroupName: deref(groupName),
			RecordBookNumber: recordBook, IsHeadman: isHeadman != nil && *isHeadman}
	}
	if isTeacher {
		u.Teacher = &TeacherProfile{DepartmentID: *deptID, DepartmentName: deref(deptName)}
	}
	u.CredentialState = "active"
	if invited {
		u.CredentialState = "invited"
	}
	return u, nil
}

// scanSummary — то же без документов (UserSummary) для списков.
func scanSummary(row pgx.Row) (User, error) {
	u, err := scanUser(row)
	return u.User, err
}

// userErr — перевод ошибок записи пользователя.
// Нарушения уникальности email и номера зачётки превращаются в 409 с
// понятными кодами для фронта; всё остальное — в общий database.Translate.
// Имена ограничений должны совпадать с теми, что в миграциях.
func userErr(err error) error {
	if database.PgCode(err) == database.CodeUnique {
		switch database.ConstraintName(err) {
		case "app_user_email_key":
			return apperr.NewConflict("email_exists", "email is already taken")
		case "student_record_book_number_key":
			return apperr.NewConflict("record_book_exists", "record book number is already taken")
		}
	}
	return database.Translate(err)
}

// Get возвращает пользователя по id; если его нет — 404 user_not_found.
func (r *Repository) Get(ctx context.Context, id int64) (AdminUser, error) {
	u, err := scanUser(r.db.QueryRow(ctx, userSelect+` WHERE u.id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return AdminUser{}, errUserNotFound
	}
	return u, database.Translate(err)
}

var errUserNotFound = apperr.NewNotFound("user_not_found")

// List возвращает страницу пользователей по фильтру.
// Условия собираются в database.Filter, значения идут параметрами запроса,
// в текст SQL пользовательский ввод не подставляется.
// Сортировка по фамилии и имени; u.id в конце делает порядок стабильным,
// чтобы при пагинации строки с одинаковым ФИО не прыгали между страницами.
func (r *Repository) List(ctx context.Context, f ListFilter) ([]User, int, error) {
	var w database.Filter

	// Фильтр по роли — по наличию строки роли; роли независимы,
	// поэтому студент-преподаватель попадает в оба фильтра.
	switch f.Role {
	case "student":
		w.Raw("s.user_id IS NOT NULL")
	case "teacher":
		w.Raw("t.user_id IS NOT NULL")
	case "admin":
		w.Raw("u.is_admin")
	}
	if f.GroupID != nil {
		w.Add("s.group_id = ?", *f.GroupID)
	}
	if f.DepartmentID != nil {
		w.Add("t.department_id = ?", *f.DepartmentID)
	}
	// Поиск подстроки без учёта регистра по фамилии, имени и email.
	if f.Q != "" {
		w.Add("(u.last_name ILIKE ? OR u.first_name ILIKE ? OR u.email ILIKE ?)", "%"+f.Q+"%")
	}

	// Сколько всего записей под фильтром — для meta.total. Считаем до того,
	// как в аргументы добавятся LIMIT/OFFSET.
	var total int
	countArgs := append([]any(nil), w.Args()...)
	if err := r.db.QueryRow(ctx, `SELECT count(*) FROM (`+userSelect+w.Where()+`) x`, countArgs...).Scan(&total); err != nil {
		return nil, 0, database.Translate(err)
	}

	// LIMIT/OFFSET тоже передаются параметрами (w.Arg возвращает плейсхолдер).
	q := userSelect + w.Where() +
		fmt.Sprintf(" ORDER BY u.last_name, u.first_name, u.id LIMIT %s OFFSET %s", w.Arg(f.Limit), w.Arg(f.Offset))
	list, err := database.Collect(ctx, r.db, scanSummary, q, w.Args()...)
	return list, total, err
}

// Create заводит пользователя и его роли одной транзакцией.
// Если вставка роли упала, app_user тоже откатывается.
func (r *Repository) Create(ctx context.Context, in CreateUser, passwordHash string) (int64, error) {
	var id int64
	err := database.InTx(ctx, r.db, func(tx pgx.Tx) error {
		var err error
		id, err = createUserTx(ctx, tx, in, passwordHash)
		return err
	})
	return id, err
}

// referencesExist проверяет группу и кафедру заранее, чтобы ответить
// 400 invalid_reference, а не общей ошибкой внешнего ключа.
func referencesExist(ctx context.Context, tx pgx.Tx, st *StudentRole, te *TeacherRole) error {
	var ok bool
	if st != nil {
		if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM study_group WHERE id = $1)`, st.GroupID).Scan(&ok); err != nil {
			return database.Translate(err)
		}
		if !ok {
			return errBadReference.WithDetails(map[string]any{"fields": map[string][]string{"student.group_id": {"group does not exist"}}})
		}
	}
	if te != nil {
		if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM department WHERE id = $1)`, te.DepartmentID).Scan(&ok); err != nil {
			return database.Translate(err)
		}
		if !ok {
			return errBadReference.WithDetails(map[string]any{"fields": map[string][]string{"teacher.department_id": {"department does not exist"}}})
		}
	}
	return nil
}

// createUserTx — вставка пользователя и его ролей внутри уже открытой транзакции.
// Используется и администратором (POST /users), и при одобрении заявки,
// чтобы создать пользователя и закрыть заявку одной транзакцией.
// Пароль приходит уже хешированным; пустая строка — пользователь приглашён.
func createUserTx(ctx context.Context, tx pgx.Tx, in CreateUser, passwordHash string) (int64, error) {
	if err := referencesExist(ctx, tx, in.Student, in.Teacher); err != nil {
		return 0, err
	}
	var id int64
	err := tx.QueryRow(ctx, `
		INSERT INTO app_user (last_name, first_name, middle_names, email, password_hash, passport, snils, avatar_file_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`,
		in.LastName, in.FirstName, in.MiddleNames, in.Email, passwordHash, in.Passport, in.Snils, in.AvatarFileID).Scan(&id)
	if err != nil {
		return 0, userErr(err)
	}
	if in.Student != nil {
		if err := upsertStudent(ctx, tx, id, *in.Student); err != nil {
			return 0, err
		}
	}
	if in.Teacher != nil {
		if err := upsertTeacher(ctx, tx, id, *in.Teacher); err != nil {
			return 0, err
		}
	}
	return id, nil
}

// upsertStudent задаёт или заменяет роль студента целиком (StudentWrite — полный объект).
func upsertStudent(ctx context.Context, tx pgx.Tx, userID int64, r StudentRole) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO student (user_id, group_id, record_book_number, is_headman) VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id) DO UPDATE
		SET group_id = EXCLUDED.group_id, record_book_number = EXCLUDED.record_book_number,
		    is_headman = EXCLUDED.is_headman`,
		userID, r.GroupID, r.RecordBookNumber, r.IsHeadman)
	return userErr(err)
}

// upsertTeacher задаёт или заменяет роль преподавателя.
func upsertTeacher(ctx context.Context, tx pgx.Tx, userID int64, r TeacherRole) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO teacher (user_id, department_id) VALUES ($1, $2)
		ON CONFLICT (user_id) DO UPDATE SET department_id = EXCLUDED.department_id`, userID, r.DepartmentID)
	return userErr(err)
}

// removeRole снимает роль. На строку роли ссылаются учебные данные (ответы студента,
// курсы и материалы преподавателя) без ON DELETE CASCADE, поэтому при наличии
// таких данных БД отвечает нарушением внешнего ключа — это и есть запрет из документации.
func removeRole(ctx context.Context, tx pgx.Tx, table string, userID int64) error {
	_, err := tx.Exec(ctx, `DELETE FROM `+table+` WHERE user_id = $1`, userID)
	if database.PgCode(err) == database.CodeFK {
		if table == "student" {
			return apperr.NewConflict("student_role_in_use", "student role has related academic data and cannot be removed")
		}
		return apperr.NewConflict("teacher_role_in_use", "teacher role has a teaching load and cannot be removed")
	}
	return database.Translate(err)
}

// UpdateProfile частично обновляет свой профиль. Меняются только переданные ключи;
// null в middle_names и avatar_file_id очищает значение.
func (r *Repository) UpdateProfile(ctx context.Context, id int64, in UpdateProfile) error {
	var set setList
	field(&set, "last_name", in.LastName)
	field(&set, "first_name", in.FirstName)
	field(&set, "middle_names", in.MiddleNames)
	field(&set, "avatar_file_id", in.AvatarFileID)
	q := `UPDATE app_user SET ` + set.sql() + ` WHERE id = ` + set.arg(id) // до чтения set.args
	tag, err := r.db.Exec(ctx, q, set.args...)
	return database.Affected(tag, err, "user_not_found")
}

// AdminUpdate частично обновляет пользователя и его роли одной транзакцией.
func (r *Repository) AdminUpdate(ctx context.Context, id int64, in AdminUpdate) error {
	return database.InTx(ctx, r.db, func(tx pgx.Tx) error {
		// 1. Блокируем строку пользователя: параллельная правка дождётся конца транзакции.
		var isAdmin, isActive bool
		err := tx.QueryRow(ctx, `SELECT is_admin, is_active FROM app_user WHERE id = $1 FOR UPDATE`, id).Scan(&isAdmin, &isActive)
		if errors.Is(err, pgx.ErrNoRows) {
			return errUserNotFound
		}
		if err != nil {
			return database.Translate(err)
		}

		// 2. Нельзя заблокировать или разжаловать последнего активного администратора.
		// Блокируем всех активных админов, чтобы два запроса не прошли проверку одновременно.
		losesAdmin := (in.IsActive.Value != nil && !*in.IsActive.Value) || (in.IsAdmin.Value != nil && !*in.IsAdmin.Value)
		if isAdmin && isActive && losesAdmin {
			rows, err := tx.Query(ctx, `SELECT id FROM app_user WHERE is_admin AND is_active FOR UPDATE`)
			if err != nil {
				return database.Translate(err)
			}
			ids, err := pgx.CollectRows(rows, pgx.RowTo[int64])
			if err != nil {
				return database.Translate(err)
			}
			if len(ids) <= 1 {
				return apperr.NewConflict("last_active_admin", "cannot block or demote the last active administrator")
			}
		}

		// 3. Общие поля учётки — только переданные ключи.
		var set setList
		field(&set, "last_name", in.LastName)
		field(&set, "first_name", in.FirstName)
		field(&set, "middle_names", in.MiddleNames)
		field(&set, "email", in.Email)
		field(&set, "passport", in.Passport)
		field(&set, "snils", in.Snils)
		field(&set, "avatar_file_id", in.AvatarFileID)
		field(&set, "is_active", in.IsActive)
		field(&set, "is_admin", in.IsAdmin)
		if len(set.cols) > 0 {
			q := `UPDATE app_user SET ` + set.sql() + ` WHERE id = ` + set.arg(id) // до чтения set.args
			if _, err := tx.Exec(ctx, q, set.args...); err != nil {
				return userErr(err)
			}
		}

		// 4. Роли: объект задаёт или заменяет роль, null снимает.
		if err := referencesExist(ctx, tx, in.Student.Value, in.Teacher.Value); err != nil {
			return err
		}
		switch {
		case in.Student.Value != nil:
			err = upsertStudent(ctx, tx, id, *in.Student.Value)
		case in.Student.Set:
			err = removeRole(ctx, tx, "student", id)
		}
		if err != nil {
			return err
		}
		switch {
		case in.Teacher.Value != nil:
			err = upsertTeacher(ctx, tx, id, *in.Teacher.Value)
		case in.Teacher.Set:
			err = removeRole(ctx, tx, "teacher", id)
		}
		// Сессии заблокированного пользователя отдельно отзывать не нужно:
		// проверка сессии в middleware и refresh смотрят на app_user.is_active.
		return err
	})
}

// setList собирает "col = $n, ..." для частичного UPDATE из полей Field.
// Имена колонок — константы из кода, значения идут параметрами.
type setList struct {
	cols []string
	args []any
}

func (s *setList) arg(v any) string {
	s.args = append(s.args, v)
	return "$" + strconv.Itoa(len(s.args))
}

// field добавляет колонку, если ключ был в запросе; Value == nil записывает NULL.
func field[T any](s *setList, col string, f Field[T]) {
	if f.Set {
		s.cols = append(s.cols, col+" = "+s.arg(f.Value))
	}
}

func (s *setList) sql() string { return strings.Join(s.cols, ", ") }

// PasswordHash возвращает хеш пароля — для проверки старого пароля
// при смене пароля или email. Наружу (в User) хеш никогда не попадает.
func (r *Repository) PasswordHash(ctx context.Context, id int64) (string, error) {
	var h string
	err := r.db.QueryRow(ctx, `SELECT password_hash FROM app_user WHERE id = $1`, id).Scan(&h)
	return h, database.Translate(err)
}

// SetPassword записывает новый хеш пароля.
func (r *Repository) SetPassword(ctx context.Context, id int64, hash string) error {
	tag, err := r.db.Exec(ctx, `UPDATE app_user SET password_hash = $2 WHERE id = $1`, id, hash)
	return database.Affected(tag, err, "user_not_found")
}

// SetEmail меняет email. Ошибка проходит через userErr, чтобы занятый email
// вернулся как 409 email_exists, а не как общая ошибка БД.
func (r *Repository) SetEmail(ctx context.Context, id int64, email string) error {
	tag, err := r.db.Exec(ctx, `UPDATE app_user SET email = $2 WHERE id = $1`, id, email)
	if err != nil {
		return userErr(err)
	}
	return database.Affected(tag, nil, "user_not_found")
}
