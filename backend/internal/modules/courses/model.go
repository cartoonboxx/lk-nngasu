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
	AcademicYearStart int       `json:"academic_year_start"`
	SemesterNumber    int       `json:"semester_number"`
	CreatedAt         time.Time `json:"created_at"`
}

type Workload struct {
	ID           int64    `json:"id"`
	CourseID     int64    `json:"course_id"`
	TeacherID    int64    `json:"teacher_id"`
	TeacherName  string   `json:"teacher_name"`
	LessonType   string   `json:"lesson_type"`
	PlannedHours *float64 `json:"planned_hours"`
}

type CourseDetail struct {
	Course
	Workloads []Workload `json:"workloads"`
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
	CurrentSemester bool   // только курсы текущего семестра
	ForTeacher      *int64 // курсы, где у этого преподавателя есть нагрузка
	ForStudent      *int64 // курсы группы этого студента
}
