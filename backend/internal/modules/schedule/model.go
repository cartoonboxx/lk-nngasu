// Package schedule владеет таблицей lesson и отдаёт расписание.
//
// В БД расписание — двухнедельный шаблон (week_number 1/2, день недели, время начала).
// Конкретные даты API получает разворачиванием шаблона на запрошенный период.
package schedule

import "time"

type Lesson struct {
	ID          int64   `json:"id"`
	WorkloadID  int64   `json:"workload_id"`
	CourseID    int64   `json:"course_id"`
	Subject     string  `json:"subject"`
	LessonType  string  `json:"lesson_type"`
	TeacherID   int64   `json:"teacher_id"`
	TeacherName string  `json:"teacher_name"`
	GroupID     int64   `json:"group_id"`
	GroupName   string  `json:"group_name"`
	WeekNumber  int     `json:"week_number"` // 1 — верхняя, 2 — нижняя
	DayOfWeek   int     `json:"day_of_week"` // 1 (Пн) .. 6 (Сб)
	StartsAt    string  `json:"starts_at"`   // "08:30"
	EndsAt      string  `json:"ends_at"`     // начало + длительность пары из конфигурации
	RoomID      int64   `json:"room_id"`
	Room        string  `json:"room"`
	Date        *string `json:"date,omitempty"` // заполняется только при запросе по датам
}

// LessonInput — редактируемые поля занятия. workload_id задаётся при создании и не меняется.
type LessonInput struct {
	WeekNumber int
	DayOfWeek  int
	StartsAt   string
	RoomID     int64
}

// LessonFilter — выборка шаблонных занятий. Без SemesterID берётся текущий семестр.
type LessonFilter struct {
	SemesterID *int64
	GroupID    *int64
	TeacherID  *int64
	StudentID  *int64 // занятия группы этого студента
	Week, Day  *int
}

// Query — временной срез: либо шаблон (week/day), либо диапазон дат (from/to).
type Query struct {
	Week, Day *int
	From, To  *time.Time
}

type GroupRef struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type TeacherRef struct {
	ID             int64  `json:"id"`
	FullName       string `json:"full_name"`
	DepartmentName string `json:"department_name"`
}

// WeekInfo — какая сегодня (или в указанную дату) неделя.
type WeekInfo struct {
	Date       string `json:"date"`
	WeekNumber int    `json:"week_number"`
	Position   string `json:"position"`    // upper | lower
	DayOfWeek  int    `json:"day_of_week"` // 1 (Пн) .. 7 (Вс)
}
