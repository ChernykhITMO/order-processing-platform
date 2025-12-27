package service

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/ChernykhITMO/order-processing-platform/orders/internal/domain"
)

type fakeProvider struct {
	createCalled int
	getCalled    int

	lastUserID int64
	lastItems  []domain.OrderItem
	lastID     int64

	createErr error
	getOrder  *domain.Order
	getErr    error
}

func (f *fakeProvider) CreateOrder(ctx context.Context, userID int64, items []domain.OrderItem) error {
	f.createCalled++
	f.lastUserID = userID
	f.lastItems = items
	return f.createErr
}

func (f *fakeProvider) GetOrderByID(ctx context.Context, id int64) (*domain.Order, error) {
	f.getCalled++
	f.lastID = id
	return f.getOrder, f.getErr
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestOrder_CreateOrder_InvalidUserID(t *testing.T) {
	fp := &fakeProvider{}
	svc := New(testLogger(), fp)

	err := svc.CreateOrder(context.Background(), 0, []domain.OrderItem{{ProductID: 1, Quantity: 1}})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if fp.createCalled != 0 {
		t.Fatalf("expected provider not called, got %d", fp.createCalled)
	}
}

func TestOrder_CreateOrder_EmptyItems(t *testing.T) {
	fp := &fakeProvider{}
	svc := New(testLogger(), fp)

	err := svc.CreateOrder(context.Background(), 1, nil)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if fp.createCalled != 0 {
		t.Fatalf("expected provider not called, got %d", fp.createCalled)
	}
}

func TestOrder_CreateOrder_ProviderError(t *testing.T) {
	fp := &fakeProvider{createErr: errors.New("db failed")}
	svc := New(testLogger(), fp)

	err := svc.CreateOrder(context.Background(), 1, []domain.OrderItem{{ProductID: 1, Quantity: 1}})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if fp.createCalled != 1 {
		t.Fatalf("expected provider called once, got %d", fp.createCalled)
	}
	if !errors.Is(err, fp.createErr) {
		t.Fatalf("expected wrapped error, got: %v", err)
	}
}
