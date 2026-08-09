package config

import "testing"

func TestLoadRequiresDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("JWT_SECRET", "01234567890123456789012345678901")
	if _, err := Load(); err == nil {
		t.Fatal("expected missing DATABASE_URL to fail")
	}
}

func TestLoadUsesSafeDefaults(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("JWT_SECRET", "01234567890123456789012345678901")
	t.Setenv("APP_ENV", "")
	t.Setenv("API_ADDR", "")
	t.Setenv("WEB_ORIGIN", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Environment != "development" || cfg.Address != ":8080" {
		t.Fatalf("unexpected defaults: %#v", cfg)
	}
}
