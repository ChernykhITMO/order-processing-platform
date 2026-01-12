package domain

import (
	"errors"
)

var (
	ErrInvalidOrderID   = errors.New("order id must be positive")
	ErrInvalidUserID    = errors.New("user id must be positive")
	ErrInvalidProductID = errors.New("product id must be positive")
	ErrInvalidQuantity  = errors.New("quantity must be at least one")
	ErrInvalidPrce      = errors.New("price must be positive")
	ErrInvalidItems     = errors.New("items must not be empty")
	ErrOrderNotFound    = errors.New("order not found")
)
