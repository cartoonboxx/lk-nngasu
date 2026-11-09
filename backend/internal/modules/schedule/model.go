// Package schedule владеет таблицей lesson и отдаёт расписание.
//
// В БД расписание — двухнедельный шаблон (week_number 1/2, день недели, время начала).
// Конкретные даты API получает разворачиванием шаблона на запрошенный период.
package schedule

import "time"

// Lesson — занятие в том виде, в каком отдаётся наружу.
// Собирается из lesson и связанных таблиц (нагрузка, дисциплина группы,
// дисциплина, группа, преподаватель, аудитория). Названия и ФИО
// подтягиваются для отображения, чтобы фронту не делать доп. запросы.
//
// В режиме шаблона это одна строка сетки; в режиме дат одно и то же
// шаблонное занятие повторяется на каждую подходящую дату с заполненным Date
// (ID при этом одинаковый).
type Lesson struct {
	ID          int64   `json:"id"`
	WorkloadID  int64   `json:"workload_id"` // строка нагрузки, из которой поставлено занятие
	CourseID    int64   `json:"course_id"`   // дисциплина группы в семестре (course_offering)
	Subject     string  `json:"subject"`     // название дисциплины
	LessonType  string  `json:"lesson_type"` // вид занятия из нагрузки (лекция, практика, лабораторная)
	TeacherID   int64   `json:"teacher_id"`
	TeacherName string  `json:"teacher_name"` // ФИО преподавателя
	GroupID     int64   `json:"group_id"`
	GroupName   string  `json:"group_name"`
	WeekNumber  int     `json:"week_number"` // 1 — верхняя, 2 — нижняя
	DayOfWeek   int     `json:"day_of_week"` // 1 (Пн) .. 6 (Сб)
	StartsAt    string  `json:"starts_at"`   // "08:30"
	EndsAt      string  `json:"ends_at"`     // начало + длительность пары из конфигурации; в БД не хранится
	RoomID      int64   `json:"room_id"`
	Room        string  `json:"room"`           // номер аудитории
	Date        *string `json:"date,omitempty"` // "YYYY-MM-DD"; заполняется только при запросе по датам
}

// LessonInput — редактируемые поля занятия. workload_id задаётся при создании и не меняется.
// Значения проверяет Service.check; StartsAt после проверки приведён к "HH:MM".
type LessonInput struct {
	WeekNumber int    // 1 — верхняя, 2 — нижняя
	DayOfWeek  int    // 1 (Пн) .. 6 (Сб)
	StartsAt   string // "08:30"
	RoomID     int64
}

// LessonFilter — выборка шаблонных занятий. Без SemesterID берётся текущий семестр.
// Все поля необязательны и сочетаются через AND; nil — без фильтра.
// GroupID и TeacherID приходят из пути публичных маршрутов,
// StudentID и TeacherID в «моём расписании» — только из JWT, не из запроса.
type LessonFilter struct {
	SemesterID *int64
	GroupID    *int64
	TeacherID  *int64
	StudentID  *int64 // занятия группы этого студента
	Week, Day  *int   // срез по сетке; заполняются сервисом из Query
}

// Query — временной срез: либо шаблон (week/day), либо диапазон дат (from/to).
// Смешивать нельзя; from и to задаются только вместе (проверяет Service.Lessons).
type Query struct {
	Week, Day *int
	From, To  *time.Time // обе границы включительно
}

// GroupRef — группа в публичном списке для выбора расписания.
type GroupRef struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// TeacherRef — преподаватель в публичном списке для выбора расписания.
// Содержит только то, что можно показать гостю: без email и прочих личных данных.
type TeacherRef struct {
	ID             int64  `json:"id"` // user_id преподавателя
	FullName       string `json:"full_name"`
	DepartmentName string `json:"department_name"`
}

// WeekInfo — какая сегодня (или в указанную дату) неделя.
type WeekInfo struct {
	Date       string `json:"date"`        // "YYYY-MM-DD", для которой посчитан ответ
	WeekNumber int    `json:"week_number"` // 1 или 2, как в Lesson.WeekNumber
	Position   string `json:"position"`    // upper | lower — то же, что WeekNumber, словом для фронта
	DayOfWeek  int    `json:"day_of_week"` // 1 (Пн) .. 7 (Вс); воскресенье есть, хотя занятий в него нет
}
