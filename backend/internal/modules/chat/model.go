// Package chat — личные и групповые чаты.
//
// Личный чат (direct) — переписка двух пользователей. Групповой (group) — один на учебную
// группу: в нём студенты группы и преподаватели, ведущие у неё курсы; администратор может
// заходить в любой групповой чат. Участники группового чата не хранятся, а вычисляются из
// таблиц student и workload — состав меняется сам при переводе студента или смене нагрузки.
//
// Доставка — опросом: клиент раз в несколько секунд просит сообщения новее последнего
// известного (GET /chats/:id/messages?after=<id>).
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
	Deleted    bool      `json:"deleted"`
	CreatedAt  time.Time `json:"created_at"`
}

// Chat — чат глазами конкретного пользователя.
type Chat struct {
	ID          int64    `json:"id"`
	Kind        string   `json:"kind"`
	Title       string   `json:"title"`    // группа — её название, личный — имя собеседника
	GroupID     *int64   `json:"group_id"` // только у группового
	Peer        *Person  `json:"peer"`     // только у личного
	LastMessage *Preview `json:"last_message"`
	Unread      int      `json:"unread"` // чужие непрочитанные сообщения
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
	Deleted    bool       `json:"deleted"`
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
