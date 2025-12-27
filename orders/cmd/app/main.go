package main

import (
	"log"
	"net"

	"github.com/ChernykhITMO/order-processing-platform/orders/internal/grpc/orders"
	"github.com/ChernykhITMO/order-processing-platform/protos/gen/ordersv1"
	"google.golang.org/grpc"
)

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal(err)
	}

	gRPCServer := grpc.NewServer()
	ordersv1.RegisterOrdersServiceServer(gRPCServer, &orders.Orders{})
	log.Println("orders gRPC listening on :50051")
	log.Fatal(gRPCServer.Serve(lis))
}
