package usecase

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ChernykhITMO/order-processing-platform/orders/internal/domain"
)

type Postgres interface {
	CreateOrder(ctx context.Context, userID int64, items []domain.OrderItem) (int64, error)
	GetOrderByID(ctx context.Context, id int64) (*domain.Order, error)
}

type Kafka interface {
	Produce(ctx context.Context, msgs ...kafka_produce.Message) error
}
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

func (o *Order) CreateOrder(ctx context.Context, userID int64, items []domain.OrderItem) (int64, error) {
	const op = "usecase.Order.CreateOrder"
	if userID <= 0 {
		return 0, fmt.Errorf("%s: %w", op, domain.ErrInvalidID)
	}

	if len(items) == 0 {
		return 0, fmt.Errorf("%s: %w", op, domain.ErrInvalidItems)
	}

	orderID, err := o.provider.CreateOrder(ctx, userID, items)
	if err != nil {
		o.log.Error("create order failed", "op", op, "user_id", userID, "err", err)
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return orderID, nil
}

func (o *Order) GetOrder(ctx context.Context, id int64) (*domain.Order, error) {
	const op = "usecase.Order.GetOrder"

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
