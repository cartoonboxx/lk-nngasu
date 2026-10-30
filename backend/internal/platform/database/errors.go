package database

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"university-api/internal/platform/apperr"
)

const (
	CodeUnique    = "23505"
	CodeFK        = "23503"
	CodeCheck     = "23514"
	CodeExclusion = "23P01"
)

func PgCode(err error) string {
	var pe *pgconn.PgError
	if errors.As(err, &pe) {
		return pe.Code
	}
	return ""
}

// ConstraintName возвращает имя нарушенного ограничения (для точных сообщений об ошибках).
func ConstraintName(err error) string {
	var pe *pgconn.PgError
	if errors.As(err, &pe) {
		return pe.ConstraintName
	}
	return ""
}

func noRows(tag pgconn.CommandTag, notFoundCode string) error {
	if tag.RowsAffected() == 0 {
		return apperr.NewNotFound(notFoundCode)
	}
	return nil
}

// Translate превращает типичные ошибки PostgreSQL в доменные apperr.Error.
// Остальные ошибки возвращаются как есть (httpx.Fail отдаст 500 и залогирует).
func Translate(err error) error {
	if err == nil {
		return nil
	}
	var ae *apperr.Error
	if errors.As(err, &ae) {
		return err
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return apperr.New(apperr.NotFound, "not_found", "record not found")
	}
	switch PgCode(err) {
	case CodeUnique:
		return apperr.New(apperr.Conflict, "already_exists", "record with these values already exists")
	case CodeFK:
		return apperr.New(apperr.Invalid, "invalid_reference", "referenced record does not exist")
	case CodeCheck:
		return apperr.New(apperr.Invalid, "constraint_violation", "data violates a constraint")
	case CodeExclusion:
		return apperr.New(apperr.Conflict, "time_conflict", "resource is already occupied at this time")
	case "22P02", "22007", "22008", "22003", "22001":
		return apperr.New(apperr.Invalid, "bad_value", "one of the values has an invalid format")
	}
	return err
}

// TranslateDelete — как Translate, но FK-ошибка при удалении означает «запись используется».
func TranslateDelete(err error) error {
	if PgCode(err) == CodeFK {
		return apperr.New(apperr.Conflict, "in_use", "record is referenced by other data")
	}
	return Translate(err)
}
