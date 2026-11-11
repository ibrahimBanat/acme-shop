package api

import "testing"

func TestRefund(t *testing.T) {
	o := &Order{ID: "o1", TotalUSD: 40, Status: "paid"}
	if err := Refund(o, 40); err != nil {
		t.Fatal(err)
	}
	if err := Refund(o, 40); err != ErrAlreadyRefunded {
		t.Fatalf("second refund: %v", err)
	}
}
