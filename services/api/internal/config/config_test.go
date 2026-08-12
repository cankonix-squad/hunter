package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadRequiresDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("DATABASE_URL_FILE", "")
	t.Setenv("JWT_SECRET", "01234567890123456789012345678901")
	t.Setenv("JWT_SECRET_FILE", "")
	t.Setenv("BOOTSTRAP_ADMIN_PASSWORD_FILE", "")
	if _, err := Load(); err == nil {
		t.Fatal("expected missing DATABASE_URL to fail")
	}
}

func TestLoadUsesSafeDefaults(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("DATABASE_URL_FILE", "")
	t.Setenv("JWT_SECRET", "01234567890123456789012345678901")
	t.Setenv("JWT_SECRET_FILE", "")
	t.Setenv("BOOTSTRAP_ADMIN_PASSWORD_FILE", "")
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

func TestSecretReadsFileAndTrimsLineEnding(t *testing.T) {
	path := writeSecret(t, "secret-value\n")
	t.Setenv("TEST_SECRET", "")
	t.Setenv("TEST_SECRET_FILE", path)

	got, err := Secret("TEST_SECRET")
	if err != nil {
		t.Fatalf("Secret() error = %v", err)
	}
	if got != "secret-value" {
		t.Fatalf("Secret() = %q", got)
	}
}

func TestSecretRejectsAmbiguousSourcesWithoutLeakingValue(t *testing.T) {
	const sensitive = "do-not-print-this-secret"
	t.Setenv("TEST_SECRET", sensitive)
	t.Setenv("TEST_SECRET_FILE", writeSecret(t, "another-secret"))

	_, err := Secret("TEST_SECRET")
	if err == nil {
		t.Fatal("expected ambiguous secret sources to fail")
	}
	if strings.Contains(err.Error(), sensitive) {
		t.Fatalf("error leaked secret: %v", err)
	}
}

func TestLoadProductionRequiresFileBackedSecrets(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("DATABASE_URL", "postgres://user:strong-password@postgres/db")
	t.Setenv("DATABASE_URL_FILE", "")
	t.Setenv("JWT_SECRET", "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	t.Setenv("JWT_SECRET_FILE", "")
	t.Setenv("BOOTSTRAP_ADMIN_EMAIL", "")
	t.Setenv("BOOTSTRAP_ADMIN_PASSWORD", "")
	t.Setenv("BOOTSTRAP_ADMIN_PASSWORD_FILE", "")

	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "DATABASE_URL_FILE") {
		t.Fatalf("expected production file requirement, got %v", err)
	}
}

func TestLoadAcceptsProductionFileBackedSecrets(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("DATABASE_URL", "")
	t.Setenv("DATABASE_URL_FILE", writeSecret(t, "postgres://user:strong-password@postgres/db\n"))
	t.Setenv("JWT_SECRET", "")
	t.Setenv("JWT_SECRET_FILE", writeSecret(t, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_\n"))
	t.Setenv("BOOTSTRAP_ADMIN_EMAIL", "")
	t.Setenv("BOOTSTRAP_ADMIN_PASSWORD", "")
	t.Setenv("BOOTSTRAP_ADMIN_PASSWORD_FILE", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Environment != "production" || cfg.DatabaseURL == "" || cfg.JWTSecret == "" {
		t.Fatalf("unexpected config: %#v", cfg)
	}
}

func TestLoadRejectsWeakProductionSecret(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("DATABASE_URL", "")
	t.Setenv("DATABASE_URL_FILE", writeSecret(t, "postgres://user:strong-password@postgres/db"))
	t.Setenv("JWT_SECRET", "")
	t.Setenv("JWT_SECRET_FILE", writeSecret(t, strings.Repeat("a", 64)))
	t.Setenv("BOOTSTRAP_ADMIN_EMAIL", "")
	t.Setenv("BOOTSTRAP_ADMIN_PASSWORD", "")
	t.Setenv("BOOTSTRAP_ADMIN_PASSWORD_FILE", "")

	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "high-entropy") {
		t.Fatalf("expected weak secret rejection, got %v", err)
	}
}

func TestLoadRejectsDevelopmentDatabaseCredentialInProduction(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("DATABASE_URL", "")
	t.Setenv("DATABASE_URL_FILE", writeSecret(t, "postgres://user:change-me-for-local-development@postgres/db"))
	t.Setenv("JWT_SECRET", "")
	t.Setenv("JWT_SECRET_FILE", writeSecret(t, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_"))
	t.Setenv("BOOTSTRAP_ADMIN_EMAIL", "")
	t.Setenv("BOOTSTRAP_ADMIN_PASSWORD", "")
	t.Setenv("BOOTSTRAP_ADMIN_PASSWORD_FILE", "")

	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "forbidden development credential") {
		t.Fatalf("expected development credential rejection, got %v", err)
	}
}

func writeSecret(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "secret")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}
