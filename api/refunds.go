package api

import "errors"

var ErrAlreadyRefunded = errors.New("order already refunded")

// Refund returns an order's money to the customer.
func Refund(o *Order, amount float64) error {
	if o.Status == StatusRefunded {
		return ErrAlreadyRefunded
	}
	o.Status = StatusRefunded
	return nil
}
