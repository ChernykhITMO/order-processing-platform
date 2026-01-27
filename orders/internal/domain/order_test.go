package domain

import (
	"errors"
	"testing"
	"time"
)

func TestNewOrder_OK(t *testing.T) {
	items := []OrderItem{
		{ProductID: 1, Quantity: 2, Price: 100},
		{ProductID: 2, Quantity: 1, Price: 50},
	}
	createdAt := time.Now().UTC()
	updatedAt := createdAt.Add(time.Minute)

	order, err := NewOrder(1, 2, string(StatusNew), items, createdAt, updatedAt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if order.TotalAmount != 250 {
		t.Fatalf("total amount mismatch: got %d, want %d", order.TotalAmount, 250)
	}
	if order.CreatedAt.IsZero() || order.UpdatedAt.IsZero() {
		t.Fatalf("timestamps should be set")
	}
}

func TestNewOrder_Invalid(t *testing.T) {
	tests := []struct {
		name      string
		orderID   int64
		userID    int64
		wantErrIs error
	}{
		{"invalid order id", 0, 1, ErrInvalidOrderID},
		{"invalid user id", 1, 0, ErrInvalidUserID},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewOrder(tt.orderID, tt.userID, string(StatusNew), nil, time.Now(), time.Now())
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !errors.Is(err, tt.wantErrIs) {
				t.Fatalf("expected error %v, got %v", tt.wantErrIs, err)
			}
		})
	}
}
