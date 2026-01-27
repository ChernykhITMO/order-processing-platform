//go:build integration
// +build integration

package api

import (
	"context"
	"database/sql"
	"io"
	"log/slog"
	"net"
	"os"
	"testing"
	"time"

	orderssvc "github.com/ChernykhITMO/order-processing-platform/orders/internal/services"
	orderspg "github.com/ChernykhITMO/order-processing-platform/orders/internal/storage/postgres"
	ordersv1 "github.com/ChernykhITMO/order-processing-proto/gen/go/opp/orders/v1"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func loadOrdersEnv(t *testing.T) {
	_ = godotenv.Load("../../../.env")
	_ = godotenv.Load("../../../orders/.env")
}

func getOrdersDSN(t *testing.T) string {
	loadOrdersEnv(t)
	dsn := os.Getenv("ORDERS_PG_DSN_TEST")
	if dsn == "" {
		t.Skip("ORDERS_PG_DSN_TEST is not set")
	}
	return dsn
}

func cleanupOrdersTables(t *testing.T, db *sql.DB) {
	const query = `TRUNCATE TABLE events, order_items, orders RESTART IDENTITY CASCADE`
	if _, err := db.Exec(query); err != nil {
		t.Fatalf("db exec: %v", err)
	}
}

func startOrdersGRPC(t *testing.T, storage *orderspg.Storage) (ordersv1.OrdersServiceClient, func()) {
	t.Helper()

	log := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
	service := orderssvc.New(log, storage)

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	server := grpc.NewServer()
	Register(server, service)

	go func() {
		_ = server.Serve(lis)
	}()

	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		server.Stop()
		_ = lis.Close()
		t.Fatalf("dial: %v", err)
	}

	cleanup := func() {
		server.Stop()
		_ = conn.Close()
		_ = lis.Close()
	}

	return ordersv1.NewOrdersServiceClient(conn), cleanup
}

func TestOrdersGRPC_Integration(t *testing.T) {
	dsn := getOrdersDSN(t)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	cleanupOrdersTables(t, db)
	defer cleanupOrdersTables(t, db)

	storage, err := orderspg.New(dsn)
	if err != nil {
		t.Fatalf("new storage: %v", err)
	}
	defer func() {
		_ = storage.Close()
	}()

	client, cleanup := startOrdersGRPC(t, storage)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	createResp, err := client.CreateOrder(ctx, &ordersv1.CreateOrderRequest{
		UserId: 1,
		Items: []*ordersv1.OrderItem{{
			ProductId: 10,
			Quantity:  2,
			Price:     &ordersv1.Money{Money: 100},
		}},
	})
	if err != nil {
		t.Fatalf("create order: %v", err)
	}
	if createResp.OrderId == 0 {
		t.Fatalf("expected order id")
	}

	getResp, err := client.GetOrder(ctx, &ordersv1.GetOrderRequest{OrderId: createResp.OrderId})
	if err != nil {
		t.Fatalf("get order: %v", err)
	}
	if getResp.Order == nil {
		t.Fatalf("expected order")
	}
	if getResp.Order.UserId != 1 {
		t.Fatalf("user id: got %d", getResp.Order.UserId)
	}
	if len(getResp.Order.Items) != 1 || getResp.Order.Items[0].ProductId != 10 {
		t.Fatalf("items mismatch")
	}
}
