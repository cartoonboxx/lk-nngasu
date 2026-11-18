// createadmin — резервная серверная команда; основной сценарий — /setup и приглашения владельца.
//
//	ADMIN_EMAIL=admin@example.com ADMIN_PASSWORD='...' \
//	DATABASE_URL=postgres://... go run ./cmd/createadmin
//
// Необязательно: ADMIN_LAST_NAME, ADMIN_FIRST_NAME.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/mail"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"

	"university-api/internal/platform/database"
)

func must(key string) string {
	v := os.Getenv(key)
	if v == "" {
		slog.Error("missing environment variable", "name", key)
		os.Exit(1)
	}
	return v
}

func orDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	dbURL := must("DATABASE_URL")
	email := strings.ToLower(strings.TrimSpace(must("ADMIN_EMAIL")))
	password := must("ADMIN_PASSWORD") // из env, а не из флага: аргументы видны в `ps`

	if _, err := mail.ParseAddress(email); err != nil {
		slog.Error("ADMIN_EMAIL is not a valid email")
		os.Exit(1)
	}
	if len(password) < 8 || len(password) > 72 {
		slog.Error("ADMIN_PASSWORD must be 8..72 bytes long")
		os.Exit(1)
	}
	lastName := orDefault("ADMIN_LAST_NAME", "Администратор")
	firstName := orDefault("ADMIN_FIRST_NAME", "Системный")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pool, err := database.NewPool(ctx, dbURL)
	if err != nil {
		slog.Error("database", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		slog.Error("bcrypt", "err", err)
		os.Exit(1)
	}

	var id int64
	err = database.InTx(ctx, pool, func(tx pgx.Tx) error {
		var completed *time.Time
		if err := tx.QueryRow(ctx, `SELECT completed_at FROM admin_setup WHERE id FOR UPDATE`).Scan(&completed); err != nil {
			return err
		}
		if err := tx.QueryRow(ctx, `INSERT INTO app_user
			(last_name, first_name, email, password_hash, is_admin, can_manage_admins)
			VALUES ($1, $2, $3, $4, true, $5) RETURNING id`,
			lastName, firstName, email, string(hash), completed == nil).Scan(&id); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, `INSERT INTO admin_access_audit (target_id, target_email, action)
			VALUES ($1, $2, 'cli_create')`, id, email)
		return err
	})
	if err != nil {
		if database.PgCode(err) == database.CodeUnique {
			slog.Error("this email is already taken")
		} else {
			slog.Error("insert", "err", err)
		}
		os.Exit(1)
	}
	fmt.Printf("admin created: id=%d email=%s\n", id, email)
}
