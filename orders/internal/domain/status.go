package domain

type Status string

const (
	StatusUnspecified      string = "StatusUnspecified"
	StatusNew              string = "StatusNew"
	StatusPaymentPending   string = "StatusPaymentPending"
	StatusPaid             string = "StatusPaid"
	StatusInventoryPending string = "StatusInventoryPending"
	StatusCompleted        string = "StatusCompleted"
	StatusCanceled         string = "StatusCanceled"
)

func StatusFromDB(status string) string {
	switch status {
	case "NEW":
		return StatusNew
	case "PAYMENT_PENDING":
		return StatusPaymentPending
	case "PAID":
		return StatusPaid
	case "INVENTORY_PENDING":
		return StatusInventoryPending
	case "COMPLETED":
		return StatusCompleted
	case "CANCELED":
		return StatusCanceled
	case "UNSPECIFIED":
		return StatusUnspecified
	}

	return StatusUnspecified
}

func StatusToDB(status string) string {
	switch status {
	case StatusNew:
		return "NEW"
	case StatusPaymentPending:
		return "PAYMENT_PENDING"
	case StatusPaid:
		return "PAID"
	case StatusInventoryPending:
		return "INVENTORY_PENDING"
	case StatusCompleted:
		return "COMPLETED"
	case StatusCanceled:
		return "CANCELED"
	case StatusUnspecified:
		return "UNSPECIFIED"
	}

	return "UNSPECIFIED"
}
