package main

import (
	"log"
	"net/http"

	"github.com/ChernykhITMO/order-processing-platform/gateway/internal/handlers"
	"github.com/ChernykhITMO/order-processing-platform/protos/gen/ordersv1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	ordersAddr := "localhost:50051"

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
		Orders: ordersv1.NewOrdersServiceClient(conn),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/orders", gw.HandleOrders)
	mux.HandleFunc("/orders/", gw.HandleOrderById)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	log.Fatal(srv.ListenAndServe())

}
