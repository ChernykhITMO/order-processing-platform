package dto

import "time"

type PaymentSucceeded struct {
	OrderID     int64     `json:"order_id"`
	UserID      int64     `json:"user_id"`
	TotalAmount int64     `json:"total_amount"`
	ProcessedAt time.Time `json:"processed_at"`
}

type PaymentFailed struct {
	OrderID     int64     `json:"order_id"`
	UserID      int64     `json:"user_id"`
	TotalAmount int64     `json:"total_amount"`
	Reason      string    `json:"reason"`
	ProcessedAt time.Time `json:"processed_at"`
}
