package database

import (
	"fmt"
	"strings"
)

// Filter — маленький построитель WHERE с позиционными параметрами.
type Filter struct {
	conds []string
	args  []any
}

// Add добавляет условие; каждый "?" заменяется на номер параметра ($1, $2, ...).
// Один Add = один аргумент, даже если "?" в условии несколько.
func (f *Filter) Add(cond string, arg any) {
	f.args = append(f.args, arg)
	f.conds = append(f.conds, strings.ReplaceAll(cond, "?", fmt.Sprintf("$%d", len(f.args))))
}

// Raw добавляет условие без параметров.
func (f *Filter) Raw(cond string) { f.conds = append(f.conds, cond) }

// Arg регистрирует параметр и возвращает его плейсхолдер (для LIMIT/OFFSET).
func (f *Filter) Arg(v any) string {
	f.args = append(f.args, v)
	return fmt.Sprintf("$%d", len(f.args))
}

func (f *Filter) Where() string {
	if len(f.conds) == 0 {
		return ""
	}
	return " WHERE " + strings.Join(f.conds, " AND ")
}

func (f *Filter) Args() []any { return f.args }
