package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/cankonix/cankonix-oi/services/api/internal/config"
	"github.com/cankonix/cankonix-oi/services/api/internal/notifications"
	"github.com/cankonix/cankonix-oi/services/api/internal/platform/database"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("invalid configuration", "error", err)
		os.Exit(1)
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	db, err := database.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("database unavailable", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	emailEnabled, _ := strconv.ParseBool(os.Getenv("EMAIL_DELIVERY_ENABLED"))
	webhookEnabled, _ := strconv.ParseBool(os.Getenv("WEBHOOK_DELIVERY_ENABLED"))
	if !emailEnabled && !webhookEnabled {
		slog.Info("notifier disabled")
		<-ctx.Done()
		return
	}
	owner := value("NOTIFIER_ID", "notifier-local-1")
	var emailDispatcher *notifications.Dispatcher
	emailDryRun, _ := strconv.ParseBool(os.Getenv("EMAIL_DRY_RUN"))
	var sender notifications.EmailSender = notifications.DryRunSender{}
	if emailEnabled && !emailDryRun {
		if cfg.Environment == "production" && os.Getenv("SMTP_PASSWORD") != "" {
			slog.Error("invalid SMTP secret configuration", "error", "SMTP_PASSWORD_FILE must be used in production")
			os.Exit(1)
		}
		port, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))
		smtpPassword, secretErr := config.Secret("SMTP_PASSWORD")
		if secretErr != nil {
			slog.Error("invalid SMTP secret configuration", "error", secretErr)
			os.Exit(1)
		}
		smtpConfig := notifications.SMTPConfig{Host: os.Getenv("SMTP_HOST"), Port: port, Username: os.Getenv("SMTP_USERNAME"), Password: smtpPassword, From: os.Getenv("SMTP_FROM"), Timeout: 20 * time.Second}
		if err := smtpConfig.Validate(); err != nil {
			slog.Error("invalid SMTP configuration", "error", err)
			os.Exit(1)
		}
		sender = notifications.SMTPSender{Config: smtpConfig}
	}
	if emailEnabled {
		emailDispatcher = notifications.NewDispatcher(db, sender, owner)
	}
	var webhookDispatcher *notifications.WebhookDispatcher
	webhookDryRun, _ := strconv.ParseBool(os.Getenv("WEBHOOK_DRY_RUN"))
	webhookURL := os.Getenv("WEBHOOK_URL")
	var webhookSender notifications.WebhookSender = notifications.DryRunWebhookSender{}
	if webhookEnabled && !webhookDryRun {
		if cfg.Environment == "production" && os.Getenv("WEBHOOK_SECRET") != "" {
			slog.Error("invalid webhook secret configuration", "error", "WEBHOOK_SECRET_FILE must be used in production")
			os.Exit(1)
		}
		webhookSecret, secretErr := config.Secret("WEBHOOK_SECRET")
		if secretErr != nil {
			slog.Error("invalid webhook secret configuration", "error", secretErr)
			os.Exit(1)
		}
		webhookConfig := notifications.WebhookConfig{URL: webhookURL, Secret: webhookSecret, Timeout: 20 * time.Second}
		if err := webhookConfig.Validate(); err != nil {
			slog.Error("invalid webhook configuration", "error", err)
			os.Exit(1)
		}
		webhookSender = notifications.HTTPWebhookSender{Config: webhookConfig}
	}
	if webhookEnabled {
		if webhookDryRun && webhookURL == "" {
			webhookURL = "dry-run://webhook"
		}
		webhookDispatcher = notifications.NewWebhookDispatcher(db, webhookSender, owner, webhookURL)
	}
	slog.Info("notifier started", "email_enabled", emailEnabled, "email_dry_run", emailDryRun, "webhook_enabled", webhookEnabled, "webhook_dry_run", webhookDryRun)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		processed := false
		if emailDispatcher != nil {
			done, runErr := emailDispatcher.RunOne(ctx)
			processed = processed || done
			if runErr != nil && ctx.Err() == nil {
				slog.Warn("email delivery failed", "error", runErr)
			}
		}
		if webhookDispatcher != nil {
			done, runErr := webhookDispatcher.RunOne(ctx)
			processed = processed || done
			if runErr != nil && ctx.Err() == nil {
				slog.Warn("webhook delivery failed", "error", runErr)
			}
		}
		if processed {
			continue
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
func value(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
