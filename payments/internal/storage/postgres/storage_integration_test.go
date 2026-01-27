//go:build integration
// +build integration

package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/ChernykhITMO/order-processing-platform/payments/internal/domain"
	"github.com/ChernykhITMO/order-processing-platform/payments/internal/domain/events"
	"github.com/ChernykhITMO/order-processing-platform/payments/internal/ports"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

func loadPaymentsEnv(t *testing.T) {
	_ = godotenv.Load("../../../.env")
	_ = godotenv.Load("../../../payments/.env")
}

func getPaymentsDSN(t *testing.T) string {
	loadPaymentsEnv(t)
	dsn := os.Getenv("PAYMENTS_PG_DSN_TEST")
	if dsn == "" {
		t.Skip("PAYMENTS_PG_DSN_TEST is not set")
	}
	return dsn
}

func cleanupPaymentsTables(t *testing.T, db *sql.DB) {
	const query = `TRUNCATE TABLE payments, events, processed_events RESTART IDENTITY CASCADE`
	if _, err := db.Exec(query); err != nil {
		t.Fatalf("db exec: %v", err)
	}
}

func TestPaymentsStorage_RunInTxAndOutbox_Integration(t *testing.T) {
	dsn := getPaymentsDSN(t)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	cleanupPaymentsTables(t, db)
	defer cleanupPaymentsTables(t, db)

	storage, err := New(dsn)
	if err != nil {
		t.Fatalf("new storage: %v", err)
	}
	defer func() {
		_ = storage.Close()
	}()

	ctx := context.Background()

	const (
		orderID     int64 = 200
		userID      int64 = 10
		totalAmount int64 = 990
		eventID     int64 = 123
		eventType         = "event-status"
	)

	if err := storage.RunInTx(ctx, func(tx ports.StorageTx) error {
		ok, err := tx.TryMarkProcessed(ctx, eventID)
		if err != nil {
			return err
		}
		if !ok {
			t.Fatalf("expected event to be marked processed")
		}

		if err := tx.UpsertPayment(ctx, orderID, userID, totalAmount, domain.StatusPaymentPending); err != nil {
			return err
		}
		if err := tx.UpdatePaymentStatus(ctx, orderID, domain.StatusSucceeded); err != nil {
			return err
		}

		payload, err := json.Marshal(events.PaymentStatus{
			OrderID:     orderID,
			UserID:      userID,
			OrderStatus: domain.StatusSucceeded,
		})
		if err != nil {
			return err
		}

		return tx.SaveEvent(ctx, eventType, payload, orderID)
	}); err != nil {
		t.Fatalf("run in tx: %v", err)
	}

	if err := storage.RunInTx(ctx, func(tx ports.StorageTx) error {
		ok, err := tx.TryMarkProcessed(ctx, eventID)
		if err != nil {
			return err
		}
		if ok {
			t.Fatalf("expected event to be already processed")
		}
		return nil
	}); err != nil {
		t.Fatalf("run in tx: %v", err)
	}

	var status string
	if err := db.QueryRow(`SELECT status FROM payments WHERE order_id = $1`, orderID).Scan(&status); err != nil {
		t.Fatalf("select payment: %v", err)
	}
	if status != domain.StatusSucceeded {
		t.Fatalf("payment status: got %s, want %s", status, domain.StatusSucceeded)
	}

	var processedCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM processed_events WHERE event_id = $1`, eventID).Scan(&processedCount); err != nil {
		t.Fatalf("select processed_events: %v", err)
	}
	if processedCount != 1 {
		t.Fatalf("processed_events count: got %d, want %d", processedCount, 1)
	}

	evt, err := storage.GetNewEvent(ctx)
	if err != nil {
		t.Fatalf("get new event: %v", err)
	}
	if evt.EventID == 0 {
		t.Fatalf("expected event id to be set")
	}
	if evt.OrderID != orderID {
		t.Fatalf("event order id: got %d, want %d", evt.OrderID, orderID)
	}

	if err := storage.MarkSent(ctx, evt.EventID); err != nil {
		t.Fatalf("mark sent: %v", err)
	}

	evt, err = storage.GetNewEvent(ctx)
	if err != nil {
		t.Fatalf("get new event: %v", err)
	}
	if evt.EventID != 0 {
		t.Fatalf("expected no new events")
	}
}

func TestPaymentsStorage_LockedEvents_Integration(t *testing.T) {
	dsn := getPaymentsDSN(t)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	cleanupPaymentsTables(t, db)
	defer cleanupPaymentsTables(t, db)

	storage, err := New(dsn)
	if err != nil {
		t.Fatalf("new storage: %v", err)
	}
	defer func() {
		_ = storage.Close()
	}()

	payload, err := json.Marshal(events.PaymentStatus{OrderID: 1, UserID: 2, OrderStatus: domain.StatusSucceeded})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	if _, err := db.Exec(`INSERT INTO events (event_type, payload, aggregate_id, created_at) VALUES ($1, $2, $3, $4)`,
		"event-status", payload, 1, time.Now()); err != nil {
		t.Fatalf("insert event: %v", err)
	}

	if _, err := db.Exec(`UPDATE events SET locked_at = now() - interval '2 minutes' WHERE id = 1`); err != nil {
		t.Fatalf("update locked_at: %v", err)
	}

	evt, err := storage.GetNewEvent(context.Background())
	if err != nil {
		t.Fatalf("get new event: %v", err)
	}
	if evt.EventID == 0 {
		t.Fatalf("expected stale locked event to be picked")
	}
}
