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
	config    Config
	stop      chan struct{}
	stopOnce  sync.Once
}

// New creates a RateLimiter with the given rate limit (tokens replenished per second)
// and bucket size (maximum burst). Additional behavior is configured through the With*
// options. Returns an error if the configuration is invalid. Callers must call Stop when
// done to release the background cleanup goroutine.
func New(rateLimit, bucket int, opts ...Option) (*RateLimiter, error) {
	cfg := &Config{
		rateLimit: rateLimit,
		bucket:    bucket,
	}
	for _, opt := range append(defaultOptions(), opts...) {
		opt(cfg)
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	rl := &RateLimiter{
		visitors:  make(map[string]*visitor),
		rateLimit: rate.Limit(cfg.rateLimit),
		config:    *cfg,
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

// AllowN reports whether the visitor identified by key may make a request that costs n tokens.
// Returns ErrRateLimited or ErrBanned if the request is denied.
func (rl *RateLimiter) AllowN(key string, n int) error {
	now := time.Now()
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, ok := rl.visitors[key]
	if !ok {
		v = &visitor{
			limiter:     rate.NewLimiter(rl.rateLimit, rl.config.bucket),
			windowStart: now,
		}
		rl.visitors[key] = v
	}
	v.lastSeen = now

	if now.Before(v.bannedUntil) {
		return ErrBanned
	}

	if rl.config.banning != nil && time.Since(v.windowStart) > rl.config.banning.Window {
		v.violations = 0
		v.windowStart = now
	}

	if !v.limiter.AllowN(now, n) {
		if rl.config.banning != nil {
			v.violations++
			if v.violations >= rl.config.banning.Threshold {
				v.bannedUntil = now.Add(rl.config.banning.Duration)
				v.violations = 0
			}
		}
		return ErrRateLimited
	}

	return nil
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.config.cleanupInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			rl.mu.Lock()
			for ip, v := range rl.visitors {
				if time.Since(v.lastSeen) > rl.config.idleTimeout {
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
// Safe to call more than once.
func (rl *RateLimiter) Stop() {
	rl.stopOnce.Do(func() { close(rl.stop) })
}
