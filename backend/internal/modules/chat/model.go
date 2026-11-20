// Package chat — личные и групповые чаты.
//
// Личный чат (direct) — переписка двух пользователей. Групповой (group) — один на учебную
// группу: в нём студенты группы и преподаватели, ведущие у неё курсы; администратор может
// заходить в любой групповой чат. Участники группового чата не хранятся, а вычисляются из
// таблиц student и workload — состав меняется сам при переводе студента или смене нагрузки.
//
// REST отдаёт контакты, список чатов и историю (GET /chats/:id/messages?before=<id>).
// Отправка — команда WebSocket message.send (протокол lk-chat.v1, см. ws.go).
// После COMMIT участникам чата уходит событие message.created.
// Редактирование и удаление отправленных сообщений недоступны.
package chat

import "time"

const (
	KindDirect = "direct"
	KindGroup  = "group"

	// MaxFiles — вложений в одном сообщении (как в ответах на задания).
	MaxFiles = 5
	// MaxBodyRunes — длина текста сообщения.
	MaxBodyRunes = 4000
	// DefaultLimit / MaxLimit — размер страницы истории.
	DefaultLimit = 50
	MaxLimit     = 100
)

type FileRef struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Person struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// Preview — последнее сообщение в списке чатов.
type Preview struct {
	ID         int64     `json:"id"`
	SenderID   int64     `json:"sender_id"`
	SenderName string    `json:"sender_name"`
	Body       string    `json:"body"`
	HasFiles   bool      `json:"has_files"`
	CreatedAt  time.Time `json:"created_at"`
}

// Chat — чат глазами конкретного пользователя.
type Chat struct {
	UnreadCount       int       `json:"unread_count"`
	LastReadMessageID int64     `json:"last_read_message_id"`
	ID                int64     `json:"id"`
	Kind              string    `json:"kind"`
	Title             string    `json:"title"`    // группа — её название, личный — имя собеседника
	GroupID           *int64    `json:"group_id"` // только у группового
	Peer              *Person   `json:"peer"`     // только у личного
	CreatedAt         time.Time `json:"created_at"`
	LastMessage       *Preview  `json:"last_message"`
}

// Contact — пользователь, которому можно написать (GET /contacts).
type Contact struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Avatar      *Avatar `json:"avatar"`
	Description *string `json:"description"` // кафедра преподавателя или группа студента
}

type Avatar struct {
	ID  int64  `json:"id"`
	URL string `json:"url"`
}

// MessagePage — страница истории: items по возрастанию id; next_before — курсор
// для следующей (более старой) страницы, null когда старше сообщений нет.
type MessagePage struct {
	Items      []Message `json:"items"`
	HasMore    bool      `json:"has_more"`
	NextBefore *int64    `json:"next_before"`
}

// ListQuery — q, limit, offset для /contacts и /chats.
type ListQuery struct {
	Q             string
	Limit, Offset int
}

type Message struct {
	ID         int64      `json:"id"`
	ChatID     int64      `json:"chat_id"`
	SenderID   int64      `json:"sender_id"`
	SenderName string     `json:"sender_name"`
	Body       string     `json:"body"`
	Files      []FileRef  `json:"files"`
	CreatedAt  time.Time  `json:"created_at"`
	EditedAt   *time.Time `json:"edited_at"`
}

// chatRow — строка таблицы chat (для проверок доступа).
type chatRow struct {
	ID       int64
	Kind     string
	GroupID  *int64
	UserLow  *int64
	UserHigh *int64
}

// userInfo — то, что нужно знать о собеседнике для правил «с кем можно переписываться».
type userInfo struct {
	ID      int64
	Name    string
	Active  bool
	Admin   bool
	Teacher bool
	GroupID *int64 // у студента
}

// ReadPosition синхронизирует прочтение между устройствами одного пользователя.
type ReadPosition struct {
	ChatID            int64 `json:"chat_id"`
	LastReadMessageID int64 `json:"last_read_message_id"`
}
