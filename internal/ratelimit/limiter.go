package ratelimit

import (
	"sync"
	"time"
)

type bucket struct {
	windowStart time.Time
	count       int
}

// Limiter enforces per-team requests-per-minute limits.
type Limiter struct {
	mu      sync.Mutex
	buckets map[string]bucket
}

func NewLimiter() *Limiter {
	return &Limiter{buckets: make(map[string]bucket)}
}

// Allow consumes one request from the caller's current minute bucket.
func (l *Limiter) Allow(team string, requestsPerMinute int, now time.Time) bool {
	if requestsPerMinute <= 0 {
		return true
	}
	window := now.UTC().Truncate(time.Minute)

	l.mu.Lock()
	defer l.mu.Unlock()

	b := l.buckets[team]
	if b.windowStart.IsZero() || b.windowStart.Before(window) {
		b.windowStart = window
		b.count = 0
	}
	if b.count >= requestsPerMinute {
		l.buckets[team] = b
		return false
	}
	b.count++
	l.buckets[team] = b
	return true
}
