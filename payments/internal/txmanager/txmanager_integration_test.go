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
	_ = godotenv.Load("../../../payments/.env")

	dsn := os.Getenv("PAYMENTS_PG_DSN_TEST")
	if dsn == "" {
		t.Skip("PAYMENTS_PG_DSN_TEST is not set")
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer pool.Close()

	if _, err := pool.Exec(context.Background(), `TRUNCATE TABLE payments, events, processed_events RESTART IDENTITY CASCADE`); err != nil {
		t.Fatalf("cleanup tables: %v", err)
	}

	mgr := New(pool)
	expectedErr := errors.New("rollback me")

	err = mgr.WithinTransaction(context.Background(), func(ctx context.Context, tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `INSERT INTO payments (order_id, user_id, total_amount, status) VALUES ($1, $2, $3, $4)`, 1, 2, 300, "pending"); err != nil {
			return err
		}
		return expectedErr
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}

	var count int
	if err := pool.QueryRow(context.Background(), `SELECT COUNT(*) FROM payments`).Scan(&count); err != nil {
		t.Fatalf("count payments: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected rollback, got %d persisted payments", count)
	}
}
