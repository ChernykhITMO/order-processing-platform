package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/ChernykhITMO/order-processing-platform/notifications/internal/config"
	"github.com/ChernykhITMO/order-processing-platform/notifications/internal/controller"
	"github.com/ChernykhITMO/order-processing-platform/notifications/internal/kafka_consume"
	"github.com/ChernykhITMO/order-processing-platform/notifications/internal/usecase"
	redis_storage "github.com/ChernykhITMO/order-processing-platform/notifications/storage/redis"
)

const (
	topicStatus    = "status-topic"
	env            = "env"
	sessionTimeout = 30 * time.Second
)

var address = []string{"localhost:9092"}

func main() {
	cfg := config.Config{
		Addr: "localhost:6379",
		DB:   0,
	}

	logger := setupLogger(envLocal)

	storage := redis_storage.New(cfg)
	uc := usecase.New(storage)
	sender := controller.NewSender(uc, logger)

	consumer, err := kafka_consume.NewConsumer(address, sender, topicStatus, sessionTimeout, logger)
	if err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		defer wg.Done()
		consumer.Start(ctx)
	}()

	<-ctx.Done()
	if err := consumer.Stop(); err != nil {
		logger.Error("consumer stopping: ", err)
	}
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
