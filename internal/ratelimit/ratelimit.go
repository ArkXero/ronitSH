// Package ratelimit provides a simple in-memory per-key rate limiter.
package ratelimit

import (
	"sync"
	"time"
)

// Limiter allows at most one action per key within the configured window.
type Limiter struct {
	mu      sync.Mutex
	window  time.Duration
	records map[string]time.Time
}

// New creates a Limiter with the given window duration.
func New(window time.Duration) *Limiter {
	return &Limiter{
		window:  window,
		records: make(map[string]time.Time),
	}
}

// Allow returns true and records the attempt if the key has not acted within
// the window. Returns false (without recording) if the key is still throttled.
func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if last, ok := l.records[key]; ok {
		if time.Since(last) < l.window {
			return false
		}
	}
	l.records[key] = time.Now()
	return true
}

// Remaining returns how long until the key can act again.
// Returns 0 if the key is not throttled.
func (l *Limiter) Remaining(key string) time.Duration {
	l.mu.Lock()
	defer l.mu.Unlock()

	last, ok := l.records[key]
	if !ok {
		return 0
	}
	rem := l.window - time.Since(last)
	if rem < 0 {
		return 0
	}
	return rem
}
