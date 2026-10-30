// Package content владеет таблицами topic и material (темы курса и учебные материалы).
package content

import "time"

type Material struct {
	ID          int64     `json:"id"`
	TopicID     int64     `json:"topic_id"`
	CourseID    int64     `json:"-"`
	Title       string    `json:"title"`
	FileID      *int64    `json:"file_id"`
	FileName    *string   `json:"file_name"`
	ExternalURL *string   `json:"external_url"`
	Position    int       `json:"position"`
	CreatedBy   int64     `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
}

type Topic struct {
	ID          int64      `json:"id"`
	CourseID    int64      `json:"course_id"`
	Title       string     `json:"title"`
	Description *string    `json:"description"`
	Position    int        `json:"position"`
	Materials   []Material `json:"materials"`
}

// Position == nil при создании означает «в конец списка», при обновлении — «не менять».
type TopicInput struct {
	Title       string
	Description *string
	Position    *int
}

type MaterialInput struct {
	Title       string
	FileID      *int64
	ExternalURL *string
	Position    *int
}
