package middleware

import (
	"net/http"
	"sync"
	"time"
)

type IPRateLimiter struct {
	ips map[string]*rateLimiter
	mu  sync.RWMutex
	rps int // requests per second
}

type rateLimiter struct {
	count    int
	lastTime time.Time
}

func NewIPRateLimiter(rps int) *IPRateLimiter {
	return &IPRateLimiter{
		ips: make(map[string]*rateLimiter),
		rps: rps,
	}
}

func (i *IPRateLimiter) Allow(ip string) bool {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter, exists := i.ips[ip]
	now := time.Now()

	if !exists {
		i.ips[ip] = &rateLimiter{count: 1, lastTime: now}
		return true
	}

	if now.Sub(limiter.lastTime) > time.Second {
		limiter.count = 1
		limiter.lastTime = now
		return true
	}

	if limiter.count >= i.rps {
		return false
	}

	limiter.count++
	return true
}

func RateLimit(rps int) func(http.Handler) http.Handler {
	limiter := NewIPRateLimiter(rps)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !limiter.Allow(r.RemoteAddr) {
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
