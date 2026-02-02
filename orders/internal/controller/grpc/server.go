package api

import (
	"context"

	"github.com/ChernykhITMO/order-processing-platform/orders/internal/dto"
	"github.com/ChernykhITMO/order-processing-platform/orders/internal/mapper"
	"github.com/ChernykhITMO/order-processing-platform/orders/internal/services"
	ordersv1 "github.com/ChernykhITMO/order-processing-proto/gen/go/opp/orders/v1"
	"google.golang.org/grpc"
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

func toStatus(err error) error {
	code, msg := mapper.MapDomainError(err)
	return status.Error(code, msg)
}

func (s *serverAPI) CreateOrder(
	ctx context.Context,
	req *ordersv1.CreateOrderRequest) (*ordersv1.CreateOrderResponse, error) {
	var input dto.CreateOrderInput

	items, err := mapper.MapToCreateItems(req.Items)
	if err != nil {
		return nil, toStatus(err)
	}

	input.UserID = req.UserId
	input.Items = items

	output, err := s.order.CreateOrder(ctx, input)
	if err != nil {
		return nil, toStatus(err)
	}

	return &ordersv1.CreateOrderResponse{OrderId: output.ID}, nil
}

func (s *serverAPI) GetOrder(ctx context.Context, req *ordersv1.GetOrderRequest) (*ordersv1.GetOrderResponse, error) {
	input := dto.GetOrderInput{ID: req.OrderId}
	output, err := s.order.GetOrder(ctx, input)
	if err != nil {
		return nil, toStatus(err)
	}

	return &ordersv1.GetOrderResponse{
		Order: mapper.MapToProto(output.Order),
	}, nil
}
