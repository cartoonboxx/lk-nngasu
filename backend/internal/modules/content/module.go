package content

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"university-api/internal/platform/middleware"
)

func Register(r *gin.RouterGroup, pool *pgxpool.Pool, guard CourseGuard, blobs Blobs) {
	h := &Handler{svc: NewService(newRepository(pool), guard, blobs)}
	staff := middleware.RequireRole(middleware.RoleTeacher, middleware.RoleAdmin)

	r.GET("/courses/:id/topics", h.ListTopics)
	r.POST("/courses/:id/topics", staff, h.CreateTopic)
	r.PUT("/topics/:id", staff, h.UpdateTopic)
	r.DELETE("/topics/:id", staff, h.DeleteTopic)

	r.POST("/topics/:id/materials", staff, h.CreateMaterial)
	r.PUT("/materials/:id", staff, h.UpdateMaterial)
	r.DELETE("/materials/:id", staff, h.DeleteMaterial)
}
