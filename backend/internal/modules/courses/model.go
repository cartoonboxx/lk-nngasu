// Package courses владеет таблицами course_offering и workload.
// Курс = дисциплина + учебная группа + семестр; студенты курса — это студенты группы.
package courses

import "time"

type Course struct {
	ID                int64     `json:"id"`
	SubjectID         int64     `json:"subject_id"`
	SubjectName       string    `json:"subject_name"`
	GroupID           int64     `json:"group_id"`
	GroupName         string    `json:"group_name"`
	SemesterID        int64     `json:"semester_id"`
	SemesterName      string    `json:"semester_name"`
	AcademicYearStart int       `json:"academic_year_start"`
	SemesterNumber    int       `json:"semester_number"`
	WorkloadCount     int       `json:"workload_count"` // число назначений нагрузки (для таблицы администратора)
	CreatedAt         time.Time `json:"created_at"`
}

// Workload — назначение преподавателя и вида занятия на курс.
type Workload struct {
	ID           int64    `json:"id"`
	CourseID     int64    `json:"course_id"`
	TeacherID    int64    `json:"teacher_id"`
	TeacherName  string   `json:"teacher_name"`
	LessonType   string   `json:"lesson_type"`   // lecture | practice | laboratory | consultation | exam
	PlannedHours *float64 `json:"planned_hours"` // академические часы; null — не задано
	// ScheduledHours — часы двухнедельного шаблона расписания: сумма минут занятий / 45,
	// округление до 2 знаков. Это не итог за семестр.
	ScheduledHours float64 `json:"scheduled_hours"`
}

// CourseDetail — ответ GET /courses/:id (страница предмета).
type CourseDetail struct {
	Course
	Workloads []Workload `json:"workloads"`

	// Topics — темы курса, в каждой материалы и задания. Собирается в app из модулей
	// content и assignments (см. TopicsLoader), поэтому тип здесь — any.
	Topics any `json:"topics"`
}

type CourseStudent struct {
	StudentID        int64   `json:"student_id"`
	FullName         string  `json:"full_name"`
	RecordBookNumber *string `json:"record_book_number"`
	IsHeadman        bool    `json:"is_headman"`
}

type CourseFilter struct {
	GroupID         *int64
	SemesterID      *int64
	CurrentSemester bool // только курсы текущего семестра
	// Viewer — не администратор: видит курсы своей группы (как студент) и курсы своей
	// нагрузки (как преподаватель), объединённо и без дублей. nil — все курсы (админ).
	Viewer        *int64
	Limit, Offset int
}

// ScheduleConflict — занятие, с которым пересекается расписание нагрузки
// после смены преподавателя (формат как в модуле schedule).
type ScheduleConflict struct {
	LessonID   int64    `json:"lesson_id"`
	Resources  []string `json:"resources"`
	WeekNumber int      `json:"week_number"`
	DayOfWeek  int      `json:"day_of_week"`
	StartsAt   string   `json:"starts_at"`
	EndsAt     string   `json:"ends_at"`
}
