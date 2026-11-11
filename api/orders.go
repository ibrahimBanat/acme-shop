package api

// Order is a customer's purchase.
type Order struct {
	ID       string
	Customer string
	TotalUSD float64
	Status   string // placed, paid, shipped, refunded
}
