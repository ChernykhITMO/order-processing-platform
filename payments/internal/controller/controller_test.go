package controller

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/ChernykhITMO/order-processing-platform/payments/internal/domain"
	"github.com/ChernykhITMO/order-processing-platform/payments/internal/domain/events"
	"github.com/ChernykhITMO/order-processing-platform/payments/internal/dto"
	"github.com/ChernykhITMO/order-processing-platform/payments/internal/ports"
	"github.com/ChernykhITMO/order-processing-platform/payments/internal/services"
)

type storageMock struct {
	tx *txMock
}

func (m *storageMock) RunInTx(ctx context.Context, fn func(tx ports.StorageTx) error) error {
	return fn(m.tx)
}

func (m *storageMock) GetNewEvent(ctx context.Context) (events.PaymentStatus, error) {
	return events.PaymentStatus{}, nil
}

func (m *storageMock) MarkSent(ctx context.Context, id int64) error {
	return nil
}

func (m *storageMock) Close() error {
	return nil
}

type txMock struct {
	tryMarkCalled int
	upsertCalled  int
	updateCalled  int
	saveCalled    int

	savedPayload []byte
}

func (m *txMock) UpsertPayment(ctx context.Context, orderID, userID, totalAmount int64, status string) error {
	m.upsertCalled++
	return nil
}

func (m *txMock) UpdatePaymentStatus(ctx context.Context, orderID int64, status string) error {
	m.updateCalled++
	return nil
}

func (m *txMock) TryMarkProcessed(ctx context.Context, eventId int64) (bool, error) {
	m.tryMarkCalled++
	return true, nil
}

func (m *txMock) SaveEvent(ctx context.Context, eventType string, payload []byte, aggregateID int64) error {
	m.saveCalled++
	m.savedPayload = payload
	return nil
}

func TestController_HandleMessage_InvalidJSON(t *testing.T) {
	st := &storageMock{tx: &txMock{}}
	log := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
	svc := services.New(st, log, "event-status")
	ctrl := NewController(*svc, st, nil, "", log)

	if err := ctrl.HandleMessage(context.Background(), []byte("{")); err == nil {
		t.Fatalf("expected error")
	}
}

func TestController_HandleMessage_Success(t *testing.T) {
	tx := &txMock{}
	st := &storageMock{tx: tx}
	log := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
	svc := services.New(st, log, "event-status")
	ctrl := NewController(*svc, st, nil, "", log)

	input := dto.OrderCreated{EventID: 1, OrderID: 2, UserID: 3, TotalAmount: 100}
	payload, _ := json.Marshal(input)

	if err := ctrl.HandleMessage(context.Background(), payload); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tx.tryMarkCalled != 1 || tx.upsertCalled != 1 || tx.updateCalled != 1 || tx.saveCalled != 1 {
		t.Fatalf("unexpected calls: try=%d upsert=%d update=%d save=%d", tx.tryMarkCalled, tx.upsertCalled, tx.updateCalled, tx.saveCalled)
	}

	var ev events.PaymentStatus
	if err := json.Unmarshal(tx.savedPayload, &ev); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if ev.OrderStatus != domain.StatusSucceeded {
		t.Fatalf("status: got %s, want %s", ev.OrderStatus, domain.StatusSucceeded)
	}
}

func TestController_HandleMessage_ServiceError(t *testing.T) {
	st := &storageMock{tx: &txMock{}}
	log := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
	svc := services.New(st, log, "event-status")
	ctrl := NewController(*svc, st, nil, "", log)

	input := dto.OrderCreated{EventID: 0, OrderID: 2, UserID: 3, TotalAmount: 100}
	payload, _ := json.Marshal(input)

	if err := ctrl.HandleMessage(context.Background(), payload); err == nil || !errors.Is(err, domain.ErrInvalidEventID) {
		t.Fatalf("expected invalid event id error, got %v", err)
	}
}
