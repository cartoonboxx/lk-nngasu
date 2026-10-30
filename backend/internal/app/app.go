package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"university-api/internal/config"
	"university-api/internal/modules/assignments"
	"university-api/internal/modules/auth"
	"university-api/internal/modules/content"
	"university-api/internal/modules/courses"
	"university-api/internal/modules/files"
	"university-api/internal/modules/schedule"
	"university-api/internal/modules/structure"
	"university-api/internal/modules/users"
	"university-api/internal/platform/access"
	"university-api/internal/platform/database"
	"university-api/internal/platform/middleware"
)

// Run собирает зависимости и запускает HTTP-сервер.
// Это единственное место, где модули знают друг о друге.
func Run(ctx context.Context, cfg config.Config) error {
	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	storage, err := files.NewStorage(cfg.UploadDir)
	if err != nil {
		return err
	}

	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger())
	for _, path := range []string{"/healthz", "/api/healthz"} {
		r.GET(path, func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	}

	secret := []byte(cfg.JWTSecret)
	v1 := r.Group("/api/v1")

	// Публичные маршруты.
	auth.Register(v1, pool, secret, cfg.AccessTTL)

	// Всё, что ниже, требует JWT.
	protected := v1.Group("", middleware.Auth(secret))
	checker := access.NewChecker(pool)

	cal := schedule.NewCalendar(cfg.Location, cfg.WeekOneDate, cfg.LessonDuration)

	structure.Register(protected, pool)
	courses.Register(protected, pool, checker)
	users.Register(protected, pool, checker)
	schedule.Register(v1, protected, pool, cal)
	content.Register(protected, pool, checker, storage)
	assignments.Register(protected, pool, checker, storage)
	files.Register(protected, pool, storage, cfg.MaxUploadBytes)

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		slog.Info("listening", "addr", cfg.HTTPAddr)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}
}
