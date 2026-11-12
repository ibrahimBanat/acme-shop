package api

import "errors"

// ErrNoCustomer is returned when a checkout has no customer to rate-limit by.
var ErrNoCustomer = errors.New("checkout requires a customer")

// checkoutLimiter caps checkout attempts per customer.
var checkoutLimiter = NewLimiter(DefaultCheckoutLimit, DefaultCheckoutWindow)

// Checkout places an order, rejecting customers who exceed
// DefaultCheckoutLimit attempts per DefaultCheckoutWindow.
// o.Customer must come from the authenticated session, not client input.
func Checkout(o *Order) error {
	return checkout(checkoutLimiter, o)
}

func checkout(l *Limiter, o *Order) error {
	if o.Customer == "" {
		return ErrNoCustomer
	}
	if !l.Allow(o.Customer) {
		return ErrRateLimited
	}
	o.Status = "placed"
	return nil
}
