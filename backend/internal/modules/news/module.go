package news

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"university-api/internal/platform/middleware"
)

// Register: public — лента без авторизации, protected — редактор (только администратор).
// Пути редактора отделены (/admin/news), чтобы черновики не попали в публичную ленту случайно.
// Возвращает Feed — свежие новости для главной страницы (GET /dashboard).
func Register(public, protected *gin.RouterGroup, pool *pgxpool.Pool) *Feed {
	h := &Handler{repo: newRepository(pool)}

	public.GET("/news", h.Feed)
	public.GET("/news/:id", h.Show)

	admin := protected.Group("/admin/news", middleware.RequireRole(middleware.RoleAdmin))
	admin.GET("", h.AdminList)
	admin.POST("", h.Create)
	admin.GET("/:id", h.AdminGet)
	admin.PUT("/:id", h.Replace)
	admin.PATCH("/:id", h.SetStatus)
	admin.DELETE("/:id", h.Delete)

	return &Feed{repo: h.repo}
}

// Feed — публичная лента для других страниц.
type Feed struct{ repo *Repository }

// Latest — n последних опубликованных новостей в порядке ленты.
func (f *Feed) Latest(ctx context.Context, n int) ([]NewsSummary, error) {
	list, _, err := f.repo.List(ctx, ListQuery{Public: true, Limit: n})
	if err != nil {
		return nil, err
	}
	return summaries(list), nil
}
