// Package filekey описывает формат ключа файла в хранилище: "<user_id>/<random>/<имя файла>".
//
// В таблице file хранится только storage_key, поэтому всё остальное выводится из него:
// владелец — это папка пользователя (первый сегмент), имя для скачивания — последний.
package filekey

import (
	"crypto/rand"
	"encoding/hex"
	"strconv"
	"strings"
)

// Prefix — префикс всех файлов пользователя ("42/"). Используется и в Go, и в SQL (starts_with).
func Prefix(userID int64) string { return strconv.FormatInt(userID, 10) + "/" }

// New создаёт уникальный ключ в папке пользователя. name должен быть уже очищен (см. files.CleanName).
func New(userID int64, name string) (string, error) {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return Prefix(userID) + hex.EncodeToString(b) + "/" + name, nil
}

// Name возвращает имя файла для скачивания.
func Name(key string) string {
	if i := strings.LastIndexByte(key, '/'); i >= 0 {
		return key[i+1:]
	}
	return key
}

// Owner возвращает id владельца файла.
func Owner(key string) (int64, bool) {
	head, _, ok := strings.Cut(key, "/")
	if !ok {
		return 0, false
	}
	id, err := strconv.ParseInt(head, 10, 64)
	return id, err == nil && id > 0
}
