package dto

type CreateOrderInput struct {
	UserID int64             `json:"user_id"`
	Items  []CreateOrderItem `json:"items"`
}

type CreateOrderItem struct {
	ProductID int64 `json:"product_id"`
	Quantity  int32 `json:"quantity"`
	Price     int64 `json:"price"`
}

type CreateOrderOutput struct {
	ID int64 `json:"id"`
}
