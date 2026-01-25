package main

import (
	"context"
	"log"
	"log/slog"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

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

	dsn := os.Getenv("PAYMENTS_PG_DSN")
	if dsn == "" {
		log.Fatal("PAYMENTS_PG_DSN is empty")
	}

	kafkaBrokers := parseKafkaBrokers(mustGetEnv("KAFKA_BROKERS"))
	if len(kafkaBrokers) == 0 {
		log.Fatal("KAFKA_BROKERS is empty")
	}

	if os.Getenv(envKey) == "" && os.Getenv("ENV") != "" {
		_ = os.Setenv(envKey, os.Getenv("ENV"))
	}
	log := setupLogger(mustGetEnv(envKey))
	log.Info("service starting")

	application, err := app.New(log, config.Config{
		DBDSN:         dsn,
		KafkaBrokers:  kafkaBrokers,
		TopicOrder:    mustGetEnv("KAFKA_TOPIC_ORDER"),
		TopicStatus:   mustGetEnv("KAFKA_TOPIC_STATUS"),
		EventType:     mustGetEnv("KAFKA_EVENT_TYPE"),
		ConsumerGroup: mustGetEnv("KAFKA_CONSUMER_GROUP"),
		SenderPeriod:  mustGetEnvDuration("KAFKA_SENDER_PERIOD"),
	})

	log.Info("config",
		slog.String("db", sanitizeDSN(dsn)),
		slog.Any("brokers", kafkaBrokers),
	)

	if err != nil {
		log.Error("failed to create application", slog.Any("err", err))
		return
	}
	defer application.Stop()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

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

func mustGetEnvDuration(key string) time.Duration {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("%s is empty", key)
	}
	parsed, err := time.ParseDuration(val)
	if err != nil {
		log.Fatalf("%s is invalid duration: %v", key, err)
	}
	return parsed
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
