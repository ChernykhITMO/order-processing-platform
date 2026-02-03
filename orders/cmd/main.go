package main

import (
	"context"
	"log/slog"
	"net/url"
	"os"
	"os/signal"
	"sync"
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

	cfg, err := config.Load(envKey, gRPCAddrKey, dsnKey, kafkaBrokersKey, kafkaTopicKey, kafkaPeriodKey)
	if err != nil {
		log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
		log.Error("config load failed", slog.Any("err", err))
		os.Exit(1)
	}

	log := setupLogger(cfg.Env)

	application, err := app.New(log, cfg.GRPC.Port, cfg.DB.DSN, cfg.Kafka.Brokers, cfg.Kafka.Topic, cfg.Kafka.Period)
	if err != nil {
		log.Error("app init failed", slog.Any("err", err))
		os.Exit(1)
	}

	log.With(
		slog.String("dsn", sanitizeDSN(cfg.DB.DSN)),
		slog.String("kafka_topic", cfg.Kafka.Topic),
	).Info("starting application")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errCh := make(chan error, 1)
	go func() {
		errCh <- application.GRPCSrv.Run()
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		application.StartEventSender(ctx)
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	select {
	case <-stop:
	case err := <-errCh:
		if err != nil {
			log.Error("gRPC server stopped with error", slog.Any("err", err))
		}
	}
	cancel()

	wg.Wait()
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
