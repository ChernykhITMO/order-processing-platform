package dto

import "time"

type OrderCreated struct {
	OrderID     int64     `json:"order_id"`
	UserID      int64     `json:"user_id"`
	TotalAmount int64     `json:"total_amount"`
	CreatedAt   time.Time `json:"created_at"`
}
