package mapper

import (
	"errors"
	"testing"

	"github.com/ChernykhITMO/order-processing-platform/orders/internal/domain"
	"github.com/ChernykhITMO/order-processing-platform/orders/internal/dto"
	ordersv1 "github.com/ChernykhITMO/order-processing-proto/gen/go/opp/orders/v1"
)

func TestMapInputItems_OK(t *testing.T) {
	input := []dto.CreateOrderItem{
		{ProductID: 10, Quantity: 2, Price: 100},
		{ProductID: 20, Quantity: 1, Price: 50},
	}

	items, err := MapInputItems(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("items count mismatch: got %d, want %d", len(items), 2)
	}
	if items[0].ProductID != 10 || items[0].Quantity != 2 || items[0].Price != 100 {
		t.Fatalf("unexpected first item: %+v", items[0])
	}
}

func TestMapInputItems_Invalid(t *testing.T) {
	input := []dto.CreateOrderItem{
		{ProductID: 0, Quantity: 1, Price: 10},
	}

	_, err := MapInputItems(input)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, domain.ErrInvalidProductID) {
		t.Fatalf("expected error %v, got %v", domain.ErrInvalidProductID, err)
	}
}

func TestMapProtoItems_OK(t *testing.T) {
	input := []*ordersv1.OrderItem{
		{ProductId: 1, Quantity: 2, Price: &ordersv1.Money{Money: 100}},
		{ProductId: 2, Quantity: 1, Price: &ordersv1.Money{Money: 50}},
	}

	items, err := MapProtoItems(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("items count mismatch: got %d, want %d", len(items), 2)
	}
	if items[1].ProductID != 2 || items[1].Quantity != 1 || items[1].Price != 50 {
		t.Fatalf("unexpected second item: %+v", items[1])
	}
}

func TestMapProtoItems_InvalidPrice(t *testing.T) {
	input := []*ordersv1.OrderItem{
		{ProductId: 1, Quantity: 2, Price: nil},
	}

	_, err := MapProtoItems(input)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, domain.ErrInvalidPrice) {
		t.Fatalf("expected error %v, got %v", domain.ErrInvalidPrice, err)
	}
}

func TestMapToCreateItems_OK(t *testing.T) {
	input := []*ordersv1.OrderItem{
		{ProductId: 1, Quantity: 2, Price: &ordersv1.Money{Money: 100}},
		{ProductId: 2, Quantity: 1, Price: nil},
	}

	items, err := MapToCreateItems(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("items count mismatch: got %d, want %d", len(items), 2)
	}
	if items[1].Price != 0 {
		t.Fatalf("expected zero price for nil money, got %d", items[1].Price)
	}
}
