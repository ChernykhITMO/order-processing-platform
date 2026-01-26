package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/ChernykhITMO/order-processing-platform/notifications/internal/dto"
	"github.com/ChernykhITMO/order-processing-platform/notifications/internal/mapper"
	"github.com/ChernykhITMO/order-processing-platform/notifications/internal/services"
)

type Sender struct {
	uc  *services.Notification
	log *slog.Logger
}

func NewSender(uc *services.Notification, log *slog.Logger) *Sender {
	return &Sender{
		uc:  uc,
		log: log,
	}
}

func (h *Sender) HandleMessage(message []byte) error {
	const op = "controller.HandleMessage"
	log := h.log.With(slog.String("op", op))

	log.Debug("starting handle message")

	var payment dto.Payment

	if err := json.Unmarshal(message, &payment); err != nil {
		log.Error("unmarshal failed", slog.Any("err", err))
		return fmt.Errorf("%s: %w", op, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	input := mapper.MapToInput(payment)

	if err := h.uc.SaveNotification(ctx, input); err != nil {
		log.Error("save notification failed", slog.Any("err", err))
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Debug("handle message successful")
	return nil
}
