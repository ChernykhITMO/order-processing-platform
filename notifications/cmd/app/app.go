package app

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/ChernykhITMO/order-processing-platform/notifications/internal/config"
	"github.com/ChernykhITMO/order-processing-platform/notifications/internal/controller"
	"github.com/ChernykhITMO/order-processing-platform/notifications/internal/kafka_consume"
	"github.com/ChernykhITMO/order-processing-platform/notifications/internal/usecase"
	redis_storage "github.com/ChernykhITMO/order-processing-platform/notifications/storage/redis"
)

type App struct {
	log      *slog.Logger
	consumer *kafka_consume.Consumer
	storage  *redis_storage.Storage
}

func New(log *slog.Logger, cfg config.Config) (*App, error) {
	const op = "app.New"

	storage := redis_storage.New(cfg)
	uc := usecase.New(storage)
	sender := controller.NewSender(uc, log)

	consumer, err := kafka_consume.NewConsumer(cfg.KafkaBrokers, sender, cfg.TopicStatus, cfg.ConsumerGroup, cfg.SessionTimeout, log)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &App{
		log:      log,
		consumer: consumer,
		storage:  storage,
	}, nil
}

func (a *App) MustRun(ctx context.Context) {
	if err := a.run(ctx); err != nil {
		panic(err)
	}
}

func (a *App) run(ctx context.Context) error {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		a.consumer.Start(ctx)
	}()

	<-ctx.Done()

	if err := a.consumer.Stop(); err != nil {
		a.log.Error("consumer stopping", slog.Any("err", err))
	}
	if err := a.storage.Close(); err != nil {
		a.log.Error("storage closing", slog.Any("err", err))
	}

	wg.Wait()
	return nil
}
