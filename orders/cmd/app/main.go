package main

import (
	"log"
	"log/slog"
	"net"
	"os"

	gw "github.com/ChernykhITMO/order-processing-platform/orders/internal/controller/grpc"
	"github.com/ChernykhITMO/order-processing-platform/orders/internal/storage/postgres"
	"github.com/ChernykhITMO/order-processing-platform/orders/internal/usecase"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("env not loaded")
	}

	addr := envOr("ORDERS_GRPC_ADDR", ":50051")
	dsn := os.Getenv("ORDERS_PG_DSN")
	if dsn == "" {
		log.Fatal("ORDERS_PG_DSN is required")
	}

	storage, err := postgres.New(dsn)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	uc := usecase.New(logger, storage)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	srv := grpc.NewServer()
	ordersv1.RegisterOrdersServiceServer(srv, gw.NewServer(uc))

	log.Printf("orders gRPC listening on %s", addr)
	log.Fatal(srv.Serve(lis))
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
