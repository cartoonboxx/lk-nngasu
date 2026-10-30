package users

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"university-api/internal/platform/middleware"
)

func Register(r *gin.RouterGroup, pool *pgxpool.Pool, files FileOwnership) {
	h := &Handler{svc: NewService(newRepository(pool), files)}
	admin := middleware.RequireRole(middleware.RoleAdmin)
	staff := middleware.RequireRole(middleware.RoleTeacher, middleware.RoleAdmin)

	r.GET("/me", h.Me)
	r.PATCH("/me", h.UpdateMe)
	r.POST("/me/password", h.ChangePassword)
	r.POST("/me/email", h.ChangeEmail)

	r.GET("/users", admin, h.List)
	r.POST("/users", admin, h.Create)
	r.GET("/users/:id", staff, h.Get)
	r.PATCH("/users/:id", admin, h.AdminUpdate)

	r.GET("/groups/:id/students", staff, h.GroupStudents)
}
