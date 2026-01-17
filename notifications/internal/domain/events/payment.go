package events

type ID int64

type Status string

type Payment struct {
	OrderID     ID     `json:"order_id"`
	UserID      ID     `json:"user_id"`
	OrderStatus Status `json:"order_status"`
}
