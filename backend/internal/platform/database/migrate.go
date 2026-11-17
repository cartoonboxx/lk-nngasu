package database

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// migrateLockKey — ключ advisory-блокировки: два экземпляра API, стартующие
// одновременно, не применят одну миграцию дважды.
const migrateLockKey = 7_100_200_300

// baselineMarkers — признаки миграций 001–010, применённых вручную до появления
// таблицы schema_migrations. Для базы без этой таблицы Migrate отмечает
// применёнными миграции подряд, пока признак находится, и не запускает их повторно.
// Новым миграциям (011+) признаки не нужны: их применяет только Migrate.
var baselineMarkers = map[int]string{
	1:  `to_regclass('public.app_user') IS NOT NULL`,
	2:  `to_regclass('public.chat') IS NOT NULL`,
	3:  `to_regclass('public.registration_request') IS NOT NULL`,
	4:  `to_regclass('public.auth_session') IS NOT NULL`,
	5:  `EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'registration_request' AND column_name = 'last_name')`,
	6:  `EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'app_user' AND column_name = 'passport')`,
	7:  `to_regclass('public.mail_outbox') IS NOT NULL`,
	8:  `EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'department' AND column_name = 'archived')`,
	9:  `EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'lesson' AND column_name = 'ends_at')`,
	10: `EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'file' AND column_name = 'mime_type')`,
}

var (
	migrationName = regexp.MustCompile(`^(\d+)_[\w-]+\.sql$`)
	// txLine — собственные BEGIN/COMMIT файла: Migrate оборачивает файл в свою транзакцию.
	txLine = regexp.MustCompile(`(?im)^\s*(BEGIN|COMMIT)\s*;\s*$`)
)

type migration struct {
	version int
	name    string
	sql     string
}

// Migrate применяет к базе ещё не применённые миграции из fsys по возрастанию номера.
// Каждая миграция и её отметка в schema_migrations выполняются в одной транзакции:
// при ошибке база остаётся как была, а сервер не стартует.
func Migrate(ctx context.Context, pool *pgxpool.Pool, fsys fs.FS) error {
	list, err := readMigrations(fsys)
	if err != nil {
		return err
	}

	conn, err := pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, `SELECT pg_advisory_lock($1)`, migrateLockKey); err != nil {
		return err
	}
	defer conn.Exec(context.Background(), `SELECT pg_advisory_unlock($1)`, migrateLockKey) //nolint:errcheck

	if _, err := conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version    int PRIMARY KEY,
			name       text NOT NULL,
			applied_at timestamptz NOT NULL DEFAULT now()
		)`); err != nil {
		return fmt.Errorf("schema_migrations: %w", err)
	}

	applied := map[int]bool{}
	rows, err := conn.Query(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			rows.Close()
			return err
		}
		applied[v] = true
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	if len(applied) == 0 {
		if err := baseline(ctx, conn, list, applied); err != nil {
			return err
		}
	}

	for _, m := range list {
		if applied[m.version] {
			continue
		}
		body := txLine.ReplaceAllString(m.sql, "")
		script := "BEGIN;\n" + body + "\n;\nINSERT INTO schema_migrations (version, name) VALUES (" +
			strconv.Itoa(m.version) + ", " + quote(m.name) + ");\nCOMMIT;"
		// Простой протокол: файл может содержать несколько команд.
		if _, err := conn.Conn().PgConn().Exec(ctx, script).ReadAll(); err != nil {
			conn.Conn().PgConn().Exec(context.Background(), "ROLLBACK").ReadAll() //nolint:errcheck
			return fmt.Errorf("migration %s: %w", m.name, err)
		}
		slog.Info("migration applied", "name", m.name)
	}
	return nil
}

// baseline отмечает миграции, применённые вручную до появления schema_migrations:
// идёт по номерам подряд, пока в базе находится признак миграции.
// Пустая база (нет app_user) ничего не отмечает — к ней применятся все файлы.
func baseline(ctx context.Context, conn *pgxpool.Conn, list []migration, applied map[int]bool) error {
	for _, m := range list {
		marker, known := baselineMarkers[m.version]
		if !known {
			return nil
		}
		var present bool
		if err := conn.QueryRow(ctx, `SELECT `+marker).Scan(&present); err != nil {
			return fmt.Errorf("baseline %s: %w", m.name, err)
		}
		if !present {
			return nil
		}
		if _, err := conn.Exec(ctx,
			`INSERT INTO schema_migrations (version, name) VALUES ($1, $2)`, m.version, m.name); err != nil {
			return err
		}
		applied[m.version] = true
		slog.Info("migration already present, marked as applied", "name", m.name)
	}
	return nil
}

func readMigrations(fsys fs.FS) ([]migration, error) {
	names, err := fs.Glob(fsys, "*.sql")
	if err != nil {
		return nil, err
	}
	var list []migration
	seen := map[int]string{}
	for _, name := range names {
		m := migrationName.FindStringSubmatch(name)
		if m == nil {
			return nil, fmt.Errorf("migration %q: name must look like 011_add_news.sql", name)
		}
		v, _ := strconv.Atoi(m[1])
		if prev, dup := seen[v]; dup {
			return nil, fmt.Errorf("migrations %s and %s have the same number", prev, name)
		}
		seen[v] = name
		body, err := fs.ReadFile(fsys, name)
		if err != nil {
			return nil, err
		}
		list = append(list, migration{version: v, name: name, sql: string(body)})
	}
	sort.Slice(list, func(i, j int) bool { return list[i].version < list[j].version })
	return list, nil
}

func quote(s string) string { return "'" + strings.ReplaceAll(s, "'", "''") + "'" }
