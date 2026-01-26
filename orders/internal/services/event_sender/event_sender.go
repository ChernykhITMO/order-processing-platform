package event_sender

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/ChernykhITMO/order-processing-platform/orders/internal/services"
)

type Kafka interface {
	Produce(ctx context.Context, message []byte, topic string) error
}

type Sender struct {
	storage  services.Postgres
	producer Kafka
	log      *slog.Logger
}

func New(storage services.Postgres, producer Kafka, log *slog.Logger) *Sender {
	return &Sender{
		storage:  storage,
		producer: producer,
		log:      log,
	}
}

func (s *Sender) StartProcessEvents(ctx context.Context, handlePeriod time.Duration, topic string) {
	const op = "services.event_sender.StartProcessEvents"

	log := s.log.With(
		slog.String("op", op),
		slog.String("topic", topic))

	ticker := time.NewTicker(handlePeriod)

	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			log.Info("stopping event processing")
			return
		case <-ticker.C:
		}

		event, eventID, err := s.storage.GetNewEvent(ctx)
		if err != nil {
			log.Error("failed to get new event", slog.Any("err", err))
			continue
		}
		event.EventID = eventID
		if event.EventID == 0 {
			continue
		}

		message, err := json.Marshal(&event)
		if err != nil {
			log.Error("marshal event failed", slog.Any("err", err))
			continue
		}

		if err := s.producer.Produce(ctx, message, topic); err != nil {
			log.Error("kafka_produce produce failed", slog.Any("err", err))
			continue
		}

		if err := s.storage.MarkSent(ctx, eventID); err != nil {
			log.Error("mark event sent failed", slog.Any("err", err))
			continue
		}
	}
}
