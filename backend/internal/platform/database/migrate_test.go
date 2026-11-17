package database

import (
	"strings"
	"testing"
	"testing/fstest"
)

func TestReadMigrationsOrderAndNames(t *testing.T) {
	list, err := readMigrations(fstest.MapFS{
		"010_b.sql": {Data: []byte("SELECT 10")},
		"002_a.sql": {Data: []byte("SELECT 2")},
		"embed.go":  {Data: []byte("package migrations")},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 || list[0].version != 2 || list[1].version != 10 {
		t.Fatalf("want 002, 010 in order, got %+v", list)
	}
}

func TestReadMigrationsRejectsBadNames(t *testing.T) {
	for name, fsys := range map[string]fstest.MapFS{
		"no number": {"add_news.sql": {}},
		"duplicate": {"011_a.sql": {}, "011_b.sql": {}},
	} {
		if _, err := readMigrations(fsys); err == nil {
			t.Errorf("%s: want error", name)
		}
	}
}

func TestOwnTransactionLinesAreStripped(t *testing.T) {
	got := txLine.ReplaceAllString("BEGIN;\nALTER TABLE x ADD y int;\n  commit ;\n-- BEGIN; в комментарии не трогаем\n", "")
	if strings.Contains(strings.ToUpper(got), "COMMIT") || strings.HasPrefix(got, "BEGIN") {
		t.Fatalf("transaction lines left: %q", got)
	}
	if !strings.Contains(got, "ALTER TABLE x ADD y int;") || !strings.Contains(got, "-- BEGIN;") {
		t.Fatalf("statements lost: %q", got)
	}
}
