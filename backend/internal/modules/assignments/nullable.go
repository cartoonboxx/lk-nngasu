package assignments

import (
	"bytes"
	"encoding/json"

	"university-api/internal/platform/apperr"
)

// nullable — поле, ключ которого обязателен, а значение может быть null
// (description, deadline_at, answer_text, comment в контракте). Обычный указатель
// не отличает отсутствующий ключ от null, а контракт требует 400 на пропущенный ключ.
type nullable[T any] struct {
	Set   bool
	Value *T
}

// UnmarshalJSON вызывается только при наличии ключа.
func (n *nullable[T]) UnmarshalJSON(b []byte) error {
	n.Set = true
	if bytes.Equal(bytes.TrimSpace(b), []byte("null")) {
		n.Value = nil
		return nil
	}
	var v T
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	n.Value = &v
	return nil
}

// fieldErr — 400 invalid_request с указанием поля, в формате ошибок binding.
func fieldErr(field, problem string) error {
	return apperr.NewInvalid("invalid_request", "request body is invalid").
		WithDetails(map[string]any{"fields": map[string][]string{field: {problem}}})
}

// requireKeys проверяет, что все обязательные nullable-ключи переданы.
func requireKeys(keys map[string]bool) error {
	for _, name := range []string{"description", "deadline_at", "answer_text", "comment"} {
		if set, checked := keys[name]; checked && !set {
			return fieldErr(name, "required (null is allowed)")
		}
	}
	return nil
}
