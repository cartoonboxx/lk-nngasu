package courses

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"university-api/internal/platform/access"
	"university-api/internal/platform/middleware"
)

// Register регистрирует маршруты курсов.
//
// topics — загрузка тем курса для страницы предмета (темы → материалы + задания).
// Темы и задания принадлежат модулям content и assignments, поэтому courses получает
// готовую функцию из app, а не импортирует эти пакеты. nil — страница без тем.
//
// Возвращает сервис: app использует его для страницы оценок ЛК.
func Register(r *gin.RouterGroup, pool *pgxpool.Pool, checker *access.Checker, topics TopicsLoader) *Service {
	repo := newRepository(pool)
	h := &Handler{svc: NewService(repo, checker, topics), repo: repo}
	admin := middleware.RequireRole(middleware.RoleAdmin)

	r.GET("/courses", h.ListCourses)
	// Предметы текущего семестра для ЛК (= /courses?current=true).
	// Справочник дисциплин для админки — в structure по пути /catalog/subjects.
	r.GET("/subjects", h.Subjects)
	r.GET("/subjects/:id", h.GetCourse) // страница предмета; id — id курса из списка /subjects
	r.POST("/courses", admin, h.CreateCourse)
	r.GET("/courses/:id", h.GetCourse)
	r.DELETE("/courses/:id", admin, h.DeleteCourse)
	r.GET("/courses/:id/students", h.ListStudents)

	r.POST("/courses/:id/workloads", admin, h.AddWorkload)
	r.DELETE("/workloads/:id", admin, h.DeleteWorkload)

	return h.svc
}
