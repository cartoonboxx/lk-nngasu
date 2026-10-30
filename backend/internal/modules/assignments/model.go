// Package assignments владеет таблицами assignment, assignment_file, submission, submission_file.
// Оценка хранится прямо в submission (отдельной таблицы оценок нет): один ответ на пару
// «задание — студент», проверка преподавателем переводит его в статус reviewed.
package assignments

import "time"

// MaxFiles — сколько файлов можно прикрепить к заданию или к ответу (FR-HW-02, FR-HW-06).
const MaxFiles = 5

// Статусы, которые видит клиент. Первые три совпадают с enum submission_state в БД;
// missing вычисляется: ответа нет, а дедлайн прошёл.
const (
	StatusNotSubmitted = "not_submitted" // не прикреплено (дедлайн не просрочен)
	StatusMissing      = "missing"       // отсутствует (дедлайн просрочен)
	StatusSubmitted    = "submitted"     // принято на проверку
	StatusReviewed     = "reviewed"      // оценено
)

type FileRef struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// MySubmission — сводка по ответу студента на задание (в списке и карточке задания).
type MySubmission struct {
	Status      string     `json:"status"`
	SubmittedAt *time.Time `json:"submitted_at"`
	IsLate      bool       `json:"is_late"`
	Grade       *float64   `json:"grade"`
}

// Stats — сводка для преподавателя: сколько студентов сдало и сколько проверено.
type Stats struct {
	StudentsTotal int `json:"students_total"`
	Submitted     int `json:"submitted"` // отправили ответ (включая уже проверенные)
	Reviewed      int `json:"reviewed"`
}

type Assignment struct {
	ID          int64         `json:"id"`
	CourseID    int64         `json:"course_id"`
	TopicID     int64         `json:"topic_id"`
	CreatedBy   int64         `json:"created_by"`
	Title       string        `json:"title"`
	Description *string       `json:"description"`
	DeadlineAt  *time.Time    `json:"deadline_at"`
	PublishedAt time.Time     `json:"published_at"`
	Files       []FileRef     `json:"files"`
	Mine        *MySubmission `json:"mine,omitempty"`  // только для студента
	Stats       *Stats        `json:"stats,omitempty"` // только для преподавателя и админа
}

type AssignmentInput struct {
	TopicID     int64
	Title       string
	Description *string
	DeadlineAt  time.Time
	FileIDs     []int64
}

type Submission struct {
	ID            int64      `json:"id"`
	AssignmentID  int64      `json:"assignment_id"`
	CourseID      int64      `json:"course_id"`
	StudentID     int64      `json:"student_id"`
	StudentName   string     `json:"student_name"`
	Status        string     `json:"status"` // not_submitted (черновик) | submitted | reviewed
	AnswerText    *string    `json:"answer_text"`
	SubmittedAt   *time.Time `json:"submitted_at"`
	IsLate        bool       `json:"is_late"`
	ReviewComment *string    `json:"review_comment"`
	ReviewedBy    *int64     `json:"reviewed_by"`
	ReviewedAt    *time.Time `json:"reviewed_at"`
	Grade         *float64   `json:"grade"`
	Files         []FileRef  `json:"files"`
}

type SubmissionInput struct {
	AnswerText *string
	FileIDs    []int64
	Submit     bool // true — отправить на проверку; false — сохранить черновик
}

// SubmissionRow — строка списка ответов: по одной на каждого студента группы, даже без ответа.
type SubmissionRow struct {
	StudentID        int64      `json:"student_id"`
	StudentName      string     `json:"student_name"`
	RecordBookNumber *string    `json:"record_book_number"`
	SubmissionID     *int64     `json:"submission_id"`
	Status           string     `json:"status"` // not_submitted | missing | submitted | reviewed
	SubmittedAt      *time.Time `json:"submitted_at"`
	IsLate           bool       `json:"is_late"`
	Grade            *float64   `json:"grade"`
}

// SubmissionFilter — фильтры списка ответов (nil — не фильтровать).
type SubmissionFilter struct {
	Status  *string // not_submitted | missing | submitted | reviewed
	Late    *bool   // true — сдано после дедлайна; false — сдано вовремя
	HasWork *bool   // true — есть отправленная работа; false — работы нет
}

type ReviewInput struct {
	Grade   float64 // 0.0 … 5.0 с шагом 0.1
	Comment *string
}

// GradeRow — одна выставленная оценка.
type GradeRow struct {
	AssignmentID    int64     `json:"assignment_id"`
	AssignmentTitle string    `json:"assignment_title"`
	CourseID        int64     `json:"course_id"`
	SubjectName     string    `json:"subject_name"`
	StudentID       int64     `json:"student_id"`
	StudentName     string    `json:"student_name"`
	Grade           float64   `json:"grade"`
	ReviewComment   *string   `json:"review_comment"`
	ReviewedAt      time.Time `json:"reviewed_at"`
}

type GradeFilter struct {
	StudentID *int64
	CourseID  *int64
}
