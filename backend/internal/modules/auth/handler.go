package auth

import (
	"github.com/gin-gonic/gin"

	"university-api/internal/platform/httpx"
)

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type Handler struct{ svc *Service }

// POST /auth/login
func (h *Handler) Login(c *gin.Context) {
	var req loginRequest
	if !httpx.Bind(c, &req) {
		return
	}
	token, err := h.svc.Login(c.Request.Context(), req.Email, req.Password)
	httpx.Reply(c, token, err)
}
