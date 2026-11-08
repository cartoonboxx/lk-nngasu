package schedule

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"university-api/internal/platform/middleware"
)

// Register собирает пакет schedule и вешает его маршруты на две группы:
//   - public — без авторизации: по ТЗ гость может смотреть расписание
//     групп и преподавателей, не входя в систему;
//   - protected — за JWT: личное расписание и правка занятий.
//
// cal передаётся сервису для работы с учебным календарём (какая сейчас неделя,
// даты занятий). Вынесен в интерфейс, чтобы в тестах подставлять фиксированную дату.
//
// Возвращает сервис: app использует его для главной страницы ЛК (пары на сегодня).
func Register(public, protected *gin.RouterGroup, pool *pgxpool.Pool, cal Calendar) *Service {
	h := &Handler{svc: NewService(newRepository(pool), cal)}
	admin := middleware.RequireRole(middleware.RoleAdmin)

	// Публичный просмотр.
	public.GET("/schedule/week", h.Week)              // информация о текущей учебной неделе
	public.GET("/schedule/groups", h.Groups)          // группы, для которых можно открыть расписание (для выбора на сайте)
	public.GET("/schedule/teachers", h.Teachers)      // преподаватели, для которых можно открыть расписание
	public.GET("/schedule/groups/:id", h.ByGroup)     // расписание группы
	public.GET("/schedule/teachers/:id", h.ByTeacher) // расписание преподавателя

	// Личное расписание: для студента — расписание его группы,
	// для преподавателя — его занятия. Кто есть кто, определяется по JWT.
	protected.GET("/schedule/me", h.Mine)

	// Страница расписания ЛК: своё расписание или ?group_id=, по датам ?date_from=&date_to=;
	// без дат — текущая неделя.
	protected.GET("/schedule", h.Overview)

	// Правка расписания: только администратор.
	protected.POST("/schedule/lessons", admin, h.CreateLesson)
	protected.PUT("/schedule/lessons/:id", admin, h.UpdateLesson)
	protected.DELETE("/schedule/lessons/:id", admin, h.DeleteLesson)

	return h.svc
}
