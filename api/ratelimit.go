package api

import (
	"errors"
	"sync"
	"time"
)

// ErrRateLimited is returned when a customer has made too many requests.
var ErrRateLimited = errors.New("rate limit exceeded")

// Default checkout limit: 10 requests per customer per minute.
const (
	DefaultCheckoutLimit  = 10
	DefaultCheckoutWindow = time.Minute
)

// Limiter is an in-memory, per-key sliding-window rate limiter.
// It is safe for concurrent use. Limits apply per process only.
type Limiter struct {
	limit  int
	window time.Duration
	now    func() time.Time

	mu        sync.Mutex
	hits      map[string][]time.Time
	lastSweep time.Time
}

// NewLimiter allows up to limit requests per key within any window.
func NewLimiter(limit int, window time.Duration) *Limiter {
	return &Limiter{
		limit:  limit,
		window: window,
		now:    time.Now,
		hits:   make(map[string][]time.Time),
	}
}

// Allow records a request for key and reports whether it is within the limit.
// Rejected requests are not recorded.
func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	cutoff := now.Add(-l.window)

	// Drop idle keys once per window so memory stays bounded.
	if now.Sub(l.lastSweep) >= l.window {
		for k, ts := range l.hits {
			if len(ts) == 0 || !ts[len(ts)-1].After(cutoff) {
				delete(l.hits, k)
			}
		}
		l.lastSweep = now
	}

	ts := l.hits[key]
	i := 0
	for i < len(ts) && !ts[i].After(cutoff) {
		i++
	}
	ts = ts[i:]

	if len(ts) >= l.limit {
		l.hits[key] = ts
		return false
	}
	l.hits[key] = append(ts, now)
	return true
}
