// Package files владеет таблицей file и хранилищем файлов на диске.
//
// Ключ файла "<user_id>/<random>/<имя>" несёт в себе владельца и имя (см. platform/filekey),
// поэтому в таблице file больше нет колонок. Загрузка и привязка разделены: сначала
// POST /files возвращает id, затем id передаётся в задание, ответ или материал.
package files

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Register(r *gin.RouterGroup, pool *pgxpool.Pool, storage *Storage, maxBytes int64) {
	h := &Handler{repo: newRepository(pool), storage: storage, maxBytes: maxBytes}

	r.POST("/files", h.Upload)
	r.GET("/files/:id", h.Download)
	r.GET("/files/:id/meta", h.Meta)
	r.DELETE("/files/:id", h.Delete)
}
