package api

import (
	"context"
	"fmt"

	"github.com/ChernykhITMO/order-processing-platform/orders/internal/dto"
	"github.com/ChernykhITMO/order-processing-platform/orders/internal/mapper"
	"github.com/ChernykhITMO/order-processing-platform/orders/internal/services"
	ordersv1 "github.com/ChernykhITMO/order-processing-proto/gen/go/opp/orders/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type serverAPI struct {
	ordersv1.UnimplementedOrdersServiceServer
	order *services.Order
}

func Register(gRPC *grpc.Server, order *services.Order) {
	srvAPI := &serverAPI{order: order}
	ordersv1.RegisterOrdersServiceServer(gRPC, srvAPI)
}

func (s *serverAPI) CreateOrder(ctx context.Context, req *ordersv1.CreateOrderRequest) (*ordersv1.CreateOrderResponse, error) {
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

	output, err := s.order.CreateOrder(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &ordersv1.CreateOrderResponse{OrderId: output.ID}, nil
}

func (s *serverAPI) GetOrder(ctx context.Context, req *ordersv1.GetOrderRequest) (*ordersv1.GetOrderResponse, error) {
	const op = "server.GetOrder"
	input := dto.GetOrderInput{ID: req.OrderId}
	output, err := s.order.GetOrder(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &ordersv1.GetOrderResponse{
		Order: mapper.MapToProto(output.Order),
	}, nil

}
func (s *serverAPI) ListOrders(ctx context.Context, req *ordersv1.ListOrdersRequest) (*ordersv1.ListOrdersResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method ListOrders not implemented")
}
func (s *serverAPI) CancelOrder(ctx context.Context, req *ordersv1.CancelOrderRequest) (*ordersv1.CancelOrderResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method CancelOrder not implemented")
}
