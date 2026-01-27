package services

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/ChernykhITMO/order-processing-platform/orders/internal/domain"
	"github.com/ChernykhITMO/order-processing-platform/orders/internal/domain/events"
	"github.com/ChernykhITMO/order-processing-platform/orders/internal/dto"
)

func TestOrdersService_Create(t *testing.T) {
	errDB := errors.New("db")
	validInput := dto.CreateOrderInput{
		UserID: 1,
		Items: []dto.CreateOrderItem{
			{ProductID: 10, Quantity: 2, Price: 100},
		},
	}
	invalidInput := dto.CreateOrderInput{
		UserID: 1,
		Items:  []dto.CreateOrderItem{},
	}

	tests := []struct {
		name            string
		input           dto.CreateOrderInput
		mockErr         error
		wantErr         bool
		wantCreateCalls int
		wantID          int64
	}{
		{"ok", validInput, nil, false, 1, 42},
		{"validation error", invalidInput, nil, true, 0, 0},
		{"repo error", validInput, errDB, true, 1, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &postgresMock{
				createErr:     tt.mockErr,
				createOrderID: 42,
			}
			log := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
			svc := New(log, mock)

			got, err := svc.CreateOrder(context.Background(), tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if mock.createCalled != tt.wantCreateCalls {
				t.Fatalf("CreateOrder calls: got %d, want %d", mock.createCalled, tt.wantCreateCalls)
			}

			if !tt.wantErr && got.ID != tt.wantID {
				t.Fatalf("CreateOrder ID: got %d, want %d", got.ID, tt.wantID)
			}
		})
	}
}

func TestOrdersService_Get(t *testing.T) {
	errDB := errors.New("db")
	tests := []struct {
		name      string
		input     dto.GetOrderInput
		mockOrder *domain.Order
		mockErr   error
		wantErrIs error
		wantOrder bool
		wantCalls int
	}{
		{
			name:  "ok",
			input: dto.GetOrderInput{ID: 10},
			mockOrder: &domain.Order{
				ID:     10,
				UserID: 1,
			},
			wantOrder: true,
			wantCalls: 1,
		},
		{
			name:      "invalid id",
			input:     dto.GetOrderInput{ID: 0},
			wantErrIs: domain.ErrInvalidOrderID,
			wantCalls: 0,
		},
		{
			name:      "not found",
			input:     dto.GetOrderInput{ID: 11},
			mockErr:   sql.ErrNoRows,
			wantErrIs: domain.ErrOrderNotFound,
			wantCalls: 1,
		},
		{
			name:      "repo error",
			input:     dto.GetOrderInput{ID: 12},
			mockErr:   errDB,
			wantErrIs: errDB,
			wantCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &postgresMock{
				getOrder: tt.mockOrder,
				getErr:   tt.mockErr,
			}
			log := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
			svc := New(log, mock)

			_, err := svc.GetOrder(context.Background(), tt.input)
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

			if mock.getCalled != tt.wantCalls {
				t.Fatalf("GetOrderByID calls: got %d, want %d", mock.getCalled, tt.wantCalls)
			}
		})
	}
}

type postgresMock struct {
	createCalled  int
	createUserID  int64
	createItems   []domain.OrderItem
	createErr     error
	createOrderID int64

	getCalled int
	getOrder  *domain.Order
	getErr    error
}

func (m *postgresMock) CreateOrder(ctx context.Context, userID int64, items []domain.OrderItem) (int64, error) {
	m.createCalled++
	m.createUserID = userID
	m.createItems = items
	if m.createErr != nil {
		return 0, m.createErr
	}
	return m.createOrderID, nil
}

func (m *postgresMock) GetOrderByID(ctx context.Context, id int64) (*domain.Order, error) {
	m.getCalled++
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.getOrder, nil
}

func (m *postgresMock) GetNewEvent(ctx context.Context) (events.OrderCreated, int64, error) {
	return events.OrderCreated{}, 0, nil
}

func (m *postgresMock) MarkSent(ctx context.Context, eventID int64) error {
	return nil
}

func (m *postgresMock) Close() error {
	return nil
}
