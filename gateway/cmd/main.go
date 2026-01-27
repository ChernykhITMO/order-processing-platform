package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ChernykhITMO/order-processing-platform/gateway/internal/handlers"
	ordersv1 "github.com/ChernykhITMO/order-processing-proto/gen/go/opp/orders/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	_ "github.com/ChernykhITMO/order-processing-platform/gateway/docs"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

//	@title			Order Processing Platform API Gateway
//	@version		1.0
//	@description	API Gateway for Order Processing Platform

// @host		localhost:8080
// @BasePath	/
func main() {
	ordersAddr := os.Getenv("ORDERS_GRPC_ADDR")
	if ordersAddr == "" {
		panic("orders address is empty")
	}

	conn, err := grpc.NewClient(
		ordersAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Println(err)
		}
	}()

	gw := &handlers.Gateway{
		Orders:         ordersv1.NewOrdersServiceClient(conn),
		RequestTimeout: 2 * time.Second,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/orders", gw.HandleOrders)
	mux.HandleFunc("/orders/", gw.HandleOrderById)
	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Println("shutdown error:", err)
		}
	case err := <-errCh:
		log.Fatal(err)
	}
}
