package ratelimiter

import (
	"errors"
	"net/http"
	"time"
)

// BanOptions defines the policy for banning visitors that repeatedly exceed the rate limit.
// All fields are required when Banning is non-nil.
type BanOptions struct {
	Threshold int           // number of violations within Window before a ban is issued
	Window    time.Duration // sliding window over which violations are counted
	Duration  time.Duration // how long a ban lasts
}

// Options controls the behavior of a RateLimiter. RateLimit and Bucket are required;
// all other fields have defaults applied by New when left at zero.
type Options struct {
	RateLimit       int                        // tokens replenished per second (required)
	Bucket          int                        // maximum burst size (required)
	IdleTimeout     time.Duration              // evict visitors inactive for this long (default 5m)
	CleanupInterval time.Duration              // how often to evict idle visitors (default 1m)
	Banning         *BanOptions                // optional; nil disables banning
	KeyFunc         func(*http.Request) string // extracts the rate limit key from a request (default RealIPKey)
}

func (o *Options) validate() error {
	if o.RateLimit <= 0 {
		return errors.New("rateLimit must be a positive number of requests per second")
	}
	if o.Bucket <= 0 {
		return errors.New("bucket must be a positive burst size")
	}
	if o.Banning != nil {
		if o.Banning.Threshold <= 0 {
			return errors.New("ban threshold must be greater than 0 when banning is enabled")
		}
		if o.Banning.Window <= 0 {
			return errors.New("ban window must be greater than 0 when banning is enabled")
		}
		if o.Banning.Duration <= 0 {
			return errors.New("ban duration must be greater than 0 when banning is enabled")
		}
	}
	return nil
}
