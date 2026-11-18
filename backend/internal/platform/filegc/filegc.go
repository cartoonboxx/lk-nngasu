// Package filegc удаляет записи о файлах, на которые больше никто не ссылается.
//
// Таблица file — общий реестр, а ссылаются на неё material, assignment_file, submission_file,
// chat_message_file и аватар пользователя. Когда ссылка исчезает (удалили задание, заменили
// вложение), файл должен исчезнуть и с диска — иначе хранилище растёт вечно.
// Загрузку, которую так и не привязали, через сутки удаляет Sweeper.
package filegc

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"

	"university-api/internal/platform/database"
)

// Unreferenced — SQL-условие «на файл f никто не ссылается». Единственный список ссылок
// на file: новую таблицу с file_id нужно добавить сюда, иначе её файлы будут удаляться.
const Unreferenced = `
	NOT EXISTS (SELECT 1 FROM material          WHERE file_id = f.id)
	AND NOT EXISTS (SELECT 1 FROM assignment_file   WHERE file_id = f.id)
	AND NOT EXISTS (SELECT 1 FROM submission_file   WHERE file_id = f.id)
	AND NOT EXISTS (SELECT 1 FROM chat_message_file WHERE file_id = f.id)
	AND NOT EXISTS (SELECT 1 FROM app_user          WHERE avatar_file_id = f.id)`

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
		WHERE f.id = ANY($1::bigint[]) AND `+Unreferenced+`
		RETURNING f.storage_key`, ids)
}

// Sweeper раз в Every удаляет непривязанные файлы старше Age (забытые загрузки).
type Sweeper struct {
	DB     database.DB
	Remove func(key string) error // удаление с диска
	Every  time.Duration
	Age    time.Duration
}

// Run работает до отмены ctx. Ошибки только пишутся в лог: следующий проход повторит попытку.
func (s Sweeper) Run(ctx context.Context) {
	t := time.NewTicker(s.Every)
	defer t.Stop()
	for {
		if n, err := s.Sweep(ctx); err != nil {
			slog.Error("file sweep", "err", err)
		} else if n > 0 {
			slog.Info("forgotten uploads removed", "count", n)
		}
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
	}
}

// Sweep — один проход; возвращает, сколько файлов удалено. Записи удаляются одним
// запросом, а файлы с диска — после него: упавшее удаление с диска оставит лишь
// файл без записи, но не запись без файла.
func (s Sweeper) Sweep(ctx context.Context) (int, error) {
	keys, err := database.Collect(ctx, s.DB, scanKey, `
		DELETE FROM file f
		WHERE f.created_at < now() - make_interval(secs => $1) AND `+Unreferenced+`
		RETURNING f.storage_key`, s.Age.Seconds())
	if err != nil {
		return 0, err
	}
	for _, k := range keys {
		_ = s.Remove(k)
	}
	return len(keys), nil
}
