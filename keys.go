package ratelimiter

import (
	"net"
	"net/http"
)

func RemoteAddrKey(r *http.Request) string {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return ip
}

func HeaderKey(header string) func(*http.Request) string {
	return func(r *http.Request) string {
		return r.Header.Get(header)
	}
}

func RealIPKey(r *http.Request) string {
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	return RemoteAddrKey(r)
}

func ForwrdedForKey(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}

	return RemoteAddrKey(r)
}

func CFConnectingIPKey(r *http.Request) string {
	if ip := r.Header.Get("CF-Connecting-IP"); ip != "" {
		return ip
	}

	return RemoteAddrKey(r)
}
