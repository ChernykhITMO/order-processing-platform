package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ChernykhITMO/order-processing-platform/notifications/internal/domain"
	"github.com/ChernykhITMO/order-processing-platform/notifications/internal/domain/events"
	"github.com/ChernykhITMO/order-processing-platform/notifications/internal/dto"
	"github.com/ChernykhITMO/order-processing-platform/notifications/internal/mapper"
)

type Redis interface {
	SaveNotification(ctx context.Context, key string, value events.Payment) error
	GetNotification(ctx context.Context, key string) (events.Payment, error)
}

type Notification struct {
	storage Redis
	log     *slog.Logger
}

func New(storage Redis, log *slog.Logger) *Notification {
	return &Notification{
		storage: storage,
		log:     log,
	}
}

func (n *Notification) SaveNotification(ctx context.Context, input dto.SaveInput) error {
	const op = "services.Save"

	log := n.log.With(slog.String("op", op))
	if err := validateInput(input); err != nil {
		return fmt.Errorf("%s: validated %w", op, err)
	}

	log = log.With(
		slog.Int64("order_id", input.OrderID),
		slog.Int64("user_id", input.UserID),
		slog.String("status", input.Status),
	)

	payment := mapper.MapToDomainSave(input)

	if err := n.storage.SaveNotification(ctx, input.Key, payment); err != nil {
		log.Error("save notification failed", slog.Any("err", err))
		return fmt.Errorf("%s: redis saved %w", op, err)
	}

	log.Debug("save notification is successful")

	return nil
}

func (n *Notification) GetNotification(ctx context.Context, input dto.GetInput) (dto.GetOutput, error) {
	const op = "services.Get"
	var output dto.GetOutput

	log := n.log.With(slog.String("op", op))

	if input.Key == "" {
		return output, fmt.Errorf("%s: validated %w", op, domain.ErrIsEmptyKey)
	}

	payment, err := n.storage.GetNotification(ctx, input.Key)
	if err != nil {
		log.Error("get notification failed", slog.Any("err", err))
		return output, fmt.Errorf("%s: redis got %w", op, err)
	}

	output = mapper.MapToDTOGet(payment)

	log.Debug("get notification is successful")
	return output, nil
}

func validateInput(input dto.SaveInput) error {
	if input.Key == "" {
		return domain.ErrIsEmptyKey
	}

	if input.UserID <= 0 {
		return domain.ErrInvalidUserID
	}

	if input.OrderID <= 0 {
		return domain.ErrInvalidOrderID
	}

	if input.Status != domain.StatusFailed && input.Status != domain.StatusSucceeded {
		return domain.ErrInvalidStatus
	}

	return nil
}
