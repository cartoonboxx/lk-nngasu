package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
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
	Location       *time.Location // часовой пояс вуза
	WeekOneDate    time.Time      // любая дата первой (верхней) недели; нулевая — чётность по ISO-неделям
	LessonDuration time.Duration  // длительность пары (в БД хранится только время начала)
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
	if cfg.AccessTTL, err = time.ParseDuration(getenv("ACCESS_TTL", "12h")); err != nil {
		return Config{}, fmt.Errorf("ACCESS_TTL: %w", err)
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
