package grpc

import (
	"context"
	"fmt"

	"github.com/ChernykhITMO/order-processing-platform/orders/internal/dto"
	"github.com/ChernykhITMO/order-processing-platform/orders/internal/mapper"
	"github.com/ChernykhITMO/order-processing-platform/orders/internal/services"
	ordersv1 "github.com/ChernykhITMO/order-processing-proto/gen/go/opp/orders/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type gRPCServer struct {
	ordersv1.UnimplementedOrdersServiceServer
	usecase *services.Order
}

func NewServer(uc *services.Order) *gRPCServer {
	return &gRPCServer{
		usecase: uc,
	}
}

func (g *gRPCServer) CreateOrder(ctx context.Context, req *ordersv1.CreateOrderRequest) (*ordersv1.CreateOrderResponse, error) {
	const op = "server.CreateOrder"
	var input dto.CreateOrderInput

	items, err := mapper.MapToCreateItems(req.Items)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	input.UserID = req.UserId
	input.Items = items
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	output, err := g.usecase.CreateOrder(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &ordersv1.CreateOrderResponse{OrderId: output.ID}, nil
}

func (g *gRPCServer) GetOrder(ctx context.Context, req *ordersv1.GetOrderRequest) (*ordersv1.GetOrderResponse, error) {
	const op = "server.GetOrder"
	input := dto.GetOrderInput{ID: req.OrderId}
	output, err := g.usecase.GetOrder(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &ordersv1.GetOrderResponse{
		Order: mapper.MapToProto(output.Order),
	}, nil

}
func (g *gRPCServer) ListOrders(ctx context.Context, req *ordersv1.ListOrdersRequest) (*ordersv1.ListOrdersResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method ListOrders not implemented")
}
func (g *gRPCServer) CancelOrder(ctx context.Context, req *ordersv1.CancelOrderRequest) (*ordersv1.CancelOrderResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method CancelOrder not implemented")
}
