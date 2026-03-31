package postgres

import (
	"context"

	"github.com/ChernykhITMO/order-processing-platform/orders/internal/domain"
	"github.com/ChernykhITMO/order-processing-platform/orders/internal/domain/events"
)

type Repository interface {
	CreateOrder(ctx context.Context, userID int64, items []domain.OrderItem) (orderID int64, err error)
	GetOrderByID(ctx context.Context, id int64) (*domain.Order, error)
	GetNewEvent(ctx context.Context) (events.OrderCreated, int64, error)
	MarkSent(ctx context.Context, eventID int64) error
	Ping(ctx context.Context) error
	Close() error
}
