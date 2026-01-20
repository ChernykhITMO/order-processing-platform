package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/ChernykhITMO/order-processing-platform/payments/internal/domain/events"
	"github.com/ChernykhITMO/order-processing-platform/payments/internal/dto"
)

const (
	statusPending   = "pending"
	statusSucceeded = "succeeded"
	statusFailed    = "failed"
)

type Producer interface {
	Produce(ctx context.Context, message []byte, topic string) error
}

type Handler struct {
	storage  Storage
	producer Producer
	topic    string
	log      *slog.Logger
}

func NewHandler(storage Storage, producer Producer, topic string, log *slog.Logger) *Handler {
	return &Handler{storage: storage, producer: producer, topic: topic, log: log}
}

func (h *Handler) HandleMessage(message []byte) error {
	const op = "handler.HandleMessage"
	log := h.log.With(slog.String("op", op))

	log.Info("handler started", slog.String("op", op))

	var input dto.OrderCreated
	if err := json.Unmarshal(message, &input); err != nil {
		log.Error("decode message", slog.String("error", err.Error()))
		return fmt.Errorf("%s: decode message: %w", op, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.storage.UpsertPayment(ctx, input.OrderID, input.UserID, input.TotalAmount, statusPending); err != nil {
		return fmt.Errorf("%s: persist payment: %w", op, err)
	}

	if input.OrderID%2 == 0 {
		if err := h.storage.UpdatePaymentStatus(ctx, input.OrderID, statusSucceeded); err != nil {
			return fmt.Errorf("%s: update payment status: %w", op, err)
		}
		event := events.PaymentStatus{
			OrderID:     input.OrderID,
			UserID:      input.UserID,
			OrderStatus: statusSucceeded,
		}
		payload, err := json.Marshal(&event)
		if err != nil {
			return fmt.Errorf("%s: encode success event: %w", op, err)
		}
		if err := h.producer.Produce(ctx, payload, h.topic); err != nil {
			return fmt.Errorf("%s: produce success event: %w", op, err)
		}
		return nil
	}

	if err := h.storage.UpdatePaymentStatus(ctx, input.OrderID, statusFailed); err != nil {
		return fmt.Errorf("%s: update payment status: %w", op, err)
	}
	event := events.PaymentStatus{
		OrderID:     input.OrderID,
		UserID:      input.UserID,
		OrderStatus: statusFailed,
	}
	payload, err := json.Marshal(&event)
	if err != nil {
		return fmt.Errorf("%s: encode failed event: %w", op, err)
	}
	if err := h.producer.Produce(ctx, payload, h.topic); err != nil {
		return fmt.Errorf("%s: produce failed event: %w", op, err)
	}

	_, err := h.storage.TryMarkProcessed(ctx, input.EventID)
	if err != nil {
		log.Error("marking event", slog.String("error", err.Error()))
		return fmt.Errorf("%s: mark event: %w", op, err)
	}
	return nil
}
