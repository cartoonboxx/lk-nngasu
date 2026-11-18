// Package assignments владеет таблицами assignment, assignment_file, submission, submission_file.
// Оценка хранится прямо в submission (отдельной таблицы оценок нет): один ответ на пару
// «задание — студент», проверка преподавателем переводит его в статус reviewed.
package assignments

import "time"

// MaxFiles — сколько файлов можно прикрепить к заданию или к ответу (AssignmentInput, SubmissionInput).
const MaxFiles = 5

// Статусы, которые видит клиент. Первые три совпадают с enum submission_state в БД;
// missing вычисляется: ответа нет, а дедлайн прошёл.
const (
	StatusNotSubmitted = "not_submitted" // не прикреплено (дедлайн не просрочен)
	StatusMissing      = "missing"       // отсутствует (дедлайн просрочен)
	StatusSubmitted    = "submitted"     // принято на проверку
	StatusReviewed     = "reviewed"      // оценено
)

// FileInfo — файл задания или ответа (схема FileInfo, как у материалов).
type FileInfo struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	MimeType  string `json:"mime_type"`
	SizeBytes int64  `json:"size_bytes"`
	key       string // ключ хранения: по нему досчитывается размер старых файлов без size_bytes
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

// Assignment — задание (схема Assignment). В одной теме — не более одного задания.
type Assignment struct {
	ID          int64         `json:"id"`
	CourseID    int64         `json:"course_id"`
	TopicID     int64         `json:"topic_id"`
	CreatedBy   int64         `json:"-"` // автор (в контракте не отдаётся)
	Title       string        `json:"title"`
	Description *string       `json:"description"`
	DeadlineAt  *time.Time    `json:"deadline_at"` // null — без срока
	PublishedAt time.Time     `json:"published_at"`
	Files       []FileInfo    `json:"files"`
	Mine        *MySubmission `json:"mine"`  // свой статус студента; null для преподавателя и админа
	Stats       *Stats        `json:"stats"` // агрегаты для преподавателя курса и админа; иначе null
}

// AssignmentDetail — карточка задания (GET /assignments/:id): студенту — его полный ответ
// или null, если ответа ещё нет; преподавателю и админу — всегда null (работы — отдельным запросом).
type AssignmentDetail struct {
	Assignment
	Submission *Submission `json:"submission"`
}

type AssignmentInput struct {
	TopicID     int64
	Title       string
	Description *string
	DeadlineAt  *time.Time // nil — без срока
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
	Files         []FileInfo `json:"files"`
}

type SubmissionInput struct {
	AnswerText *string
	FileIDs    []int64
	Submit     bool // true — отправить на проверку; false — сохранить черновик (только пока не отправлено)
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
	Status        *string // not_submitted | missing | submitted | reviewed
	Late          *bool   // true — сдано после дедлайна; false — только отправленные вовремя
	HasWork       *bool   // true — есть отправленная работа; false — работы нет (черновик не в счёт)
	Limit, Offset int
}

type ReviewInput struct {
	Grade   float64 // 0.0 … 5.0 с шагом 0.1
	Comment *string
}

// GradeRow — строка таблицы оценок: одно задание одного студента, даже если ответа нет
// (status not_submitted/missing, submission_id = null). Черновик не считается работой.
type GradeRow struct {
	AssignmentID    int64      `json:"assignment_id"`
	AssignmentTitle string     `json:"assignment_title"`
	TopicTitle      string     `json:"topic_title"`
	CourseID        int64      `json:"course_id"`
	SubjectName     string     `json:"subject_name"`
	SemesterID      int64      `json:"semester_id"`
	SemesterName    string     `json:"semester_name"`
	StudentID       int64      `json:"student_id"`
	StudentName     string     `json:"student_name"`
	SubmissionID    *int64     `json:"submission_id"`
	Status          string     `json:"status"` // not_submitted | missing | submitted | reviewed
	SubmittedAt     *time.Time `json:"submitted_at"`
	Grade           *float64   `json:"grade"`
	ReviewComment   *string    `json:"review_comment"`
	ReviewedAt      *time.Time `json:"reviewed_at"`
}

// GradeFilter — выборка строк оценок; nil/пусто — без фильтра.
type GradeFilter struct {
	StudentID     *int64
	CourseID      *int64
	SemesterID    *int64
	Status        *string
	Q             string // поиск по названию задания, темы и дисциплины
	Limit, Offset int
}
