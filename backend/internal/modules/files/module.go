// Package files владеет таблицей file и хранилищем файлов на диске.
// Другие модули (задания, ответы, материалы) файлы не хранят, а ссылаются на них по id.
//
// Ключ файла "<user_id>/<random>/<имя>" несёт в себе владельца и исходное имя
// (разбор — в platform/filekey), поэтому других колонок в таблице file нет.
//
// Загрузка и привязка разделены: сначала POST /files сохраняет файл и возвращает id,
// затем этот id передаётся в задание, ответ или материал.
package files

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Register собирает зависимости модуля files и регистрирует его маршруты в группе r.
//
// pool     — общий пул подключений к Postgres, создаётся один раз в main;
// storage  — хранилище файлов на диске, создаётся в main;
// maxBytes — максимальный размер загружаемого файла в байтах.
func Register(r *gin.RouterGroup, pool *pgxpool.Pool, storage *Storage, maxBytes int64) {
	// Отдельного сервиса нет: handler работает напрямую
	// с репозиторием (таблица file) и хранилищем (файлы на диске).
	h := &Handler{repo: newRepository(pool), storage: storage, maxBytes: maxBytes}

	// Загрузить файл на диск и создать запись в file; возвращает id.
	r.POST("/files", h.Upload)

	// Скачать содержимое файла по id.
	r.GET("/files/:id", h.Download)

	// Получить метаданные файла (без содержимого).
	r.GET("/files/:id/meta", h.Meta)

	// Удалить файл: запись в file и сам файл на диске.
	r.DELETE("/files/:id", h.Delete)
}
