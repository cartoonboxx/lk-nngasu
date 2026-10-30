package schedule

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"university-api/internal/platform/middleware"
)

// Register: public — без авторизации (гость смотрит расписание по ТЗ), protected — с JWT.
func Register(public, protected *gin.RouterGroup, pool *pgxpool.Pool, cal Calendar) {
	h := &Handler{svc: NewService(newRepository(pool), cal)}
	admin := middleware.RequireRole(middleware.RoleAdmin)

	public.GET("/schedule/week", h.Week)
	public.GET("/schedule/groups", h.Groups)
	public.GET("/schedule/teachers", h.Teachers)
	public.GET("/schedule/groups/:id", h.ByGroup)
	public.GET("/schedule/teachers/:id", h.ByTeacher)

	protected.GET("/schedule/me", h.Mine)

	protected.POST("/schedule/lessons", admin, h.CreateLesson)
	protected.PUT("/schedule/lessons/:id", admin, h.UpdateLesson)
	protected.DELETE("/schedule/lessons/:id", admin, h.DeleteLesson)
}
