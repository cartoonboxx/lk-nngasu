// Package migrations встраивает SQL-миграции в бинарник API,
// чтобы сервер применял их сам при старте (см. database.Migrate).
package migrations

import "embed"

// FS — все файлы NNN_name.sql этой папки.
//
//go:embed *.sql
var FS embed.FS
