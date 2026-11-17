package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"
	"university-api/internal/modules/chat"
	"university-api/internal/platform/middleware"
	"university-api/migrations"

	"github.com/gin-gonic/gin"

	"university-api/internal/config"
	"university-api/internal/modules/assignments"
	"university-api/internal/modules/auth"
	"university-api/internal/modules/content"
	"university-api/internal/modules/courses"
	"university-api/internal/modules/files"
	"university-api/internal/modules/profile"
	"university-api/internal/modules/schedule"
	"university-api/internal/modules/structure"
	"university-api/internal/modules/users"
	"university-api/internal/platform/access"
	"university-api/internal/platform/database"
	"university-api/internal/platform/mail"
	"university-api/internal/platform/ratelimit"
)

// Run собирает зависимости и запускает HTTP-сервер.
// Это единственное место, где модули знают друг о друге.
func Run(ctx context.Context, cfg config.Config) error {
	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	if cfg.MigrateOnStart {
		if err := database.Migrate(ctx, pool, migrations.FS); err != nil {
			return err
		}
	}

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
	limiter := ratelimit.New()

	// Письма: через SMTP, а без настроек — в лог сервера (для разработки).
	var mailer mail.Sender = mail.Log{}
	if cfg.SMTP.Host != "" {
		mailer = mail.SMTP{Host: cfg.SMTP.Host, Port: cfg.SMTP.Port, User: cfg.SMTP.User,
			Password: cfg.SMTP.Password, From: cfg.SMTP.From}
	}

	// Фоновая отправка писем из очереди mail_outbox; останавливается вместе с сервером.
	go mail.NewWorker(pool, mailer).Run(ctx)

	// Публичные маршруты авторизации: вход, CSRF, refresh, logout, активация.
	authSvc := auth.Register(v1, pool, auth.Options{
		Secret: secret, AccessTTL: cfg.AccessTTL, CookieSecure: cfg.CookieSecure,
		AllowedOrigins: cfg.AllowedOrigins, AppURL: cfg.AppURL,
	}, limiter)

	// Всё, что ниже, требует JWT действующей сессии.
	protected := v1.Group("", middleware.Auth(secret, authSvc))
	checker := access.NewChecker(pool)

	bells := make([]schedule.Bell, 0, len(cfg.LessonBells))
	for _, b := range cfg.LessonBells {
		bells = append(bells, schedule.Bell{Start: b[0], End: b[1]})
	}
	cal := schedule.NewCalendar(cfg.Location, cfg.WeekOneDate, cfg.LessonDuration).WithBells(bells)

	// Каждый модуль регистрируется ровно один раз: повторная регистрация
	// тех же путей роняет gin при старте ("handlers are already registered").

	// Модули, чьи сервисы нужны сборным страницам ЛК (см. cabinet.go).
	// content и assignments — раньше courses: courses получает от них загрузчик тем.
	contentSvc := content.Register(protected, pool, checker, storage)
	assignSvc := assignments.Register(protected, pool, checker, storage)
	coursesSvc := courses.Register(protected, pool, checker, courseTopics(contentSvc, assignSvc), int(cfg.LessonDuration.Minutes()))
	schedule.Register(v1, protected, pool, cal)

	// Остальные модули.
	chat.Register(protected, pool, checker, storage)
	files.Register(protected, pool, storage, cfg.MaxUploadBytes)
	structure.Register(protected, pool)                         // v1 — публичные списки кафедр и групп для регистрации
	users.Register(v1, protected, pool, checker, users.Options{ // v1 — POST /auth/register
		TermsVersion:  cfg.TermsVersion,
		Inviter:       authSvc, // создание пользователя без пароля → письмо
		RegisterLimit: limiter.Middleware("register", 10, 10*time.Minute),
		PasswordLimit: limiter.Middleware("password", 5, 15*time.Minute),
	})
	profile.Register(protected, pool, "/api/v1/files/") // GET /auth/me

	// Сборные страницы ЛК: /dashboard и /grades.
	registerCabinet(protected, coursesSvc, assignSvc)

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
