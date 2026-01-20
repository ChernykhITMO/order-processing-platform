package event_sender

import (
	"context"
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

	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				log.Info("stopping event processing")
				return
			case <-ticker.C:
				//
			}

			s.producer.Produce(ctx)
		}
	}()
}
