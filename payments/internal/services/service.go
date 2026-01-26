package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/ChernykhITMO/order-processing-platform/payments/internal/domain"
	"github.com/ChernykhITMO/order-processing-platform/payments/internal/domain/events"
	"github.com/ChernykhITMO/order-processing-platform/payments/internal/dto"
	"github.com/ChernykhITMO/order-processing-platform/payments/internal/ports"
)

type Service struct {
	storage   ports.Storage
	log       *slog.Logger
	eventType string
}

func New(storage ports.Storage, log *slog.Logger, eventType string) *Service {
	return &Service{
		storage:   storage,
		log:       log,
		eventType: eventType,
	}
}

func (s *Service) HandleOrderCreated(ctx context.Context, input dto.OrderCreated) error {
	const op = "services.HandleOrderCreated"

	log := s.log.With(
		slog.String("op", op),
		slog.Int64("order_id", input.OrderID),
		slog.Int64("user_id", input.UserID),
		slog.Int64("event_id", input.EventID),
	)

	return s.storage.RunInTx(ctx, func(tx ports.StorageTx) error {
		if input.EventID == 0 {
			return fmt.Errorf("%s: %w", op, domain.ErrInvalidEventID)
		}

		ok, err := tx.TryMarkProcessed(ctx, input.EventID)
		if err != nil {
			log.Error("try mark processed failed", slog.Any("err", err))
			return fmt.Errorf("%s: %w", op, err)
		}

		if !ok {
			return nil
		}

		if err := tx.UpsertPayment(ctx, input.OrderID, input.UserID, input.TotalAmount, domain.StatusPaymentPending); err != nil {
			log.Error("upsert payment failed", slog.Any("err", err))
			return fmt.Errorf("%s: persist payment: %w", op, err)
		}

		if input.OrderID%2 == 0 {
			if err := tx.UpdatePaymentStatus(ctx, input.OrderID, domain.StatusSucceeded); err != nil {
				log.Error("update payment status failed", slog.Any("err", err))
				return fmt.Errorf("%s: update payment status: %w", op, err)
			}
			event := events.PaymentStatus{
				OrderID:     input.OrderID,
				UserID:      input.UserID,
				OrderStatus: domain.StatusSucceeded,
			}
			payload, err := json.Marshal(&event)
			if err != nil {
				log.Error("marshal event failed", slog.Any("err", err))
				return fmt.Errorf("%s: encode success event: %w", op, err)
			}

			if err := tx.SaveEvent(ctx, s.eventType, payload, input.OrderID); err != nil {
				log.Error("save event failed", slog.Any("err", err))
				return fmt.Errorf("%s: kafka_produce produce: %w", op, err)
			}
		} else {
			if err := tx.UpdatePaymentStatus(ctx, input.OrderID, domain.StatusFailed); err != nil {
				log.Error("update payment status failed", slog.Any("err", err))
				return fmt.Errorf("%s: update payment status: %w", op, err)
			}
			event := events.PaymentStatus{
				OrderID:     input.OrderID,
				UserID:      input.UserID,
				OrderStatus: domain.StatusFailed,
			}
			payload, err := json.Marshal(&event)
			if err != nil {
				log.Error("marshal event failed", slog.Any("err", err))
				return fmt.Errorf("%s: encode failed event: %w", op, err)
			}

			if err := tx.SaveEvent(ctx, s.eventType, payload, input.OrderID); err != nil {
				log.Error("save event failed", slog.Any("err", err))
				return fmt.Errorf("%s: save event: %w", op, err)
			}
		}
		return nil
	})
}
