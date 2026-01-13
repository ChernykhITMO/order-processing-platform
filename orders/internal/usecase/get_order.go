package usecase

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ChernykhITMO/order-processing-platform/orders/internal/domain"
	"github.com/ChernykhITMO/order-processing-platform/orders/internal/dto"
)

func (o *Order) GetOrder(ctx context.Context, input dto.GetOrderInput) (dto.GetOrderOutput, error) {
	const op = "usecase.Order.GetOrder"

	var output dto.GetOrderOutput

	if input.ID <= 0 {
		return output, fmt.Errorf("%s: %w", op, domain.ErrInvalidOrderID)
	}

	order, err := o.postgres.GetOrderByID(ctx, input.ID)
	if err != nil {
		o.log.Error("get order failed", "op", op, "order_id", input.ID, "err", err)

		if errors.Is(err, sql.ErrNoRows) {
			return output, fmt.Errorf("%s: %w", op, domain.ErrOrderNotFound)
		}

		return output, fmt.Errorf("%s: %w", op, err)
	}

	output.Order = *order
	return output, nil
}
