package api

import "errors"

var ErrAlreadyRefunded = errors.New("order already refunded")

// Refund returns an order's money to the customer.
func Refund(o *Order, amount float64) error {
	if o.Status == "refunded" {
		return ErrAlreadyRefunded
	}
	o.Status = "refunded"
	return nil
}
