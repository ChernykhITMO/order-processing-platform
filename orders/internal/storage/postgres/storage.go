package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/ChernykhITMO/order-processing-platform/orders/internal/domain"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.postgres.New"

	db, err := sql.Open("pgx", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) CreateOrder(ctx context.Context, userID int64, items []domain.OrderItem) (int64, error) {
	const op = "storage.postgres.CreateOrder"

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	status := domain.StatusToDB(domain.StatusNew)
	const insertOrder = `
		INSERT INTO orders (user_id, status) VALUES ($1, $2)
		RETURNING id
	`

	var orderID int64
	if err := tx.QueryRowContext(ctx, insertOrder, userID, string(status)).Scan(&orderID); err != nil {
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

	return orderID, tx.Commit()
}

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

	rows, err := s.db.QueryContext(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var (
		orderID   int64
		userID    int64
		status    string
		createdAt time.Time
		updatedAt time.Time
		productID sql.NullInt64
		quantity  sql.NullInt32
		price     sql.NullInt64
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
		return nil, fmt.Errorf("%s: %w", op, sql.ErrNoRows)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	order, err := domain.NewOrder(
		orderID, userID, domain.StatusFromDB(status),
		items, createdAt, updatedAt)

	return order, err
}
