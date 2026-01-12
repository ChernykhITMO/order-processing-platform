//go:build integration
// +build integration

package postgres

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/ChernykhITMO/order-processing-platform/orders/internal/domain"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func getTestDSN(t *testing.T) string {
	dsn := os.Getenv("ORDERS_PG_DSN")
	if dsn == "" {
		t.Skip("ORDERS_PG_DSN is not set")
	}
	return dsn
}

func cleanupTables(t *testing.T, db *sql.DB) {
	_, err := db.Exec("TRUNCATE TABLE order_items, orders RESTART IDENTITY CASCADE")
	if err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}
}

func TestStorage_CreateAndGetOrder(t *testing.T) {
	dsn := getTestDSN(t)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	cleanupTables(t, db)
	defer cleanupTables(t, db)

	storage, err := New(dsn)
	if err != nil {
		t.Fatalf("new storage: %v", err)
	}

	items := []domain.OrderItem{
		{ProductID: 1, Quantity: 2, Price: 100},
		{ProductID: 2, Quantity: 3, Price: 50},
	}

	orderID, err := storage.CreateOrder(context.Background(), 10, items)
	if err != nil {
		t.Fatalf("create order: %v", err)
	}
	if orderID == 0 {
		t.Fatalf("expected non-zero order id")
	}

	order, err := storage.GetOrderByID(context.Background(), orderID)
	if err != nil {
		t.Fatalf("get order: %v", err)
	}

	if order.UserID != 10 {
		t.Fatalf("user_id mismatch: got %d", order.UserID)
	}
	if len(order.Items) != 2 {
		t.Fatalf("items count mismatch: got %d", len(order.Items))
	}

	var total int64
	for _, it := range order.Items {
		total += it.Price * int64(it.Quantity)
	}
	if order.TotalAmount != total {
		t.Fatalf("total mismatch: got %d, want %d", order.TotalAmount, total)
	}
}
