package api

import "errors"

// ErrAlreadyRefunded is returned by Refund when the order's status is
// already "refunded".
var ErrAlreadyRefunded = errors.New("order already refunded")

// Refund returns an order's money to the customer.
//
// It returns ErrAlreadyRefunded if the order is already refunded;
// otherwise it sets the order's Status to "refunded".
//
// Note: amount is currently ignored. Refund does not support partial
// refunds, does not validate amount against TotalUSD, does not check
// that the order was paid, and does not call any payment provider.
func Refund(o *Order, amount float64) error {
	if o.Status == "refunded" {
		return ErrAlreadyRefunded
	}
	o.Status = "refunded"
	return nil
}
