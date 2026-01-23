package app

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/ChernykhITMO/order-processing-platform/payments/internal/config"
	"github.com/ChernykhITMO/order-processing-platform/payments/internal/controller"
	"github.com/ChernykhITMO/order-processing-platform/payments/internal/kafka"
	"github.com/ChernykhITMO/order-processing-platform/payments/internal/kafka_consume"
	"github.com/ChernykhITMO/order-processing-platform/payments/internal/services"
	"github.com/ChernykhITMO/order-processing-platform/payments/internal/storage/postgres"
)

type App struct {
	log      *slog.Logger
	consumer *kafka_consume.Consumer
	producer *kafka.Producer
}

func New(log *slog.Logger, cfg config.Config) (*App, error) {
	const op = "app.New"

	storage, err := postgres.New(cfg.DBDSN)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	producer, err := kafka.NewProducer(cfg.KafkaBrokers)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	service := services.New(storage, log, cfg.TopicStatus, cfg.EventType)
	ctrl := controller.NewController(*service, storage, producer, cfg.TopicStatus, log)
	consumer, err := kafka_consume.NewConsumer(ctrl, cfg.KafkaBrokers, cfg.TopicOrder, cfg.ConsumerGroup, log)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &App{
		log:      log,
		consumer: consumer,
		producer: producer,
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

	a.producer.Close()
	wg.Wait()
	return nil
}
