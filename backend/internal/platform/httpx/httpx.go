package httpx

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"university-api/internal/platform/apperr"
)

func statusOf(k apperr.Kind) int {
	switch k {
	case apperr.Invalid:
		return http.StatusBadRequest
	case apperr.Unauthorized:
		return http.StatusUnauthorized
	case apperr.Forbidden:
		return http.StatusForbidden
	case apperr.NotFound:
		return http.StatusNotFound
	case apperr.Conflict:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

// Fail — единственное место, где доменная ошибка превращается в HTTP-ответ.
func Fail(c *gin.Context, err error) {
	var ae *apperr.Error
	if !errors.As(err, &ae) {
		slog.Error("unhandled error", "path", c.FullPath(), "err", err)
		ae = apperr.New(apperr.Internal, "internal_error", "internal server error")
	}
	c.AbortWithStatusJSON(statusOf(ae.Kind), gin.H{
		"error": gin.H{"code": ae.Code, "message": ae.Message},
	})
}

func BadRequest(c *gin.Context, err error) {
	Fail(c, apperr.New(apperr.Invalid, "invalid_request", err.Error()))
}
