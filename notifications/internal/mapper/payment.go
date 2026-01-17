package mapper

import (
	"github.com/ChernykhITMO/order-processing-platform/notifications/internal/domain/events"
	"github.com/ChernykhITMO/order-processing-platform/notifications/internal/dto"
)

func MapToDomainSave(input dto.SaveInput) events.Payment {
	var payment events.Payment

	payment.UserID = events.ID(input.UserID)
	payment.OrderID = events.ID(input.OrderID)
	payment.OrderStatus = events.Status(input.Status)

	return payment
}

func MapToDTOGet(payment events.Payment) dto.GetOutput {
	var output dto.GetOutput
	output.UserID = payment.UserID
	output.OrderID = payment.OrderID
	output.OrderStatus = payment.OrderStatus

	return output
}
