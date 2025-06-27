package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimiter stores rate limiters for different clients
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     r,
		burst:    b,
	}
}

// GetLimiter returns the rate limiter for a client
func (rl *RateLimiter) GetLimiter(clientID string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[clientID]
	if !exists {
		limiter = rate.NewLimiter(rl.rate, rl.burst)
		rl.limiters[clientID] = limiter
	}

	return limiter
}

// RateLimit middleware for rate limiting requests
func (rl *RateLimiter) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Use client IP as identifier
		clientID := c.ClientIP()
		limiter := rl.GetLimiter(clientID)

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate limit exceeded",
				"retry_after": time.Second.String(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// CreateRateLimitMiddleware creates a rate limit middleware with specified limits
func CreateRateLimitMiddleware(requestsPerSecond int, burst int) gin.HandlerFunc {
	rl := NewRateLimiter(rate.Limit(requestsPerSecond), burst)
	return rl.RateLimit()
}
