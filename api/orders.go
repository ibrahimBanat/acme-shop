package api

// Order is a customer's purchase.
type Order struct {
	ID       string
	Customer string
	TotalUSD float64
	Status   string // placed, paid, shipped, refunded

	// RefundReason is the free-text reason given when the order was refunded.
	// It may contain customer personal data (names, contact details, etc.),
	// so treat it like Customer: never write it to logs, metrics or error
	// messages.
	RefundReason string
}
