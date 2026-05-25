package ratelimiter

import (
	"errors"
	"net/http"
	"time"
)

type BanOptions struct {
	Threshold int
	Window    time.Duration
	Duration  time.Duration
}

type Options struct {
	RateLimit       int
	Bucket          int
	IdleTimeout     time.Duration
	CleanupInterval time.Duration
	Banning         *BanOptions
	KeyFunc         func(*http.Request) string
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
