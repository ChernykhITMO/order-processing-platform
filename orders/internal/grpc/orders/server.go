package orders

import (
	"context"
	"math/rand"

	"github.com/ChernykhITMO/order-processing-platform/protos/gen/ordersv1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type gRPCServer struct {
	ordersv1.UnimplementedOrdersServiceServer
}

func (g *gRPCServer) CreateOrder(ctx context.Context, req *ordersv1.CreateOrderRequest) (*ordersv1.CreateOrderResponse, error) {
	return &ordersv1.CreateOrderResponse{OrderId: rand.Int63(), Order: nil}, nil
}
func (g *gRPCServer) GetOrder(ctx context.Context, req *ordersv1.GetOrderRequest) (*ordersv1.GetOrderResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method GetOrder not implemented")
}
func (g *gRPCServer) ListOrders(ctx context.Context, req *ordersv1.ListOrdersRequest) (*ordersv1.ListOrdersResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method ListOrders not implemented")
}
func (g *gRPCServer) CancelOrder(ctx context.Context, req *ordersv1.CancelOrderRequest) (*ordersv1.CancelOrderResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method CancelOrder not implemented")
}
