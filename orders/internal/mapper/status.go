package mapper

import (
	"github.com/ChernykhITMO/order-processing-platform/orders/internal/domain"
	ordersv1 "github.com/ChernykhITMO/order-processing-proto/gen/go/opp/orders/v1"
)

func MapStatusToProto(status domain.Status) ordersv1.OrderStatus {
	switch status {
	case "StatusNew":
		return ordersv1.OrderStatus_NEW
	case "StatusPaymentPending":
		return ordersv1.OrderStatus_PAYMENT_PENDING
	case "StatusPaid":
		return ordersv1.OrderStatus_PAID
	case "StatusInventoryPending":
		return ordersv1.OrderStatus_INVENTORY_PENDING
	case "StatusCompleted":
		return ordersv1.OrderStatus_COMPLETED
	case "StatusCanceled":
		return ordersv1.OrderStatus_CANCELED
	}

	return ordersv1.OrderStatus_UNSPECIFIED
}
