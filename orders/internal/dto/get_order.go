package dto

import "github.com/ChernykhITMO/order-processing-platform/orders/internal/domain"

type GetOrderInput struct {
	ID int64
}

type GetOrderOutput struct {
	domain.Order
}
