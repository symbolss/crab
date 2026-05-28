package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/family/crab-server/pkg/response"
)

// RateLimiter implements a simple in-memory rate limiter
type RateLimiter struct {
	mu       sync.RWMutex
	visitors map[string]*visitor
	limit    int           // max requests per window
	window   time.Duration // time window
}

type visitor struct {
	count    int
	lastSeen time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		limit:    limit,
		window:   window,
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// Allow checks if a request from the given IP is allowed
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		rl.visitors[ip] = &visitor{
			count:    1,
			lastSeen: time.Now(),
		}
		return true
	}

	// Reset counter if window has passed
	if time.Since(v.lastSeen) > rl.window {
		v.count = 1
		v.lastSeen = time.Now()
		return true
	}

	// Check limit
	if v.count >= rl.limit {
		return false
	}

	v.count++
	v.lastSeen = time.Now()
	return true
}

// cleanup removes old entries periodically
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > rl.window {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// RateLimit returns a middleware that limits requests per IP
func (rl *RateLimiter) RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getRealIP(r)
		if !rl.Allow(ip) {
			response.Error(w, http.StatusTooManyRequests, "RATE_LIMITED", "too many requests, please try again later")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// getRealIP extracts the real IP from the request
func getRealIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For may contain multiple IPs, use the first one
		if len(xff) > 0 {
			return xff
		}
	}
	// Fall back to RemoteAddr
	return r.RemoteAddr
}
