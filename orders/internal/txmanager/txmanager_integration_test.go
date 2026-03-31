//go:build integration
// +build integration

package txmanager

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func TestManagerRollback_Integration(t *testing.T) {
	_ = godotenv.Load("../../../.env")
	_ = godotenv.Load("../../../orders/.env")

	dsn := os.Getenv("ORDERS_PG_DSN_TEST")
	if dsn == "" {
		t.Skip("ORDERS_PG_DSN_TEST is not set")
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer pool.Close()

	if _, err := pool.Exec(context.Background(), `TRUNCATE TABLE events, order_items, orders RESTART IDENTITY CASCADE`); err != nil {
		t.Fatalf("cleanup tables: %v", err)
	}

	mgr := New(pool)
	expectedErr := errors.New("rollback me")

	err = mgr.WithinTransaction(context.Background(), func(ctx context.Context, tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `INSERT INTO orders (user_id, status, created_at, updated_at) VALUES ($1, $2, NOW(), NOW())`, 42, "new"); err != nil {
			return err
		}
		return expectedErr
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}

	var count int
	if err := pool.QueryRow(context.Background(), `SELECT COUNT(*) FROM orders`).Scan(&count); err != nil {
		t.Fatalf("count orders: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected rollback, got %d persisted orders", count)
	}
}
