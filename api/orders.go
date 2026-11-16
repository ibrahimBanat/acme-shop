package api

// Order statuses.
const (
	StatusPlaced   = "placed"
	StatusPaid     = "paid"
	StatusShipped  = "shipped"
	StatusRefunded = "refunded"
)

// Order is a customer's purchase.
type Order struct {
	ID       string
	Customer string
	TotalUSD float64
	Status   string // one of the Status* constants
}
