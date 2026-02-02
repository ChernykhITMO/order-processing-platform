package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

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

	kafkaBrokers := parseKafkaBrokers(mustGetEnv("KAFKA_BROKERS"))
	if len(kafkaBrokers) == 0 {
		log.Fatal("KAFKA_BROKERS is empty")
	}

	if os.Getenv(envKey) == "" && os.Getenv("ENV") != "" {
		_ = os.Setenv(envKey, os.Getenv("ENV"))
	}

	log := setupLogger(mustGetEnv(envKey))

	cfg := config.Config{
		Addr:           mustGetEnv("REDIS_ADDR"),
		DB:             mustGetEnvInt("REDIS_DB"),
		TTL:            getEnvDuration("REDIS_TTL", 0),
		SessionTimeout: getEnvDuration("SESSION_TIMEOUT", 30*time.Second),
		KafkaBrokers:   kafkaBrokers,
		TopicStatus:    mustGetEnv("KAFKA_TOPIC_STATUS"),
		ConsumerGroup:  mustGetEnv("KAFKA_CONSUMER_GROUP"),
		ReadTimeout:    getEnvDuration("READ_TIMEOUT", time.Second),
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
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	application.MustRun(ctx)
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

func parseKafkaBrokers(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func mustGetEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("%s is empty", key)
	}
	return val
}

func mustGetEnvInt(key string) int {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("%s is empty", key)
	}
	parsed, err := strconv.Atoi(val)
	if err != nil {
		log.Fatalf("%s is invalid int: %v", key, err)
	}
	return parsed
}

func getEnvDuration(key string, def time.Duration) time.Duration {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	parsed, err := time.ParseDuration(val)
	if err != nil {
		return def
	}
	return parsed
}
