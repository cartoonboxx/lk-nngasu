package httpx

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

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
	case apperr.Gone:
		return http.StatusGone
	case apperr.TooLarge:
		return http.StatusRequestEntityTooLarge
	case apperr.Unsupported:
		return http.StatusUnsupportedMediaType
	case apperr.TooManyRequests:
		return http.StatusTooManyRequests
	case apperr.Unavailable:
		return http.StatusServiceUnavailable
	case apperr.UpgradeRequired:
		return http.StatusUpgradeRequired
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
	body := gin.H{"code": ae.Code, "message": ae.Message}
	if len(ae.Details) > 0 {
		body["details"] = ae.Details
	}
	if ae.RetryAfter > 0 {
		c.Header("Retry-After", strconv.Itoa(ae.RetryAfter))
	}
	c.AbortWithStatusJSON(statusOf(ae.Kind), gin.H{"error": body})
}

func BadRequest(c *gin.Context, err error) {
	Fail(c, apperr.New(apperr.Invalid, "invalid_request", err.Error()))
}
