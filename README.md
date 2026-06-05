# go-ratelimiter

A simple, framework-agnostic rate limiter for Go with optional IP banning.

## Install

```sh
go get github.com/haze/go-ratelimiter
```

## Usage

### Basic rate limiting

```go
rl, err := ratelimiter.New(10, 10) // rate limit (tokens/sec), bucket (burst)
if err != nil {
    log.Fatal(err)
}
defer rl.Stop()

if err := rl.Allow("192.0.2.1"); err != nil {
    // err is ErrRateLimited
}
```

### With banning

Visitors that hit the rate limit too many times get banned for a duration.

```go
rl, err := ratelimiter.New(10, 10,
    ratelimiter.WithBanning(ratelimiter.BanOptions{
        Threshold: 5,           // violations before ban
        Window:    time.Minute, // violation tracking window
        Duration:  15 * time.Minute,
    }),
)
```

`Allow` returns `ErrRateLimited` or `ErrBanned` so callers can distinguish:

```go
switch err := rl.Allow(key); {
case err == nil:
    // allowed
case errors.Is(err, ratelimiter.ErrRateLimited):
    w.WriteHeader(http.StatusTooManyRequests)
case errors.Is(err, ratelimiter.ErrBanned):
    w.WriteHeader(http.StatusForbidden)
}

```

### Key functions

Built-in helpers for extracting keys from requests:

| Function | Source |
|---|---|
| `RealIPKey` | `X-Real-IP`, fallback to `RemoteAddr` |
| `ForwardedForKey` | `X-Forwarded-For`, fallback to `RemoteAddr` |
| `CFConnectingIPKey` | `CF-Connecting-IP`, fallback to `RemoteAddr` |
| `RemoteAddrKey` | `r.RemoteAddr` |
| `HeaderKey(h)` | any header |

## Configuration

`New(rateLimit, bucket int, opts ...Option)` takes the two required values
positionally; everything else is configured with `With*` options.

| Argument | Description |
|---|---|
| `rateLimit` | tokens replenished per second (required) |
| `bucket` | burst capacity (required) |

| Option | Default | Description |
|---|---|---|
| `WithIdleTimeout(d)` | 5m | evict visitors inactive for this long |
| `WithCleanupInterval(d)` | 1m | how often to run eviction |
| `WithKeyFunc(f)` | `RealIPKey` | extracts the rate limit key from a request |
| `WithBanning(b)` | disabled | enable banning, see `BanOptions` |
