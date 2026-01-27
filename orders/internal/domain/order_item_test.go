package domain

import (
	"errors"
	"testing"
)

func TestNewOrderItem_OK(t *testing.T) {
	item, err := NewOrderItem(10, 2, 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.ProductID != 10 || item.Quantity != 2 || item.Price != 100 {
		t.Fatalf("unexpected item: %+v", item)
	}
}

func TestNewOrderItem_Invalid(t *testing.T) {
	tests := []struct {
		name      string
		productID int64
		quantity  int32
		price     int64
		wantErrIs error
	}{
		{"invalid product", 0, 1, 10, ErrInvalidProductID},
		{"invalid quantity", 1, 0, 10, ErrInvalidQuantity},
		{"invalid price", 1, 1, 0, ErrInvalidPrice},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewOrderItem(tt.productID, tt.quantity, tt.price)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !errors.Is(err, tt.wantErrIs) {
				t.Fatalf("expected error %v, got %v", tt.wantErrIs, err)
			}
		})
	}
}
