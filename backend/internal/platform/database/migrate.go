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

// baselineMarkers — признаки миграций 001–011, применённых вручную до появления
// таблицы schema_migrations. Для базы без этой таблицы Migrate отмечает
// применёнными миграции подряд, пока признак находится, и не запускает их повторно.
// Учитывает также SQL, применённый docker-entrypoint-initdb.d при создании базы.
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
	11: `to_regclass('public.admin_setup') IS NOT NULL`,
	12: `EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'assignment_topic_unique')`,
	13: `EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'file' AND column_name = 'created_at')`,
	14: `to_regclass('public.news') IS NOT NULL`,
	15: `NOT EXISTS (SELECT 1 FROM chat_message WHERE deleted_at IS NOT NULL)`,
	16: `NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'department' AND column_name = 'archived')`,
	17: `NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'topic' AND column_name = 'position')`,
}

// renamedMigrations — миграции, сменившие номер при слиянии веток (011_admin_setup занял 011).
// Базы, где они уже применены под старыми номерами, переводятся на новые до запуска,
// иначе 011_admin_setup считался бы применённым. Порядок — от больших номеров к меньшим.
var renamedMigrations = []struct {
	oldVersion int
	oldName    string
	newVersion int
	newName    string
}{
	{14, "014_chat_hard_delete.sql", 15, "015_chat_hard_delete.sql"},
	{13, "013_news.sql", 14, "014_news.sql"},
	{12, "012_file_created_at.sql", 13, "013_file_created_at.sql"},
	{11, "011_assignment_per_topic.sql", 12, "012_assignment_per_topic.sql"},
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

	for _, r := range renamedMigrations {
		if _, err := conn.Exec(ctx,
			`UPDATE schema_migrations SET version = $3, name = $4 WHERE version = $1 AND name = $2`,
			r.oldVersion, r.oldName, r.newVersion, r.newName); err != nil {
			return fmt.Errorf("schema_migrations rename %s: %w", r.oldName, err)
		}
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
