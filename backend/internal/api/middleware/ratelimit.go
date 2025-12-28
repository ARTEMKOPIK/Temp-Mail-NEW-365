package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter implements a simple in-memory rate limiter
type RateLimiter struct {
	requests map[string]*clientInfo
	mu       sync.RWMutex
	maxReqs  int
	window   time.Duration
}

type clientInfo struct {
	count      int
	windowStart time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(maxRequests int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string]*clientInfo),
		maxReqs:  maxRequests,
		window:   window,
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// Middleware returns a Gin middleware function
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		if !rl.Allow(clientIP) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
				"message": "too many requests, please try again later",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Allow checks if a request from the given IP is allowed
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	
	info, exists := rl.requests[ip]
	if !exists {
		rl.requests[ip] = &clientInfo{
			count:      1,
			windowStart: now,
		}
		return true
	}

	// Check if window has expired
	if now.Sub(info.windowStart) > rl.window {
		info.count = 1
		info.windowStart = now
		return true
	}

	// Increment counter
	info.count++

	// Check if limit exceeded
	return info.count <= rl.maxReqs
}

// cleanup periodically removes old entries
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.window)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, info := range rl.requests {
			if now.Sub(info.windowStart) > rl.window*2 {
				delete(rl.requests, ip)
			}
		}
		rl.mu.Unlock()
	}
}

