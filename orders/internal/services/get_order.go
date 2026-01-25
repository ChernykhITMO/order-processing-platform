package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/ChernykhITMO/order-processing-platform/orders/internal/domain"
	"github.com/ChernykhITMO/order-processing-platform/orders/internal/dto"
)

func (o *Order) GetOrder(ctx context.Context, input dto.GetOrderInput) (dto.GetOrderOutput, error) {
	const op = "services.Order.GetOrder"

	log := o.log.With(
		slog.String("op", op),
		slog.Int64("order_id", input.ID))

	var output dto.GetOrderOutput

	if input.ID <= 0 {
		return output, fmt.Errorf("%s: %w", op, domain.ErrInvalidOrderID)
	}

	order, err := o.postgres.GetOrderByID(ctx, input.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return output, fmt.Errorf("%s: %w", op, domain.ErrOrderNotFound)
		}
		log.Error("get order failed", slog.Any("err", err))
		return output, fmt.Errorf("%s: %w", op, err)
	}

	output.Order = *order
	return output, nil
}
