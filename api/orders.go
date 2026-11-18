package api

// Order is a customer's purchase.
type Order struct {
	// ID uniquely identifies the order.
	ID string
	// Customer identifies who placed the order.
	Customer string
	// TotalUSD is the order total in US dollars, stored as a float.
	TotalUSD float64
	// Status is the order's lifecycle state: one of
	// "placed", "paid", "shipped", or "refunded".
	Status string
}
