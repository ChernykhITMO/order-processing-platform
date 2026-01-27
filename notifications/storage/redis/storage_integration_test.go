//go:build integration
// +build integration

package redis_storage

import (
	"context"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/ChernykhITMO/order-processing-platform/notifications/internal/config"
	"github.com/ChernykhITMO/order-processing-platform/notifications/internal/domain"
	"github.com/ChernykhITMO/order-processing-platform/notifications/internal/domain/events"
	"github.com/joho/godotenv"
)

func loadNotificationsEnv(t *testing.T) {
	_ = godotenv.Load("../../../.env")
	_ = godotenv.Load("../../../notifications/.env")
}

func redisConfigFromEnv(t *testing.T, ttl time.Duration) config.Config {
	loadNotificationsEnv(t)

	addr := os.Getenv("REDIS_ADDR_TEST")
	if addr == "" {
		t.Skip("REDIS_ADDR_TEST is not set")
	}

	db := 0
	if value := os.Getenv("REDIS_DB"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			t.Fatalf("REDIS_DB: %v", err)
		}
		db = parsed
	}

	return config.Config{
		Addr: addr,
		DB:   db,
		TTL:  ttl,
	}
}

func TestRedisStorage_SaveGet_Integration(t *testing.T) {
	cfg := redisConfigFromEnv(t, time.Minute)
	st := New(cfg)
	defer func() {
		_ = st.Close()
	}()

	ctx := context.Background()
	if err := st.Ping(ctx); err != nil {
		t.Fatalf("redis ping: %v", err)
	}

	key := "test-payment-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	value := events.Payment{OrderID: 10, UserID: 20, OrderStatus: domain.StatusSucceeded}

	if err := st.SaveNotification(ctx, key, value); err != nil {
		t.Fatalf("save notification: %v", err)
	}

	got, err := st.GetNotification(ctx, key)
	if err != nil {
		t.Fatalf("get notification: %v", err)
	}

	if got.OrderID != value.OrderID || got.UserID != value.UserID || got.OrderStatus != value.OrderStatus {
		t.Fatalf("unexpected value: got %+v, want %+v", got, value)
	}
}

func TestRedisStorage_TTL_Integration(t *testing.T) {
	cfg := redisConfigFromEnv(t, 50*time.Millisecond)
	st := New(cfg)
	defer func() {
		_ = st.Close()
	}()

	ctx := context.Background()
	key := "test-ttl-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	value := events.Payment{OrderID: 11, UserID: 21, OrderStatus: domain.StatusSucceeded}

	if err := st.SaveNotification(ctx, key, value); err != nil {
		t.Fatalf("save notification: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	if _, err := st.GetNotification(ctx, key); err == nil || err != domain.ErrNotFound {
		t.Fatalf("expected not found after ttl, got %v", err)
	}
}

func TestRedisStorage_TTLDisabled_Integration(t *testing.T) {
	cfg := redisConfigFromEnv(t, 0)
	st := New(cfg)
	defer func() {
		_ = st.Close()
	}()

	ctx := context.Background()
	key := "test-no-ttl-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	value := events.Payment{OrderID: 12, UserID: 22, OrderStatus: domain.StatusFailed}

	if err := st.SaveNotification(ctx, key, value); err != nil {
		t.Fatalf("save notification: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	if _, err := st.GetNotification(ctx, key); err != nil {
		t.Fatalf("expected value to persist, got %v", err)
	}
}

func TestRedisStorage_NotFound_Integration(t *testing.T) {
	cfg := redisConfigFromEnv(t, time.Minute)
	st := New(cfg)
	defer func() {
		_ = st.Close()
	}()

	ctx := context.Background()
	key := "missing-" + strconv.FormatInt(time.Now().UnixNano(), 10)

	if _, err := st.GetNotification(ctx, key); err == nil || err != domain.ErrNotFound {
		t.Fatalf("expected not found error, got %v", err)
	}
}
