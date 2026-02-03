package main

import (
	"context"
	"errors"
	"log/slog"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/ChernykhITMO/order-processing-platform/payments/cmd/app"
	"github.com/ChernykhITMO/order-processing-platform/payments/internal/config"
	"github.com/joho/godotenv"
)

const (
	envKey = "env"

	envLocal = "local"
	envDev   = "dev"
)

func main() {
	_ = godotenv.Load()

	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	if os.Getenv(envKey) == "" && os.Getenv("ENV") != "" {
		_ = os.Setenv(envKey, os.Getenv("ENV"))
	}
	envVal := os.Getenv(envKey)
	if envVal == "" {
		log.Error("config load failed", slog.Any("err", errors.New("env is empty")))
		os.Exit(1)
	}
	log = setupLogger(envVal)
	log.Info("service starting")

	cfg, err := config.Load()
	if err != nil {
		log.Error("config load failed", slog.Any("err", err))
		os.Exit(1)
	}

	application, err := app.New(log, cfg)

	log.Info("config",
		slog.String("db", sanitizeDSN(cfg.DBDSN)),
		slog.Any("brokers", cfg.KafkaBrokers),
	)

	if err != nil {
		log.Error("failed to create application", slog.Any("err", err))
		os.Exit(1)
	}
	defer application.Stop()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := application.Run(ctx); err != nil {
		log.Error("application stopped with error", slog.Any("err", err))
		os.Exit(1)
	}
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(slog.NewTextHandler(
			os.Stdout,
			&slog.HandlerOptions{Level: slog.LevelDebug}))
	case envDev:
		log = slog.New(slog.NewJSONHandler(
			os.Stdout,
			&slog.HandlerOptions{Level: slog.LevelDebug}))
	default:
		log = slog.New(slog.NewTextHandler(
			os.Stdout,
			&slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return log
}


func sanitizeDSN(dsn string) string {
	u, err := url.Parse(dsn)
	if err != nil {
		return dsn
	}
	if u.User != nil {
		user := u.User.Username()
		u.User = url.User(user)
	}
	return u.String()
}
