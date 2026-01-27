package mapper

import (
	"github.com/ChernykhITMO/order-processing-platform/orders/internal/domain"
	ordersv1 "github.com/ChernykhITMO/order-processing-proto/gen/go/opp/orders/v1"
)

func MapStatusToProto(status domain.Status) ordersv1.OrderStatus {
	switch status {
	case domain.StatusNew:
		return ordersv1.OrderStatus_new
	default:
	}
	return ordersv1.OrderStatus_unspecified
}
