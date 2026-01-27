//go:build integration
// +build integration

package kafka_consume

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ChernykhITMO/order-processing-platform/payments/internal/controller"
	"github.com/ChernykhITMO/order-processing-platform/payments/internal/domain"
	"github.com/ChernykhITMO/order-processing-platform/payments/internal/dto"
	kafkaproduce "github.com/ChernykhITMO/order-processing-platform/payments/internal/kafka_produce"
	"github.com/ChernykhITMO/order-processing-platform/payments/internal/services"
	"github.com/ChernykhITMO/order-processing-platform/payments/internal/storage/postgres"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

func loadPaymentsEnv(t *testing.T) {
	_ = godotenv.Load("../../../.env")
	_ = godotenv.Load("../../../payments/.env")
}

func parseCSV(value string) []string {
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		item := strings.TrimSpace(p)
		if item == "" {
			continue
		}
		out = append(out, item)
	}
	return out
}

func cleanupPaymentsTables(t *testing.T, db *sql.DB) {
	t.Helper()
	const query = `TRUNCATE TABLE payments, events, processed_events RESTART IDENTITY CASCADE`
	if _, err := db.Exec(query); err != nil {
		t.Fatalf("db exec: %v", err)
	}
}

func TestPaymentsKafkaConsumer_Integration(t *testing.T) {
	loadPaymentsEnv(t)

	dsn := os.Getenv("PAYMENTS_PG_DSN")
	if dsn == "" {
		t.Skip("PAYMENTS_PG_DSN is not set")
	}

	brokers := parseCSV(os.Getenv("KAFKA_TEST_BROKERS"))
	if len(brokers) == 0 {
		brokers = parseCSV(os.Getenv("KAFKA_BROKERS"))
	}
	if len(brokers) == 0 {
		t.Skip("KAFKA_TEST_BROKERS or KAFKA_BROKERS is not set")
	}

	topic := os.Getenv("KAFKA_TEST_TOPIC_ORDER")
	if topic == "" {
		topic = os.Getenv("KAFKA_TOPIC_ORDER")
	}
	if topic == "" {
		t.Skip("KAFKA_TEST_TOPIC_ORDER or KAFKA_TOPIC_ORDER is not set")
	}

	eventType := os.Getenv("KAFKA_EVENT_TYPE")
	if eventType == "" {
		eventType = "event-status"
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	cleanupPaymentsTables(t, db)
	defer cleanupPaymentsTables(t, db)

	storage, err := postgres.New(dsn)
	if err != nil {
		t.Fatalf("new storage: %v", err)
	}
	defer func() {
		_ = storage.Close()
	}()

	log := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
	svc := services.New(storage, log, eventType)
	ctrl := controller.NewController(svc, storage, nil, "", log)

	consumer, err := NewConsumer(ctrl, brokers, topic, "test-group-"+time.Now().Format("150405.000"), log)
	if err != nil {
		t.Fatalf("new consumer: %v", err)
	}
	defer func() {
		_ = consumer.Stop()
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go consumer.Start(ctx)

	producer, err := kafkaproduce.NewProducer(brokers)
	if err != nil {
		t.Fatalf("new producer: %v", err)
	}
	defer producer.Close()

	orderID := int64(200)
	event := dto.OrderCreated{
		EventID:     time.Now().UnixNano(),
		OrderID:     orderID,
		UserID:      10,
		TotalAmount: 999,
		CreatedAt:   time.Now(),
	}

	payload, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("marshal event: %v", err)
	}

	produceCtx, cancelProduce := context.WithTimeout(context.Background(), 5*time.Second)
	if err := producer.Produce(produceCtx, payload, topic); err != nil {
		cancelProduce()
		t.Fatalf("produce: %v", err)
	}
	cancelProduce()

	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		var status string
		err := db.QueryRow(`SELECT status FROM payments WHERE order_id = $1`, orderID).Scan(&status)
		if err == nil && status == domain.StatusSucceeded {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}

	var status string
	if err := db.QueryRow(`SELECT status FROM payments WHERE order_id = $1`, orderID).Scan(&status); err != nil {
		t.Fatalf("select status: %v", err)
	}
	if status != domain.StatusSucceeded {
		t.Fatalf("payment status: got %s, want %s", status, domain.StatusSucceeded)
	}

	var processedCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM processed_events WHERE event_id = $1`, event.EventID).Scan(&processedCount); err != nil {
		t.Fatalf("processed_events count: %v", err)
	}
	if processedCount != 1 {
		t.Fatalf("processed_events count: got %d, want %d", processedCount, 1)
	}

	var outboxCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM events WHERE aggregate_id = $1`, orderID).Scan(&outboxCount); err != nil {
		t.Fatalf("events count: %v", err)
	}
	if outboxCount == 0 {
		t.Fatalf("expected outbox event")
	}
}
