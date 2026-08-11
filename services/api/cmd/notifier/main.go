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
	enabled, _ := strconv.ParseBool(os.Getenv("EMAIL_DELIVERY_ENABLED"))
	if !enabled {
		slog.Info("notifier disabled")
		<-ctx.Done()
		return
	}
	dryRun, _ := strconv.ParseBool(os.Getenv("EMAIL_DRY_RUN"))
	var sender notifications.EmailSender = notifications.DryRunSender{}
	if !dryRun {
		port, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))
		sender = notifications.SMTPSender{Config: notifications.SMTPConfig{Host: os.Getenv("SMTP_HOST"), Port: port, Username: os.Getenv("SMTP_USERNAME"), Password: os.Getenv("SMTP_PASSWORD"), From: os.Getenv("SMTP_FROM"), Timeout: 20 * time.Second}}
	}
	dispatcher := notifications.NewDispatcher(db, sender, value("NOTIFIER_ID", "notifier-local-1"))
	slog.Info("notifier started", "dry_run", dryRun)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		processed, runErr := dispatcher.RunOne(ctx)
		if runErr != nil && ctx.Err() == nil {
			slog.Warn("email delivery failed", "error", runErr)
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
