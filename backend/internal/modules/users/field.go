package users

import (
	"bytes"
	"encoding/json"
)

// Field — поле PATCH-запроса с тремя состояниями:
//
//	ключа нет в JSON      → Set = false            (не менять)
//	"key": null           → Set = true, Value = nil (очистить / снять)
//	"key": значение       → Set = true, Value = &v  (записать)
//
// Обычный указатель различает только два состояния: null и отсутствие ключа
// для него одинаковы. Документация требует «null очищает отчество», «null снимает роль».
type Field[T any] struct {
	Set   bool
	Value *T
}

// UnmarshalJSON вызывается только когда ключ есть в JSON, поэтому Set = true.
func (f *Field[T]) UnmarshalJSON(b []byte) error {
	f.Set = true
	if bytes.Equal(bytes.TrimSpace(b), []byte("null")) {
		f.Value = nil
		return nil
	}
	// Свой декодер: обычный json.Unmarshal пропустил бы лишние ключи внутри
	// вложенного объекта (например, student), а контракт требует 400.
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	var v T
	if err := dec.Decode(&v); err != nil {
		return err
	}
	f.Value = &v
	return nil
}

// Null — ключ передан со значением null.
func (f Field[T]) Null() bool { return f.Set && f.Value == nil }

// mapField переносит состояние поля, преобразуя значение (тело запроса → тип сервиса).
func mapField[A, B any](f Field[A], conv func(A) B) Field[B] {
	out := Field[B]{Set: f.Set}
	if f.Value != nil {
		v := conv(*f.Value)
		out.Value = &v
	}
	return out
}
