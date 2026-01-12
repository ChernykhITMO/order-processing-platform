package grpc

import (
	"context"
	"fmt"

	"github.com/ChernykhITMO/order-processing-platform/orders/internal/mapper"
	"github.com/ChernykhITMO/order-processing-platform/orders/internal/usecase"
	ordersv1 "github.com/ChernykhITMO/order-processing-proto/gen/go/opp/orders/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type gRPCServer struct {
	ordersv1.UnimplementedOrdersServiceServer
	usecase *usecase.Order
}

func NewServer(uc *usecase.Order) *gRPCServer {
	return &gRPCServer{
		usecase: uc,
	}
}

func (g *gRPCServer) CreateOrder(ctx context.Context, req *ordersv1.CreateOrderRequest) (*ordersv1.CreateOrderResponse, error) {
	const op = "server.CreateOrder"
	items, err := mapper.MapProtoItems(req.Items)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	orderID, err := g.usecase.CreateOrder(ctx, req.UserId, items)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &ordersv1.CreateOrderResponse{OrderId: orderID}, nil
}

func (g *gRPCServer) GetOrder(ctx context.Context, req *ordersv1.GetOrderRequest) (*ordersv1.GetOrderResponse, error) {
	const op = "server.GetOrder"

	order, err := g.usecase.GetOrder(ctx, req.OrderId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &ordersv1.GetOrderResponse{
		Order: mapper.MapToProto(*order),
	}, nil

}
func (g *gRPCServer) ListOrders(ctx context.Context, req *ordersv1.ListOrdersRequest) (*ordersv1.ListOrdersResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method ListOrders not implemented")
}
func (g *gRPCServer) CancelOrder(ctx context.Context, req *ordersv1.CancelOrderRequest) (*ordersv1.CancelOrderResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method CancelOrder not implemented")
}
