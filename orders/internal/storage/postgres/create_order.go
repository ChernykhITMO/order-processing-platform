package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ChernykhITMO/order-processing-platform/orders/internal/domain"
	"github.com/ChernykhITMO/order-processing-platform/orders/internal/domain/events"
)

const (
	orderCreated string = "order created"
)

func (s *Storage) CreateOrder(
	ctx context.Context,
	userID int64,
	items []domain.OrderItem) (orderID int64, err error) {
	const op = "storage.postgres.CreateOrder"

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	const insertOrder = `
		INSERT INTO orders (user_id, status) VALUES ($1, $2)
		RETURNING id, created_at
	`

	var createdAt time.Time

	if err := tx.QueryRowContext(ctx, insertOrder, userID, domain.StatusNew).Scan(&orderID, &createdAt); err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	const insertOrderItem = `
		INSERT INTO order_items (order_id, product_id, quantity, price)
		VALUES ($1,$2, $3, $4);
	`
	for i := 0; i < len(items); i++ {
		_, err = tx.ExecContext(
			ctx, insertOrderItem, orderID, items[i].ProductID,
			items[i].Quantity, items[i].Price)
		if err != nil {
			return 0, fmt.Errorf("%s: %w", op, err)
		}
	}

	var totalAmount int64
	const querySum = `SELECT SUM(price *quantity) FROM order_items WHERE order_id = $1`
	if err := tx.QueryRowContext(ctx, querySum, orderID).Scan(&totalAmount); err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	evt := events.OrderCreated{
		OrderID:     domain.ID(orderID),
		UserID:      domain.ID(userID),
		TotalAmount: domain.Money(totalAmount),
		CreatedAt:   createdAt,
	}

	payload, err := json.Marshal(evt)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	if err := s.saveEvent(ctx, tx, orderCreated, payload, int64(evt.OrderID), createdAt); err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return orderID, tx.Commit()
}

func (s *Storage) saveEvent(
	ctx context.Context,
	tx *sql.Tx,
	eventType string,
	payload []byte,
	aggregateID int64,
	createdAt time.Time) error {
	const op = "storage.postgres.saveEvent"

	const query = `
		INSERT INTO events (event_type, payload, aggregate_id, created_at)
		VALUES ($1, $2, $3, $4)
	`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer func() {
		_ = stmt.Close()
	}()

	_, err = stmt.ExecContext(ctx, eventType, payload, aggregateID, createdAt)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
