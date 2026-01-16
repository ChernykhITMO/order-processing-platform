package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/ChernykhITMO/order-processing-platform/payments/internal/handler"
	"github.com/ChernykhITMO/order-processing-platform/payments/internal/kafka"
	"github.com/ChernykhITMO/order-processing-platform/payments/internal/kafka_consume"
	"github.com/ChernykhITMO/order-processing-platform/payments/internal/storage/postgres"
	"github.com/joho/godotenv"
)

var address = []string{"localhost:9092"}

const (
	topicOrder  = "order-topic"
	topicStatus = "status-topic"

	envLocal = "local"
	envDev   = "dev"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("env not loaded")
	}

	p, err := kafka.NewProducer(address)
	if err != nil {
		log.Fatal(err)
	}

	DSN := os.Getenv("PAYMENTS_PG_DSN")
	if DSN == "" {
		log.Fatal("PAYMENTS_PG_DSN is empty")
	}
	storage, err := postgres.New(DSN)
	if err != nil {
		log.Fatal(err)
	}

	logger := setupLogger(envLocal)
	logger.Info("service starting")

	h := handler.NewHandler(storage, p, topicStatus, logger)
	c, err := kafka_consume.NewConsumer(h, address, topicOrder, "my-group", logger)
	if err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		defer wg.Done()
		c.Start(ctx)
	}()

	<-ctx.Done()
	if err := c.Stop(); err != nil {
		logger.Error("consumer stopping: ", err)
	}

	p.Close()
	wg.Wait()
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
	}

	return log
}
