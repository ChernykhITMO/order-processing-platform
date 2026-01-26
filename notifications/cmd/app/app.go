package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ChernykhITMO/order-processing-platform/notifications/internal/config"
	"github.com/ChernykhITMO/order-processing-platform/notifications/internal/controller"
	"github.com/ChernykhITMO/order-processing-platform/notifications/internal/kafka_consume"
	"github.com/ChernykhITMO/order-processing-platform/notifications/internal/services"
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
	uc := services.New(storage, log)
	sender := controller.NewSender(uc, log)

	consumer, err := kafka_consume.NewConsumer(
		cfg.KafkaBrokers, sender,
		cfg.TopicStatus, cfg.ConsumerGroup,
		cfg.SessionTimeout, cfg.ReadTimeout, log,
	)

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
	const op = "app.run"

	log := a.log.With(slog.String("op", op))

	log.Debug("running application")
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- a.consumer.Start(ctx)
	}()

	select {
	case <-ctx.Done():
	case err := <-errCh:
		if err != nil {
			cancel()
			return err
		}
	}

	if err := a.consumer.Stop(); err != nil {
		log.Error("consumer stopped", slog.Any("err", err))
	}
	if err := a.storage.Close(); err != nil {
		log.Error("storage closed", slog.Any("err", err))
	}

	log.Debug("stopping application")
	return nil
}
