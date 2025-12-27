package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ChernykhITMO/order-processing-platform/orders/internal/domain"
)

type Order struct {
	log      *slog.Logger
	provider OrderProvider
}

func New(log *slog.Logger, provider OrderProvider) *Order {
	return &Order{
		log:      log,
		provider: provider,
	}
}

type OrderProvider interface {
	CreateOrder(ctx context.Context, userID int64, items []domain.OrderItem) error
	GetOrderByID(ctx context.Context, id int64) (*domain.Order, error)
}

func (o *Order) CreateOrder(ctx context.Context, userID int64, items []domain.OrderItem) error {
	const op = "service.Order.CreateOrder"
	if userID <= 0 {
		return fmt.Errorf("%s: %w", op, domain.ErrInvalidID)
	}

	if len(items) == 0 {
		return fmt.Errorf("%s: %w", op, domain.ErrInvalidItems)
	}

	if err := o.provider.CreateOrder(ctx, userID, items); err != nil {
		o.log.Error("create order failed", "op", op, "user_id", userID, "err", err)
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (o *Order) GetOrder(ctx context.Context, id int64) (*domain.Order, error) {
	const op = "service.Order.GetOrder"

	if id <= 0 {
		return nil, fmt.Errorf("%s: %w", op, domain.ErrInvalidID)
	}

	order, err := o.provider.GetOrderByID(ctx, id)
	if err != nil {
		o.log.Error("get order failed", "op", op, "order_id", id, "err", err)
		return nil, fmt.Errorf("%s: %w", op, domain.ErrInvalidItems)
	}

	return order, nil
}
