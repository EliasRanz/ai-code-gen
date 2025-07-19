package gateway_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/EliasRanz/ai-code-gen/internal/gateway"
)

func TestRateLimitMiddleware_Process(t *testing.T) {
	// Create middleware with very low limits for testing
	middleware := gateway.NewRateLimitMiddleware(1, 1) // 1 request per second, burst of 1

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest("GET", "/test", nil)
	c.Request = req

	ctx := gateway.WrapGinContext(c)

	// First request should succeed
	nextCalled := false
	next := func() error {
		nextCalled = true
		return nil
	}

	err := middleware.Process(ctx, next)
	assert.NoError(t, err)
	assert.True(t, nextCalled)

	// Second request should be rate limited (same client IP)
	nextCalled = false
	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Request = req
	ctx2 := gateway.WrapGinContext(c2)

	err = middleware.Process(ctx2, next)
	assert.Error(t, err)
	assert.False(t, nextCalled)
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)
}

func TestRateLimitMiddleware_CheckLimit(t *testing.T) {
	middleware := gateway.NewRateLimitMiddleware(2, 2) // 2 requests per second, burst of 2

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/test", nil)
	ctx := gateway.WrapGinContext(c)

	// First check should succeed
	err := middleware.CheckLimit(ctx, "client1")
	assert.NoError(t, err)

	// Second check should succeed (within burst limit)
	err = middleware.CheckLimit(ctx, "client1")
	assert.NoError(t, err)

	// Third check should fail (exceeds burst)
	err = middleware.CheckLimit(ctx, "client1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rate limit exceeded")
}

func TestRateLimitMiddleware_GetLimitInfo(t *testing.T) {
	middleware := gateway.NewRateLimitMiddleware(10, 5) // 10 rps, burst 5

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/test", nil)
	ctx := gateway.WrapGinContext(c)

	// Get limit info for new client
	info, err := middleware.GetLimitInfo(ctx, "client1")
	assert.NoError(t, err)
	assert.NotNil(t, info)

	// Should have tokens remaining initially
	assert.GreaterOrEqual(t, info.Remaining, 0)
	assert.False(t, info.ResetTime.IsZero())
	assert.False(t, info.WindowStart.IsZero())
}

func TestRateLimitMiddleware_Configuration(t *testing.T) {
	middleware := gateway.NewRateLimitMiddleware(100, 10)

	// Test configuration
	config := middleware.GetConfig()
	assert.Equal(t, "rate-limit", config.GetName())
	assert.True(t, config.IsEnabled())

	params := config.GetParameters()
	assert.Equal(t, float64(100), params["requests_per_second"])
	assert.Equal(t, 10, params["burst"])

	// Test health check
	assert.NoError(t, middleware.HealthCheck())

	// Test validation
	assert.NoError(t, middleware.ValidateConfig())

	// Test metadata
	assert.Equal(t, "rate-limit", middleware.GetName())
	assert.Equal(t, 20, middleware.GetOrder())
}

func TestRateLimitMiddleware_InvalidConfiguration(t *testing.T) {
	// Test with zero rate
	middleware := gateway.NewRateLimitMiddleware(0, 1)
	assert.Error(t, middleware.HealthCheck())
	assert.Error(t, middleware.ValidateConfig())

	// Test with zero burst
	middleware = gateway.NewRateLimitMiddleware(1, 0)
	assert.Error(t, middleware.HealthCheck())
	assert.Error(t, middleware.ValidateConfig())

	// Test with negative values
	middleware = gateway.NewRateLimitMiddleware(-1, -1)
	assert.Error(t, middleware.HealthCheck())
	assert.Error(t, middleware.ValidateConfig())
}

func TestRateLimitMiddleware_DifferentClients(t *testing.T) {
	middleware := gateway.NewRateLimitMiddleware(1, 1) // Very restrictive limits

	gin.SetMode(gin.TestMode)

	// Create request from client 1
	w1 := httptest.NewRecorder()
	c1, _ := gin.CreateTestContext(w1)
	req1, _ := http.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = "192.168.1.1:8080"
	c1.Request = req1
	ctx1 := gateway.WrapGinContext(c1)

	// Create request from client 2
	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	req2, _ := http.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "192.168.1.2:8080"
	c2.Request = req2
	ctx2 := gateway.WrapGinContext(c2)

	next := func() error { return nil }

	// Both clients should be able to make their first request
	err1 := middleware.Process(ctx1, next)
	assert.NoError(t, err1)

	err2 := middleware.Process(ctx2, next)
	assert.NoError(t, err2)

	// But second requests should be rate limited
	w1 = httptest.NewRecorder()
	c1, _ = gin.CreateTestContext(w1)
	c1.Request = req1
	ctx1 = gateway.WrapGinContext(c1)

	err1 = middleware.Process(ctx1, next)
	assert.Error(t, err1)
}

func TestRateLimitMiddleware_Recovery(t *testing.T) {
	// Test that rate limit recovers over time
	middleware := gateway.NewRateLimitMiddleware(10, 1) // 10 rps, burst 1

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest("GET", "/test", nil)
	c.Request = req
	ctx := gateway.WrapGinContext(c)

	next := func() error { return nil }

	// First request should succeed
	err := middleware.Process(ctx, next)
	assert.NoError(t, err)

	// Second request should fail (burst exceeded)
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = req
	ctx = gateway.WrapGinContext(c)

	err = middleware.Process(ctx, next)
	assert.Error(t, err)

	// Wait for rate limiter to recover (small delay)
	time.Sleep(150 * time.Millisecond)

	// Should be able to make request again
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = req
	ctx = gateway.WrapGinContext(c)

	err = middleware.Process(ctx, next)
	assert.NoError(t, err)
}

func TestRateLimitMiddleware_Headers(t *testing.T) {
	middleware := gateway.NewRateLimitMiddleware(100, 10)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Use very restrictive limiter to trigger rate limiting
	middleware = gateway.NewRateLimitMiddleware(1, 0) // No burst, 1 rps

	req, _ := http.NewRequest("GET", "/test", nil)
	c.Request = req
	ctx := gateway.WrapGinContext(c)

	next := func() error { return nil }

	// Make request that should be rate limited
	err := middleware.Process(ctx, next)
	assert.Error(t, err)

	// Check rate limit headers were set
	assert.Equal(t, "1", w.Header().Get("X-RateLimit-Limit"))
	assert.Equal(t, "0", w.Header().Get("X-RateLimit-Remaining"))
	assert.NotEmpty(t, w.Header().Get("X-RateLimit-Reset"))
}
