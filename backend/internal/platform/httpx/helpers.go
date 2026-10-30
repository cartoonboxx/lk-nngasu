package httpx

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"university-api/internal/platform/apperr"
)

// ---- разбор запроса ----

// ID разбирает числовой path-параметр; при ошибке сам пишет 400.
func ID(c *gin.Context, name string) (int64, bool) {
	id, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil || id <= 0 {
		Fail(c, apperr.NewInvalid("invalid_id", "invalid path parameter: "+name))
		return 0, false
	}
	return id, true
}

func invalidQuery(c *gin.Context, name string) {
	Fail(c, apperr.NewInvalid("invalid_query", "invalid query parameter: "+name))
}

// OptInt64 разбирает необязательный числовой query-параметр.
func OptInt64(c *gin.Context, name string) (*int64, bool) {
	raw := c.Query(name)
	if raw == "" {
		return nil, true
	}
	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		invalidQuery(c, name)
		return nil, false
	}
	return &v, true
}

// OptBool разбирает необязательный query-параметр "true"/"false".
func OptBool(c *gin.Context, name string) (*bool, bool) {
	switch c.Query(name) {
	case "":
		return nil, true
	case "true":
		v := true
		return &v, true
	case "false":
		v := false
		return &v, true
	}
	invalidQuery(c, name)
	return nil, false
}

// OptDate разбирает необязательный query-параметр YYYY-MM-DD (дата без времени, в UTC).
func OptDate(c *gin.Context, name string) (*time.Time, bool) {
	raw := c.Query(name)
	if raw == "" {
		return nil, true
	}
	d, err := time.Parse(time.DateOnly, raw)
	if err != nil {
		invalidQuery(c, name)
		return nil, false
	}
	return &d, true
}

// Bind разбирает JSON-тело и валидирует binding-тегами; при ошибке сам пишет 400.
func Bind(c *gin.Context, dst any) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		BadRequest(c, err)
		return false
	}
	return true
}

// Page возвращает limit/offset из query (по умолчанию 50, максимум 200).
func Page(c *gin.Context) (limit, offset int) {
	limit, offset = 50, 0
	if v, err := strconv.Atoi(c.Query("limit")); err == nil && v > 0 {
		limit = v
	}
	if limit > 200 {
		limit = 200
	}
	if v, err := strconv.Atoi(c.Query("offset")); err == nil && v > 0 {
		offset = v
	}
	return limit, offset
}

// ---- ответы ----

func OK(c *gin.Context, data any)      { c.JSON(http.StatusOK, gin.H{"data": data}) }
func Created(c *gin.Context, data any) { c.JSON(http.StatusCreated, gin.H{"data": data}) }
func NoContent(c *gin.Context)         { c.Status(http.StatusNoContent) }

// Reply отвечает 200 с data либо переводит err в ошибочный ответ — типичный хвост любого хендлера.
func Reply(c *gin.Context, data any, err error) {
	if err != nil {
		Fail(c, err)
		return
	}
	OK(c, data)
}

// ReplyCreated — как Reply, но с кодом 201.
func ReplyCreated(c *gin.Context, data any, err error) {
	if err != nil {
		Fail(c, err)
		return
	}
	Created(c, data)
}

// ReplyEmpty отвечает 204 либо ошибкой.
func ReplyEmpty(c *gin.Context, err error) {
	if err != nil {
		Fail(c, err)
		return
	}
	NoContent(c)
}
