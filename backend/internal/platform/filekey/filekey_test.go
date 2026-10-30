package filekey

import "testing"

func TestKeyRoundTrip(t *testing.T) {
	key, err := New(42, "отчёт.pdf")
	if err != nil {
		t.Fatal(err)
	}
	if got := Name(key); got != "отчёт.pdf" {
		t.Errorf("Name = %q", got)
	}
	if id, ok := Owner(key); !ok || id != 42 {
		t.Errorf("Owner = %d, %v", id, ok)
	}
	if key[:len(Prefix(42))] != Prefix(42) {
		t.Errorf("key %q must start with %q", key, Prefix(42))
	}
	other, _ := New(42, "отчёт.pdf")
	if other == key {
		t.Error("keys must be unique")
	}
}

func TestOwnerRejectsGarbage(t *testing.T) {
	for _, k := range []string{"", "abc", "x/y/z", "0/a/b", "-3/a/b", "7"} {
		if id, ok := Owner(k); ok {
			t.Errorf("Owner(%q) = %d, want !ok", k, id)
		}
	}
}
