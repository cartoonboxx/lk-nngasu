package users

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"university-api/internal/platform/middleware"
)

// Register собирает пакет users (репозиторий → сервис → хендлеры)
// и вешает его маршруты на r.
//
// r должна быть уже закрыта middleware аутентификации: все маршруты
// пакета рассчитывают, что middleware.Me(c) вернёт текущего пользователя.
// Здесь поверх этого задаются только ограничения по ролям.
//
// files нужен сервису, чтобы проверять, что файл аватара принадлежит
// пользователю, который его ставит.
func Register(r *gin.RouterGroup, pool *pgxpool.Pool, files FileOwnership) {
	h := &Handler{svc: NewService(newRepository(pool), files)}

	// Ограничения по ролям, навешиваются на отдельные маршруты.
	admin := middleware.RequireRole(middleware.RoleAdmin)
	staff := middleware.RequireRole(middleware.RoleTeacher, middleware.RoleAdmin)

	// Своя учётка: доступно любому вошедшему пользователю.
	// Id берётся из сессии, так что чужой профиль через эти маршруты не изменить.
	r.GET("/me", h.Me)
	r.PATCH("/me", h.UpdateMe)
	r.POST("/me/password", h.ChangePassword)
	r.POST("/me/email", h.ChangeEmail)

	// Управление пользователями: только администратор.
	r.GET("/users", admin, h.List)
	r.POST("/users", admin, h.Create)
	r.PATCH("/users/:id", admin, h.AdminUpdate)

	// Просмотр карточки пользователя: преподаватели и администраторы.
	// Студенты чужие карточки не видят.
	r.GET("/users/:id", staff, h.Get)

	// Состав группы: преподаватели и администраторы (журнал, ведомости).
	r.GET("/groups/:id/students", staff, h.GroupStudents)
}
