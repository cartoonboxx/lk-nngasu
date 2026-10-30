package files

import (
	"bytes"
	"path"
	"strings"
	"unicode/utf8"
)

// fileType — допустимый тип файла: расширение → MIME и проверка «магических» первых байт.
// Расширение в имени легко подделать, поэтому содержимое должно ему соответствовать.
type fileType struct {
	mime  string
	magic func(head []byte) bool
}

func prefix(p ...byte) func([]byte) bool {
	return func(head []byte) bool { return bytes.HasPrefix(head, p) }
}

func zipMagic(head []byte) bool {
	return bytes.HasPrefix(head, []byte("PK\x03\x04")) || bytes.HasPrefix(head, []byte("PK\x05\x06"))
}

func webpMagic(head []byte) bool {
	return len(head) >= 12 && bytes.HasPrefix(head, []byte("RIFF")) && bytes.Equal(head[8:12], []byte("WEBP"))
}

// attachmentTypes — вложения заданий, ответов и материалов (FR-HW-02, FR-HW-06).
var attachmentTypes = map[string]fileType{
	".pdf":  {"application/pdf", prefix('%', 'P', 'D', 'F', '-')},
	".docx": {"application/vnd.openxmlformats-officedocument.wordprocessingml.document", zipMagic},
	".pptx": {"application/vnd.openxmlformats-officedocument.presentationml.presentation", zipMagic},
	".xlsx": {"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", zipMagic},
	".xls":  {"application/vnd.ms-excel", prefix(0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1)},
	".zip":  {"application/zip", zipMagic},
	".rar":  {"application/vnd.rar", prefix('R', 'a', 'r', '!', 0x1A, 0x07)},
}

// avatarTypes — картинки профиля.
var avatarTypes = map[string]fileType{
	".jpg":  {"image/jpeg", prefix(0xFF, 0xD8, 0xFF)},
	".jpeg": {"image/jpeg", prefix(0xFF, 0xD8, 0xFF)},
	".png":  {"image/png", prefix(0x89, 'P', 'N', 'G', '\r', '\n', 0x1A, '\n')},
	".webp": {"image/webp", webpMagic},
}

func extOf(name string) string { return strings.ToLower(path.Ext(name)) }

// mimeFor возвращает MIME по расширению имени файла (для скачивания).
func mimeFor(name string) string {
	ext := extOf(name)
	if t, ok := attachmentTypes[ext]; ok {
		return t.mime
	}
	if t, ok := avatarTypes[ext]; ok {
		return t.mime
	}
	return "application/octet-stream"
}

const maxNameRunes = 150

// CleanName делает из имени, присланного клиентом, безопасное имя файла: без путей,
// управляющих символов и знаков, недопустимых в файловых системах; расширение сохраняется.
func CleanName(name string) string {
	name = path.Base(strings.ReplaceAll(name, `\`, "/"))

	var b strings.Builder
	for _, r := range name {
		switch {
		case r == utf8.RuneError, r < 0x20, r == 0x7f:
			// выбрасываем
		case strings.ContainsRune(`/\:*?"<>|`, r):
			b.WriteRune('_')
		default:
			b.WriteRune(r)
		}
	}
	out := strings.TrimLeft(strings.TrimSpace(b.String()), ".")

	if runes := []rune(out); len(runes) > maxNameRunes {
		ext := []rune(path.Ext(out))
		if len(ext) > 10 {
			ext = nil
		}
		out = string(runes[:maxNameRunes-len(ext)]) + string(ext)
	}
	if out == "" {
		return "file"
	}
	return out
}
