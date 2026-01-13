package usecase

import (
	"context"
	"log/slog"

	"github.com/ChernykhITMO/order-processing-platform/orders/internal/domain"
)

type Postgres interface {
	CreateOrder(ctx context.Context, userID int64, items []domain.OrderItem) (int64, error)
	GetOrderByID(ctx context.Context, id int64) (*domain.Order, error)
}

type Kafka interface {
	Produce(ctx context.Context, message []byte, topic string) error
}
type Order struct {
	log      slog.Logger
	postgres Postgres
	kafka    Kafka
}

func New(log *slog.Logger, postgres Postgres, kafka Kafka) *Order {
	return &Order{
		log:      *log,
		postgres: postgres,
		kafka:    kafka,
	}
}
