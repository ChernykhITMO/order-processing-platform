package mapper

import (
	"testing"

	"github.com/ChernykhITMO/order-processing-platform/orders/internal/domain"
	ordersv1 "github.com/ChernykhITMO/order-processing-proto/gen/go/opp/orders/v1"
)

func TestMapStatusToProto_New(t *testing.T) {
	got := MapStatusToProto(domain.StatusNew)
	if got != ordersv1.OrderStatus_new {
		t.Fatalf("expected %v, got %v", ordersv1.OrderStatus_new, got)
	}
}

func TestMapStatusToProto_Unknown(t *testing.T) {
	got := MapStatusToProto(domain.Status("unknown"))
	if got != ordersv1.OrderStatus_unspecified {
		t.Fatalf("expected %v, got %v", ordersv1.OrderStatus_unspecified, got)
	}
}
