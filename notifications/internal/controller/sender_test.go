package controller

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/ChernykhITMO/order-processing-platform/notifications/internal/domain"
	"github.com/ChernykhITMO/order-processing-platform/notifications/internal/domain/events"
	"github.com/ChernykhITMO/order-processing-platform/notifications/internal/dto"
	"github.com/ChernykhITMO/order-processing-platform/notifications/internal/services"
)

type redisMock struct {
	saveCalled int
	savedKey   string
	savedValue events.Payment
	saveErr    error
}

func (m *redisMock) SaveNotification(ctx context.Context, key string, value events.Payment) error {
	m.saveCalled++
	m.savedKey = key
	m.savedValue = value
	return m.saveErr
}

func (m *redisMock) GetNotification(ctx context.Context, key string) (events.Payment, error) {
	return events.Payment{}, domain.ErrNotFound
}

func TestSender_HandleMessage_InvalidJSON(t *testing.T) {
	st := &redisMock{}
	log := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
	uc := services.New(st, log)
	sender := NewSender(uc, log)

	if err := sender.HandleMessage([]byte("{")); err == nil {
		t.Fatalf("expected error")
	}
}

func TestSender_HandleMessage_Success(t *testing.T) {
	st := &redisMock{}
	log := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
	uc := services.New(st, log)
	sender := NewSender(uc, log)

	payment := dto.Payment{OrderID: 10, UserID: 20, OrderStatus: domain.StatusSucceeded}
	payload, _ := json.Marshal(payment)

	if err := sender.HandleMessage(payload); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if st.saveCalled != 1 {
		t.Fatalf("expected SaveNotification to be called")
	}
	if st.savedKey != "10" {
		t.Fatalf("expected key 10, got %s", st.savedKey)
	}
	if st.savedValue.OrderID != 10 || st.savedValue.UserID != 20 {
		t.Fatalf("unexpected payment: %+v", st.savedValue)
	}
}

func TestSender_HandleMessage_SaveError(t *testing.T) {
	st := &redisMock{saveErr: errors.New("redis")}
	log := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
	uc := services.New(st, log)
	sender := NewSender(uc, log)

	payment := dto.Payment{OrderID: 10, UserID: 20, OrderStatus: domain.StatusSucceeded}
	payload, _ := json.Marshal(payment)

	if err := sender.HandleMessage(payload); err == nil {
		t.Fatalf("expected error")
	}
}
