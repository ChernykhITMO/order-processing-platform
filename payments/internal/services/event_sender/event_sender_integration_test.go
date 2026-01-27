//go:build integration
// +build integration

package event_sender

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

	"github.com/ChernykhITMO/order-processing-platform/payments/internal/domain"
	"github.com/ChernykhITMO/order-processing-platform/payments/internal/domain/events"
	kafkaproduce "github.com/ChernykhITMO/order-processing-platform/payments/internal/kafka_produce"
	"github.com/ChernykhITMO/order-processing-platform/payments/internal/storage/postgres"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

func loadPaymentsEnv(t *testing.T) {
	_ = godotenv.Load("../../../../.env")
	_ = godotenv.Load("../../../../payments/.env")
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
	const query = `TRUNCATE TABLE payments, events, processed_events RESTART IDENTITY CASCADE`
	if _, err := db.Exec(query); err != nil {
		t.Fatalf("db exec: %v", err)
	}
}

func TestPaymentsOutboxSender_Integration(t *testing.T) {
	loadPaymentsEnv(t)

	dsn := os.Getenv("PAYMENTS_PG_DSN_TEST")
	if dsn == "" {
		t.Skip("PAYMENTS_PG_DSN_TEST is not set")
	}

	brokers := parseCSV(os.Getenv("KAFKA_TEST_BROKERS"))
	if len(brokers) == 0 {
		brokers = parseCSV(os.Getenv("KAFKA_BROKERS"))
	}
	if len(brokers) == 0 {
		t.Skip("KAFKA_TEST_BROKERS or KAFKA_BROKERS is not set")
	}

	topic := os.Getenv("KAFKA_TEST_TOPIC_STATUS")
	if topic == "" {
		topic = os.Getenv("KAFKA_TOPIC_STATUS")
	}
	if topic == "" {
		t.Skip("KAFKA_TEST_TOPIC_STATUS or KAFKA_TOPIC_STATUS is not set")
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

	orderID := int64(5001)
	payload, err := json.Marshal(events.PaymentStatus{OrderID: orderID, UserID: 10, OrderStatus: domain.StatusSucceeded})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	var eventID int64
	if err := db.QueryRow(`INSERT INTO events (event_type, payload, aggregate_id, created_at) VALUES ($1, $2, $3, NOW()) RETURNING id`,
		"event-status", payload, orderID).Scan(&eventID); err != nil {
		t.Fatalf("insert event: %v", err)
	}

	storage, err := postgres.New(dsn)
	if err != nil {
		t.Fatalf("new storage: %v", err)
	}
	defer func() {
		_ = storage.Close()
	}()

	producer, err := kafkaproduce.NewProducer(brokers)
	if err != nil {
		t.Fatalf("new producer: %v", err)
	}
	defer producer.Close()

	log := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
	sender := New(storage, producer, log, topic)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = sender.StartProcessEvents(ctx, 50*time.Millisecond)
	}()

	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  strings.Join(brokers, ","),
		"group.id":           "test-group-" + time.Now().Format("150405.000"),
		"auto.offset.reset":  "earliest",
		"session.timeout.ms": 6000,
	})
	if err != nil {
		t.Fatalf("new consumer: %v", err)
	}
	defer func() {
		_ = consumer.Close()
	}()

	if err := consumer.Subscribe(topic, nil); err != nil {
		t.Fatalf("subscribe: %v", err)
	}

	received := false
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		msg, err := consumer.ReadMessage(500 * time.Millisecond)
		if err != nil {
			if kerr, ok := err.(kafka.Error); ok && kerr.Code() == kafka.ErrTimedOut {
				continue
			}
			continue
		}

		if msg == nil {
			continue
		}

		var got events.PaymentStatus
		if err := json.Unmarshal(msg.Value, &got); err != nil {
			continue
		}
		if got.OrderID != orderID {
			continue
		}

		received = true
		cancel()
		break
	}

	if !received {
		t.Fatalf("expected message in kafka, got none")
	}

	var sentCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM events WHERE id = $1 AND sent_at IS NOT NULL`, eventID).Scan(&sentCount); err != nil {
		t.Fatalf("select sent: %v", err)
	}
	if sentCount != 1 {
		t.Fatalf("expected event to be marked sent")
	}
}
