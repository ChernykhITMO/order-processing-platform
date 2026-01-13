package usecase

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ChernykhITMO/order-processing-platform/orders/internal/domain"
	"github.com/ChernykhITMO/order-processing-platform/orders/internal/domain/events"
	"github.com/ChernykhITMO/order-processing-platform/orders/internal/dto"
	"github.com/ChernykhITMO/order-processing-platform/orders/internal/mapper"
)

func (o *Order) CreateOrder(ctx context.Context, input dto.CreateOrderInput) (dto.CreateOrderOutput, error) {
	const op = "usecase.Order.CreateOrder"
	var output dto.CreateOrderOutput

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
		o.log.Error("create order failed", "op", op, "user_id", input.UserID, "err", err)
		return output, fmt.Errorf("%s: %w", op, err)
	}

	output.ID = orderID

	var inputGetOrder dto.GetOrderInput
	inputGetOrder.ID = orderID

	outputGetOrder, err := o.GetOrder(ctx, inputGetOrder)
	if err != nil {
		return output, fmt.Errorf("%s: %w", op, err)
	}

	orderCreated := events.OrderCreated{
		OrderID:     outputGetOrder.ID,
		UserID:      outputGetOrder.UserID,
		TotalAmount: outputGetOrder.TotalAmount,
		CreatedAt:   outputGetOrder.CreatedAt,
	}

	jsonMsg, err := json.Marshal(&orderCreated)
	if err != nil {
		return output, fmt.Errorf("%s: %w", op, err)
	}

	if err := o.kafka.Produce(ctx, jsonMsg, domain.KafkaTopic); err != nil {
		return output, fmt.Errorf("%s: %w", op, err)
	}
	return output, nil
}
