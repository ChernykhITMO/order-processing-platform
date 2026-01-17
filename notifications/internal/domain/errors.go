package domain

import "errors"

var (
	ErrIsEmptyKey = errors.New("key is empty")

	ErrInvalidOrderID = errors.New("order id must be positive")
	ErrInvalidUserID  = errors.New("user id must be positive")

	ErrInvalidStatus = errors.New("status must be succeeded or failed")

	ErrNotFound = errors.New("object found")

	ErrInvalidProductID = errors.New("product id must be positive")
	ErrInvalidQuantity  = errors.New("quantity must be at least one")
	ErrInvalidPrice     = errors.New("price must be positive")
	ErrInvalidItems     = errors.New("items must not be empty")
	ErrOrderNotFound    = errors.New("order not found")
	ErrUnknownType      = errors.New("unknown type")
)
