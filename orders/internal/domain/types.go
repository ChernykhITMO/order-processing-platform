package domain

import (
	"errors"
	"time"
)

var (
	ErrInvalidID    = errors.New("id must be positive")
	ErrInvalidItems = errors.New("items must not be empty")
)

type Status int32

const (
	StatusUnspecified      Status = 0
	StatusNew              Status = 1
	StatusPaymentPending   Status = 2
	StatusPaid             Status = 3
	StatusInventoryPending Status = 4
	StatusCompleted        Status = 5
	StatusCanceled         Status = 6
)

type OrderItem struct {
	ProductID int64
	Quantity  int32
	Price     Money
}

type Order struct {
	ID          int64
	UserID      int64
	Status      Status
	Items       []OrderItem
	TotalAmount Money
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
type Money struct {
	Units int64
	Nanos int32
}
