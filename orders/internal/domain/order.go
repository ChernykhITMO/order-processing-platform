package domain

import (
	"fmt"
	"time"
)

type ID int64

type Order struct {
	ID          ID
	UserID      ID
	Status      Status
	Items       []OrderItem
	TotalAmount Money
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewOrder(
	orderID int64, userID int64, status string,
	items []OrderItem, createdAt time.Time, updatedAt time.Time) (*Order, error) {
	const op = "domain.Order.New"

	o := &Order{
		ID:        ID(orderID),
		UserID:    ID(userID),
		Status:    Status(status),
		Items:     items,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	o.calculate()

	if err := o.validate(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return o, nil
}

func (o *Order) validate() error {
	const op = "domain.order.validate"

	if o.ID <= 0 {
		return fmt.Errorf("%s: %w", op, ErrInvalidOrderID)
	}

	if o.UserID <= 0 {
		return fmt.Errorf("%s: %w", op, ErrInvalidUserID)
	}

	return nil
}

func (o *Order) calculate() {
	o.TotalAmount = 0
	for _, it := range o.Items {
		o.TotalAmount += it.Price * Money(it.Quantity)
	}
}
