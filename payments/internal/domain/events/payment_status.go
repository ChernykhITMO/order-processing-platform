package events

type PaymentStatus struct {
	EventID     int64  `json:"event_id"`
	OrderID     int64  `json:"order_id"`
	UserID      int64  `json:"user_id"`
	OrderStatus string `json:"order_status"`
}
