package services

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/ChernykhITMO/order-processing-platform/notifications/internal/domain"
	"github.com/ChernykhITMO/order-processing-platform/notifications/internal/domain/events"
	"github.com/ChernykhITMO/order-processing-platform/notifications/internal/dto"
)

func TestNotificationService_Save(t *testing.T) {
	errDB := errors.New("db")
	validInput := dto.SaveInput{
		Key:     "10",
		OrderID: 10,
		UserID:  20,
		Status:  domain.StatusSucceeded,
	}

	tests := []struct {
		name        string
		input       dto.SaveInput
		storageErr  error
		wantErrIs   error
		wantCalls   int
		wantKey     string
		wantOrderID int64
	}{
		{"ok", validInput, nil, nil, 1, "10", 10},
		{"empty key", dto.SaveInput{Key: "", OrderID: 10, UserID: 20, Status: domain.StatusSucceeded}, nil, domain.ErrIsEmptyKey, 0, "", 0},
		{"invalid user", dto.SaveInput{Key: "10", OrderID: 10, UserID: 0, Status: domain.StatusSucceeded}, nil, domain.ErrInvalidUserID, 0, "", 0},
		{"invalid order", dto.SaveInput{Key: "10", OrderID: 0, UserID: 20, Status: domain.StatusSucceeded}, nil, domain.ErrInvalidOrderID, 0, "", 0},
		{"invalid status", dto.SaveInput{Key: "10", OrderID: 10, UserID: 20, Status: "unknown"}, nil, domain.ErrInvalidStatus, 0, "", 0},
		{"storage error", validInput, errDB, errDB, 1, "10", 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := &redisMock{saveErr: tt.storageErr}
			log := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
			svc := New(st, log)

			err := svc.SaveNotification(context.Background(), tt.input)
			if tt.wantErrIs != nil {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if !errors.Is(err, tt.wantErrIs) {
					t.Fatalf("expected error %v, got %v", tt.wantErrIs, err)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if st.saveCalled != tt.wantCalls {
				t.Fatalf("SaveNotification calls: got %d, want %d", st.saveCalled, tt.wantCalls)
			}
			if tt.wantCalls > 0 && st.savedKey != tt.wantKey {
				t.Fatalf("saved key: got %s, want %s", st.savedKey, tt.wantKey)
			}
			if tt.wantCalls > 0 && st.savedPayment.OrderID != events.ID(tt.wantOrderID) {
				t.Fatalf("saved order id: got %d, want %d", st.savedPayment.OrderID, tt.wantOrderID)
			}
		})
	}
}

func TestNotificationService_Get(t *testing.T) {
	errDB := errors.New("db")
	found := events.Payment{
		OrderID:     10,
		UserID:      20,
		OrderStatus: domain.StatusSucceeded,
	}

	tests := []struct {
		name       string
		input      dto.GetInput
		storageVal events.Payment
		storageErr error
		wantErrIs  error
		wantCalls  int
	}{
		{"ok", dto.GetInput{Key: "10"}, found, nil, nil, 1},
		{"empty key", dto.GetInput{Key: ""}, events.Payment{}, nil, domain.ErrIsEmptyKey, 0},
		{"not found", dto.GetInput{Key: "10"}, events.Payment{}, domain.ErrNotFound, domain.ErrNotFound, 1},
		{"storage error", dto.GetInput{Key: "10"}, events.Payment{}, errDB, errDB, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := &redisMock{getVal: tt.storageVal, getErr: tt.storageErr}
			log := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
			svc := New(st, log)

			got, err := svc.GetNotification(context.Background(), tt.input)
			if tt.wantErrIs != nil {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if !errors.Is(err, tt.wantErrIs) {
					t.Fatalf("expected error %v, got %v", tt.wantErrIs, err)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if st.getCalled != tt.wantCalls {
				t.Fatalf("GetNotification calls: got %d, want %d", st.getCalled, tt.wantCalls)
			}
			if tt.wantErrIs == nil && got.OrderID != found.OrderID {
				t.Fatalf("unexpected order id: got %d, want %d", got.OrderID, found.OrderID)
			}
		})
	}
}

type redisMock struct {
	saveCalled   int
	savedKey     string
	savedPayment events.Payment
	saveErr      error

	getCalled int
	getKey    string
	getVal    events.Payment
	getErr    error
}

func (m *redisMock) SaveNotification(ctx context.Context, key string, value events.Payment) error {
	m.saveCalled++
	m.savedKey = key
	m.savedPayment = value
	return m.saveErr
}

func (m *redisMock) GetNotification(ctx context.Context, key string) (events.Payment, error) {
	m.getCalled++
	m.getKey = key
	if m.getErr != nil {
		return events.Payment{}, m.getErr
	}
	return m.getVal, nil
}
