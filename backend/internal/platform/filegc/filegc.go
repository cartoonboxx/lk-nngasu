// Package filegc удаляет записи о файлах, на которые больше никто не ссылается.
//
// Таблица file — общий реестр, а ссылаются на неё material, assignment_file, submission_file
// и аватар пользователя. Когда ссылка исчезает (удалили задание, заменили вложение),
// файл должен исчезнуть и с диска — иначе хранилище растёт вечно.
package filegc

import (
	"context"

	"github.com/jackc/pgx/v5"

	"university-api/internal/platform/database"
)

func scanKey(row pgx.Row) (string, error) {
	var key string
	err := row.Scan(&key)
	return key, err
}

// Purge удаляет из ids те файлы, на которые не осталось ссылок, и возвращает их ключи.
// Вызывать внутри той же транзакции, что удалила ссылки; удалять файлы с диска —
// после коммита (по возвращённым ключам).
func Purge(ctx context.Context, db database.DB, ids []int64) ([]string, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	return database.Collect(ctx, db, scanKey, `
		DELETE FROM file f
		WHERE f.id = ANY($1::bigint[])
		  AND NOT EXISTS (SELECT 1 FROM material        WHERE file_id = f.id)
		  AND NOT EXISTS (SELECT 1 FROM assignment_file WHERE file_id = f.id)
		  AND NOT EXISTS (SELECT 1 FROM submission_file WHERE file_id = f.id)
		  AND NOT EXISTS (SELECT 1 FROM app_user        WHERE avatar_file_id = f.id)
		RETURNING f.storage_key`, ids)
}
