// Package content владеет таблицами topic и material (темы курса и учебные материалы).
// Читать темы может любой, у кого есть доступ к курсу; создавать, менять и удалять —
// только преподаватели и администраторы. Файлы материалов загружаются через модуль files,
// здесь хранятся только их id.
package content

// Material — учебный материал внутри темы: файл и/или внешняя ссылка.
type Material struct {
	ID       int64 `json:"id"`
	TopicID  int64 `json:"topic_id"`
	CourseID int64 `json:"-"` // курс темы — нужен сервису для проверки прав, клиенту не отдаётся

	// Источник материала: файл, ссылка или оба (хотя бы один обязателен).
	File        *FileInfo `json:"file"`         // привязанный файл; null — только ссылка
	ExternalURL *string   `json:"external_url"` // только http(s)

	FileID  *int64 `json:"-"` // id привязанного файла — для правил редактирования
	fileKey string // ключ хранения: размер старых файлов берётся с диска
}

// Topic — тема курса с вложенными материалами.
// FileInfo — метаданные файла: имя, тип и размер. Скачивание — GET /files/:id
// с авторизацией (приватные файлы фронт получает через fetch → Blob).
type FileInfo struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	MimeType  string `json:"mime_type"`
	SizeBytes int64  `json:"size_bytes"`
}

type Topic struct {
	ID          int64   `json:"id"`
	CourseID    int64   `json:"course_id"`
	Title       string  `json:"title"`
	Description *string `json:"description"` // nil — описания нет

	// Materials — материалы темы по id.
	// Никогда не nil: тема без материалов отдаётся как "materials": [].
	Materials []Material `json:"materials"`
}

// TopicInput — входные данные для создания и изменения темы.
// Поля должны совпадать с topicRequest в handler: он приводится к TopicInput напрямую.
//
// Description == nil при обновлении стирает описание (обновление — полная замена).
type TopicInput struct {
	Title       string
	Description *string
}

// MaterialInput — входные данные для создания и изменения материала.
// Поля должны совпадать с materialRequest в handler: он приводится к MaterialInput напрямую.
//
// Нужен FileID или ExternalURL (можно оба) — проверяет сервис.
// При обновлении непереданные FileID и ExternalURL обнуляются.
type MaterialInput struct {
	FileID      *int64
	ExternalURL *string
}
