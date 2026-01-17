package dto

import "github.com/ChernykhITMO/order-processing-platform/notifications/internal/domain/events"

type GetInput struct {
	Key string
}

type GetOutput struct {
	events.Payment
}
