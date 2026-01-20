package services

import (
	"context"
	"log/slog"

	"github.com/ChernykhITMO/order-processing-platform/orders/internal/domain"
	"github.com/ChernykhITMO/order-processing-platform/orders/internal/domain/events"
)

type Postgres interface {
	CreateOrder(
		ctx context.Context,
		userID int64,
		items []domain.OrderItem) (orderID int64, err error)

	GetOrderByID(ctx context.Context, id int64) (*domain.Order, error)
	GetNewEvent(ctx context.Context) (events.OrderCreated, int64, error)
	MarkSent(ctx context.Context, eventID int64) error
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
