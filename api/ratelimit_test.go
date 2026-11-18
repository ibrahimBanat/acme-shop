package api

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type fakeClock struct{ t time.Time }

func (c *fakeClock) now() time.Time          { return c.t }
func (c *fakeClock) advance(d time.Duration) { c.t = c.t.Add(d) }

func newTestLimiter() (*RateLimiter, *fakeClock) {
	c := &fakeClock{t: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	l := NewCheckoutLimiter()
	l.now = c.now
	return l, c
}

func TestRateLimiterAllowsTenThenBlocks(t *testing.T) {
	l, _ := newTestLimiter()
	for i := 0; i < 10; i++ {
		if _, err := l.Allow("alice"); err != nil {
			t.Fatalf("request %d: %v", i+1, err)
		}
	}
	wait, err := l.Allow("alice")
	if err != ErrRateLimited {
		t.Fatalf("11th request: got %v, want ErrRateLimited", err)
	}
	if wait != time.Minute {
		t.Fatalf("retry after: got %v, want 1m", wait)
	}
}

func TestRateLimiterPerCustomer(t *testing.T) {
	l, _ := newTestLimiter()
	for i := 0; i < 10; i++ {
		l.Allow("alice")
	}
	if _, err := l.Allow("bob"); err != nil {
		t.Fatalf("bob blocked by alice's usage: %v", err)
	}
}

func TestRateLimiterWindowSlides(t *testing.T) {
	l, c := newTestLimiter()
	// 5 requests at t=0, 5 at t=30s.
	for i := 0; i < 5; i++ {
		l.Allow("alice")
	}
	c.advance(30 * time.Second)
	for i := 0; i < 5; i++ {
		l.Allow("alice")
	}
	if _, err := l.Allow("alice"); err != ErrRateLimited {
		t.Fatalf("t=30s: got %v, want ErrRateLimited", err)
	}
	// At t=61s the first 5 have expired, the second 5 have not.
	c.advance(31 * time.Second)
	for i := 0; i < 5; i++ {
		if _, err := l.Allow("alice"); err != nil {
			t.Fatalf("t=61s request %d: %v", i+1, err)
		}
	}
	if _, err := l.Allow("alice"); err != ErrRateLimited {
		t.Fatalf("t=61s 6th: got %v, want ErrRateLimited", err)
	}
}

func TestRateLimiterRejectedRequestsDoNotCount(t *testing.T) {
	l, c := newTestLimiter()
	for i := 0; i < 10; i++ {
		l.Allow("alice")
	}
	c.advance(30 * time.Second)
	for i := 0; i < 50; i++ {
		l.Allow("alice") // all rejected
	}
	c.advance(31 * time.Second)
	if _, err := l.Allow("alice"); err != nil {
		t.Fatalf("after window: %v", err)
	}
}

func TestRateLimiterSweepsIdleKeys(t *testing.T) {
	l, c := newTestLimiter()
	l.Allow("alice")
	c.advance(2 * time.Minute)
	l.Allow("bob")
	if _, ok := l.hits["alice"]; ok {
		t.Fatal("idle key alice not swept")
	}
}

func TestRateLimiterConcurrent(t *testing.T) {
	l := NewCheckoutLimiter()
	var ok int64
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := l.Allow("alice"); err == nil {
				atomic.AddInt64(&ok, 1)
			}
		}()
	}
	wg.Wait()
	if ok != 10 {
		t.Fatalf("allowed %d, want 10", ok)
	}
}

func TestRateLimitMiddleware(t *testing.T) {
	l, _ := newTestLimiter()
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h := RateLimitMiddleware(l, func(r *http.Request) string {
		return r.Header.Get("X-Customer-ID")
	}, next)

	do := func(customer string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/checkout", nil)
		if customer != "" {
			req.Header.Set("X-Customer-ID", customer)
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec
	}

	for i := 0; i < 10; i++ {
		if rec := do("alice"); rec.Code != http.StatusOK {
			t.Fatalf("request %d: status %d", i+1, rec.Code)
		}
	}
	rec := do("alice")
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("11th: status %d, want 429", rec.Code)
	}
	if got := rec.Header().Get("Retry-After"); got != "60" {
		t.Fatalf("Retry-After: got %q, want 60", got)
	}

	// Anonymous requests fall back to client IP and are limited too.
	for i := 0; i < 10; i++ {
		if rec := do(""); rec.Code != http.StatusOK {
			t.Fatalf("anon request %d: status %d", i+1, rec.Code)
		}
	}
	if rec := do(""); rec.Code != http.StatusTooManyRequests {
		t.Fatalf("anon 11th: status %d, want 429", rec.Code)
	}
}
