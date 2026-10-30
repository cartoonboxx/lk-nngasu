package courses

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"university-api/internal/platform/access"
	"university-api/internal/platform/middleware"
)

func Register(r *gin.RouterGroup, pool *pgxpool.Pool, checker *access.Checker) {
	repo := newRepository(pool)
	h := &Handler{svc: NewService(repo, checker), repo: repo}
	admin := middleware.RequireRole(middleware.RoleAdmin)

	r.GET("/courses", h.ListCourses)
	r.POST("/courses", admin, h.CreateCourse)
	r.GET("/courses/:id", h.GetCourse)
	r.DELETE("/courses/:id", admin, h.DeleteCourse)
	r.GET("/courses/:id/students", h.ListStudents)

	r.POST("/courses/:id/workloads", admin, h.AddWorkload)
	r.DELETE("/workloads/:id", admin, h.DeleteWorkload)
}
