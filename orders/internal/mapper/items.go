package mapper

import (
	"fmt"

	"github.com/ChernykhITMO/order-processing-platform/orders/internal/domain"
	"github.com/ChernykhITMO/order-processing-platform/orders/internal/dto"
	ordersv1 "github.com/ChernykhITMO/order-processing-proto/gen/go/opp/orders/v1"
)

func MapCreateItems(items []dto.CreateOrderItem) ([]domain.OrderItem, error) {
	const op = "mapper.CreateItems"
	res := make([]domain.OrderItem, 0, len(items))

	for _, it := range items {
		o, err := domain.NewOrderItem(it.ProductID, it.Quantity, it.Price)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		res = append(res, o)
	}

	return res, nil
}

func MapProtoItems(items []*ordersv1.OrderItem) ([]domain.OrderItem, error) {
	const op = "mapper.ProtoItems"

	res := make([]domain.OrderItem, 0, len(items))

	for _, it := range items {
		var price int64
		if it.GetPrice() != nil {
			price = it.GetPrice().GetMoney()
		}
		o, err := domain.NewOrderItem(it.ProductId, it.Quantity, price)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		res = append(res, o)
	}
	return res, nil
}
