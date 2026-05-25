package ratelimiter

import "errors"

var (
	// ErrRateLimited is returned when a visitor exceeds the allowed request rate.
	ErrRateLimited = errors.New("too many requests")
	// ErrBanned is returned when a visitor has been banned for repeated violations.
	ErrBanned = errors.New("banned due to repeated violations")
)
