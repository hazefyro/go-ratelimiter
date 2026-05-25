package ratelimiter

import (
	"time"

	"golang.org/x/time/rate"
)

type visitor struct {
	limiter     *rate.Limiter
	lastSeen    time.Time
	violations  int
	windowStart time.Time
	bannedUntil time.Time
}

type VisitorStats struct {
	Banned      bool
	BannedUntil time.Time
	Violations  int
	LastSeen    time.Time
}
