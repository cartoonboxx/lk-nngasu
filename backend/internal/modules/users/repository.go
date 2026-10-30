package users

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"university-api/internal/platform/apperr"
	"university-api/internal/platform/database"
)

type Repository struct{ db *pgxpool.Pool }

func newRepository(db *pgxpool.Pool) *Repository { return &Repository{db: db} }

var userSelect = `
	SELECT u.id, u.last_name, u.first_name, u.middle_names, ` + database.FullName("u") + `,
	       u.email, u.avatar_file_id, u.is_active,
	       CASE WHEN u.is_admin THEN 'admin'
	            WHEN t.user_id IS NOT NULL THEN 'teacher'
	            WHEN s.user_id IS NOT NULL THEN 'student'
	            ELSE 'none' END,
	       s.group_id, g.name, s.record_book_number, s.is_headman, t.department_id, d.name
	FROM app_user u
	LEFT JOIN teacher t        ON t.user_id = u.id
	LEFT JOIN department d     ON d.id = t.department_id
	LEFT JOIN student s        ON s.user_id = u.id
	LEFT JOIN study_group g    ON g.id = s.group_id`

func scanUser(row pgx.Row) (User, error) {
	var u User
	err := row.Scan(&u.ID, &u.LastName, &u.FirstName, &u.MiddleNames, &u.FullName,
		&u.Email, &u.AvatarFileID, &u.IsActive, &u.Role,
		&u.GroupID, &u.GroupName, &u.RecordBookNumber, &u.IsHeadman, &u.DepartmentID, &u.DepartmentName)
	return u, err
}

// userErr уточняет нарушения уникальности по имени ограничения.
func userErr(err error) error {
	if database.PgCode(err) == database.CodeUnique {
		switch database.ConstraintName(err) {
		case "app_user_email_key":
			return apperr.NewConflict("email_taken", "email is already taken")
		case "student_record_book_number_key":
			return apperr.NewConflict("record_book_taken", "record book number is already taken")
		}
	}
	return database.Translate(err)
}

func (r *Repository) Get(ctx context.Context, id int64) (User, error) {
	u, err := scanUser(r.db.QueryRow(ctx, userSelect+` WHERE u.id = $1`, id))
	return u, database.Translate(err)
}

func (r *Repository) List(ctx context.Context, f ListFilter) ([]User, error) {
	var w database.Filter
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
	if f.Q != "" {
		w.Add("(u.last_name ILIKE ? OR u.first_name ILIKE ? OR u.email ILIKE ?)", "%"+f.Q+"%")
	}
	q := userSelect + w.Where() +
		fmt.Sprintf(" ORDER BY u.last_name, u.first_name, u.id LIMIT %s OFFSET %s", w.Arg(f.Limit), w.Arg(f.Offset))
	return database.Collect(ctx, r.db, scanUser, q, w.Args()...)
}

// Create заводит пользователя и его профиль (student/teacher) одной транзакцией.
func (r *Repository) Create(ctx context.Context, in CreateUser, passwordHash string) (int64, error) {
	var id int64
	err := database.InTx(ctx, r.db, func(tx pgx.Tx) error {
		err := tx.QueryRow(ctx, `
			INSERT INTO app_user (last_name, first_name, middle_names, email, password_hash, is_admin)
			VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
			in.LastName, in.FirstName, in.MiddleNames, in.Email, passwordHash, in.Role == "admin").Scan(&id)
		if err != nil {
			return userErr(err)
		}
		switch in.Role {
		case "student":
			_, err = tx.Exec(ctx,
				`INSERT INTO student (user_id, group_id, record_book_number, is_headman) VALUES ($1, $2, $3, $4)`,
				id, *in.GroupID, in.RecordBookNumber, in.IsHeadman)
		case "teacher":
			_, err = tx.Exec(ctx, `INSERT INTO teacher (user_id, department_id) VALUES ($1, $2)`, id, *in.DepartmentID)
		}
		return userErr(err)
	})
	return id, err
}

func (r *Repository) UpdateProfile(ctx context.Context, id int64, in UpdateProfile) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE app_user SET
			last_name      = COALESCE($2, last_name),
			first_name     = COALESCE($3, first_name),
			middle_names   = COALESCE($4, middle_names),
			avatar_file_id = COALESCE($5, avatar_file_id)
		WHERE id = $1`, id, in.LastName, in.FirstName, in.MiddleNames, in.AvatarFileID)
	return database.Affected(tag, err, "user_not_found")
}

func (r *Repository) AdminUpdate(ctx context.Context, id int64, in AdminUpdate, passwordHash *string) error {
	return database.InTx(ctx, r.db, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, `
			UPDATE app_user SET
				last_name     = COALESCE($2, last_name),
				first_name    = COALESCE($3, first_name),
				middle_names  = COALESCE($4, middle_names),
				email         = COALESCE($5, email),
				password_hash = COALESCE($6, password_hash),
				is_active     = COALESCE($7, is_active),
				is_admin      = COALESCE($8, is_admin)
			WHERE id = $1`,
			id, in.LastName, in.FirstName, in.MiddleNames, in.Email, passwordHash, in.IsActive, in.IsAdmin)
		if err != nil {
			return userErr(err)
		}
		if tag.RowsAffected() == 0 {
			return apperr.NewNotFound("user_not_found")
		}

		if in.GroupID != nil || in.IsHeadman != nil || in.RecordBookNumber != nil {
			tag, err = tx.Exec(ctx, `
				UPDATE student SET
					group_id           = COALESCE($2, group_id),
					is_headman         = COALESCE($3, is_headman),
					record_book_number = COALESCE($4, record_book_number)
				WHERE user_id = $1`, id, in.GroupID, in.IsHeadman, in.RecordBookNumber)
			if err != nil {
				return userErr(err)
			}
			if tag.RowsAffected() == 0 {
				return apperr.NewInvalid("not_a_student", "user is not a student")
			}
		}
		if in.DepartmentID != nil {
			tag, err = tx.Exec(ctx, `UPDATE teacher SET department_id = $2 WHERE user_id = $1`, id, *in.DepartmentID)
			if err != nil {
				return database.Translate(err)
			}
			if tag.RowsAffected() == 0 {
				return apperr.NewInvalid("not_a_teacher", "user is not a teacher")
			}
		}
		return nil
	})
}

func (r *Repository) PasswordHash(ctx context.Context, id int64) (string, error) {
	var h string
	err := r.db.QueryRow(ctx, `SELECT password_hash FROM app_user WHERE id = $1`, id).Scan(&h)
	return h, database.Translate(err)
}

func (r *Repository) SetPassword(ctx context.Context, id int64, hash string) error {
	tag, err := r.db.Exec(ctx, `UPDATE app_user SET password_hash = $2 WHERE id = $1`, id, hash)
	return database.Affected(tag, err, "user_not_found")
}

func (r *Repository) SetEmail(ctx context.Context, id int64, email string) error {
	tag, err := r.db.Exec(ctx, `UPDATE app_user SET email = $2 WHERE id = $1`, id, email)
	if err != nil {
		return userErr(err)
	}
	return database.Affected(tag, nil, "user_not_found")
}
