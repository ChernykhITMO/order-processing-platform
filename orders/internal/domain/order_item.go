package domain

import "fmt"

type Money int64
type OrderItem struct {
	ProductID ID
	Quantity  int32
	Price     Money
}

func NewOrderItem(productID int64, quantity int32, price int64) (OrderItem, error) {
	const op = "domain.OrderItem.New"

	o := OrderItem{
		ProductID: ID(productID),
		Quantity:  quantity,
		Price:     Money(price),
	}

	if err := o.validate(); err != nil {
		return OrderItem{}, fmt.Errorf("%s: %w", op, err)
	}

	return o, nil
}

func (o *OrderItem) validate() error {
	const op = "domain.OrderItem.validate"

	if o.ProductID <= 0 {
		return fmt.Errorf("%s: %w", op, ErrInvalidProductID)
	}

	if o.Quantity <= 0 {
		return fmt.Errorf("%s: %w", op, ErrInvalidQuantity)
	}

	if o.Price <= 0 {
		return fmt.Errorf("%s: %w", op, ErrInvalidPrice)
	}

	return nil
}
