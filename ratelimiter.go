// Package ratelimiter provides a token-bucket rate limiter with optional IP banning.
package ratelimiter

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter tracks per-key request rates and optionally bans repeat violators.
type RateLimiter struct {
	mu        sync.RWMutex
	rateLimit rate.Limit
	visitors  map[string]*visitor
	options   Options
	stop      chan struct{}
}

// New creates a RateLimiter from the given options. Returns an error if options are invalid.
// Callers must call Stop when done to release the background cleanup goroutine.
func New(options *Options) (*RateLimiter, error) {
	if options == nil {
		options = &Options{}
	}
	if options.KeyFunc == nil {
		options.KeyFunc = RealIPKey
	}
	if options.CleanupInterval == 0 {
		options.CleanupInterval = time.Minute
	}
	if options.IdleTimeout == 0 {
		options.IdleTimeout = 5 * time.Minute
	}
	if err := options.validate(); err != nil {
		return nil, err
	}

	rl := &RateLimiter{
		visitors:  make(map[string]*visitor),
		rateLimit: rate.Limit(options.RateLimit),
		options:   *options,
		stop:      make(chan struct{}),
	}

	go rl.cleanup()

	return rl, nil
}

// Allow reports whether the visitor identified by key may make a request.
// Returns ErrRateLimited or ErrBanned if the request is denied.
func (rl *RateLimiter) Allow(key string) error {
	return rl.AllowN(key, 1)
}

// AllowN is like Allow but consumes n tokens at once.
func (rl *RateLimiter) AllowN(key string, n int) error {
	now := time.Now()
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, ok := rl.visitors[key]
	if !ok {
		v = &visitor{
			limiter:     rate.NewLimiter(rl.rateLimit, rl.options.Bucket),
			windowStart: now,
		}
		rl.visitors[key] = v
	}
	v.lastSeen = now

	if now.Before(v.bannedUntil) {
		return ErrBanned
	}

	if rl.options.Banning != nil && time.Since(v.windowStart) > rl.options.Banning.Window {
		v.violations = 0
		v.windowStart = now
	}

	if !v.limiter.AllowN(now, n) {
		if rl.options.Banning != nil {
			v.violations++
			if v.violations >= rl.options.Banning.Threshold {
				v.bannedUntil = now.Add(rl.options.Banning.Duration)
				v.violations = 0
			}
		}
		return ErrRateLimited
	}

	return nil
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.options.CleanupInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			rl.mu.Lock()
			for ip, v := range rl.visitors {
				if time.Since(v.lastSeen) > rl.options.IdleTimeout {
					delete(rl.visitors, ip)
				}
			}
			rl.mu.Unlock()
		case <-rl.stop:
			return
		}
	}
}

// Stats returns a snapshot of all active visitors keyed by their rate limit key.
func (rl *RateLimiter) Stats() map[string]VisitorStats {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	now := time.Now()
	result := make(map[string]VisitorStats, len(rl.visitors))
	for key, v := range rl.visitors {
		result[key] = VisitorStats{
			Banned:      now.Before(v.bannedUntil),
			BannedUntil: v.bannedUntil,
			Violations:  v.violations,
			LastSeen:    v.lastSeen,
		}
	}
	return result
}

// Stop shuts down the background cleanup goroutine. Must be called when the RateLimiter is no longer needed.
func (rl *RateLimiter) Stop() {
	close(rl.stop)
}
