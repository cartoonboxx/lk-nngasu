package users

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"university-api/internal/platform/apperr"
	"university-api/internal/platform/database"
	"university-api/internal/platform/filekey"
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
// Роль одна и вычисляется (database.RoleSQL): admin, иначе teacher, иначе student.
// Профиль другой роли (строка, оставшаяся от старых данных) наружу не отдаётся.
// Порядок колонок должен совпадать с scanUser.
var userSelect = `
	SELECT u.id, u.last_name, u.first_name, u.middle_names, ` + database.FullName("u") + `,
	       u.email, ` + database.FileInfoJSON("af") + `, u.is_active, u.is_admin, u.can_manage_admins,
	       ` + database.RoleSQL("u", "t") + `,
	       s.user_id IS NOT NULL, s.group_id, g.name, s.record_book_number, s.is_headman,
	       t.user_id IS NOT NULL, t.department_id, d.name,
	       u.passport, u.snils
	FROM app_user u
	LEFT JOIN file af          ON af.id = u.avatar_file_id
	LEFT JOIN teacher t        ON t.user_id = u.id
	LEFT JOIN department d     ON d.id = t.department_id
	LEFT JOIN student s        ON s.user_id = u.id
	LEFT JOIN study_group g    ON g.id = s.group_id`

// likeEscape экранирует спецсимволы шаблона LIKE (\\, %, _).
func likeEscape(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}

// scanUser читает одну строку userSelect в AdminUser.
// Подходит и для QueryRow, и для строк из database.Collect.
func scanUser(row pgx.Row) (AdminUser, error) {
	var (
		u                    AdminUser
		avatar               []byte
		isStudent, isTeacher bool
		groupID, deptID      *int64
		groupName, deptName  *string
		isHeadman            *bool
		recordBook           *string
	)
	err := row.Scan(&u.ID, &u.LastName, &u.FirstName, &u.MiddleNames, &u.FullName,
		&u.Email, &avatar, &u.IsActive, &u.IsAdmin, &u.CanManageAdmins, &u.Role,
		&isStudent, &groupID, &groupName, &recordBook, &isHeadman,
		&isTeacher, &deptID, &deptName,
		&u.Passport, &u.Snils)
	if err != nil {
		return AdminUser{}, err
	}
	if u.Avatar, err = database.ParseFileInfo(avatar); err != nil {
		return AdminUser{}, err
	}
	if isStudent && u.Role == RoleStudent {
		u.Student = &StudentProfile{GroupID: *groupID, GroupName: deref(groupName),
			RecordBookNumber: recordBook, IsHeadman: isHeadman != nil && *isHeadman}
	}
	if isTeacher && u.Role == RoleTeacher {
		u.Teacher = &TeacherProfile{DepartmentID: *deptID, DepartmentName: deref(deptName)}
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

	// Фильтр по единственной роли (студенты включают старост).
	if f.Role != "" {
		w.Add(database.RoleSQL("u", "t")+" = ?", f.Role)
	}
	// is_headman=false — студенты без признака старосты; пользователи без student не попадают.
	if f.IsHeadman != nil {
		w.Add("s.user_id IS NOT NULL AND s.is_headman = ?", *f.IsHeadman)
	}
	if f.IsActive != nil {
		w.Add("u.is_active = ?", *f.IsActive)
	}
	if f.GroupID != nil {
		w.Add("s.group_id = ?", *f.GroupID)
	}
	if f.DepartmentID != nil {
		w.Add("t.department_id = ?", *f.DepartmentID)
	}
	// Поиск без учёта регистра по полному ФИО (в том числе «Иванова Анна») и email.
	// Спецсимволы LIKE экранируются, чтобы «_» и «%» искались буквально.
	if f.Q != "" {
		w.Add("("+database.FullName("u")+" ILIKE ? OR u.email ILIKE ?)", "%"+likeEscape(f.Q)+"%")
	}

	// Сколько всего записей под фильтром — для meta.total. Считаем до того,
	// как в аргументы добавятся LIMIT/OFFSET.
	var total int
	countArgs := append([]any(nil), w.Args()...)
	if err := r.db.QueryRow(ctx, `SELECT count(*) FROM (`+userSelect+w.Where()+`) x`, countArgs...).Scan(&total); err != nil {
		return nil, 0, database.Translate(err)
	}

	// LIMIT/OFFSET тоже передаются параметрами (w.Arg возвращает плейсхолдер).
	// Сортировка по ФИО (full_name), затем по id — порядок стабилен при пагинации.
	dir := "ASC"
	if f.Desc {
		dir = "DESC"
	}
	q := userSelect + w.Where() +
		fmt.Sprintf(" ORDER BY u.last_name %[1]s, u.first_name %[1]s, u.middle_names %[1]s NULLS FIRST, u.id %[1]s LIMIT %s OFFSET %s",
			dir, w.Arg(f.Limit), w.Arg(f.Offset))
	list, err := database.Collect(ctx, r.db, scanSummary, q, w.Args()...)
	return list, total, err
}

// Create заводит пользователя и его роли одной транзакцией.
// after выполняется в той же транзакции после вставки (приглашение);
// ошибка в нём откатывает и пользователя.
func (r *Repository) Create(ctx context.Context, in CreateUser, passwordHash string, after func(tx pgx.Tx, id int64) error) (int64, error) {
	var id int64
	err := database.InTx(ctx, r.db, func(tx pgx.Tx) error {
		var err error
		if id, err = createUserTx(ctx, tx, in, passwordHash); err != nil {
			return err
		}
		if after != nil {
			return after(tx, id)
		}
		return nil
	})
	return id, err
}

// referencesExist проверяет группу и кафедру заранее, чтобы ответить
// 400 invalid_reference, а не общей ошибкой внешнего ключа.
//
// referencesExist — группа и кафедра роли существуют. FOR SHARE не даёт удалить их
// параллельно с этой транзакцией.
func referencesExist(ctx context.Context, tx pgx.Tx, st *StudentRole, te *TeacherRole) error {
	var found bool
	if st != nil {
		err := tx.QueryRow(ctx, `SELECT true FROM study_group WHERE id = $1 FOR SHARE`, st.GroupID).Scan(&found)
		if errors.Is(err, pgx.ErrNoRows) {
			return errBadReference.WithDetails(map[string]any{"fields": map[string][]string{
				"student.group_id": {"group does not exist"}}})
		}
		if err != nil {
			return database.Translate(err)
		}
	}
	if te != nil {
		err := tx.QueryRow(ctx, `SELECT true FROM department WHERE id = $1 FOR SHARE`, te.DepartmentID).Scan(&found)
		if errors.Is(err, pgx.ErrNoRows) {
			return errBadReference.WithDetails(map[string]any{"fields": map[string][]string{
				"teacher.department_id": {"department does not exist"}}})
		}
		if err != nil {
			return database.Translate(err)
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
		INSERT INTO app_user (last_name, first_name, middle_names, email, password_hash, passport, snils, avatar_file_id, is_admin)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`,
		in.LastName, in.FirstName, in.MiddleNames, in.Email, passwordHash, in.Passport, in.Snils, in.AvatarFileID,
		in.Role == RoleAdmin).Scan(&id)
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
		return apperr.NewConflict("role_in_use", table+" role has related academic data or teaching load and cannot be removed").
			WithDetails(map[string]any{"role": table})
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

// errInvalidRole — роль и профили student/teacher не согласованы.
func errInvalidRole(field, problem string) error {
	return apperr.NewInvalid("invalid_role", "role and student/teacher profiles do not match").
		WithDetails(map[string]any{"fields": map[string][]string{field: {problem}}})
}

var errOwnerRequired = apperr.NewForbidden("owner_required", "only the system owner can grant or revoke administrator rights")

// AdminUpdate частично обновляет пользователя и его роль одной транзакцией.
//
// Роль одна: при смене role профиль новой роли нужно передать (если его ещё нет),
// профиль прежней снимается автоматически; снять профиль с учебными данными нельзя
// (409 role_in_use). Выдача роли admin здесь запрещена, снять её может только владелец; последнего
// активного администратора нельзя заблокировать или лишить роли (409 last_active_admin).
func (r *Repository) AdminUpdate(ctx context.Context, id, actorID int64, in AdminUpdate) error {
	return database.InTx(ctx, r.db, func(tx pgx.Tx) error {
		// 1. Блокируем строку пользователя: параллельная правка дождётся конца транзакции.
		var isAdmin, isActive, isOwner, hasStudent, hasTeacher bool
		err := tx.QueryRow(ctx, `
			SELECT is_admin, is_active, can_manage_admins,
			       EXISTS (SELECT 1 FROM student WHERE user_id = u.id),
			       EXISTS (SELECT 1 FROM teacher WHERE user_id = u.id)
			FROM app_user u WHERE id = $1 FOR UPDATE`, id).Scan(&isAdmin, &isActive, &isOwner, &hasStudent, &hasTeacher)
		if errors.Is(err, pgx.ErrNoRows) {
			return errUserNotFound
		}
		if err != nil {
			return database.Translate(err)
		}
		if isOwner && actorID != id {
			return apperr.NewForbidden("owner_protected", "other administrators cannot edit the system owner")
		}

		current := RoleStudent
		switch {
		case isAdmin:
			current = RoleAdmin
		case hasTeacher:
			current = RoleTeacher
		}
		role := current
		if in.Role.Value != nil {
			role = *in.Role.Value
		}

		if role == RoleAdmin && !isAdmin {
			return apperr.NewForbidden("admin_invitation_required", "administrator access must be granted through administrator management")
		}

		// 2. Профили согласованы с ролью: свой — передан или уже есть, чужой — не передан.
		needStudent, needTeacher := role == RoleStudent, role == RoleTeacher
		switch {
		case !needStudent && in.Student.Value != nil:
			return errInvalidRole("student", "must be null unless role is student")
		case !needTeacher && in.Teacher.Value != nil:
			return errInvalidRole("teacher", "must be null unless role is teacher")
		case needStudent && (in.Student.Null() || (!in.Student.Set && !hasStudent)):
			return errInvalidRole("student", "is required for role student")
		case needTeacher && (in.Teacher.Null() || (!in.Teacher.Set && !hasTeacher)):
			return errInvalidRole("teacher", "is required for role teacher")
		}

		// 3. Права администратора: выдать или снять — только владелец; владелец свою роль
		// не меняет и себя не блокирует; последний активный администратор остаётся.
		blocking := in.IsActive.Value != nil && !*in.IsActive.Value
		leavingAdmin := isAdmin && role != RoleAdmin
		if (role == RoleAdmin) != isAdmin {
			owner, err := ownerTx(ctx, tx, actorID)
			if err != nil {
				return err
			}
			if !owner {
				return errOwnerRequired
			}
		}
		if isOwner && (blocking || leavingAdmin) {
			return apperr.NewConflict("owner_protected", "the system owner cannot be blocked or lose administrator rights")
		}
		if isAdmin && isActive && (blocking || leavingAdmin) {
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

		// 4. Общие поля учётки — только переданные ключи.
		var set setList
		field(&set, "last_name", in.LastName)
		field(&set, "first_name", in.FirstName)
		field(&set, "middle_names", in.MiddleNames)
		field(&set, "email", in.Email)
		field(&set, "passport", in.Passport)
		field(&set, "snils", in.Snils)
		field(&set, "avatar_file_id", in.AvatarFileID)
		field(&set, "is_active", in.IsActive)
		if role != current {
			admin := role == RoleAdmin
			field(&set, "is_admin", Field[bool]{Set: true, Value: &admin})
		}
		if len(set.cols) > 0 {
			q := `UPDATE app_user SET ` + set.sql() + ` WHERE id = ` + set.arg(id) // до чтения set.args
			if _, err := tx.Exec(ctx, q, set.args...); err != nil {
				return userErr(err)
			}
		}

		// 5. Профили: свой задаётся или заменяется, чужие снимаются.
		if err := referencesExist(ctx, tx, in.Student.Value, in.Teacher.Value); err != nil {
			return err
		}
		if in.Student.Value != nil {
			err = upsertStudent(ctx, tx, id, *in.Student.Value)
		} else if !needStudent && hasStudent {
			err = removeRole(ctx, tx, "student", id)
		}
		if err != nil {
			return err
		}
		if in.Teacher.Value != nil {
			err = upsertTeacher(ctx, tx, id, *in.Teacher.Value)
		} else if !needTeacher && hasTeacher {
			err = removeRole(ctx, tx, "teacher", id)
		}
		if err != nil {
			return err
		}

		// 6. Блокировка и смена роли или профиля отзывают сессии:
		// права в старом access-токене больше не верны, пользователь входит заново.
		switch {
		case blocking:
			return revokeSessions(ctx, tx, id, "", "blocked")
		case role != current || in.Student.Set || in.Teacher.Set:
			return revokeSessions(ctx, tx, id, "", "roles_changed")
		}
		return nil
	})
}

// ownerTx — actor — владелец системы; читается в транзакции изменения прав.
func ownerTx(ctx context.Context, tx pgx.Tx, actorID int64) (bool, error) {
	var ok bool
	err := tx.QueryRow(ctx, `SELECT can_manage_admins FROM app_user WHERE id = $1 FOR SHARE`, actorID).Scan(&ok)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return ok, database.Translate(err)
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

// GroupVisible — группа существует и доступна: администратору любая,
// преподавателю — если он ведёт у неё хотя бы один курс (есть нагрузка).
func (r *Repository) GroupVisible(ctx context.Context, groupID, userID int64, isAdmin bool) (bool, error) {
	var ok bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS (SELECT 1 FROM study_group g WHERE g.id = $1 AND ($3 OR EXISTS (
			SELECT 1 FROM workload w JOIN course_offering co ON co.id = w.course_id
			WHERE co.group_id = g.id AND w.teacher_id = $2)))`, groupID, userID, isAdmin).Scan(&ok)
	return ok, database.Translate(err)
}

// GroupStudents — страница состава группы, по ФИО, затем по id.
func (r *Repository) GroupStudents(ctx context.Context, groupID int64, limit, offset int) ([]CourseStudent, int, error) {
	var total int
	if err := r.db.QueryRow(ctx, `SELECT count(*) FROM student WHERE group_id = $1`, groupID).Scan(&total); err != nil {
		return nil, 0, database.Translate(err)
	}
	list, err := database.Collect(ctx, r.db, func(row pgx.Row) (CourseStudent, error) {
		var s CourseStudent
		err := row.Scan(&s.StudentID, &s.FullName, &s.RecordBookNumber, &s.IsHeadman)
		return s, err
	}, `
		SELECT u.id, `+database.FullName("u")+`, s.record_book_number, s.is_headman
		FROM student s JOIN app_user u ON u.id = s.user_id
		WHERE s.group_id = $1
		ORDER BY u.last_name, u.first_name, u.middle_names NULLS FIRST, u.id
		LIMIT $2 OFFSET $3`, groupID, limit, offset)
	return list, total, err
}

// errFileNotFound — файла с таким id нет (404 по документации PATCH /me).
var errFileNotFound = apperr.NewNotFound("file_not_found")

// errInvalidFile — файл есть, но чужой или не картинка.
var errInvalidFile = apperr.NewInvalid("invalid_file", "avatar must be your own uploaded JPEG, PNG or WebP image")

// avatarExts — допустимые расширения аватара (JPEG, PNG, WebP), как в модуле files.
var avatarExts = map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true}

// CheckAvatar проверяет файл для аватара: существует (иначе 404 file_not_found),
// загружен этим пользователем и является картинкой (иначе 400 invalid_file).
// Владелец и имя файла хранятся в ключе "<user_id>/<random>/<имя>".
func (r *Repository) CheckAvatar(ctx context.Context, userID, fileID int64) error {
	var key string
	err := r.db.QueryRow(ctx, `SELECT storage_key FROM file WHERE id = $1`, fileID).Scan(&key)
	if errors.Is(err, pgx.ErrNoRows) {
		return errFileNotFound
	}
	if err != nil {
		return database.Translate(err)
	}
	if owner, _ := filekey.Owner(key); owner != userID {
		return errInvalidFile
	}
	if !avatarExts[strings.ToLower(path.Ext(filekey.Name(key)))] {
		return errInvalidFile
	}
	return nil
}

// PasswordHash возвращает хеш пароля — для проверки старого пароля
// при смене пароля или email. Наружу (в User) хеш никогда не попадает.
func (r *Repository) PasswordHash(ctx context.Context, id int64) (string, error) {
	var h string
	err := r.db.QueryRow(ctx, `SELECT password_hash FROM app_user WHERE id = $1`, id).Scan(&h)
	return h, database.Translate(err)
}

// SetPassword записывает новый хеш пароля и отзывает все сессии пользователя,
// включая текущую: после смены пароля нужен повторный вход.
func (r *Repository) SetPassword(ctx context.Context, id int64, hash string) error {
	return database.InTx(ctx, r.db, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, `UPDATE app_user SET password_hash = $2 WHERE id = $1`, id, hash)
		if err := database.Affected(tag, err, "user_not_found"); err != nil {
			return err
		}
		return revokeSessions(ctx, tx, id, "", "password_changed")
	})
}

// revokeSessions отзывает активные сессии пользователя, кроме keepSid ("" — все).
// Таблица auth_session принадлежит модулю auth, но отзыв должен попасть в ту же
// транзакцию, что и блокировка или смена пароля, поэтому запрос выполняется здесь.
// Отозванная сессия не проходит проверку в middleware и не обновляется через refresh.
func revokeSessions(ctx context.Context, tx pgx.Tx, userID int64, keepSid, reason string) error {
	_, err := tx.Exec(ctx, `
		UPDATE auth_session SET revoked_at = now(), revoke_reason = $3
		WHERE user_id = $1 AND revoked_at IS NULL AND id::text <> $2`, userID, keepSid, reason)
	return database.Translate(err)
}
