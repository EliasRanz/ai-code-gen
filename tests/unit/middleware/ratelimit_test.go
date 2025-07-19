package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"

	"github.com/EliasRanz/ai-code-gen/internal/gateway"
)

func TestNewRateLimiter(t *testing.T) {
	tests := []struct {
		name        string
		rateLimit   rate.Limit
		burst       int
		expectRate  rate.Limit
		expectBurst int
	}{
		{
			name:        "Standard rate limiter",
			rateLimit:   rate.Limit(10),
			burst:       5,
			expectRate:  rate.Limit(10),
			expectBurst: 5,
		},
		{
			name:        "High rate limiter",
			rateLimit:   rate.Limit(100),
			burst:       20,
			expectRate:  rate.Limit(100),
			expectBurst: 20,
		},
		{
			name:        "Low rate limiter",
			rateLimit:   rate.Limit(1),
			burst:       1,
			expectRate:  rate.Limit(1),
			expectBurst: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rl := gateway.NewRateLimiter(tt.rateLimit, tt.burst)
			assert.NotNil(t, rl)

			// Create a test client to verify limiter behavior
			limiter := rl.GetLimiter("test-client")
			assert.NotNil(t, limiter)
		})
	}
}

func TestRateLimiter_GetLimiter(t *testing.T) {
	rl := gateway.NewRateLimiter(rate.Limit(10), 5)

	// Test getting limiter for same client returns same instance
	limiter1 := rl.GetLimiter("client1")
	limiter2 := rl.GetLimiter("client1")
	// Should be the exact same pointer
	assert.True(t, limiter1 == limiter2, "Same client should get same limiter instance")

	// Test getting limiter for different client returns different instance
	limiter3 := rl.GetLimiter("client2")
	assert.True(t, limiter1 != limiter3, "Different clients should get different limiter instances")
}

func TestRateLimit_AllowedRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create rate limiter allowing 2 requests per second with burst of 2
	rl := gateway.NewRateLimiter(rate.Limit(2), 2)

	router := gin.New()
	router.Use(rl.RateLimit())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	// First request should succeed
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = "192.168.1.1:8080"
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Second request should succeed (burst allows)
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "192.168.1.1:8080"
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
}

func TestRateLimit_ExceedsLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create very restrictive rate limiter: 1 request per second, burst of 1
	rl := gateway.NewRateLimiter(rate.Limit(1), 1)

	router := gin.New()
	router.Use(rl.RateLimit())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	// First request should succeed
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = "192.168.1.1:8080"
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Second request should be rate limited
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "192.168.1.1:8080"
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)

	// Verify error response
	assert.Contains(t, w2.Body.String(), "rate limit exceeded")
	assert.Contains(t, w2.Body.String(), "retry_after")
}

func TestRateLimit_DifferentClients(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create rate limiter: 1 request per second, burst of 1
	rl := gateway.NewRateLimiter(rate.Limit(1), 1)

	router := gin.New()
	router.Use(rl.RateLimit())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	// Client 1 first request should succeed
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = "192.168.1.1:8080"
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Client 2 first request should also succeed (different limiter)
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "192.168.1.2:8080"
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)

	// Client 1 second request should be rate limited
	req3 := httptest.NewRequest("GET", "/test", nil)
	req3.RemoteAddr = "192.168.1.1:8080"
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusTooManyRequests, w3.Code)

	// Client 2 second request should also be rate limited
	req4 := httptest.NewRequest("GET", "/test", nil)
	req4.RemoteAddr = "192.168.1.2:8080"
	w4 := httptest.NewRecorder()
	router.ServeHTTP(w4, req4)
	assert.Equal(t, http.StatusTooManyRequests, w4.Code)
}

func TestCreateRateLimitMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	middleware := gateway.CreateRateLimitMiddleware(10, 5)
	assert.NotNil(t, middleware)

	router := gin.New()
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	// Should allow initial requests up to burst limit
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:8080"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code, "Request %d should succeed", i+1)
	}

	// Next request should be rate limited
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:8080"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}

func TestRateLimit_RecoveryAfterTime(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create rate limiter with 10 requests per second, burst of 1
	// This means after burst, we need to wait 0.1 seconds for next token
	rl := gateway.NewRateLimiter(rate.Limit(10), 1)

	router := gin.New()
	router.Use(rl.RateLimit())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	// First request should succeed
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = "192.168.1.1:8080"
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Second request should be rate limited
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "192.168.1.1:8080"
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)

	// Wait for rate limiter to recover (0.11 seconds to be safe)
	time.Sleep(110 * time.Millisecond)

	// Third request should succeed again
	req3 := httptest.NewRequest("GET", "/test", nil)
	req3.RemoteAddr = "192.168.1.1:8080"
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusOK, w3.Code)
}

func TestRateLimit_ConcurrentClients(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create rate limiter: 5 requests per second, burst of 2
	rl := gateway.NewRateLimiter(rate.Limit(5), 2)

	router := gin.New()
	router.Use(rl.RateLimit())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	// Test multiple clients can each use their burst allowance
	clients := []string{"192.168.1.1:8080", "192.168.1.2:8080", "192.168.1.3:8080"}

	for _, clientAddr := range clients {
		// Each client should be able to make 2 requests (burst limit)
		for i := 0; i < 2; i++ {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = clientAddr
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code, "Client %s request %d should succeed", clientAddr, i+1)
		}

		// Third request should be rate limited for each client
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = clientAddr
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusTooManyRequests, w.Code, "Client %s third request should be rate limited", clientAddr)
	}
}
