package api

import (
	"errors"
	"strings"
	"testing"
)

func TestRefund(t *testing.T) {
	o := &Order{ID: "o1", TotalUSD: 40, Status: "paid"}
	if err := Refund(o, 40, "damaged in transit"); err != nil {
		t.Fatal(err)
	}
	if o.Status != "refunded" {
		t.Fatalf("status = %q, want refunded", o.Status)
	}
	if o.RefundReason != "damaged in transit" {
		t.Fatalf("reason = %q", o.RefundReason)
	}
	if err := Refund(o, 40, "second try"); !errors.Is(err, ErrAlreadyRefunded) {
		t.Fatalf("second refund: %v", err)
	}
	if o.RefundReason != "damaged in transit" {
		t.Fatalf("reason overwritten: %q", o.RefundReason)
	}
}

func TestRefundReasonValidation(t *testing.T) {
	tests := []struct {
		name       string
		reason     string
		wantErr    error
		wantReason string
	}{
		{"empty", "", ErrReasonRequired, ""},
		{"whitespace only", " \t\n ", ErrReasonRequired, ""},
		{"trimmed", "  wrong size  ", nil, "wrong size"},
		{"exactly 200", strings.Repeat("a", 200), nil, strings.Repeat("a", 200)},
		{"201", strings.Repeat("a", 201), ErrReasonTooLong, ""},
		{"200 multibyte", strings.Repeat("é", 200), nil, strings.Repeat("é", 200)},
		{"201 multibyte", strings.Repeat("é", 201), ErrReasonTooLong, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &Order{ID: "o1", TotalUSD: 40, Status: "paid"}
			err := Refund(o, 40, tt.reason)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr != nil {
				if o.Status != "paid" || o.RefundReason != "" {
					t.Fatalf("order mutated on error: %+v", o)
				}
				return
			}
			if o.Status != "refunded" || o.RefundReason != tt.wantReason {
				t.Fatalf("got status %q reason %q", o.Status, o.RefundReason)
			}
		})
	}
}

func TestRefundErrorsDoNotLeakReason(t *testing.T) {
	// The reason may contain personal data; errors must not echo it.
	pii := "jane.doe@example.com " + strings.Repeat("x", 200)
	o := &Order{ID: "o1", TotalUSD: 40, Status: "paid"}
	err := Refund(o, 40, pii)
	if !errors.Is(err, ErrReasonTooLong) {
		t.Fatalf("err = %v, want ErrReasonTooLong", err)
	}
	if strings.Contains(err.Error(), "jane.doe") {
		t.Fatalf("error leaks reason: %q", err)
	}

	o = &Order{ID: "o2", TotalUSD: 40, Status: "refunded", RefundReason: "first"}
	err = Refund(o, 40, "jane.doe@example.com asked again")
	if !errors.Is(err, ErrAlreadyRefunded) {
		t.Fatalf("err = %v, want ErrAlreadyRefunded", err)
	}
	if strings.Contains(err.Error(), "jane.doe") {
		t.Fatalf("error leaks reason: %q", err)
	}
}

func TestRefundInvalidReasonOnRefundedOrder(t *testing.T) {
	o := &Order{ID: "o1", TotalUSD: 40, Status: "refunded", RefundReason: "first"}
	if err := Refund(o, 40, ""); !errors.Is(err, ErrReasonRequired) {
		t.Fatalf("err = %v, want ErrReasonRequired", err)
	}
	if o.RefundReason != "first" {
		t.Fatalf("reason changed: %q", o.RefundReason)
	}
}
