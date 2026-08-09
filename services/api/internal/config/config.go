package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	Environment string
	Address     string
	DatabaseURL string
	WebOrigin   string
	JWTSecret   string
	AccessTTL   time.Duration
	RefreshTTL  time.Duration
	AdminEmail  string
	AdminPass   string
	AdminName   string
}

func Load() (Config, error) {
	cfg := Config{
		Environment: value("APP_ENV", "development"),
		Address:     value("API_ADDR", ":8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		WebOrigin:   value("WEB_ORIGIN", "http://localhost:3000"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		AdminEmail:  os.Getenv("BOOTSTRAP_ADMIN_EMAIL"),
		AdminPass:   os.Getenv("BOOTSTRAP_ADMIN_PASSWORD"),
		AdminName:   value("BOOTSTRAP_ADMIN_NAME", "CANKONIX Administrator"),
	}
	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	if len(cfg.JWTSecret) < 32 {
		return Config{}, fmt.Errorf("JWT_SECRET must contain at least 32 characters")
	}
	var err error
	if cfg.AccessTTL, err = time.ParseDuration(value("ACCESS_TOKEN_TTL", "15m")); err != nil || cfg.AccessTTL <= 0 {
		return Config{}, fmt.Errorf("ACCESS_TOKEN_TTL must be a positive duration")
	}
	if cfg.RefreshTTL, err = time.ParseDuration(value("REFRESH_TOKEN_TTL", "168h")); err != nil || cfg.RefreshTTL <= 0 {
		return Config{}, fmt.Errorf("REFRESH_TOKEN_TTL must be a positive duration")
	}
	return cfg, nil
}

func value(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
