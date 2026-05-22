package ratelimiter

import "errors"

var (
	ErrRateLimited = errors.New("too many requests")
	ErrBanned      = errors.New("banned due to repeated violations")
)
