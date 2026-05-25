package ratelimiter

import (
	"net"
	"net/http"
)

// RemoteAddrKey extracts the IP address from r.RemoteAddr.
func RemoteAddrKey(r *http.Request) string {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return ip
}

// HeaderKey returns a KeyFunc that uses the value of a specific HTTP header.
func HeaderKey(header string) func(*http.Request) string {
	return func(r *http.Request) string {
		return r.Header.Get(header)
	}
}

// RealIPKey extracts the IP from X-Real-IP, falling back to RemoteAddr.
func RealIPKey(r *http.Request) string {
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	return RemoteAddrKey(r)
}

// ForwardedForKey extracts the IP from X-Forwarded-For, falling back to RemoteAddr.
func ForwardedForKey(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}

	return RemoteAddrKey(r)
}

// CFConnectingIPKey extracts the IP from CF-Connecting-IP (Cloudflare), falling back to RemoteAddr.
func CFConnectingIPKey(r *http.Request) string {
	if ip := r.Header.Get("CF-Connecting-IP"); ip != "" {
		return ip
	}

	return RemoteAddrKey(r)
}
