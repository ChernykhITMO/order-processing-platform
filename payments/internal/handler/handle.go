package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/ChernykhITMO/order-processing-platform/payments/internal/dto"
)

const (
	statusPending   = "pending"
	statusSucceeded = "succeeded"
	statusFailed    = "failed"
)

type Storage interface {
	UpsertPayment(ctx context.Context, orderID, userID, totalAmount int64, status string) error
	UpdatePaymentStatus(ctx context.Context, orderID int64, status string) error
}

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
	h.log.Info("handler started", slog.String("op", op))
	var input dto.OrderCreated
	if err := json.Unmarshal(message, &input); err != nil {
		return fmt.Errorf("%s: decode message: %w", op, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.storage.UpsertPayment(ctx, input.OrderID, input.UserID, input.TotalAmount, statusPending); err != nil {
		return fmt.Errorf("%s: persist payment: %w", op, err)
	}

	now := time.Now().UTC()
	if input.OrderID%2 == 0 {
		if err := h.storage.UpdatePaymentStatus(ctx, input.OrderID, statusSucceeded); err != nil {
			return fmt.Errorf("%s: update payment status: %w", op, err)
		}
		event := dto.PaymentSucceeded{
			OrderID:     input.OrderID,
			UserID:      input.UserID,
			TotalAmount: input.TotalAmount,
			ProcessedAt: now,
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
	event := dto.PaymentFailed{
		OrderID:     input.OrderID,
		UserID:      input.UserID,
		TotalAmount: input.TotalAmount,
		Reason:      "stub payment failure",
		ProcessedAt: now,
	}
	payload, err := json.Marshal(&event)
	if err != nil {
		return fmt.Errorf("%s: encode failed event: %w", op, err)
	}
	if err := h.producer.Produce(ctx, payload, h.topic); err != nil {
		return fmt.Errorf("%s: produce failed event: %w", op, err)
	}

	log.Printf("%s: order_id=%d status=%s", op, input.OrderID, statusFailed)
	return nil
}
