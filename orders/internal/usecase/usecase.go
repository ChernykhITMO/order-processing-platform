package usecase

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/ChernykhITMO/order-processing-platform/orders/internal/domain"
)

type Postgres interface {
	CreateOrder(ctx context.Context, userID int64, items []domain.OrderItem) (int64, error)
	GetOrderByID(ctx context.Context, id int64) (*domain.Order, error)
}

type Kafka interface {
	// Produce(ctx context.Context, msgs ...kafka_produce.Message) error
}
type Order struct {
	log      slog.Logger
	postgres Postgres
}

func New(log *slog.Logger, postgres Postgres) *Order {
	return &Order{
		log:      *log,
		postgres: postgres,
	}
}

func (o *Order) CreateOrder(ctx context.Context, userID int64, items []domain.OrderItem) (int64, error) {
	const op = "usecase.Order.CreateOrder"
	if userID <= 0 {
		return 0, fmt.Errorf("%s: %w", op, domain.ErrInvalidUserID)
	}

	if len(items) == 0 {
		return 0, fmt.Errorf("%s: %w", op, domain.ErrInvalidItems)
	}

	orderID, err := o.postgres.CreateOrder(ctx, userID, items)
	if err != nil {
		o.log.Error("create order failed", "op", op, "user_id", userID, "err", err)
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return orderID, nil
}

func (o *Order) GetOrder(ctx context.Context, id int64) (*domain.Order, error) {
	const op = "usecase.Order.GetOrder"

	if id <= 0 {
		return nil, fmt.Errorf("%s: %w", op, domain.ErrInvalidOrderID)
	}

	order, err := o.postgres.GetOrderByID(ctx, id)
	if err != nil {
		o.log.Error("get order failed", "op", op, "order_id", id, "err", err)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, domain.ErrOrderNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return order, nil
}
