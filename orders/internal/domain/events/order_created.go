package events

import (
	"time"

	"github.com/ChernykhITMO/order-processing-platform/orders/internal/domain"
)

type OrderCreated struct {
	EventID     int64        `json:"event_id"`
	OrderID     domain.ID    `json:"order_id"`
	UserID      domain.ID    `json:"user_id"`
	TotalAmount domain.Money `json:"total_amount"`
	CreatedAt   time.Time    `json:"created_at"`
}
