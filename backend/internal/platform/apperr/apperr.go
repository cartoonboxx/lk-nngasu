// Package apperr содержит доменные ошибки, не зависящие от HTTP.
// Сервисы возвращают *Error, а в код ответа его превращает httpx.Fail.
package apperr

import (
	"errors"
	"fmt"
)

type Kind int

const (
	Internal Kind = iota
	Invalid
	Unauthorized
	Forbidden
	NotFound
	Conflict
	Gone            // 410: например, истёкшее приглашение
	TooLarge        // 413: файл больше лимита
	Unsupported     // 415: недопустимый тип файла
	TooManyRequests // 429: превышен лимит частоты запросов
	Unavailable     // 503: недоступна зависимость
	UpgradeRequired // 426: нужен WebSocket
)

type Error struct {
	Kind    Kind
	Code    string // машинное имя: "room_busy", "lesson_not_found"
	Message string // человекочитаемое описание

	// Details — необязательные подробности, уходят клиенту как error.details:
	// fields для ошибок валидации полей, conflicts для расписания и т.п.
	Details map[string]any

	// RetryAfter — для 429: через сколько секунд можно повторить (заголовок Retry-After).
	RetryAfter int
}

// WithDetails возвращает копию ошибки с подробностями для клиента.
func (e *Error) WithDetails(details map[string]any) *Error {
	cp := *e
	cp.Details = details
	return &cp
}

func (e *Error) Error() string { return fmt.Sprintf("%s: %s", e.Code, e.Message) }

func New(kind Kind, code, message string) *Error {
	return &Error{Kind: kind, Code: code, Message: message}
}

func NewInvalid(code, message string) *Error      { return New(Invalid, code, message) }
func NewUnauthorized(code, message string) *Error { return New(Unauthorized, code, message) }
func NewForbidden(code, message string) *Error    { return New(Forbidden, code, message) }
func NewConflict(code, message string) *Error     { return New(Conflict, code, message) }

// NewNotFound — «запись не найдена» с кодом конкретной сущности: "lesson_not_found".
func NewNotFound(code string) *Error { return New(NotFound, code, "record not found") }

func NewGone(code, message string) *Error        { return New(Gone, code, message) }
func NewTooLarge(code, message string) *Error    { return New(TooLarge, code, message) }
func NewUnsupported(code, message string) *Error { return New(Unsupported, code, message) }

// NewRateLimited — 429 rate_limited; retryAfter — секунды до следующей попытки.
func NewRateLimited(retryAfter int) *Error {
	e := New(TooManyRequests, "rate_limited", "too many requests, try again later")
	e.RetryAfter = retryAfter
	return e
}

// IsNotFound сообщает, что err — ошибка «не найдено».
func IsNotFound(err error) bool {
	var e *Error
	return errors.As(err, &e) && e.Kind == NotFound
}
