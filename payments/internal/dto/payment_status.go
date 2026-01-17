package dto

type PaymentStatus struct {
	OrderID     int64  `json:"order_id"`
	UserID      int64  `json:"user_id"`
	OrderStatus string `json:"order_status"`
}
