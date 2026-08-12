package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/cankonix/cankonix-oi/services/api/internal/collector"
	"github.com/cankonix/cankonix-oi/services/api/internal/config"
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
	owner := os.Getenv("COLLECTOR_ID")
	if owner == "" {
		owner = "collector-local-1"
	}
	runner := collector.NewRunner(db, collector.NewRegistry(collector.FixtureAdapter{}, collector.NewPublicSPSEAdapter(nil)), owner)
	schedulerEnabled, _ := strconv.ParseBool(os.Getenv("COLLECTOR_SCHEDULER_ENABLED"))
	monitor := collector.NewMonitor(schedulerEnabled)
	healthServer := &http.Server{Addr: value("COLLECTOR_HEALTH_ADDR", ":8081"), Handler: monitor.Handler(), ReadHeaderTimeout: 3 * time.Second, ReadTimeout: 5 * time.Second, WriteTimeout: 5 * time.Second, IdleTimeout: 30 * time.Second}
	serverErrors := make(chan error, 1)
	go func() { serverErrors <- healthServer.ListenAndServe() }()
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = healthServer.Shutdown(shutdownCtx)
	}()
	scheduler := collector.NewScheduler(db)
	scheduleTicker := time.NewTicker(time.Minute)
	defer scheduleTicker.Stop()
	if schedulerEnabled {
		result, scheduleErr := scheduler.Tick(ctx)
		monitor.ObserveSchedule(result.Queued, scheduleErr)
		if scheduleErr != nil {
			slog.Error("initial source scheduling failed", "error", scheduleErr)
		} else if result.Queued > 0 {
			slog.Info("scheduled source runs", "eligible", result.Eligible, "queued", result.Queued)
		}
	}
	slog.Info("collector started", "owner", owner, "scheduler_enabled", schedulerEnabled)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case serverErr := <-serverErrors:
			if serverErr != nil && serverErr != http.ErrServerClosed {
				slog.Error("collector health server failed", "error", serverErr)
				os.Exit(1)
			}
		default:
		}
		processed, err := runner.ExecuteOne(ctx)
		monitor.ObserveIteration(processed, err)
		if err != nil && ctx.Err() == nil {
			slog.Error("collector iteration failed", "error", err)
		}
		if processed {
			continue
		}
		select {
		case <-ctx.Done():
			return
		case <-scheduleTicker.C:
			if schedulerEnabled {
				result, scheduleErr := scheduler.Tick(ctx)
				monitor.ObserveSchedule(result.Queued, scheduleErr)
				if scheduleErr != nil {
					slog.Error("source scheduling failed", "error", scheduleErr)
				} else if result.Queued > 0 {
					slog.Info("scheduled source runs", "eligible", result.Eligible, "queued", result.Queued)
				}
			}
		case <-ticker.C:
		}
	}
}

func value(key, fallback string) string {
	if configured := os.Getenv(key); configured != "" {
		return configured
	}
	return fallback
}
