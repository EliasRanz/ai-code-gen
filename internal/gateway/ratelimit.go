package gateway

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimitMiddlewareImpl provides rate limiting functionality
type RateLimitMiddlewareImpl struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
	name     string
	order    int
}

// NewRateLimitMiddleware creates a new rate limiting middleware
func NewRateLimitMiddleware(requestsPerSecond, burst int) *RateLimitMiddlewareImpl {
	return &RateLimitMiddlewareImpl{
		limiters: make(map[string]*rate.Limiter),
		rate:     rate.Limit(requestsPerSecond),
		burst:    burst,
		name:     "rate-limit",
		order:    20, // Run early but after logging
	}
}

// Process implements middleware processing with rate limiting
func (r *RateLimitMiddlewareImpl) Process(ctx Context, next Next) error {
	clientID := ctx.ClientIP()
	limiter := r.getLimiter(clientID)

	if !limiter.Allow() {
		return r.handleRateLimit(ctx, clientID)
	}

	return next()
}

// CheckLimit checks if the client is within rate limits
func (r *RateLimitMiddlewareImpl) CheckLimit(ctx Context, identifier string) error {
	limiter := r.getLimiter(identifier)
	if !limiter.Allow() {
		return fmt.Errorf("rate limit exceeded for %s", identifier)
	}
	return nil
}

// GetLimitInfo returns current rate limit information
func (r *RateLimitMiddlewareImpl) GetLimitInfo(ctx Context, identifier string) (*LimitInfo, error) {
	limiter := r.getLimiter(identifier)

	// Calculate tokens remaining (approximation)
	tokens := limiter.Tokens()
	remaining := int(tokens)
	if remaining < 0 {
		remaining = 0
	}

	return &LimitInfo{
		Remaining:   remaining,
		ResetTime:   time.Now().Add(time.Second), // Simplified - would need proper window tracking
		WindowStart: time.Now().Truncate(time.Second),
	}, nil
}

// GetConfig returns middleware configuration
func (r *RateLimitMiddlewareImpl) GetConfig() MiddlewareConfig {
	return NewBasicMiddlewareConfig(r.name, true, map[string]interface{}{
		"requests_per_second": float64(r.rate),
		"burst":               r.burst,
		"limiter_count":       len(r.limiters),
	})
}

// GetName returns middleware name
func (r *RateLimitMiddlewareImpl) GetName() string {
	return r.name
}

// GetOrder returns middleware execution order
func (r *RateLimitMiddlewareImpl) GetOrder() int {
	return r.order
}

// HealthCheck validates middleware health
func (r *RateLimitMiddlewareImpl) HealthCheck() error {
	if r.rate <= 0 {
		return fmt.Errorf("invalid rate limit configuration: rate=%v", r.rate)
	}
	if r.burst <= 0 {
		return fmt.Errorf("invalid rate limit configuration: burst=%v", r.burst)
	}
	return nil
}

// ValidateConfig validates middleware configuration
func (r *RateLimitMiddlewareImpl) ValidateConfig() error {
	return r.HealthCheck()
}

// getLimiter gets or creates a rate limiter for a client
func (r *RateLimitMiddlewareImpl) getLimiter(clientID string) *rate.Limiter {
	r.mu.Lock()
	defer r.mu.Unlock()

	limiter, exists := r.limiters[clientID]
	if !exists {
		limiter = rate.NewLimiter(r.rate, r.burst)
		r.limiters[clientID] = limiter

		// Clean up old limiters periodically (simplified approach)
		if len(r.limiters) > 1000 {
			r.cleanupLimiters()
		}
	}

	return limiter
}

// handleRateLimit handles rate limit exceeded cases
func (r *RateLimitMiddlewareImpl) handleRateLimit(ctx Context, clientID string) error {
	if ginCtx, ok := ctx.(*GinContextWrapper); ok {
		ginCtx.Header("X-RateLimit-Limit", fmt.Sprintf("%.0f", float64(r.rate)))
		ginCtx.Header("X-RateLimit-Remaining", "0")
		ginCtx.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Second).Unix()))

		ginCtx.JSON(http.StatusTooManyRequests, map[string]interface{}{
			"error":       "rate limit exceeded",
			"retry_after": time.Second.String(),
			"client_id":   clientID,
		})
		ginCtx.Abort()
	}

	return fmt.Errorf("rate limit exceeded for client %s", clientID)
}

// cleanupLimiters removes half of the limiters (simplified cleanup)
func (r *RateLimitMiddlewareImpl) cleanupLimiters() {
	count := 0
	target := len(r.limiters) / 2

	for clientID := range r.limiters {
		if count >= target {
			break
		}
		delete(r.limiters, clientID)
		count++
	}
}
