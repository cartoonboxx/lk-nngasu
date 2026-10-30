package schedule

import (
	"time"
)

// Calendar — правила календаря: какая неделя верхняя и сколько длится пара.
type Calendar struct {
	loc           *time.Location
	weekOneMonday time.Time // понедельник первой недели; нулевое значение — чётность ISO-недель
	lessonLen     time.Duration
}

// NewCalendar: weekOneDate — любая дата первой (верхней) недели; нулевая — верхней считается
// нечётная ISO-неделя.
func NewCalendar(loc *time.Location, weekOneDate time.Time, lessonLen time.Duration) Calendar {
	if loc == nil {
		loc = time.UTC
	}
	c := Calendar{loc: loc, lessonLen: lessonLen}
	if !weekOneDate.IsZero() {
		c.weekOneMonday = mondayOf(civil(weekOneDate))
	}
	return c
}

// civil отбрасывает время суток: возвращает полночь UTC той же календарной даты.
// Все вычисления недель идут по датам, поэтому переход на летнее время на результат не влияет.
func civil(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// isoDay: понедельник = 1 … воскресенье = 7.
func isoDay(d time.Time) int {
	if wd := int(d.Weekday()); wd != 0 {
		return wd
	}
	return 7
}

func mondayOf(d time.Time) time.Time { return d.AddDate(0, 0, -(isoDay(d) - 1)) }

// Today — сегодняшняя дата в часовом поясе вуза.
func (c Calendar) Today() time.Time { return civil(time.Now().In(c.loc)) }

// WeekNumber возвращает 1 (верхняя) или 2 (нижняя) неделю для даты.
func (c Calendar) WeekNumber(date time.Time) int {
	monday := mondayOf(civil(date))
	if c.weekOneMonday.IsZero() {
		if _, w := monday.ISOWeek(); w%2 == 1 {
			return 1
		}
		return 2
	}
	weeks := int(monday.Sub(c.weekOneMonday).Hours()/24) / 7
	return 1 + ((weeks%2)+2)%2
}

// Week описывает неделю, на которую приходится дата.
func (c Calendar) Week(date time.Time) WeekInfo {
	date = civil(date)
	n := c.WeekNumber(date)
	pos := "upper"
	if n == 2 {
		pos = "lower"
	}
	return WeekInfo{Date: date.Format(time.DateOnly), WeekNumber: n, Position: pos, DayOfWeek: isoDay(date)}
}

// EndTime — время окончания пары "HH:MM" по времени начала "HH:MM".
func (c Calendar) EndTime(start string) string {
	t, err := time.Parse("15:04", start)
	if err != nil {
		return start
	}
	return t.Add(c.lessonLen).Format("15:04")
}

// parseClock проверяет время "H:MM"/"HH:MM" и приводит его к виду "HH:MM".
func parseClock(s string) (string, bool) {
	t, err := time.Parse("15:04", s)
	if err != nil {
		return "", false
	}
	return t.Format("15:04"), true
}

// expand разворачивает шаблон на даты from..to включительно: каждое занятие появляется
// в те даты, чьи неделя и день недели совпадают с его week_number/day_of_week.
func (c Calendar) expand(lessons []Lesson, from, to time.Time) []Lesson {
	out := make([]Lesson, 0)
	for d := civil(from); !d.After(civil(to)); d = d.AddDate(0, 0, 1) {
		week, day := c.WeekNumber(d), isoDay(d)
		date := d.Format(time.DateOnly)
		for _, l := range lessons {
			if l.WeekNumber == week && l.DayOfWeek == day {
				l.Date = &date // копия l: у каждого вхождения свой указатель на строку
				out = append(out, l)
			}
		}
	}
	return out
}
