package ratelimiter

import (
	"errors"
	"net/http"
	"time"
)

type Options struct {
	RateLimit          int
	Bucket             int
	ViolationThreshold int
	IdleTimeout        time.Duration
	ViolationWindow    time.Duration
	BanDuration        time.Duration
	CleanupInterval    time.Duration
	KeyFunc            func(*http.Request) string
}

func (o *Options) validate() error {
	if o.RateLimit <= 0 {
		return errors.New("rateLimit must be a positive number of requests per second")
	}
	if o.Bucket <= 0 {
		return errors.New("bucket must be a positive burst size")
	}
	if o.ViolationThreshold > 0 || o.ViolationWindow > 0 || o.BanDuration > 0 {
		if o.ViolationThreshold <= 0 {
			return errors.New("violation threshold must be greater than 0 when banning is enabled")
		}
		if o.ViolationWindow <= 0 {
			return errors.New("violation window must be greater than 0 when banning is enabled")
		}
		if o.BanDuration <= 0 {
			return errors.New("ban duration must be greater than 0 when banning is enabled")
		}
	}
	return nil
}
