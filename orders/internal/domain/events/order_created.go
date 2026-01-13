package events

import (
	"time"

	"github.com/ChernykhITMO/order-processing-platform/orders/internal/domain"
)

type OrderCreated struct {
	OrderID     domain.ID
	UserID      domain.ID
	TotalAmount domain.Money
	CreatedAt   time.Time
}
