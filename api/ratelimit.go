package api

import (
	"errors"
	"math"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// ErrRateLimited is returned when a customer has used up their request budget.
var ErrRateLimited = errors.New("rate limit exceeded")

// RateLimiter allows at most Limit requests per customer within any sliding
// Window. It is safe for concurrent use. State is held in memory, so each
// process enforces its own limit and counts reset on restart.
type RateLimiter struct {
	limit  int
	window time.Duration
	now    func() time.Time

	mu        sync.Mutex
	hits      map[string][]time.Time
	lastSweep time.Time
}

// NewRateLimiter returns a limiter allowing limit requests per window per key.
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		limit:  limit,
		window: window,
		now:    time.Now,
		hits:   make(map[string][]time.Time),
	}
}

// NewCheckoutLimiter returns the checkout limiter: 10 requests a minute per customer.
func NewCheckoutLimiter() *RateLimiter {
	return NewRateLimiter(10, time.Minute)
}

// Allow records a request for key. If the key is over its limit, the request
// is not recorded and Allow returns ErrRateLimited along with how long to wait
// before the next request would be accepted.
func (l *RateLimiter) Allow(key string) (time.Duration, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	cutoff := now.Add(-l.window)
	l.sweep(now, cutoff)

	recent := prune(l.hits[key], cutoff)
	if len(recent) >= l.limit {
		l.hits[key] = recent
		return recent[0].Sub(cutoff), ErrRateLimited
	}
	l.hits[key] = append(recent, now)
	return 0, nil
}

// prune drops timestamps at or before cutoff. ts is in ascending order.
func prune(ts []time.Time, cutoff time.Time) []time.Time {
	i := 0
	for i < len(ts) && !ts[i].After(cutoff) {
		i++
	}
	return ts[i:]
}

// sweep removes idle keys at most once per window so the map does not grow
// without bound. Caller must hold l.mu.
func (l *RateLimiter) sweep(now, cutoff time.Time) {
	if now.Sub(l.lastSweep) < l.window {
		return
	}
	l.lastSweep = now
	for k, ts := range l.hits {
		if ts = prune(ts, cutoff); len(ts) == 0 {
			delete(l.hits, k)
		} else {
			l.hits[k] = ts
		}
	}
}

// RateLimitMiddleware rejects requests over the limiter's budget with
// 429 Too Many Requests and a Retry-After header. keyFn returns the customer
// ID for a request; when it returns "", the client IP is used instead so
// anonymous callers are still limited.
func RateLimitMiddleware(l *RateLimiter, keyFn func(*http.Request) string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := "customer:" + keyFn(r)
		if key == "customer:" {
			key = "ip:" + clientIP(r)
		}
		if wait, err := l.Allow(key); err != nil {
			secs := int(math.Ceil(wait.Seconds()))
			if secs < 1 {
				secs = 1
			}
			w.Header().Set("Retry-After", strconv.Itoa(secs))
			http.Error(w, "too many requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
