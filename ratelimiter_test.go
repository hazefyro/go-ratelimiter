package ratelimiter_test

import (
	"errors"
	"testing"

	ratelimiter "github.com/haze/go-ratelimiter"
)

const key = "192.0.2.1"

func newLimiter(t *testing.T, opts *ratelimiter.Options) *ratelimiter.RateLimiter {
	t.Helper()
	rl, err := ratelimiter.New(opts)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(rl.Stop)
	return rl
}

func TestAllow_permits_under_limit(t *testing.T) {
	rl := newLimiter(t, &ratelimiter.Options{
		RateLimit: 10,
		Bucket:    10,
	})

	if err := rl.Allow("192.0.2.1"); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestAllow_returns_ErrRateLimited(t *testing.T) {
	rl := newLimiter(t, &ratelimiter.Options{
		RateLimit: 1,
		Bucket:    1,
	})

	if err := rl.Allow(key); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}

	if err := rl.Allow(key); !errors.Is(err, ratelimiter.ErrRateLimited) {
		t.Fatalf("expected ErrRateLimited, got %v", err)
	}

}

func TestAllowN_consumes_multiple_tokens(t *testing.T) {
	rl := newLimiter(t, &ratelimiter.Options{
		RateLimit: 10,
		Bucket:    5,
	})

	if err := rl.AllowN(key, 5); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}

}

func TestAllowN_returns_ErrRateLimited_when_insufficient_tokens(t *testing.T) {
	rl := newLimiter(t, &ratelimiter.Options{
		RateLimit: 10,
		Bucket:    5,
	})

	if err := rl.AllowN(key, 6); !errors.Is(err, ratelimiter.ErrRateLimited) {
		t.Fatalf("expected ErrRateLimited, got %v", err)
	}
}

// func TestAllow_bans_after_threshold(t *testing.T){
//     rl := newLimiter(t, &ratelimiter.Options{
// 		RateLimit: 10,
// 		Bucket:    5,
// 		ViolationThreshold: 1,
// 		ViolationWindow: 1,
// 		BanDuration: 5,
// 	})
// }
