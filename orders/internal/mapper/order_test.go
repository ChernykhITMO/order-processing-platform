package mapper

import (
	"testing"
	"time"

	"github.com/ChernykhITMO/order-processing-platform/orders/internal/domain"
	ordersv1 "github.com/ChernykhITMO/order-processing-proto/gen/go/opp/orders/v1"
)

func TestMapToProto_OK(t *testing.T) {
	items := []domain.OrderItem{
		{ProductID: 1, Quantity: 2, Price: 100},
	}
	createdAt := time.Now().UTC()
	updatedAt := createdAt.Add(time.Minute)

	order, err := domain.NewOrder(1, 2, string(domain.StatusNew), items, createdAt, updatedAt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	proto := MapToProto(*order)
	if proto.OrderId != 1 || proto.UserId != 2 {
		t.Fatalf("unexpected ids: order_id=%d user_id=%d", proto.OrderId, proto.UserId)
	}
	if proto.Status != ordersv1.OrderStatus_new {
		t.Fatalf("unexpected status: %v", proto.Status)
	}
	if len(proto.Items) != 1 {
		t.Fatalf("items count mismatch: got %d, want %d", len(proto.Items), 1)
	}
	if proto.TotalAmount.GetMoney() != int64(order.TotalAmount) {
		t.Fatalf("total mismatch: got %d, want %d", proto.TotalAmount.GetMoney(), order.TotalAmount)
	}
	if proto.CreatedAt == nil || proto.UpdatedAt == nil {
		t.Fatalf("expected timestamps to be set")
	}
}

func TestMapToProto_ZeroTimestamps(t *testing.T) {
	order := domain.Order{
		ID:     1,
		UserID: 2,
		Status: domain.StatusNew,
		Items: []domain.OrderItem{
			{ProductID: 1, Quantity: 2, Price: 100},
		},
	}

	proto := MapToProto(order)
	if proto.CreatedAt != nil || proto.UpdatedAt != nil {
		t.Fatalf("expected nil timestamps for zero time")
	}
}
