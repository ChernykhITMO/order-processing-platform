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
	service services.Service
	log     *slog.Logger
}

func NewController(service services.Service, storage ports.Storage, producer Producer, topic string, log *slog.Logger) *Controller {
	return &Controller{
		service: service,
		log:     log,
	}
}

func (h *Controller) HandleMessage(parentCtx context.Context, message []byte) error {
	const op = "controller.HandleMessage"
	log := h.log.With(slog.String("op", op))

	var input dto.OrderCreated
	if err := json.Unmarshal(message, &input); err != nil {
		log.Error("decode message", slog.Any("err", err))
		return fmt.Errorf("%s: decode message: %w", op, err)
	}

	ctx, cancel := context.WithTimeout(parentCtx, 5*time.Second)
	defer cancel()

	if err := h.service.HandleOrderCreated(ctx, input); err != nil {
		log.Error("handle failed", slog.Any("err", err))
		return fmt.Errorf("%s: handle message: %w", op, err)
	}

	return nil
}
