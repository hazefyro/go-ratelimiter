// Package ratelimiter provides a token-bucket rate limiter with optional IP banning.
//
// Create a limiter with [New], then call [RateLimiter.Allow] or [RateLimiter.AllowN] for each
// incoming request. Both methods return nil when the request is permitted, or a sentinel
// error ([ErrRateLimited], [ErrBanned]) that callers can inspect with errors.Is.
//
// Example (basic):
//
//	rl, err := ratelimiter.New(10, 10)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer rl.Stop()
//
//	switch err := rl.Allow("192.0.2.1"); {
//	case err == nil:
//	    // allow
//	case errors.Is(err, ratelimiter.ErrRateLimited):
//	    w.WriteHeader(http.StatusTooManyRequests)
//	case errors.Is(err, ratelimiter.ErrBanned):
//	    w.WriteHeader(http.StatusForbidden)
//	}
//
// Example (with banning):
//
//	rl, err := ratelimiter.New(10, 10,
//	    ratelimiter.WithBanning(ratelimiter.BanOptions{
//	        Threshold: 5,
//	        Window:    time.Minute,
//	        Duration:  15 * time.Minute,
//	    }),
//	)
//
// Key extraction is handled by a [KeyFunc]. Built-in options include [RealIPKey],
// [ForwardedForKey], [CFConnectingIPKey], [RemoteAddrKey], and [HeaderKey].
package ratelimiter
