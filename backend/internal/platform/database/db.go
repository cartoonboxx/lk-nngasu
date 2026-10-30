package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// DB — то, что умеют и пул соединений, и транзакция. Репозитории принимают DB там,
// где один и тот же запрос выполняется и вне транзакции, и внутри неё.
type DB interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// Beginner — пул, способный открыть транзакцию.
type Beginner interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

// InTx выполняет fn в транзакции: коммит при успехе, откат при ошибке или панике.
func InTx(ctx context.Context, b Beginner, fn func(tx pgx.Tx) error) error {
	tx, err := b.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck // после Commit откат — no-op
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// Collect выполняет запрос и собирает все строки в слайс (пустой, а не nil — для JSON "[]").
func Collect[T any](ctx context.Context, db DB, scan func(pgx.Row) (T, error), sql string, args ...any) ([]T, error) {
	rows, err := db.Query(ctx, sql, args...)
	if err != nil {
		return nil, Translate(err)
	}
	defer rows.Close()

	out := make([]T, 0)
	for rows.Next() {
		v, err := scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, Translate(rows.Err())
}

// Affected переводит результат UPDATE/DELETE в ошибку: сбой БД или «ни одной строки».
func Affected(tag pgconn.CommandTag, err error, notFoundCode string) error {
	if err != nil {
		return Translate(err)
	}
	return noRows(tag, notFoundCode)
}

// AffectedDelete — как Affected, но FK-нарушение означает «запись используется».
func AffectedDelete(tag pgconn.CommandTag, err error, notFoundCode string) error {
	if err != nil {
		return TranslateDelete(err)
	}
	return noRows(tag, notFoundCode)
}

// FullName — SQL-выражение «Фамилия Имя Отчество» для таблицы app_user с алиасом alias.
func FullName(alias string) string {
	return fmt.Sprintf("%[1]s.last_name || ' ' || %[1]s.first_name || COALESCE(' ' || %[1]s.middle_names, '')", alias)
}

// FileName — SQL-выражение «имя файла для скачивания» по таблице file с алиасом alias:
// последний сегмент storage_key (см. internal/platform/filekey).
func FileName(alias string) string {
	return fmt.Sprintf("substring(%s.storage_key from '[^/]*$')", alias)
}
