package ports

import (
	"context"

	"github.com/ChernykhITMO/order-processing-platform/payments/internal/domain/events"
)

type StorageTx interface {
	UpsertPayment(ctx context.Context, orderID, userID, totalAmount int64, status string) error
	UpdatePaymentStatus(ctx context.Context, orderID int64, status string) error
	TryMarkProcessed(ctx context.Context, eventId int64) (bool, error)
	SaveEvent(ctx context.Context, eventType string, payload []byte, aggregateID int64) error
}

type Storage interface {
	RunInTx(ctx context.Context, fn func(tx StorageTx) error) error
	GetNewEvent(ctx context.Context) (events.PaymentStatus, error)
	MarkSent(ctx context.Context, id int64) error
}
