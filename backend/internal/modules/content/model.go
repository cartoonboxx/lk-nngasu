// Package content владеет таблицами topic и material (темы курса и учебные материалы).
// Читать темы может любой, у кого есть доступ к курсу; создавать, менять и удалять —
// только преподаватели и администраторы. Файлы материалов загружаются через модуль files,
// здесь хранятся только их id.
package content

import "time"

// Material — учебный материал внутри темы: файл и/или внешняя ссылка.
type Material struct {
	ID       int64  `json:"id"`
	TopicID  int64  `json:"topic_id"`
	CourseID int64  `json:"-"` // курс темы — нужен сервису для проверки прав, клиенту не отдаётся
	Title    string `json:"title"`

	// Источник материала: файл, ссылка или оба (хотя бы один обязателен).
	FileID      *int64  `json:"file_id"`      // nil — у материала нет файла
	FileName    *string `json:"file_name"`    // имя файла, извлекается из ключа; nil, если файла нет
	ExternalURL *string `json:"external_url"` // только http(s)

	Position  int       `json:"position"`   // порядок внутри темы (с 1); при равенстве — по id
	CreatedBy int64     `json:"created_by"` // преподаватель, добавивший материал
	CreatedAt time.Time `json:"created_at"`
}

// Topic — тема курса с вложенными материалами.
type Topic struct {
	ID          int64   `json:"id"`
	CourseID    int64   `json:"course_id"`
	Title       string  `json:"title"`
	Description *string `json:"description"` // nil — описания нет
	Position    int     `json:"position"`    // порядок в курсе (с 1); при равенстве — по id

	// Materials — материалы темы, отсортированные по position.
	// Никогда не nil: тема без материалов отдаётся как "materials": [].
	Materials []Material `json:"materials"`
}

// TopicInput — входные данные для создания и изменения темы.
// Поля должны совпадать с topicRequest в handler: он приводится к TopicInput напрямую.
//
// Position == nil при создании означает «в конец списка», при обновлении — «не менять».
// Description == nil при обновлении стирает описание (обновление — полная замена).
type TopicInput struct {
	Title       string
	Description *string
	Position    *int
}

// MaterialInput — входные данные для создания и изменения материала.
// Поля должны совпадать с materialRequest в handler: он приводится к MaterialInput напрямую.
//
// Нужен FileID или ExternalURL (можно оба) — проверяет сервис.
// При обновлении непереданные FileID и ExternalURL обнуляются;
// Position == nil — как у темы: при создании в конец, при обновлении не менять.
type MaterialInput struct {
	Title       string
	FileID      *int64
	ExternalURL *string
	Position    *int
}
