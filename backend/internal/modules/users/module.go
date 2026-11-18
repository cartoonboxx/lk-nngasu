package users

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"university-api/internal/platform/middleware"
)

// Register собирает пакет users (репозиторий → сервис → хендлеры)
// и вешает его маршруты.
//
// public — группа без аутентификации: только отправка заявки на регистрацию.
// protected должна быть уже закрыта middleware аутентификации: её маршруты
// рассчитывают, что middleware.Me(c) вернёт текущего пользователя.
// Здесь поверх этого задаются только ограничения по ролям.
//
// files нужен сервису, чтобы проверять, что файл аватара принадлежит
// пользователю, который его ставит.
//
// opts — версия условий для регистрации и отправка приглашений (см. Options).
func Register(public, r *gin.RouterGroup, pool *pgxpool.Pool, files FileOwnership, opts Options) {
	h := &Handler{svc: NewService(newRepository(pool), files, opts)}

	// Заявка на регистрацию — без входа. Учётку создаёт администратор, одобряя заявку.
	register := []gin.HandlerFunc{h.SubmitRegistration}
	if opts.RegisterLimit != nil {
		register = append([]gin.HandlerFunc{opts.RegisterLimit}, register...)
	}
	public.POST("/auth/register", register...)

	// Ограничения по ролям, навешиваются на отдельные маршруты.
	admin := middleware.RequireRole(middleware.RoleAdmin)
	staff := middleware.RequireRole(middleware.RoleTeacher, middleware.RoleAdmin)

	// Своя учётка: доступно любому вошедшему пользователю.
	// Id берётся из сессии, так что чужой профиль через эти маршруты не изменить.
	r.GET("/me", h.Me)
	r.PATCH("/me", h.UpdateMe)
	password := []gin.HandlerFunc{h.ChangePassword}
	if opts.PasswordLimit != nil {
		password = append([]gin.HandlerFunc{opts.PasswordLimit}, password...)
	}
	r.POST("/me/password", password...)

	// Управление пользователями: только администратор.
	r.GET("/users", admin, h.List)
	r.POST("/users", admin, h.Create)
	r.PATCH("/users/:id", admin, h.AdminUpdate)
	r.POST("/users/:id/invitations", admin, h.ResendInvitation) // повторно отправить приглашение задать пароль

	// Заявки на регистрацию: только администратор.
	r.GET("/registration-requests", admin, h.ListRegistrations)                // очередь заявок
	r.GET("/registration-requests/:id", admin, h.GetRegistration)              // одна заявка
	r.POST("/registration-requests/:id/decision", admin, h.DecideRegistration) // одобрить или отклонить

	// Полная карточка с паспортом и СНИЛС — только администратор.
	// Преподаватель видит студентов через состав группы или курса, без документов.
	r.GET("/users/:id", admin, h.Get)

	// Состав группы: администратор — любая группа, преподаватель — группа,
	// у которой он ведёт курс (проверяет сервис). Студентам недоступно.
	r.GET("/groups/:id/students", staff, h.GroupStudents)
}
