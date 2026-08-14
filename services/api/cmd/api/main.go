package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/cankonix/cankonix-oi/services/api/internal/collector"
	"github.com/cankonix/cankonix-oi/services/api/internal/config"
	"github.com/cankonix/cankonix-oi/services/api/internal/httpapi"
	"github.com/cankonix/cankonix-oi/services/api/internal/identity"
	"github.com/cankonix/cankonix-oi/services/api/internal/platform/database"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		healthcheck()
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "provision-user" {
		provisionUser(os.Args[2:])
		return
	}

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
	identityService := identity.New(db, cfg.JWTSecret, cfg.AccessTTL, cfg.RefreshTTL)
	if err := identityService.BootstrapAdmin(ctx, cfg.AdminEmail, cfg.AdminName, cfg.AdminPass); err != nil {
		slog.Error("administrator bootstrap failed", "error", err)
		os.Exit(1)
	}
	collectorRunner := collector.NewRunner(db, collector.NewRegistry(collector.NewPublicSPSEAdapter(nil)), "api-queue")

	server := &http.Server{
		Addr:              cfg.Address,
		Handler:           httpapi.New(db, identityService, collectorRunner),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	slog.Info("api started", "address", cfg.Address, "environment", cfg.Environment)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("api stopped unexpectedly", "error", err)
		os.Exit(1)
	}
}

func provisionUser(args []string) {
	flags := flag.NewFlagSet("provision-user", flag.ExitOnError)
	email := flags.String("email", "", "user email")
	name := flags.String("name", "", "display name")
	rolesValue := flags.String("roles", "", "comma-separated roles; empty creates a viewer")
	passwordFile := flags.String("password-file", "", "path to password file")
	_ = flags.Parse(args)
	if *passwordFile == "" {
		slog.Error("user provisioning failed", "error", "password-file is required")
		os.Exit(2)
	}
	passwordBytes, err := os.ReadFile(*passwordFile)
	if err != nil {
		slog.Error("user provisioning failed", "error", "password file is unavailable")
		os.Exit(1)
	}
	cfg, err := config.Load()
	if err != nil {
		slog.Error("invalid configuration", "error", err)
		os.Exit(1)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	db, err := database.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("database unavailable", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	roles := []string{}
	if strings.TrimSpace(*rolesValue) != "" {
		roles = strings.Split(*rolesValue, ",")
	}
	user, err := identity.New(db, cfg.JWTSecret, cfg.AccessTTL, cfg.RefreshTTL).ProvisionUser(
		ctx, *email, *name, strings.TrimSpace(string(passwordBytes)), roles)
	if err != nil {
		slog.Error("user provisioning failed", "error", err)
		os.Exit(1)
	}
	fmt.Printf("provisioned_user=%s email=%s roles=%s\n", user.UserID, user.Email, strings.Join(user.Roles, ","))
}

func healthcheck() {
	client := http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://127.0.0.1:8080/healthz")
	if err != nil || resp.StatusCode != http.StatusOK {
		os.Exit(1)
	}
	_ = resp.Body.Close()
}
