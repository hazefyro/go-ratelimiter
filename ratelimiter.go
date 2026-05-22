package ratelimiter

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type RateLimiter struct {
	mu        sync.Mutex
	rateLimit rate.Limit
	visitors  map[string]*visitor
	options   Options
	stop      chan struct{}
}

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

func (rl *RateLimiter) Allow(key string) error {
	return rl.AllowN(key, 1)
}

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

	if rl.options.BanDuration > 0 && time.Since(v.windowStart) > rl.options.ViolationWindow {
		v.violations = 0
		v.windowStart = now
	}

	if !v.limiter.AllowN(now, n) {
		if rl.options.BanDuration > 0 {
			v.violations++
			if v.violations >= rl.options.ViolationThreshold {
				v.bannedUntil = now.Add(rl.options.BanDuration)
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

func (rl *RateLimiter) Stop() {
	close(rl.stop)
}
