// Package news — новости: публичная лента (/news) и редактор администратора (/admin/news).
//
// Состояния: draft (черновик) ↔ published (опубликовано) → archived (архив) → draft/published.
// published_at есть только у опубликованной новости: при публикации — now(),
// при снятии с публикации или архивации — null (CHECK news_published_at_valid).
package news

import (
	"strings"
	"time"
	"unicode"
)

const (
	StatusDraft     = "draft"
	StatusPublished = "published"
	StatusArchived  = "archived"
)

// excerptRunes — длина анонса в символах.
const excerptRunes = 240

// PersonRef — автор: только id и ФИО.
type PersonRef struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// NewsSummary — элемент публичной ленты.
type NewsSummary struct {
	ID          int64      `json:"id"`
	Excerpt     string     `json:"excerpt"`
	Author      PersonRef  `json:"author"`
	PublishedAt *time.Time `json:"published_at"`
}

// NewsDetail — опубликованная новость целиком.
type NewsDetail struct {
	NewsSummary
	Body string `json:"body"`
}

// AdminNews — новость в редакторе: с черновиками и архивом, статусом и датами.
type AdminNews struct {
	NewsDetail
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Excerpt — анонс: первый абзац текста (до пустой строки) с пробелами, сжатыми до одного,
// не длиннее 240 символов; обрезанный анонс заканчивается многоточием.
func Excerpt(body string) string {
	first := strings.TrimSpace(body)
	lines := strings.Split(first, "\n")
	var para []string
	for _, l := range lines {
		if strings.TrimSpace(l) == "" {
			if len(para) > 0 {
				break
			}
			continue
		}
		para = append(para, l)
	}
	text := strings.Join(strings.FieldsFunc(strings.Join(para, " "), unicode.IsSpace), " ")
	r := []rune(text)
	if len(r) <= excerptRunes {
		return text
	}
	return strings.TrimRightFunc(string(r[:excerptRunes-1]), unicode.IsSpace) + "…"
}
