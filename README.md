# go-ratelimiter

A simple, framework-agnostic rate limiter for Go with optional IP banning.

## Install

```sh
go get github.com/haze/go-ratelimiter
```

## Usage

### Basic rate limiting

```go
rl, err := ratelimiter.New(&ratelimiter.Options{
    RateLimit: 10, // tokens per second
    Bucket:    10, // burst capacity
})
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
rl, err := ratelimiter.New(&ratelimiter.Options{
    RateLimit: 10,
    Bucket:    10,
    Banning: &ratelimiter.BanOptions{
        Threshold: 5,             // violations before ban
        Window:    time.Minute,   // violation tracking window
        Duration:  15 * time.Minute,
    },
})
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

## Options

| Field | Default | Description |
|---|---|---|
| `RateLimit` | required | tokens per second |
| `Bucket` | required | burst capacity |
| `IdleTimeout` | 5m | evict visitors inactive for this long |
| `CleanupInterval` | 1m | how often to run eviction |
| `KeyFunc` | `RealIPKey` | extracts the rate limit key from a request |
| `Banning` | nil | optional ban config, see `BanOptions` |
