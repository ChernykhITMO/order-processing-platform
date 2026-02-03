package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ChernykhITMO/order-processing-platform/gateway/internal/handlers"
	"github.com/ChernykhITMO/order-processing-platform/gateway/internal/metrics"
	"github.com/ChernykhITMO/order-processing-platform/gateway/internal/middleware"
	ordersv1 "github.com/ChernykhITMO/order-processing-proto/gen/go/opp/orders/v1"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
		log.Println("ORDERS_GRPC_ADDR is empty")
		os.Exit(1)
	}

	conn, err := grpc.NewClient(
		ordersAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

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

	metrics.Register()

	apiMux := http.NewServeMux()
	apiMux.Handle("/metrics", promhttp.Handler())
	apiMux.Handle("/orders", middleware.Instrument("gateway", "/orders", http.HandlerFunc(gw.HandleOrders)))
	apiMux.Handle("/orders/", middleware.Instrument("gateway", "/orders/{id}", http.HandlerFunc(gw.HandleOrderById)))
	apiMux.Handle("/swagger/", middleware.Instrument("gateway", "/swagger/*", httpSwagger.WrapHandler))

	apiSrv := &http.Server{
		Addr:              ":8080",
		Handler:           apiMux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      20 * time.Second,
		IdleTimeout:       time.Minute,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)

	go func() {
		if err := apiSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := apiSrv.Shutdown(shutdownCtx); err != nil {
			log.Println("http shutdown:", err)
		}

	case err := <-errCh:
		log.Fatal(err)
	}
}
