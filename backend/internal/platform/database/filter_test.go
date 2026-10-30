package database

import "testing"

func TestFilter(t *testing.T) {
	var f Filter
	f.Raw("a IS NULL")
	f.Add("b = ?", 1)
	f.Add("(c ILIKE ? OR d ILIKE ?)", "%x%")
	lim := f.Arg(10)

	if got, want := f.Where(), " WHERE a IS NULL AND b = $1 AND (c ILIKE $2 OR d ILIKE $2)"; got != want {
		t.Fatalf("where = %q, want %q", got, want)
	}
	if lim != "$3" || len(f.Args()) != 3 {
		t.Fatalf("limit placeholder %q, args %v", lim, f.Args())
	}
	if (&Filter{}).Where() != "" {
		t.Fatal("empty filter must produce empty WHERE")
	}
}
