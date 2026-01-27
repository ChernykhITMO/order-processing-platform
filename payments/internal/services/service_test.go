package services

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
)

func TestService_HandleOrderCreated_InvalidEventID(t *testing.T) {
	st := &storageMock{tx: &txMock{}}
	log := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
	svc := New(st, log, "payment-status")

	err := svc.HandleOrderCreated(context.Background(), dto.OrderCreated{
		EventID: 0,
		OrderID: 1,
		UserID:  2,
	})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, domain.ErrInvalidEventID) {
		t.Fatalf("expected error %v, got %v", domain.ErrInvalidEventID, err)
	}
	if st.tx.tryMarkCalled != 0 {
		t.Fatalf("TryMarkProcessed should not be called")
	}
}

func TestService_HandleOrderCreated_AlreadyProcessed(t *testing.T) {
	st := &storageMock{tx: &txMock{tryMarkOK: false}}
	log := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
	svc := New(st, log, "payment-status")

	err := svc.HandleOrderCreated(context.Background(), dto.OrderCreated{
		EventID:     1,
		OrderID:     2,
		UserID:      3,
		TotalAmount: 100,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if st.tx.upsertCalled != 0 || st.tx.updateCalled != 0 || st.tx.saveEventCalled != 0 {
		t.Fatalf("no writes expected when event already processed")
	}
}

func TestService_HandleOrderCreated_SuccessEven(t *testing.T) {
	st := &storageMock{tx: &txMock{tryMarkOK: true}}
	log := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
	svc := New(st, log, "payment-status")

	err := svc.HandleOrderCreated(context.Background(), dto.OrderCreated{
		EventID:     10,
		OrderID:     2,
		UserID:      3,
		TotalAmount: 100,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if st.tx.upsertCalled != 1 {
		t.Fatalf("UpsertPayment calls: got %d, want %d", st.tx.upsertCalled, 1)
	}
	if st.tx.updateCalled != 1 {
		t.Fatalf("UpdatePaymentStatus calls: got %d, want %d", st.tx.updateCalled, 1)
	}
	if st.tx.saveEventCalled != 1 {
		t.Fatalf("SaveEvent calls: got %d, want %d", st.tx.saveEventCalled, 1)
	}

	var ev events.PaymentStatus
	if err := json.Unmarshal(st.tx.savedPayload, &ev); err != nil {
		t.Fatalf("unmarshal saved payload: %v", err)
	}
	if ev.OrderStatus != domain.StatusSucceeded {
		t.Fatalf("expected status %s, got %s", domain.StatusSucceeded, ev.OrderStatus)
	}
}

func TestService_HandleOrderCreated_SuccessOdd(t *testing.T) {
	st := &storageMock{tx: &txMock{tryMarkOK: true}}
	log := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
	svc := New(st, log, "payment-status")

	err := svc.HandleOrderCreated(context.Background(), dto.OrderCreated{
		EventID:     10,
		OrderID:     3,
		UserID:      3,
		TotalAmount: 100,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var ev events.PaymentStatus
	if err := json.Unmarshal(st.tx.savedPayload, &ev); err != nil {
		t.Fatalf("unmarshal saved payload: %v", err)
	}
	if ev.OrderStatus != domain.StatusFailed {
		t.Fatalf("expected status %s, got %s", domain.StatusFailed, ev.OrderStatus)
	}
}

type storageMock struct {
	tx        *txMock
	runCalled int
}

func (m *storageMock) RunInTx(ctx context.Context, fn func(tx ports.StorageTx) error) error {
	m.runCalled++
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
	tryMarkOK bool

	tryMarkCalled   int
	upsertCalled    int
	updateCalled    int
	saveEventCalled int

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
	return m.tryMarkOK, nil
}

func (m *txMock) SaveEvent(ctx context.Context, eventType string, payload []byte, aggregateID int64) error {
	m.saveEventCalled++
	m.savedPayload = payload
	return nil
}
