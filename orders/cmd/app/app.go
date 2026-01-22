package app

import (
	"context"
	"log/slog"
	"time"

	grpcapp "github.com/ChernykhITMO/order-processing-platform/orders/cmd/app/grpc"
	"github.com/ChernykhITMO/order-processing-platform/orders/internal/kafka_produce"
	"github.com/ChernykhITMO/order-processing-platform/orders/internal/services"
	"github.com/ChernykhITMO/order-processing-platform/orders/internal/services/event_sender"
	"github.com/ChernykhITMO/order-processing-platform/orders/internal/storage/postgres"
)

type App struct {
	GRPCSrv       *grpcapp.App
	EventSender   *event_sender.Sender
	KafkaProducer *kafka_produce.Producer
	KafkaTopic    string
	KafkaPeriod   time.Duration
}

func New(
	log *slog.Logger,
	grpcPort int,
	dsn string,
	kafkaBrokers []string,
	kafkaTopic string,
	kafkaPeriod time.Duration,
) *App {

	storage, err := postgres.New(dsn)
	if err != nil {
		panic(err)
	}

	order := services.New(log, storage)

	grpcApp := grpcapp.New(log, order, grpcPort)

	var producer *kafka_produce.Producer
	var sender *event_sender.Sender
	if len(kafkaBrokers) > 0 && kafkaTopic != "" {
		producer, err = kafka_produce.NewProducer(kafkaBrokers)
		if err != nil {
			panic(err)
		}
		sender = event_sender.New(storage, producer, log)
	}

	return &App{
		GRPCSrv:       grpcApp,
		EventSender:   sender,
		KafkaProducer: producer,
		KafkaTopic:    kafkaTopic,
		KafkaPeriod:   kafkaPeriod,
	}
}

func (a *App) StartEventSender(ctx context.Context) {
	if a.EventSender == nil {
		return
	}
	period := a.KafkaPeriod
	if period <= 0 {
		period = time.Second
	}
	a.EventSender.StartProcessEvents(ctx, period, a.KafkaTopic)
}

func (a *App) Stop() {
	a.GRPCSrv.Stop()
	if a.KafkaProducer != nil {
		a.KafkaProducer.Close()
	}
}
