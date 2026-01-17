package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/ChernykhITMO/order-processing-platform/notifications/internal/dto"
	"github.com/ChernykhITMO/order-processing-platform/notifications/internal/mapper"
	"github.com/ChernykhITMO/order-processing-platform/notifications/internal/usecase"
)

type Handler struct {
	uc  *usecase.Notification
	log *slog.Logger
}

func NewHandler(uc *usecase.Notification, log *slog.Logger) *Handler {
	return &Handler{
		uc:  uc,
		log: log,
	}
}

func (h *Handler) HandleMessage(message []byte) error {
	const op = "handler.HandleMessage"
	h.log.Info("handler started", slog.String("op", op))

	var payment dto.Payment

	if err := json.Unmarshal(message, &payment); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	input := mapper.MapToInput(payment)

	if err := h.uc.SaveNotification(ctx, input); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
