package mapper

import (
	"strconv"

	"github.com/ChernykhITMO/order-processing-platform/notifications/internal/dto"
)

func MapToInput(payment dto.Payment) dto.SaveInput {
	var output dto.SaveInput

	output.Key = strconv.FormatInt(payment.OrderID, 10)
	output.OrderID = payment.OrderID
	output.UserID = payment.UserID
	output.Status = payment.OrderStatus

	return output
}
