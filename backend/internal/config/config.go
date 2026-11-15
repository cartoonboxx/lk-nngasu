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
	HTTPAddr       string
	DatabaseURL    string
	JWTSecret      string
	AccessTTL      time.Duration
	UploadDir      string
	MaxUploadBytes int64

	// Расписание.
	// Авторизация по документации API.
	CookieSecure   bool     // флаг Secure у refresh-cookie (false — только для локальной отладки по http)
	AllowedOrigins []string // кроме своего адреса, откуда можно вызывать /auth/refresh и /auth/logout
	TermsVersion   string   // поддерживаемая версия условий использования
	AppURL         string   // адрес фронтенда для ссылок в письмах (…/activation?token=…)
	SMTP           SMTP     // отправка писем; пустой Host — ссылки пишутся в лог сервера

	Location       *time.Location // часовой пояс вуза
	WeekOneDate    time.Time      // любая дата первой (верхней) недели; нулевая — чётность по ISO-неделям
	LessonDuration time.Duration  // длительность пары (в БД хранится только время начала)
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

	var err error
	// Access-токен короткий: продлевается через POST /auth/refresh по refresh-cookie.
	if cfg.AccessTTL, err = time.ParseDuration(getenv("ACCESS_TTL", "15m")); err != nil {
		return Config{}, fmt.Errorf("ACCESS_TTL: %w", err)
	}

	cfg.CookieSecure = getenv("COOKIE_SECURE", "true") != "false"
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

	mb, err := strconv.ParseInt(getenv("MAX_UPLOAD_MB", "15"), 10, 64)
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
	return cfg, nil
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
