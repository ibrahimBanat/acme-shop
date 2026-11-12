package api

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type fakeClock struct{ t time.Time }

func (c *fakeClock) now() time.Time          { return c.t }
func (c *fakeClock) advance(d time.Duration) { c.t = c.t.Add(d) }

func newTestLimiter() (*Limiter, *fakeClock) {
	c := &fakeClock{t: time.Unix(1_700_000_000, 0)}
	l := NewLimiter(DefaultCheckoutLimit, DefaultCheckoutWindow)
	l.now = c.now
	return l, c
}

func TestLimiterBlocksEleventhRequest(t *testing.T) {
	l, c := newTestLimiter()
	for i := 1; i <= 10; i++ {
		if !l.Allow("alice") {
			t.Fatalf("request %d rejected", i)
		}
		c.advance(time.Second)
	}
	if l.Allow("alice") {
		t.Fatal("11th request in window allowed")
	}
}

func TestLimiterSlidingWindow(t *testing.T) {
	l, c := newTestLimiter()
	for i := 0; i < 10; i++ {
		l.Allow("alice")
	}
	c.advance(59 * time.Second)
	if l.Allow("alice") {
		t.Fatal("allowed before window passed")
	}
	c.advance(time.Second)
	if !l.Allow("alice") {
		t.Fatal("rejected after window passed")
	}
}

func TestLimiterPerCustomer(t *testing.T) {
	l, _ := newTestLimiter()
	for i := 0; i < 10; i++ {
		l.Allow("alice")
	}
	if l.Allow("alice") {
		t.Fatal("alice over limit allowed")
	}
	if !l.Allow("bob") {
		t.Fatal("bob limited by alice's usage")
	}
}

func TestLimiterEvictsIdleCustomers(t *testing.T) {
	l, c := newTestLimiter()
	l.Allow("alice")
	l.Allow("bob")
	c.advance(2 * time.Minute)
	l.Allow("carol")
	if _, ok := l.hits["alice"]; ok {
		t.Fatal("idle customer not evicted")
	}
	if len(l.hits) != 1 {
		t.Fatalf("hits has %d keys, want 1", len(l.hits))
	}
}

func TestLimiterConcurrent(t *testing.T) {
	l, _ := newTestLimiter()
	var ok atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if l.Allow("alice") {
				ok.Add(1)
			}
		}()
	}
	wg.Wait()
	if got := ok.Load(); got != 10 {
		t.Fatalf("%d requests allowed, want 10", got)
	}
}

func TestCheckout(t *testing.T) {
	l, _ := newTestLimiter()
	for i := 0; i < 10; i++ {
		o := &Order{ID: "o", Customer: "alice"}
		if err := checkout(l, o); err != nil {
			t.Fatalf("checkout %d: %v", i+1, err)
		}
		if o.Status != "placed" {
			t.Fatalf("status %q, want placed", o.Status)
		}
	}
	o := &Order{ID: "o", Customer: "alice"}
	if err := checkout(l, o); err != ErrRateLimited {
		t.Fatalf("11th checkout: %v, want ErrRateLimited", err)
	}
	if o.Status != "" {
		t.Fatalf("rate-limited order status changed to %q", o.Status)
	}
	if err := checkout(l, &Order{ID: "o"}); err != ErrNoCustomer {
		t.Fatalf("no customer: %v, want ErrNoCustomer", err)
	}
}
