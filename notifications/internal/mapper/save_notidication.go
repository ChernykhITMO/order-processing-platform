package mapper

import (
	"strconv"

	"github.com/ChernykhITMO/order-processing-platform/notifications/internal/dto"
)

/*type SaveInput struct {
	Key     string
	OrderID int64
	UserID  int64
	Status  string
}*/

func MapToInput(payment dto.Payment) dto.SaveInput {
	var output dto.SaveInput

	output.Key = strconv.FormatInt(output.OrderID, 10)
	output.OrderID = payment.OrderID
	output.UserID = payment.UserID
	output.Status = payment.OrderStatus

	return output
}
