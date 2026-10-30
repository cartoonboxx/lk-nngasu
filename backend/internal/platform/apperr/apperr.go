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
)

type Error struct {
	Kind    Kind
	Code    string // машинное имя: "room_busy", "lesson_not_found"
	Message string // человекочитаемое описание
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

// IsNotFound сообщает, что err — ошибка «не найдено».
func IsNotFound(err error) bool {
	var e *Error
	return errors.As(err, &e) && e.Kind == NotFound
}
