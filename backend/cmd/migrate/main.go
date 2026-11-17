// Команда migrate применяет новые миграции к $DATABASE_URL и завершается.
// То же самое делает API при старте (если не задано MIGRATE_ON_START=false);
// команда нужна, чтобы обновить базу без запуска сервера.
package main

import (
	"context"
	"log/slog"
	"os"

	"university-api/internal/platform/database"
	"university-api/migrations"
)

func main() {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		slog.Error("DATABASE_URL is required")
		os.Exit(1)
	}
	ctx := context.Background()
	pool, err := database.NewPool(ctx, url)
	if err != nil {
		slog.Error("connect", "err", err)
		os.Exit(1)
	}
	defer pool.Close()
	if err := database.Migrate(ctx, pool, migrations.FS); err != nil {
		slog.Error("migrate", "err", err)
		os.Exit(1)
	}
	slog.Info("database is up to date")
}
