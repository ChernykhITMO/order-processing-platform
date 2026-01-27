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
	"github.com/joho/godotenv"
)

func getDSN(t *testing.T) string {
	_ = godotenv.Load("../../../.env")
	_ = godotenv.Load("../../../orders/.env")
	dsn := os.Getenv("ORDERS_PG_DSN_TEST")
	if dsn == "" {
		t.Skip("ORDERS_PG_DSN_TEST is not set")
	}
	return dsn
}

func cleanupTables(t *testing.T, db *sql.DB) {
	const query = `
	TRUNCATE TABLE events, order_items, orders 
    RESTART IDENTITY CASCADE
    `

	if _, err := db.Exec(query); err != nil {
		t.Fatalf("db exec: %v", err)
	}
}

func TestCreate_Integration(t *testing.T) {
	dsn := getDSN(t)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	cleanupTables(t, db)
	defer cleanupTables(t, db)

	storage, err := New(dsn)
	if err != nil {
		t.Fatalf("new storage: %v", err)
	}

	ctx := context.Background()

	// arrange
	var userID int64 = 1
	items := []domain.OrderItem{
		domain.OrderItem{ProductID: 1, Price: 100, Quantity: 1},
		domain.OrderItem{ProductID: 2, Price: 200, Quantity: 1},
		domain.OrderItem{ProductID: 3, Price: 300, Quantity: 1},
	}

	totalAmount := 600

	// act
	orderID, err := storage.CreateOrder(ctx, userID, items)

	if err != nil {
		t.Fatal(err)
	}

	if orderID <= 0 {
		t.Fatal("orderID must be postive")
	}

	order, err := storage.GetOrderByID(ctx, orderID)
	if err != nil {
		t.Fatal(err)
	}

	if int64(order.UserID) != userID {
		t.Fatalf("order's userID not equal userID")
	}

	if len(order.Items) != len(items) {
		t.Fatalf("len order's items not equal test's items")
	}

	if int(order.TotalAmount) != totalAmount {
		t.Fatalf("total amout not equal db total amount")
	}
}

func TestOutbox_Integration(t *testing.T) {
	dsn := getDSN(t)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	cleanupTables(t, db)
	defer cleanupTables(t, db)

	storage, err := New(dsn)
	if err != nil {
		t.Fatalf("new storage: %v", err)
	}

	ctx := context.Background()

	items := []domain.OrderItem{
		{ProductID: 1, Price: 100, Quantity: 1},
	}

	orderID, err := storage.CreateOrder(ctx, 1, items)
	if err != nil {
		t.Fatal(err)
	}

	event, eventID, err := storage.GetNewEvent(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if eventID <= 0 {
		t.Fatalf("eventID must be positive")
	}
	if int64(event.OrderID) != orderID {
		t.Fatalf("event order id mismatch")
	}

	if err := storage.MarkSent(ctx, eventID); err != nil {
		t.Fatal(err)
	}

	event, eventID, err = storage.GetNewEvent(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if eventID != 0 {
		t.Fatalf("expected no new events")
	}
	if event.OrderID != 0 {
		t.Fatalf("expected empty event")
	}
}

func TestMultipleOrders_Integration(t *testing.T) {
	dsn := getDSN(t)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	cleanupTables(t, db)
	defer cleanupTables(t, db)

	storage, err := New(dsn)
	if err != nil {
		t.Fatalf("new storage: %v", err)
	}

	ctx := context.Background()

	items1 := []domain.OrderItem{
		{ProductID: 10, Price: 120, Quantity: 2},
	}
	items2 := []domain.OrderItem{
		{ProductID: 11, Price: 200, Quantity: 1},
		{ProductID: 12, Price: 50, Quantity: 3},
	}

	orderID1, err := storage.CreateOrder(ctx, 5, items1)
	if err != nil {
		t.Fatal(err)
	}
	orderID2, err := storage.CreateOrder(ctx, 6, items2)
	if err != nil {
		t.Fatal(err)
	}

	order1, err := storage.GetOrderByID(ctx, orderID1)
	if err != nil {
		t.Fatal(err)
	}
	if order1.Status != domain.StatusNew {
		t.Fatalf("order1 status: got %s, want %s", order1.Status, domain.StatusNew)
	}
	if len(order1.Items) != len(items1) {
		t.Fatalf("order1 items: got %d, want %d", len(order1.Items), len(items1))
	}
	if order1.TotalAmount != 240 {
		t.Fatalf("order1 total: got %d, want %d", order1.TotalAmount, 240)
	}

	order2, err := storage.GetOrderByID(ctx, orderID2)
	if err != nil {
		t.Fatal(err)
	}
	if order2.Status != domain.StatusNew {
		t.Fatalf("order2 status: got %s, want %s", order2.Status, domain.StatusNew)
	}
	if len(order2.Items) != len(items2) {
		t.Fatalf("order2 items: got %d, want %d", len(order2.Items), len(items2))
	}
	if order2.TotalAmount != 350 {
		t.Fatalf("order2 total: got %d, want %d", order2.TotalAmount, 350)
	}
}

func TestOutbox_Locking_Integration(t *testing.T) {
	dsn := getDSN(t)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	cleanupTables(t, db)
	defer cleanupTables(t, db)

	storage, err := New(dsn)
	if err != nil {
		t.Fatalf("new storage: %v", err)
	}

	ctx := context.Background()

	if _, err := storage.CreateOrder(ctx, 1, []domain.OrderItem{{ProductID: 1, Price: 10, Quantity: 1}}); err != nil {
		t.Fatal(err)
	}
	if _, err := storage.CreateOrder(ctx, 2, []domain.OrderItem{{ProductID: 2, Price: 20, Quantity: 1}}); err != nil {
		t.Fatal(err)
	}

	_, eventID1, err := storage.GetNewEvent(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if eventID1 == 0 {
		t.Fatalf("expected first event id")
	}

	_, eventID2, err := storage.GetNewEvent(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if eventID2 == 0 {
		t.Fatalf("expected second event id")
	}
	if eventID1 == eventID2 {
		t.Fatalf("expected different event ids")
	}
}
