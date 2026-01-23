package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/ChernykhITMO/order-processing-platform/payments/internal/dto"
	"github.com/ChernykhITMO/order-processing-platform/payments/internal/ports"
	"github.com/ChernykhITMO/order-processing-platform/payments/internal/services"
)

type Producer interface {
	Produce(ctx context.Context, message []byte, topic string) error
}

type Controller struct {
	service  services.Service
	storage  ports.Storage
	producer Producer
	topic    string
	log      *slog.Logger
}

func NewController(service services.Service, storage ports.Storage, producer Producer, topic string, log *slog.Logger) *Controller {
	return &Controller{
		service:  service,
		storage:  storage,
		producer: producer,
		topic:    topic,
		log:      log,
	}
}

func (h *Controller) HandleMessage(message []byte) error {
	const op = "controller.HandleMessage"
	log := h.log.With(slog.String("op", op))

	log.Info("controller started", slog.String("op", op))

	var input dto.OrderCreated
	if err := json.Unmarshal(message, &input); err != nil {
		log.Error("decode message", slog.String("error", err.Error()))
		return fmt.Errorf("%s: decode message: %w", op, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.service.HandleOrderCreated(ctx, input); err != nil {
		log.Error("handle message", slog.String("error", err.Error()))
		return fmt.Errorf("%s: handle message: %w", op, err)
	}

	return nil
}
