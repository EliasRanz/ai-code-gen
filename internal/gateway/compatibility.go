package gateway

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/EliasRanz/ai-code-gen/internal/cache"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/time/rate"
)

// AuthServiceProxy creates a Gin-compatible middleware for auth service proxy
// This is for backward compatibility with existing tests
func AuthServiceProxy(authServiceURL string, authCache *cache.AuthCache) gin.HandlerFunc {
	middleware := NewAuthProxyMiddleware(authServiceURL, authCache)

	return func(c *gin.Context) {
		ctx := WrapGinContext(c)

		next := func() error {
			c.Next()
			return nil
		}

		if err := middleware.Process(ctx, next); err != nil {
			// Error handling is done within the middleware Process method
			c.Abort()
			return
		}
	}
}

// AuthServiceRoleProxy creates a Gin-compatible middleware for role-based auth service proxy
// This is for backward compatibility with existing tests
func AuthServiceRoleProxy(authServiceURL string, authCache *cache.AuthCache, requiredRole string) gin.HandlerFunc {
	middleware := NewAuthProxyMiddleware(authServiceURL, authCache)

	return func(c *gin.Context) {
		ctx := WrapGinContext(c)

		// First validate the token
		userContext, err := middleware.ValidateToken(ctx)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "Authentication required"})
			return
		}

		// Check role permissions
		if userContext.Role != "admin" && userContext.Role != requiredRole {
			c.AbortWithStatusJSON(403, gin.H{"error": "Insufficient permissions"})
			return
		}

		// Set user context in Gin context
		c.Set("user_id", userContext.UserID)
		c.Set("user_email", userContext.Email)
		c.Set("user_role", userContext.Role)
		c.Set("user_active", userContext.Active)

		c.Next()
	}
}

// InvalidateUserCache invalidates user cache by token
// This is for backward compatibility with existing tests
func InvalidateUserCache(ctx context.Context, authCache *cache.AuthCache, token string) error {
	if authCache == nil {
		return nil // No cache to invalidate
	}

	// Create token hash for cache key
	tokenHash := fmt.Sprintf("%x", sha256.Sum256([]byte(token)))

	return authCache.InvalidateUserContext(ctx, tokenHash)
}

// MetricsMiddleware creates a Gin-compatible middleware for metrics collection
// This is for backward compatibility with existing tests
func MetricsMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		start := time.Now()

		// Get the metrics collector
		metricsCollector := NewMetricsCollector()

		// Increment request count
		metricsCollector.IncrementRequestCount(c.Request.URL.Path, c.Request.Method)

		// Process request
		c.Next()

		// Record latency and response code
		duration := time.Since(start)
		metricsCollector.RecordLatency(c.Request.URL.Path, duration)
		metricsCollector.IncrementResponseCode(c.Writer.Status())
	})
}

// RequestID creates a Gin-compatible middleware for request ID injection
// This is for backward compatibility with existing tests
func RequestID() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)
		c.Next()
	})
}

// ErrorHandler creates a Gin-compatible middleware for error handling
// This is for backward compatibility with existing tests
func ErrorHandler() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		c.Next()

		// Handle any errors that occurred during processing
		if len(c.Errors) > 0 {
			// Find the most relevant error
			for _, ginErr := range c.Errors {
				if ginErr.Type == gin.ErrorTypeBind {
					// Handle binding/validation errors
					if !c.Writer.Written() {
						c.AbortWithStatusJSON(400, gin.H{"error": "Invalid request data"})
					}
					return
				}
			}

			// Handle other types of errors
			err := c.Errors.Last()
			if err.Type == gin.ErrorTypePublic {
				if !c.Writer.Written() {
					c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
				}
			} else {
				// General server error
				if !c.Writer.Written() {
					c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
				}
			}
		}
	})
}

// NewRateLimiter creates a new rate limiter instance
// This is for backward compatibility with existing tests
func NewRateLimiter(rps rate.Limit, burst int) *CompatibleRateLimiter {
	return NewCompatibleRateLimiter(rps, burst)
}

// CreateRateLimitMiddleware creates a Gin-compatible rate limiting middleware
// This is for backward compatibility with existing tests
func CreateRateLimitMiddleware(requestsPerSecond, burst int) gin.HandlerFunc {
	middleware := NewRateLimitMiddleware(requestsPerSecond, burst)

	return func(c *gin.Context) {
		ctx := WrapGinContext(c)

		next := func() error {
			c.Next()
			return nil
		}

		if err := middleware.Process(ctx, next); err != nil {
			c.AbortWithStatusJSON(429, gin.H{"error": "Rate limit exceeded"})
			return
		}
	}
}

// CompatibleRateLimiter wraps RateLimitMiddlewareImpl to add missing methods for backward compatibility
type CompatibleRateLimiter struct {
	*RateLimitMiddlewareImpl
}

// GetLimiter exposes the internal getLimiter method for backward compatibility
func (c *CompatibleRateLimiter) GetLimiter(clientID string) *rate.Limiter {
	return c.getLimiter(clientID)
}

// RateLimit creates a Gin middleware function for rate limiting
func (c *CompatibleRateLimiter) RateLimit() gin.HandlerFunc {
	return func(ginCtx *gin.Context) {
		ctx := WrapGinContext(ginCtx)

		next := func() error {
			ginCtx.Next()
			return nil
		}

		if err := c.Process(ctx, next); err != nil {
			ginCtx.AbortWithStatusJSON(429, gin.H{"error": "Rate limit exceeded"})
			return
		}
	}
}

// NewCompatibleRateLimiter creates a new backward-compatible rate limiter
func NewCompatibleRateLimiter(rps rate.Limit, burst int) *CompatibleRateLimiter {
	impl := NewRateLimitMiddleware(int(rps), burst)
	return &CompatibleRateLimiter{
		RateLimitMiddlewareImpl: impl,
	}
}
