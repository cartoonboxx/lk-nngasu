// Package content — темы курса и учебные материалы внутри тем.
// Читать список тем может любой, у кого есть доступ к курсу,
// создавать, менять и удалять — только преподаватели и администраторы.
package content

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"university-api/internal/platform/middleware"
)

// Register собирает зависимости модуля content и регистрирует его маршруты в группе r.
//
// pool  — общий пул подключений к Postgres, создаётся один раз в main;
// guard — проверки доступа к курсу и владения файлами (реализует access.Checker);
// blobs — удаление файлов с диска (реализует files.Storage).
//
// Модуль зависит от интерфейсов CourseGuard и Blobs, а не от пакетов access и files
// напрямую, — так модули не импортируют друг друга.
//
// Возвращает сервис: app использует его для страницы предмета (темы с материалами).
func Register(r *gin.RouterGroup, pool *pgxpool.Pool, guard CourseGuard, blobs Blobs) *Service {
	// Сборка цепочки: репозиторий → сервис → handler.
	h := &Handler{svc: NewService(newRepository(pool), guard, blobs)}

	// staff — грубая проверка по роли из JWT: пропускает только преподавателей и админов.
	// Имеет ли пользователь доступ именно к этому курсу, проверяет сервис через guard.
	staff := middleware.RequireRole(middleware.RoleTeacher, middleware.RoleAdmin)

	// Темы курса.
	r.GET("/courses/:id/topics", h.ListTopics)          // список тем курса с материалами
	r.POST("/courses/:id/topics", staff, h.CreateTopic) // создать тему в курсе
	r.PUT("/topics/:id", staff, h.UpdateTopic)          // изменить тему
	r.DELETE("/topics/:id", staff, h.DeleteTopic)       // удалить тему вместе с материалами

	// Материалы темы. Файлы загружаются заранее через POST /files,
	// здесь передаются только их id.
	r.POST("/topics/:id/materials", staff, h.CreateMaterial) // добавить материал в тему
	r.PUT("/materials/:id", staff, h.UpdateMaterial)         // изменить материал
	r.DELETE("/materials/:id", staff, h.DeleteMaterial)      // удалить материал

	return h.svc
}
