package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/ChernykhITMO/order-processing-platform/orders/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (s *Storage) GetOrderByID(ctx context.Context, id int64) (*domain.Order, error) {
	const op = "storage.postgres.GetOrderByID"

	const query = `
	SELECT 
	    o.id, o.user_id, o.status, o.created_at, o.updated_at,
	    i.product_id, i.quantity, i.price
	FROM orders AS o
	LEFT JOIN order_items AS i ON o.id = i.order_id
	WHERE o.id = $1
	ORDER BY i.id;
	`

	rows, err := s.db.Query(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer func() {
		rows.Close()
	}()

	var (
		orderID   int64
		userID    int64
		status    string
		createdAt time.Time
		updatedAt time.Time
		productID pgtype.Int8
		quantity  pgtype.Int4
		price     pgtype.Int8
		find      bool
	)

	items := make([]domain.OrderItem, 0)

	for rows.Next() {
		find = true
		if err := rows.Scan(
			&orderID, &userID, &status, &createdAt,
			&updatedAt, &productID, &quantity, &price); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		if productID.Valid {
			item, err := domain.NewOrderItem(productID.Int64, quantity.Int32, price.Int64)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", op, err)
			}
			items = append(items, item)
		}
	}

	if !find {
		return nil, fmt.Errorf("%s: %w", op, pgx.ErrNoRows)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	order, err := domain.NewOrder(
		orderID, userID, status,
		items, createdAt, updatedAt)

	return order, err
}
