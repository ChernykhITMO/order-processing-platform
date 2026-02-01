package app

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/ChernykhITMO/order-processing-platform/payments/internal/config"
	"github.com/ChernykhITMO/order-processing-platform/payments/internal/controller"
	"github.com/ChernykhITMO/order-processing-platform/payments/internal/kafka_consume"
	"github.com/ChernykhITMO/order-processing-platform/payments/internal/kafka_produce"
	"github.com/ChernykhITMO/order-processing-platform/payments/internal/ports"
	"github.com/ChernykhITMO/order-processing-platform/payments/internal/services"
	"github.com/ChernykhITMO/order-processing-platform/payments/internal/services/event_sender"
	"github.com/ChernykhITMO/order-processing-platform/payments/internal/storage/postgres"
)

type App struct {
	log          *slog.Logger
	consumer     *kafka_consume.Consumer
	producer     *kafka.Producer
	sender       *event_sender.Sender
	senderPeriod time.Duration
	storage      ports.Storage
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

	service := services.New(storage, log, cfg.EventType)
	ctrl := controller.NewController(*service, storage, producer, cfg.TopicStatus, log)
	consumer, err := kafka_consume.NewConsumer(ctrl, cfg.KafkaBrokers, cfg.TopicOrder, cfg.ConsumerGroup, log)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	sender := event_sender.New(storage, producer, log, cfg.TopicStatus)

	return &App{
		log:          log,
		consumer:     consumer,
		producer:     producer,
		sender:       sender,
		senderPeriod: cfg.SenderPeriod,
		storage:      storage,
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

	log.Info("starting application")
	var wg sync.WaitGroup
	wg.Add(2)

	go func() { defer wg.Done(); a.consumer.Start(ctx) }()

	period := a.senderPeriod
	if period <= 0 {
		period = time.Second
	}

	go func() {
		defer wg.Done()
		err := a.sender.StartProcessEvents(ctx, period)
		if err != nil {
			log.Error("start process events failed", slog.Any("err", err))
			return
		}
	}()

	<-ctx.Done()

	if err := a.consumer.Stop(); err != nil {
		log.Error("consumer stopping", slog.Any("err", err))
	}

	a.producer.Close()
	if a.storage != nil {
		_ = a.storage.Close()
	}
	wg.Wait()
	return nil
}

func (a *App) Stop() {
	const op = "app.Stop"

	a.log.With(slog.String("op", op)).
		Info("stopping payments server")
}
