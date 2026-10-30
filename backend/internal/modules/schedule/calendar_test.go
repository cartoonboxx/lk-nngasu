package schedule

import (
	"testing"
	"time"
)

func date(s string) time.Time {
	d, err := time.Parse(time.DateOnly, s)
	if err != nil {
		panic(err)
	}
	return d
}

func TestWeekNumberWithAnchor(t *testing.T) {
	// 2026-09-01 — вторник; неделя с понедельника 2026-08-31 — верхняя (1).
	c := NewCalendar(time.UTC, date("2026-09-01"), 90*time.Minute)
	cases := map[string]int{
		"2026-08-31": 1, // понедельник недели 1
		"2026-09-06": 1, // воскресенье недели 1
		"2026-09-07": 2,
		"2026-09-13": 2,
		"2026-09-14": 1,
		"2026-12-28": 2, // ровно 17 недель после якоря — нечётное число → нижняя
		"2026-08-24": 2, // неделя ДО якоря
		"2026-08-17": 1, // две недели до
	}
	for in, want := range cases {
		if got := c.WeekNumber(date(in)); got != want {
			t.Errorf("WeekNumber(%s) = %d, want %d", in, got, want)
		}
	}
}

func TestWeekNumberISOParityWithoutAnchor(t *testing.T) {
	c := NewCalendar(time.UTC, time.Time{}, 90*time.Minute)
	// 2026-01-05 — понедельник ISO-недели 2 → нижняя; 2026-01-12 — неделя 3 → верхняя.
	if got := c.WeekNumber(date("2026-01-05")); got != 2 {
		t.Errorf("ISO week 2 → %d, want 2", got)
	}
	if got := c.WeekNumber(date("2026-01-12")); got != 1 {
		t.Errorf("ISO week 3 → %d, want 1", got)
	}
}

func TestWeekInfo(t *testing.T) {
	c := NewCalendar(time.UTC, date("2026-09-01"), 90*time.Minute)
	w := c.Week(date("2026-09-09")) // среда нижней недели
	if w.WeekNumber != 2 || w.Position != "lower" || w.DayOfWeek != 3 || w.Date != "2026-09-09" {
		t.Errorf("unexpected %+v", w)
	}
	if got := c.Week(date("2026-09-06")).DayOfWeek; got != 7 {
		t.Errorf("Sunday must be day 7, got %d", got)
	}
}

func TestEndTimeAndParseClock(t *testing.T) {
	c := NewCalendar(time.UTC, time.Time{}, 90*time.Minute)
	if got := c.EndTime("08:30"); got != "10:00" {
		t.Errorf("EndTime = %s", got)
	}
	if got, ok := parseClock("8:30"); !ok || got != "08:30" {
		t.Errorf("parseClock(8:30) = %q, %v", got, ok)
	}
	for _, bad := range []string{"", "25:00", "08:60", "abc", "08-30"} {
		if _, ok := parseClock(bad); ok {
			t.Errorf("parseClock(%q) must fail", bad)
		}
	}
}

func TestExpand(t *testing.T) {
	c := NewCalendar(time.UTC, date("2026-09-01"), 90*time.Minute)
	lessons := []Lesson{
		{ID: 1, WeekNumber: 1, DayOfWeek: 2, StartsAt: "08:30"}, // верхняя неделя, вторник
		{ID: 2, WeekNumber: 2, DayOfWeek: 2, StartsAt: "10:10"}, // нижняя неделя, вторник
	}
	got := c.expand(lessons, date("2026-09-01"), date("2026-09-15")) // вт недели 1, вт недели 2, вт недели 1
	if len(got) != 3 {
		t.Fatalf("got %d occurrences, want 3: %+v", len(got), got)
	}
	wantDates := []string{"2026-09-01", "2026-09-08", "2026-09-15"}
	wantIDs := []int64{1, 2, 1}
	for i, l := range got {
		if l.Date == nil || *l.Date != wantDates[i] || l.ID != wantIDs[i] {
			t.Errorf("occurrence %d = id %d date %v, want id %d date %s", i, l.ID, l.Date, wantIDs[i], wantDates[i])
		}
	}
	if lessons[0].Date != nil {
		t.Error("expand must not mutate the template")
	}
}
