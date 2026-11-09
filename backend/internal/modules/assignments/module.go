// Package assignments — задания курса, ответы студентов, проверка и оценки.
package assignments

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"university-api/internal/platform/middleware"
)

// Register регистрирует маршруты заданий, ответов и оценок.
// Возвращает сервис: app использует его для страницы предмета и страницы оценок ЛК.
func Register(r *gin.RouterGroup, pool *pgxpool.Pool, guard Guard, blobs Blobs) *Service {
	h := &Handler{svc: NewService(newRepository(pool), guard, blobs)}
	staff := middleware.RequireRole(middleware.RoleTeacher, middleware.RoleAdmin) //определение роли пользователя
	student := middleware.RequireRole(middleware.RoleStudent)

	r.GET("/courses/:id/assignments", h.List)
	r.POST("/courses/:id/assignments", staff, h.Create)
	r.GET("/courses/:id/grades", staff, h.CourseGrades)

	r.GET("/assignments/:id", h.Get)
	r.PUT("/assignments/:id", staff, h.Update)
	r.DELETE("/assignments/:id", staff, h.Delete)

	r.GET("/assignments/:id/submissions", staff, h.ListSubmissions)
	r.GET("/assignments/:id/submissions/archive", staff, h.AssignmentArchive)
	r.GET("/assignments/:id/submission", student, h.MySubmission)
	r.PUT("/assignments/:id/submission", student, h.SaveSubmission)
	r.POST("/assignments/:id/submission", student, h.SaveSubmission) // тот же handler: путь из ТЗ фронта

	r.GET("/submissions/:id", h.GetSubmission)
	r.GET("/submissions/:id/archive", h.SubmissionArchive)
	r.PUT("/submissions/:id/review", staff, h.Review)

	r.GET("/me/grades", student, h.MyGrades)

	return h.svc
}
