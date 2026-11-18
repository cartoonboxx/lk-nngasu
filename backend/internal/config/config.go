package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AdminSetupKey  string // одноразовая первоначальная настройка; пусто — отключена
	HTTPAddr       string
	DatabaseURL    string
	JWTSecret      string
	AccessTTL      time.Duration
	UploadDir      string
	MaxUploadBytes int64
	// MigrateOnStart — применять новые миграции при старте API (MIGRATE_ON_START, по умолчанию true).
	MigrateOnStart bool

	// Расписание.
	// Авторизация по документации API.
	CookieSecure   bool     // флаг Secure у refresh-cookie (false — только для локальной отладки по http)
	AllowedOrigins []string // кроме своего адреса, откуда можно вызывать /auth/refresh и /auth/logout
	TermsVersion   string   // поддерживаемая версия условий использования
	AppURL         string   // адрес фронтенда для ссылок в письмах (…/activation?token=…)
	SMTP           SMTP     // отправка писем; пустой Host — доставка отключена

	Location       *time.Location // часовой пояс вуза
	WeekOneDate    time.Time      // любая дата первой (верхней) недели; нулевая — календарь не настроен (409)
	LessonDuration time.Duration  // длительность пары: для расчёта часов нагрузки и старых записей
	// LessonBells — сетка звонков из LESSON_BELLS ("08:00-09:30,09:45-11:15,…"):
	// занятие можно поставить только на одну из этих пар. Пусто — не проверяется.
	LessonBells [][2]string
}

// SMTP — почтовый сервер для писем-приглашений (порт 587, STARTTLS).
type SMTP struct {
	Host, Port, User, Password, From string
}

func Load() (Config, error) {
	cfg := Config{
		HTTPAddr:    getenv("HTTP_ADDR", ":8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		UploadDir:   getenv("UPLOAD_DIR", "./uploads"),
	}
	if cfg.DatabaseURL == "" || cfg.JWTSecret == "" {
		return Config{}, errors.New("DATABASE_URL and JWT_SECRET are required")
	}
	if len(cfg.JWTSecret) < 32 {
		return Config{}, errors.New("JWT_SECRET must be at least 32 characters")
	}

	cfg.AdminSetupKey = os.Getenv("ADMIN_SETUP_KEY")
	if cfg.AdminSetupKey != "" && (len(cfg.AdminSetupKey) < 32 || len(cfg.AdminSetupKey) > 256) {
		return Config{}, errors.New("ADMIN_SETUP_KEY must be 32..256 bytes long")
	}

	var err error
	// Access-токен короткий: продлевается через POST /auth/refresh по refresh-cookie.
	if cfg.AccessTTL, err = time.ParseDuration(getenv("ACCESS_TTL", "15m")); err != nil {
		return Config{}, fmt.Errorf("ACCESS_TTL: %w", err)
	}

	cfg.CookieSecure = getenv("COOKIE_SECURE", "true") != "false"
	cfg.MigrateOnStart = getenv("MIGRATE_ON_START", "true") != "false"
	for _, o := range strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",") {
		if o = strings.TrimRight(strings.TrimSpace(o), "/"); o != "" {
			cfg.AllowedOrigins = append(cfg.AllowedOrigins, o)
		}
	}
	cfg.TermsVersion = getenv("TERMS_VERSION", "2026-11-13") // как в документации и конфигурации фронта
	cfg.AppURL = strings.TrimRight(getenv("APP_URL", "http://localhost:5173"), "/")
	cfg.SMTP = SMTP{
		Host:     os.Getenv("SMTP_HOST"),
		Port:     getenv("SMTP_PORT", "587"),
		User:     os.Getenv("SMTP_USER"),
		Password: os.Getenv("SMTP_PASSWORD"),
		From:     os.Getenv("SMTP_FROM"),
	}
	if cfg.SMTP.Host != "" && cfg.SMTP.From == "" {
		cfg.SMTP.From = cfg.SMTP.User
	}

	mb, err := strconv.ParseInt(getenv("MAX_UPLOAD_MB", "20"), 10, 64)
	if err != nil || mb < 1 {
		return Config{}, errors.New("MAX_UPLOAD_MB must be a positive integer")
	}
	cfg.MaxUploadBytes = mb << 20

	if cfg.Location, err = time.LoadLocation(getenv("TIMEZONE", "Europe/Moscow")); err != nil {
		return Config{}, fmt.Errorf("TIMEZONE: %w", err)
	}
	if raw := os.Getenv("WEEK_ONE_DATE"); raw != "" {
		if cfg.WeekOneDate, err = time.Parse(time.DateOnly, raw); err != nil {
			return Config{}, errors.New("WEEK_ONE_DATE must look like 2026-09-01")
		}
	}
	mins, err := strconv.Atoi(getenv("LESSON_MINUTES", "90"))
	if err != nil || mins < 1 || mins > 600 {
		return Config{}, errors.New("LESSON_MINUTES must be between 1 and 600")
	}
	cfg.LessonDuration = time.Duration(mins) * time.Minute
	if raw := strings.TrimSpace(os.Getenv("LESSON_BELLS")); raw != "" {
		for _, pair := range strings.Split(raw, ",") {
			se := strings.Split(strings.TrimSpace(pair), "-")
			if len(se) != 2 {
				return Config{}, errors.New("LESSON_BELLS must look like 08:00-09:30,09:45-11:15")
			}
			start, err1 := time.Parse("15:04", strings.TrimSpace(se[0]))
			end, err2 := time.Parse("15:04", strings.TrimSpace(se[1]))
			if err1 != nil || err2 != nil || !end.After(start) {
				return Config{}, errors.New("LESSON_BELLS: bad pair " + pair)
			}
			cfg.LessonBells = append(cfg.LessonBells, [2]string{start.Format("15:04"), end.Format("15:04")})
		}
	}
	return cfg, nil
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
