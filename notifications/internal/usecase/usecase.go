package usecase

import (
	"context"
	"fmt"

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
}

func New(storage Redis) *Notification {
	return &Notification{
		storage: storage,
	}
}

func (n *Notification) SaveNotification(ctx context.Context, input dto.SaveInput) error {
	const op = "usecase.Save"

	if err := validateInput(input); err != nil {
		return fmt.Errorf("%s: validated %w", op, err)
	}

	payment := mapper.MapToDomainSave(input)

	if err := n.storage.SaveNotification(ctx, input.Key, payment); err != nil {
		return fmt.Errorf("%s: redis saved %w", op, err)
	}

	return nil
}

func (n *Notification) GetNotification(ctx context.Context, input dto.GetInput) (dto.GetOutput, error) {
	const op = "usecase.Get"
	var output dto.GetOutput

	if input.Key == "" {
		return output, fmt.Errorf("%s: validated %w", op, domain.ErrIsEmptyKey)
	}

	payment, err := n.storage.GetNotification(ctx, input.Key)
	if err != nil {
		return output, fmt.Errorf("%s: redis got %w", op, err)
	}

	output = mapper.MapToDTOGet(payment)

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
