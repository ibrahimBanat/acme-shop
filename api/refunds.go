package api

import (
	"errors"
	"strings"
	"unicode/utf8"
)

// MaxRefundReasonLen is the maximum length of a refund reason, in characters.
const MaxRefundReasonLen = 200

var (
	ErrAlreadyRefunded = errors.New("order already refunded")
	ErrReasonRequired  = errors.New("refund reason is required")
	ErrReasonTooLong   = errors.New("refund reason exceeds 200 characters")
)

// Refund returns an order's money to the customer and records why.
// The reason is trimmed of surrounding whitespace, must not be empty and
// must be at most MaxRefundReasonLen characters. On error the order is left
// unchanged.
func Refund(o *Order, amount float64, reason string) error {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return ErrReasonRequired
	}
	if utf8.RuneCountInString(reason) > MaxRefundReasonLen {
		return ErrReasonTooLong
	}
	if o.Status == "refunded" {
		return ErrAlreadyRefunded
	}
	o.Status = "refunded"
	o.RefundReason = reason
	return nil
}
