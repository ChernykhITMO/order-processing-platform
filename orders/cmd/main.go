package main

import (
	"context"
	"log/slog"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/ChernykhITMO/order-processing-platform/orders/cmd/app"
	"github.com/ChernykhITMO/order-processing-platform/orders/internal/config"
	"github.com/joho/godotenv"
)

const (
	envKey          = "env"
	gRPCAddrKey     = "ORDERS_GRPC_ADDR"
	dsnKey          = "ORDERS_PG_DSN"
	kafkaBrokersKey = "KAFKA_BROKERS"
	kafkaTopicKey   = "KAFKA_TOPIC"
	kafkaPeriodKey  = "KAFKA_PERIOD"

	envLocal = "local"
	envDev   = "dev"
)

func main() {
	_ = godotenv.Load()

	if os.Getenv(envKey) == "" && os.Getenv("ENV") != "" {
		_ = os.Setenv(envKey, os.Getenv("ENV"))
	}

	cfg := config.MustLoad(envKey, gRPCAddrKey, dsnKey, kafkaBrokersKey, kafkaTopicKey, kafkaPeriodKey)

	if cfg == nil {
		panic("cfg is empty")
	}

	log := setupLogger(cfg.Env)

	application := app.New(log, cfg.GRPC.Port, cfg.DB.DSN, cfg.Kafka.Brokers, cfg.Kafka.Topic, cfg.Kafka.Period)

	log.With(
		slog.String("dsn", sanitizeDSN(cfg.DB.DSN)),
		slog.String("kafka_topic", cfg.Kafka.Topic),
	).Info("starting application")

	go func() {
		application.GRPCSrv.MustRun()
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		application.StartEventSender(ctx)
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop

	cancel()
	application.Stop()
	log.Info("gracefully stopped")
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
