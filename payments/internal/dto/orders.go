package dto

import "time"

type OrderCreated struct {
	EventID     int64     `json:"event_id"`
	OrderID     int64     `json:"order_id"`
	UserID      int64     `json:"user_id"`
	TotalAmount int64     `json:"total_amount"`
	CreatedAt   time.Time `json:"created_at"`
}
