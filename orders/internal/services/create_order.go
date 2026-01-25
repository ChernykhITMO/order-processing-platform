package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ChernykhITMO/order-processing-platform/orders/internal/domain"
	"github.com/ChernykhITMO/order-processing-platform/orders/internal/dto"
	"github.com/ChernykhITMO/order-processing-platform/orders/internal/mapper"
)

func (o *Order) CreateOrder(ctx context.Context, input dto.CreateOrderInput) (dto.CreateOrderOutput, error) {
	const op = "services.Order.CreateOrder"
	var output dto.CreateOrderOutput

	log := o.log.With(
		slog.String("op", op),
		slog.Int64("user_id", input.UserID))

	if input.UserID <= 0 {
		return output, fmt.Errorf("%s: %w", op, domain.ErrInvalidUserID)
	}

	if len(input.Items) == 0 {
		return output, fmt.Errorf("%s: %w", op, domain.ErrInvalidItems)
	}

	items, err := mapper.MapInputItems(input.Items)
	if err != nil {
		return output, fmt.Errorf("%s: %w", op, err)
	}

	orderID, err := o.postgres.CreateOrder(ctx, input.UserID, items)
	if err != nil {
		log.Error("create order failed", slog.Any("err", err))
		return output, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("order created", slog.Int64("order_id", orderID))

	output.ID = orderID
	return output, nil
}
