package ratelimiter_test

import (
	"errors"
	"testing"
	"time"

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

	if err := rl.Allow(key); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestAllow_returns_ErrRateLimited(t *testing.T) {
	rl := newLimiter(t, &ratelimiter.Options{
		RateLimit: 1,
		Bucket:    1,
	})

	rl.Allow(key)

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

func TestAllow_bans_after_threshold(t *testing.T) {
	rl := newLimiter(t, &ratelimiter.Options{
		RateLimit: 1,
		Bucket:    1,
		Banning: &ratelimiter.BanOptions{
			Duration:  time.Second + 30*time.Millisecond,
			Window:    time.Second,
			Threshold: 1,
		},
	})

	for range 2 {
		rl.Allow(key)
	}

	if err := rl.Allow(key); !errors.Is(err, ratelimiter.ErrBanned) {
		t.Fatalf("expected ErrBanned, got %v", err)
	}

}

func TestAllow_returns_ErrBanned_while_banned(t *testing.T) {
	rl := newLimiter(t, &ratelimiter.Options{
		RateLimit: 1,
		Bucket:    1,
		Banning: &ratelimiter.BanOptions{
			Duration:  time.Second + 30*time.Millisecond,
			Window:    time.Second,
			Threshold: 1,
		},
	})

	for range 2 {
		rl.Allow(key)
	}

	for range 5 {
		if err := rl.Allow(key); !errors.Is(err, ratelimiter.ErrBanned) {
			t.Fatalf("expected ErrBanned, got %v", err)
		}
	}

}

func TestAllow_resets_violations_after_window(t *testing.T) {
	rl := newLimiter(t, &ratelimiter.Options{
		RateLimit: 1,
		Bucket:    1,
		Banning: &ratelimiter.BanOptions{
			Duration:  time.Second + 30*time.Millisecond,
			Window:    time.Second,
			Threshold: 2,
		},
	})

	for range 2 {
		rl.Allow(key)
	}

	time.Sleep(time.Second)

	if err := rl.Allow(key); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}

}

func TestAllow_lifts_ban_after_durationfunc(t *testing.T) {
	rl := newLimiter(t, &ratelimiter.Options{
		RateLimit: 1,
		Bucket:    1,
		Banning: &ratelimiter.BanOptions{
			Duration:  time.Second + 30*time.Millisecond,
			Window:    time.Second,
			Threshold: 1,
		},
	})

	for range 2 {
		rl.Allow(key)
	}

	if err := rl.Allow(key); !errors.Is(err, ratelimiter.ErrBanned) {
		t.Fatalf("expected ErrBanned, got %v", err)
	}

	time.Sleep(time.Second + 30*time.Millisecond)

	if err := rl.Allow(key); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}

}
