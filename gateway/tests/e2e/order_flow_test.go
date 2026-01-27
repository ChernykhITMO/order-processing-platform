package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

type createOrderRequest struct {
	UserID int64 `json:"user_id"`
	Items  []struct {
		ProductID int64 `json:"product_id"`
		Quantity  int32 `json:"quantity"`
		Price     int64 `json:"price"`
	} `json:"items"`
}

type createOrderResponse struct {
	OrderID int64 `json:"order_id"`
}

type paymentNotification struct {
	OrderID     int64  `json:"order_id"`
	UserID      int64  `json:"user_id"`
	OrderStatus string `json:"order_status"`
}

func TestE2E_OrderFlow(t *testing.T) {
	gatewayURL := os.Getenv("E2E_GATEWAY_URL")
	if gatewayURL == "" {
		gatewayURL = "http://localhost:8080"
	}

	redisAddr := os.Getenv("E2E_REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	timeout := 20 * time.Second
	if value := os.Getenv("E2E_TIMEOUT"); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			timeout = parsed
		}
	}

	reqBody := createOrderRequest{
		UserID: 1,
		Items: []struct {
			ProductID int64 `json:"product_id"`
			Quantity  int32 `json:"quantity"`
			Price     int64 `json:"price"`
		}{
			{ProductID: 10, Quantity: 2, Price: 100},
		},
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	resp, err := http.Post(gatewayURL+"/orders", "application/json", bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("create order: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}

	var createResp createOrderResponse
	if err := json.NewDecoder(resp.Body).Decode(&createResp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if createResp.OrderID == 0 {
		t.Fatalf("expected order id")
	}

	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	defer func() {
		_ = rdb.Close()
	}()

	key := strconv.FormatInt(createResp.OrderID, 10)
	_ = rdb.Del(context.Background(), key).Err()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			t.Fatalf("timeout waiting for notification")
		default:
		}

		val, err := rdb.Get(ctx, key).Result()
		if err == nil && val != "" {
			var payment paymentNotification
			if err := json.Unmarshal([]byte(val), &payment); err != nil {
				t.Fatalf("unmarshal notification: %v", err)
			}
			if payment.OrderID != createResp.OrderID {
				t.Fatalf("order id mismatch: got %d, want %d", payment.OrderID, createResp.OrderID)
			}
			if payment.UserID != reqBody.UserID {
				t.Fatalf("user id mismatch: got %d, want %d", payment.UserID, reqBody.UserID)
			}
			if payment.OrderStatus == "" {
				t.Fatalf("expected order_status to be set")
			}
			return
		}

		time.Sleep(300 * time.Millisecond)
	}
}
