package api

type OrderItem struct {
	ProductID int64 `json:"product_id"`
	Quantity  int32 `json:"quantity"`
	Price     Money `json:"price"`
}

type CreateOrderRequest struct {
	UserID int64       `json:"user_id"`
	Items  []OrderItem `json:"items"`
}

type Order struct {
	OrderID     int64       `json:"order_id"`
	UserID      int64       `json:"user_id"`
	Status      string      `json:"status"`
	Items       []OrderItem `json:"items"`
	TotalAmount Money       `json:"total_amount"`
	CreatedAt   string      `json:"created_at"`
	UpdatedAt   string      `json:"updated_at"`
}

type GetOrderResponse struct {
	Order Order `json:"order"`
}

type Money struct {
	Units int64 `json:"units"`
	Nanos int32 `json:"nanos"`
}
