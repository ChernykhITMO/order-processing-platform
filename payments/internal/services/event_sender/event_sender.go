package event_sender

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/ChernykhITMO/order-processing-platform/payments/internal/ports"
)

type Producer interface {
	Produce(ctx context.Context, message []byte, topic string) error
}

type Sender struct {
	storage  ports.Storage
	producer Producer
	log      *slog.Logger
	topic    string
}

func New(storage ports.Storage, producer Producer, log *slog.Logger, topic string) *Sender {
	return &Sender{
		storage:  storage,
		producer: producer,
		log:      log,
		topic:    topic,
	}
}

func (s *Sender) StartProcessEvents(ctx context.Context, handlePeriod time.Duration) error {
	const op = "services.event_sender.StartProcessEvents"

	log := s.log.With(slog.String("op", op))

	ticker := time.NewTicker(handlePeriod)

	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			log.Info("stopping event processing")
			return nil
		case <-ticker.C:
		}

		event, err := s.storage.GetNewEvent(ctx)
		if err != nil {
			log.Error("fetch event failed", slog.Any("err", err))
			continue
		}
		if event.EventID == 0 {
			continue
		}

		payload, err := json.Marshal(&event)
		if err != nil {
			log.Error("marshal event failed", slog.Any("err", err))
			continue
		}

		if err := s.producer.Produce(ctx, payload, s.topic); err != nil {
			log.Error("produce event failed", slog.Any("err", err))
			continue
		}

		if err := s.storage.MarkSent(ctx, event.EventID); err != nil {
			log.Error("mark sent failed", slog.Any("err", err))
			continue
		}
	}
}
