package domain

import (
	"errors"
)

var (
	ErrInvalidID    = errors.New("id must be positive")
	ErrInvalidItems = errors.New("items must not be empty")
)

type OrderItem struct {
	ProductID int64
	Quantity  int32
	Price     int64
}
