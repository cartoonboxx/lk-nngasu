package auth

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"university-api/internal/platform/httpx"
	"university-api/internal/platform/middleware"
	"university-api/internal/platform/ratelimit"
)

func (s *Service) RegisterAdminRoutes(r *gin.RouterGroup, limiter *ratelimit.Limiter) {
	h := &Handler{svc: s, opts: s.opts}
	r.GET("/admin/administrators", h.Administrators)
	write := r.Group("/admin/administrators", limiter.Middleware("manage-admins", 5, 15*time.Minute, func(c *gin.Context) string {
		return strconv.FormatInt(middleware.Me(c).UserID, 10)
	}))
	write.POST("/invitations", h.InviteAdministrator)
	write.POST("/promote", h.PromoteAdministrator)
}

func (h *Handler) SetupStatus(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	available, err := h.svc.SetupAvailable(c.Request.Context())
	httpx.Reply(c, gin.H{"available": available}, err)
}

func (h *Handler) Setup(c *gin.Context) {
	var req struct {
		AdminInput
		Key      string `json:"setup_key" binding:"required,max=256"`
		Password string `json:"password" binding:"required,min=8,max=72"`
	}
	if !httpx.Bind(c, &req) {
		return
	}
	err := h.svc.Setup(c.Request.Context(), req.Key, req.Password, req.AdminInput)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Created(c, gin.H{"created": true})
}

func (h *Handler) Administrators(c *gin.Context) {
	list, err := h.svc.Administrators(c.Request.Context(), middleware.Me(c).UserID)
	httpx.Reply(c, list, err)
}

func (h *Handler) InviteAdministrator(c *gin.Context) {
	var req struct {
		AdminInput
		CurrentPassword string `json:"current_password" binding:"required,max=72"`
	}
	if !httpx.Bind(c, &req) {
		return
	}
	httpx.ReplyEmpty(c, h.svc.InviteAdministrator(c.Request.Context(), middleware.Me(c).UserID, req.CurrentPassword, req.AdminInput))
}

func (h *Handler) PromoteAdministrator(c *gin.Context) {
	var req struct {
		Email           string `json:"email" binding:"required,email,max=255"`
		CurrentPassword string `json:"current_password" binding:"required,max=72"`
	}
	if !httpx.Bind(c, &req) {
		return
	}
	httpx.ReplyEmpty(c, h.svc.PromoteAdministrator(c.Request.Context(), middleware.Me(c).UserID, req.CurrentPassword, req.Email))
}
