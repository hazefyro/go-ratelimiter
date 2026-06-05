package ratelimiter

import (
	"errors"
	"net/http"
	"time"
)

// Config holds the resolved configuration for a RateLimiter. The required values,
// rate limit and bucket size, are passed positionally to New; everything else is
// set through the With* options.
type Config struct {
	rateLimit       int                        // tokens replenished per second (required)
	bucket          int                        // maximum burst size (required)
	idleTimeout     time.Duration              // evict visitors inactive for this long
	cleanupInterval time.Duration              // how often to evict idle visitors
	banning         *BanOptions                // nil disables banning
	keyFunc         func(*http.Request) string // extracts the rate limit key from a request
}

// Option configures a RateLimiter. Build options with the With* helpers and pass
// them to New.
type Option func(*Config)

// BanOptions defines the policy for banning visitors that repeatedly exceed the rate limit.
// All fields are required when banning is enabled.
type BanOptions struct {
	Threshold int           // number of violations within Window before a ban is issued
	Window    time.Duration // sliding window over which violations are counted
	Duration  time.Duration // how long a ban lasts
}

// WithIdleTimeout sets how long a visitor may be inactive before it is evicted (default 5m).
func WithIdleTimeout(d time.Duration) Option {
	return func(c *Config) { c.idleTimeout = d }
}

// WithCleanupInterval sets how often idle visitors are evicted (default 1m).
func WithCleanupInterval(d time.Duration) Option {
	return func(c *Config) { c.cleanupInterval = d }
}

// WithKeyFunc sets the function used to extract the rate limit key from a request (default RealIPKey).
func WithKeyFunc(f func(*http.Request) string) Option {
	return func(c *Config) { c.keyFunc = f }
}

// WithBanning enables banning of visitors that repeatedly exceed the rate limit.
func WithBanning(b BanOptions) Option {
	return func(c *Config) { c.banning = &b }
}

// defaultOptions are applied before any user-supplied options.
func defaultOptions() []Option {
	return []Option{
		WithIdleTimeout(5 * time.Minute),
		WithCleanupInterval(time.Minute),
		WithKeyFunc(RealIPKey),
	}
}

func (c *Config) validate() error {
	if c.rateLimit <= 0 {
		return errors.New("rateLimit must be a positive number of requests per second")
	}
	if c.bucket <= 0 {
		return errors.New("bucket must be a positive burst size")
	}
	if c.banning != nil {
		if c.banning.Threshold <= 0 {
			return errors.New("ban threshold must be greater than 0 when banning is enabled")
		}
		if c.banning.Window <= 0 {
			return errors.New("ban window must be greater than 0 when banning is enabled")
		}
		if c.banning.Duration <= 0 {
			return errors.New("ban duration must be greater than 0 when banning is enabled")
		}
	}
	return nil
}
