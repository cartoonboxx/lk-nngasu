// Package structure владеет справочниками учебной структуры вуза:
// кафедры, дисциплины, специальности, группы, семестры, аудитории.
// Это редко меняющиеся данные, на которые ссылаются остальные пакеты
// (пользователи, расписание, оценки).
package structure

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"university-api/internal/platform/middleware"
)

// Register собирает пакет structure и вешает его маршруты на r.
//
// r должна быть уже закрыта middleware аутентификации. Общее правило для
// всех справочников: читать может любой вошедший пользователь (данные
// нужны для форм, фильтров и расписания), изменять — только администратор.
//
// Изменение — через PUT: клиент присылает объект целиком, а не отдельные поля.
func Register(r *gin.RouterGroup, pool *pgxpool.Pool) {
	repo := newRepository(pool)
	h := &Handler{repo: repo} // сервисного слоя нет: правила простые, хендлеры работают с репозиторием напрямую
	admin := middleware.RequireRole(middleware.RoleAdmin)

	// Справочники «id + название»: чтение — всем авторизованным, правка — администратору.
	// Одинаковые CRUD-хендлеры, отличается только таблица.
	registerCatalog(r, "/departments", admin, repo.departments)
	// Справочник дисциплин — под /catalog: путь /subjects занят страницей «Мои предметы» ЛК (модуль courses).
	registerCatalog(r, "/catalog/subjects", admin, repo.subjects)

	// Специальности (направления подготовки).
	r.GET("/specializations", h.ListSpecializations)
	r.POST("/specializations", admin, h.CreateSpecialization)
	r.PUT("/specializations/:id", admin, h.UpdateSpecialization)
	r.DELETE("/specializations/:id", admin, h.DeleteSpecialization)

	// Учебные группы.
	// Состав группы (GET /groups/:id/students) регистрирует пакет users.
	// gin требует одинакового имени параметра в одной позиции пути,
	// поэтому здесь и там должно быть именно :id, иначе паника при старте.
	r.GET("/groups", h.ListGroups)
	r.GET("/groups/:id", h.GetGroup)
	r.POST("/groups", admin, h.CreateGroup)
	r.PUT("/groups/:id", admin, h.UpdateGroup)
	r.DELETE("/groups/:id", admin, h.DeleteGroup)

	// Семестры. Ровно один семестр отмечен текущим: его отдаёт /semesters/current,
	// переключает POST /semesters/:id/current (отдельное действие, а не поле в PUT,
	// чтобы снятие флага со старого и установка на новый шли одной операцией).
	r.GET("/semesters", h.ListSemesters)
	r.GET("/semesters/current", h.CurrentSemester)
	r.POST("/semesters", admin, h.CreateSemester)
	r.PUT("/semesters/:id", admin, h.UpdateSemester)
	r.POST("/semesters/:id/current", admin, h.SetCurrentSemester)
	r.DELETE("/semesters/:id", admin, h.DeleteSemester)

	// Аудитории.
	r.GET("/rooms", h.ListRooms)
	r.POST("/rooms", admin, h.CreateRoom)
	r.PUT("/rooms/:id", admin, h.UpdateRoom)
	r.DELETE("/rooms/:id", admin, h.DeleteRoom)
}

// registerCatalog вешает GET/POST/PUT/DELETE для справочника «id + название».
//
//	GET    path      — список, любому вошедшему;
//	POST   path      — создать, только admin;
//	PUT    path/:id  — переименовать, только admin;
//	DELETE path/:id  — удалить, только admin.
//
// cat задаёт таблицу справочника, хендлеры для неё строит catalogHandlers.
func registerCatalog(r *gin.RouterGroup, path string, admin gin.HandlerFunc, cat catalog) {
	list, create, update, remove := catalogHandlers(cat)
	r.GET(path, list)
	r.POST(path, admin, create)
	r.PUT(path+"/:id", admin, update)
	r.DELETE(path+"/:id", admin, remove)
}
