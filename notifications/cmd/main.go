package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/ChernykhITMO/order-processing-platform/notifications/cmd/app"
	"github.com/ChernykhITMO/order-processing-platform/notifications/internal/config"
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

	cfg, err := config.Load()
	if err != nil {
		log.Error("config load failed", slog.Any("err", err))
		os.Exit(1)
	}

	log = log.With(
		slog.String("redis_addr", cfg.Addr),
		slog.String("kafka_topic", cfg.TopicStatus),
		slog.String("kafka_consumer_group", cfg.ConsumerGroup),
		slog.Duration("session_timeout", cfg.SessionTimeout),
	)

	log.Debug("starting application")

	application, err := app.New(log, cfg)
	if err != nil {
		log.Error("application failed", slog.Any("err", err))
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

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
