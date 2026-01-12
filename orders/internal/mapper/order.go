package mapper

import (
	"github.com/ChernykhITMO/order-processing-platform/orders/internal/domain"
	ordersv1 "github.com/ChernykhITMO/order-processing-proto/gen/go/opp/orders/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func MapToProto(order domain.Order) *ordersv1.Order {
	items := make([]*ordersv1.OrderItem, 0, len(order.Items))
	for _, it := range order.Items {
		items = append(items, &ordersv1.OrderItem{
			ProductId: int64(it.ProductID),
			Quantity:  it.Quantity,
			Price:     &ordersv1.Money{Money: int64(it.Price)},
		})
	}

	statusProto := MapStatusToProto(order.Status)

	var (
		createdAt *timestamppb.Timestamp
		updatedAt *timestamppb.Timestamp
	)
	if !order.CreatedAt.IsZero() {
		createdAt = timestamppb.New(order.CreatedAt)
	}
	if !order.UpdatedAt.IsZero() {
		updatedAt = timestamppb.New(order.UpdatedAt)
	}

	return &ordersv1.Order{
		OrderId:     int64(order.ID),
		UserId:      int64(order.UserID),
		Status:      statusProto,
		Items:       items,
		TotalAmount: &ordersv1.Money{Money: int64(order.TotalAmount)},
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
}
