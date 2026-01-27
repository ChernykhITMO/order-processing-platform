//go:build integration
// +build integration

package kafka_consume

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ChernykhITMO/order-processing-platform/notifications/internal/config"
	"github.com/ChernykhITMO/order-processing-platform/notifications/internal/controller"
	"github.com/ChernykhITMO/order-processing-platform/notifications/internal/domain"
	"github.com/ChernykhITMO/order-processing-platform/notifications/internal/dto"
	"github.com/ChernykhITMO/order-processing-platform/notifications/internal/services"
	redis_storage "github.com/ChernykhITMO/order-processing-platform/notifications/storage/redis"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/joho/godotenv"
)

func loadNotificationsEnv(t *testing.T) {
	_ = godotenv.Load("../../../.env")
	_ = godotenv.Load("../../../notifications/.env")
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

func redisConfigFromEnv(t *testing.T, ttl time.Duration) config.Config {
	loadNotificationsEnv(t)

	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		t.Skip("REDIS_ADDR is not set")
	}

	db := 0
	if value := os.Getenv("REDIS_DB"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			t.Fatalf("REDIS_DB: %v", err)
		}
		db = parsed
	}

	maxRetries := 0
	if value := os.Getenv("REDIS_MAX_RETRIES"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			t.Fatalf("REDIS_MAX_RETRIES: %v", err)
		}
		maxRetries = parsed
	}

	dialTimeout := 2 * time.Second
	if value := os.Getenv("REDIS_DIAL_TIMEOUT"); value != "" {
		parsed, err := time.ParseDuration(value)
		if err != nil {
			t.Fatalf("REDIS_DIAL_TIMEOUT: %v", err)
		}
		dialTimeout = parsed
	}

	timeout := 2 * time.Second
	if value := os.Getenv("REDIS_TIMEOUT"); value != "" {
		parsed, err := time.ParseDuration(value)
		if err != nil {
			t.Fatalf("REDIS_TIMEOUT: %v", err)
		}
		timeout = parsed
	}

	return config.Config{
		Addr:        addr,
		Password:    os.Getenv("REDIS_PASSWORD"),
		User:        os.Getenv("REDIS_USER"),
		DB:          db,
		MaxRetries:  maxRetries,
		DialTimeout: dialTimeout,
		Timeout:     timeout,
		TTL:         ttl,
	}
}

func newKafkaProducer(t *testing.T, brokers []string) *kafka.Producer {
	t.Helper()

	cfg := &kafka.ConfigMap{"bootstrap.servers": strings.Join(brokers, ",")}
	producer, err := kafka.NewProducer(cfg)
	if err != nil {
		t.Fatalf("new producer: %v", err)
	}

	return producer
}

func produceKafkaMessage(t *testing.T, producer *kafka.Producer, topic string, payload []byte) {
	t.Helper()

	delivery := make(chan kafka.Event, 1)
	if err := producer.Produce(&kafka.Message{TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny}, Value: payload}, delivery); err != nil {
		t.Fatalf("produce: %v", err)
	}

	select {
	case ev := <-delivery:
		if msg, ok := ev.(*kafka.Message); ok && msg.TopicPartition.Error != nil {
			t.Fatalf("delivery error: %v", msg.TopicPartition.Error)
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("delivery timeout")
	}
}

func TestNotificationsKafkaConsumer_Integration(t *testing.T) {
	loadNotificationsEnv(t)

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

	cfg := redisConfigFromEnv(t, time.Minute)
	storage := redis_storage.New(cfg)
	defer func() {
		_ = storage.Close()
	}()

	log := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
	svc := services.New(storage, log)
	handler := controller.NewSender(svc, log)

	consumer, err := NewConsumer(brokers, handler, topic, "test-group-"+time.Now().Format("150405.000"), time.Second*6, time.Second*2, log)
	if err != nil {
		t.Fatalf("new consumer: %v", err)
	}
	defer func() {
		_ = consumer.Stop()
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = consumer.Start(ctx)
	}()

	producer := newKafkaProducer(t, brokers)
	defer producer.Close()

	payment := dto.Payment{
		OrderID:     3001,
		UserID:      4001,
		OrderStatus: domain.StatusSucceeded,
	}

	payload, err := json.Marshal(payment)
	if err != nil {
		t.Fatalf("marshal payment: %v", err)
	}

	produceKafkaMessage(t, producer, topic, payload)

	key := strconv.FormatInt(payment.OrderID, 10)
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		got, err := storage.GetNotification(context.Background(), key)
		if err == nil && got.OrderID == payment.OrderID {
			return
		}
		time.Sleep(200 * time.Millisecond)
	}

	if _, err := storage.GetNotification(context.Background(), key); err != nil {
		t.Fatalf("expected notification to be stored, got %v", err)
	}
}
