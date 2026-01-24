package dto

import (
	"time"

	ordersv1 "github.com/ChernykhITMO/order-processing-proto/gen/go/opp/orders/v1"
)

type OrderItem struct {
	ProductID int64 `json:"product_id"`
	Quantity  int32 `json:"quantity"`
	Price     int64 `json:"price"`
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
	TotalAmount int64       `json:"total_amount"`
	CreatedAt   string      `json:"created_at"`
	UpdatedAt   string      `json:"updated_at"`
}

type CreateOrderResponse struct {
	OrderID int64 `json:"order_id"`
}

type Money struct {
	Money int64 `json:"money"`
}

type GetOrderRequest struct {
	OrderID int64 `json:"order_id"`
}

type GetOrderResponse struct {
	Order Order `json:"order"`
}

func ProtoGetToDTO(response ordersv1.GetOrderResponse) GetOrderResponse {
	var output GetOrderResponse
	if response.Order == nil {
		return output
	}
	output.Order.OrderID = response.Order.OrderId
	output.Order.UserID = response.Order.UserId
	output.Order.Status = mapStatus(response.Order.Status)

	output.Order.Items = ProtoToDTOItems(response.Order.Items)
	if response.Order.TotalAmount != nil {
		output.Order.TotalAmount = response.Order.TotalAmount.Money
	}

	if response.Order.CreatedAt != nil {
		output.Order.CreatedAt = response.Order.CreatedAt.AsTime().Format(time.RFC3339)
	}
	if response.Order.UpdatedAt != nil {
		output.Order.UpdatedAt = response.Order.UpdatedAt.AsTime().Format(time.RFC3339)
	}

	return output
}

func ProtoToDTOItems(items []*ordersv1.OrderItem) []OrderItem {
	out := make([]OrderItem, 0, len(items))
	for _, it := range items {
		var price int64
		if it.Price != nil {
			price = it.Price.Money
		}
		out = append(out, OrderItem{
			ProductID: it.ProductId,
			Quantity:  it.Quantity,
			Price:     price,
		})
	}
	return out
}

func mapStatus(status ordersv1.OrderStatus) string {
	switch status {
	case ordersv1.OrderStatus_NEW:
		return "new"
	case ordersv1.OrderStatus_PAYMENT_PENDING:
		return "payment_pending"
	default:
		return "unspecified"
	}
}
