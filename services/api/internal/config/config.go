package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

const maxSecretFileSize = 64 * 1024

var developmentSecrets = []string{
	"change-me-for-local-development",
	"local-development-secret-change-me-now",
	"replace-with-at-least-32-random-characters",
	"replace-this-before-first-start",
}

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
	databaseURL, databaseFromFile, err := secret("DATABASE_URL")
	if err != nil {
		return Config{}, err
	}
	jwtSecret, jwtFromFile, err := secret("JWT_SECRET")
	if err != nil {
		return Config{}, err
	}
	adminPass, adminPassFromFile, err := secret("BOOTSTRAP_ADMIN_PASSWORD")
	if err != nil {
		return Config{}, err
	}
	cfg := Config{
		Environment: strings.ToLower(value("APP_ENV", "development")),
		Address:     value("API_ADDR", ":8080"),
		DatabaseURL: databaseURL,
		WebOrigin:   value("WEB_ORIGIN", "http://localhost:3000"),
		JWTSecret:   jwtSecret,
		AdminEmail:  os.Getenv("BOOTSTRAP_ADMIN_EMAIL"),
		AdminPass:   adminPass,
		AdminName:   value("BOOTSTRAP_ADMIN_NAME", "CANKONIX Administrator"),
	}
	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	if len(cfg.JWTSecret) < 32 {
		return Config{}, fmt.Errorf("JWT_SECRET must contain at least 32 characters")
	}
	if cfg.Environment == "production" {
		if !databaseFromFile {
			return Config{}, fmt.Errorf("DATABASE_URL_FILE is required in production")
		}
		if isDevelopmentSecret(cfg.DatabaseURL) {
			return Config{}, fmt.Errorf("DATABASE_URL_FILE contains a forbidden development credential")
		}
		if !jwtFromFile {
			return Config{}, fmt.Errorf("JWT_SECRET_FILE is required in production")
		}
		if len(cfg.JWTSecret) < 43 || distinctBytes(cfg.JWTSecret) < 12 || isDevelopmentSecret(cfg.JWTSecret) {
			return Config{}, fmt.Errorf("JWT_SECRET_FILE must contain a high-entropy secret of at least 43 characters")
		}
		if cfg.AdminPass != "" && !adminPassFromFile {
			return Config{}, fmt.Errorf("BOOTSTRAP_ADMIN_PASSWORD_FILE is required when bootstrap is enabled in production")
		}
		if cfg.AdminPass != "" && (len(cfg.AdminPass) < 20 || distinctBytes(cfg.AdminPass) < 10 || isDevelopmentSecret(cfg.AdminPass)) {
			return Config{}, fmt.Errorf("BOOTSTRAP_ADMIN_PASSWORD_FILE must contain a high-entropy secret of at least 20 characters")
		}
		if (cfg.AdminEmail == "") != (cfg.AdminPass == "") {
			return Config{}, fmt.Errorf("BOOTSTRAP_ADMIN_EMAIL and BOOTSTRAP_ADMIN_PASSWORD_FILE must be configured together")
		}
	}
	if cfg.AccessTTL, err = time.ParseDuration(value("ACCESS_TOKEN_TTL", "15m")); err != nil || cfg.AccessTTL <= 0 {
		return Config{}, fmt.Errorf("ACCESS_TOKEN_TTL must be a positive duration")
	}
	if cfg.RefreshTTL, err = time.ParseDuration(value("REFRESH_TOKEN_TTL", "168h")); err != nil || cfg.RefreshTTL <= 0 {
		return Config{}, fmt.Errorf("REFRESH_TOKEN_TTL must be a positive duration")
	}
	return cfg, nil
}

// Secret reads KEY or the contents of KEY_FILE. It never includes secret contents in errors.
func Secret(key string) (string, error) {
	value, _, err := secret(key)
	return value, err
}

func secret(key string) (string, bool, error) {
	direct := os.Getenv(key)
	path := os.Getenv(key + "_FILE")
	if direct != "" && path != "" {
		return "", false, fmt.Errorf("%s and %s_FILE cannot both be set", key, key)
	}
	if path == "" {
		return direct, false, nil
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", false, fmt.Errorf("read %s_FILE: %w", key, err)
	}
	if !info.Mode().IsRegular() || info.Size() > maxSecretFileSize {
		return "", false, fmt.Errorf("%s_FILE must reference a regular file no larger than %d bytes", key, maxSecretFileSize)
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		return "", false, fmt.Errorf("read %s_FILE: %w", key, err)
	}
	secretValue := strings.TrimRight(string(contents), "\r\n")
	if secretValue == "" {
		return "", false, fmt.Errorf("%s_FILE is empty", key)
	}
	if strings.ContainsRune(secretValue, '\x00') {
		return "", false, fmt.Errorf("%s_FILE contains invalid data", key)
	}
	return secretValue, true, nil
}

func isDevelopmentSecret(value string) bool {
	lower := strings.ToLower(value)
	for _, forbidden := range developmentSecrets {
		if strings.Contains(lower, forbidden) {
			return true
		}
	}
	return false
}

func distinctBytes(value string) int {
	seen := make(map[byte]struct{})
	for i := range len(value) {
		seen[value[i]] = struct{}{}
	}
	return len(seen)
}

func value(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
