package schedule

import (
	"time"
)

// Calendar — правила календаря: какая неделя верхняя и сколько длится пара.
// Значение без состояния: создаётся один раз при старте из конфигурации
// и передаётся в сервис. Копировать безопасно.
type Calendar struct {
	loc           *time.Location // часовой пояс вуза: в нём определяется «сегодня»
	weekOneMonday time.Time      // понедельник первой недели; нулевое значение — чётность ISO-недель
	lessonLen     time.Duration  // длительность пары, одна на все занятия
}

// NewCalendar: weekOneDate — любая дата первой (верхней) недели; нулевая — верхней считается
// нечётная ISO-неделя.
// loc == nil означает UTC. Время суток в weekOneDate не учитывается.
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
// Дата берётся в поясе самого t, поэтому «сегодня» нужно сначала перевести в пояс вуза (см. Today).
func civil(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// isoDay: понедельник = 1 … воскресенье = 7.
// В time.Weekday воскресенье — 0, поэтому его приходится переносить в конец.
func isoDay(d time.Time) int {
	if wd := int(d.Weekday()); wd != 0 {
		return wd
	}
	return 7
}

// mondayOf возвращает понедельник недели, в которую попадает дата.
func mondayOf(d time.Time) time.Time { return d.AddDate(0, 0, -(isoDay(d) - 1)) }

// Today — сегодняшняя дата в часовом поясе вуза.
// Без перевода в loc сервер в UTC считал бы «завтра» уже с 21:00 по Москве.
func (c Calendar) Today() time.Time { return civil(time.Now().In(c.loc)) }

// WeekNumber возвращает 1 (верхняя) или 2 (нижняя) неделю для даты.
//
// Если задан weekOneMonday, недели чередуются от него: чётное число недель
// от первой — верхняя, нечётное — нижняя. Работает и для дат раньше
// weekOneMonday: двойной остаток ((x%2)+2)%2 убирает отрицательный остаток,
// который в Go даёт % для отрицательных чисел.
// Иначе — по чётности ISO-номера недели.
func (c Calendar) WeekNumber(date time.Time) int {
	monday := mondayOf(civil(date))
	if c.weekOneMonday.IsZero() {
		if _, w := monday.ISOWeek(); w%2 == 1 {
			return 1
		}
		return 2
	}
	// Обе даты — понедельники в полночь UTC, так что разница — целое число недель.
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
// Некорректное время возвращается как есть, без паники: в БД оно уже
// проверено parseClock, так что этот случай не должен встречаться.
func (c Calendar) EndTime(start string) string {
	t, err := time.Parse("15:04", start)
	if err != nil {
		return start
	}
	return t.Add(c.lessonLen).Format("15:04")
}

// parseClock проверяет время "H:MM"/"HH:MM" и приводит его к виду "HH:MM".
// В раскладке "15:04" час допускает одну или две цифры, минуты — строго две;
// значения вне диапазона ("25:00", "08:60") отклоняются.
func parseClock(s string) (string, bool) {
	t, err := time.Parse("15:04", s)
	if err != nil {
		return "", false
	}
	return t.Format("15:04"), true
}

// expand разворачивает шаблон на даты from..to включительно: каждое занятие появляется
// в те даты, чьи неделя и день недели совпадают с его week_number/day_of_week.
// Результат упорядочен по датам, внутри даты — в порядке входного списка
// (то есть по времени, как его отсортировал репозиторий).
// Пустой результат — пустой срез, а не nil: в JSON уйдёт [], а не null.
func (c Calendar) expand(lessons []Lesson, from, to time.Time) []Lesson {
	out := make([]Lesson, 0)
	for d := civil(from); !d.After(civil(to)); d = d.AddDate(0, 0, 1) {
		week, day := c.WeekNumber(d), isoDay(d)
		date := d.Format(time.DateOnly) // своя переменная на каждый день
		for _, l := range lessons {
			if l.WeekNumber == week && l.DayOfWeek == day {
				// l — копия элемента, так что вход не меняется. Занятия одного дня
				// делят указатель на одну и ту же строку date; это безопасно,
				// потому что строки неизменяемы и Date дальше не меняется.
				l.Date = &date
				out = append(out, l)
			}
		}
	}
	return out
}