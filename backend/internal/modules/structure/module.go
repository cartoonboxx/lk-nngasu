package structure

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"university-api/internal/platform/middleware"
)

func Register(r *gin.RouterGroup, pool *pgxpool.Pool) {
	repo := newRepository(pool)
	h := &Handler{repo: repo}
	admin := middleware.RequireRole(middleware.RoleAdmin)

	// Справочники «id + название»: чтение — всем авторизованным, правка — администратору.
	registerCatalog(r, "/departments", admin, repo.departments)
	registerCatalog(r, "/subjects", admin, repo.subjects)

	r.GET("/specializations", h.ListSpecializations)
	r.POST("/specializations", admin, h.CreateSpecialization)
	r.PUT("/specializations/:id", admin, h.UpdateSpecialization)
	r.DELETE("/specializations/:id", admin, h.DeleteSpecialization)

	r.GET("/groups", h.ListGroups)
	r.GET("/groups/:id", h.GetGroup)
	r.POST("/groups", admin, h.CreateGroup)
	r.PUT("/groups/:id", admin, h.UpdateGroup)
	r.DELETE("/groups/:id", admin, h.DeleteGroup)

	r.GET("/semesters", h.ListSemesters)
	r.GET("/semesters/current", h.CurrentSemester)
	r.POST("/semesters", admin, h.CreateSemester)
	r.PUT("/semesters/:id", admin, h.UpdateSemester)
	r.POST("/semesters/:id/current", admin, h.SetCurrentSemester)
	r.DELETE("/semesters/:id", admin, h.DeleteSemester)

	r.GET("/rooms", h.ListRooms)
	r.POST("/rooms", admin, h.CreateRoom)
	r.PUT("/rooms/:id", admin, h.UpdateRoom)
	r.DELETE("/rooms/:id", admin, h.DeleteRoom)
}

// registerCatalog вешает GET/POST/PUT/DELETE для справочника «id + название».
func registerCatalog(r *gin.RouterGroup, path string, admin gin.HandlerFunc, cat catalog) {
	list, create, update, remove := catalogHandlers(cat)
	r.GET(path, list)
	r.POST(path, admin, create)
	r.PUT(path+"/:id", admin, update)
	r.DELETE(path+"/:id", admin, remove)
}
