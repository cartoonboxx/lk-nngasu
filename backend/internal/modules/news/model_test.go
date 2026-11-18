package news

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestExcerpt(t *testing.T) {
	cases := map[string]string{
		"Приглашаем студентов принять участие.\n\nРегистрация открыта.": "Приглашаем студентов принять участие.",
		"\n\n  Первая строка\nвторая   строка\n\nдальше":                "Первая строка вторая строка",
		"Одна строка": "Одна строка",
	}
	for in, want := range cases {
		if got := Excerpt(in); got != want {
			t.Errorf("Excerpt(%q) = %q, want %q", in, got, want)
		}
	}
	long := Excerpt(strings.Repeat("слово ", 100))
	if n := utf8.RuneCountInString(long); n > excerptRunes || !strings.HasSuffix(long, "…") {
		t.Errorf("long excerpt: %d runes, %q", n, long[len(long)-10:])
	}
}
