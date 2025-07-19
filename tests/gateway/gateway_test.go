package gateway_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/EliasRanz/ai-code-gen/internal/gateway"
)

func TestObservableGateway_New(t *testing.T) {
	// Reset metrics to prevent duplicate registration
	gateway.ResetMetricsForTesting()

	factory := gateway.NewMiddlewareFactory("http://localhost:8001", nil)

	observableGateway := gateway.NewObservableGateway(factory)

	assert.NotNil(t, observableGateway)

	// Should have default observers
	middleware := observableGateway.GetMiddleware()
	assert.Empty(t, middleware) // No middleware added yet

	// Should be able to perform health check
	err := observableGateway.HealthCheck()
	assert.NoError(t, err)
}

func TestObservableGateway_SetupMiddleware(t *testing.T) {
	// Reset metrics to prevent duplicate registration
	gateway.ResetMetricsForTesting()

	factory := gateway.NewMiddlewareFactory("http://localhost:8001", nil)
	observableGateway := gateway.NewObservableGateway(factory)

	// Create middleware configurations
	configs := []gateway.MiddlewareConfig{
		gateway.NewBasicMiddlewareConfig("logging", true, nil),
		gateway.NewBasicMiddlewareConfig("rate-limit", true, map[string]interface{}{
			"requests_per_second": 100,
			"burst":               10,
		}),
		gateway.NewBasicMiddlewareConfig("auth-proxy", true, nil),
	}

	err := observableGateway.SetupMiddleware(configs)
	require.NoError(t, err)

	// Verify middleware were added
	middleware := observableGateway.GetMiddleware()
	assert.Len(t, middleware, 3)

	// Should be ordered by priority
	assert.Equal(t, "logging", middleware[0].GetName())
	assert.Equal(t, "rate-limit", middleware[1].GetName())
	assert.Equal(t, "auth-proxy", middleware[2].GetName())
}

func TestObservableGateway_SetupMiddleware_InvalidType(t *testing.T) {
	factory := gateway.NewMiddlewareFactory("http://localhost:8001", nil)
	observableGateway := gateway.NewObservableGateway(factory)

	configs := []gateway.MiddlewareConfig{
		gateway.NewBasicMiddlewareConfig("invalid-type", true, nil),
	}

	err := observableGateway.SetupMiddleware(configs)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown middleware type")
}

func TestObservableGateway_AddRemoveObserver(t *testing.T) {
	factory := gateway.NewMiddlewareFactory("http://localhost:8001", nil)
	observableGateway := gateway.NewObservableGateway(factory)

	// Create custom observer
	customObserver := gateway.NewMetricsObserver()

	// Add observer
	err := observableGateway.AddObserver(customObserver)
	assert.NoError(t, err)

	// Remove observer
	err = observableGateway.RemoveObserver(customObserver)
	assert.NoError(t, err)
}

func TestObservableGateway_HealthCheck(t *testing.T) {
	factory := gateway.NewMiddlewareFactory("http://localhost:8001", nil)
	observableGateway := gateway.NewObservableGateway(factory)

	// Add healthy middleware
	configs := []gateway.MiddlewareConfig{
		gateway.NewBasicMiddlewareConfig("logging", true, nil),
		gateway.NewBasicMiddlewareConfig("rate-limit", true, map[string]interface{}{
			"requests_per_second": 10,
			"burst":               5,
		}),
	}

	err := observableGateway.SetupMiddleware(configs)
	require.NoError(t, err)

	// Health check should pass
	err = observableGateway.HealthCheck()
	assert.NoError(t, err)
}

func TestObservableGateway_HealthCheck_FailedMiddleware(t *testing.T) {
	factory := gateway.NewMiddlewareFactory("http://localhost:8001", nil)
	observableGateway := gateway.NewObservableGateway(factory)

	// Add unhealthy middleware (invalid rate limit config)
	configs := []gateway.MiddlewareConfig{
		gateway.NewBasicMiddlewareConfig("rate-limit", true, map[string]interface{}{
			"requests_per_second": 0, // Invalid configuration
			"burst":               0, // Invalid configuration
		}),
	}

	err := observableGateway.SetupMiddleware(configs)
	require.NoError(t, err)

	// Health check should fail
	err = observableGateway.HealthCheck()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rate-limit")
}

func TestObservableGateway_CreateGinMiddleware(t *testing.T) {
	factory := gateway.NewMiddlewareFactory("http://localhost:8001", nil)
	observableGateway := gateway.NewObservableGateway(factory)

	// Setup middleware
	configs := []gateway.MiddlewareConfig{
		gateway.NewBasicMiddlewareConfig("logging", true, nil),
	}

	err := observableGateway.SetupMiddleware(configs)
	require.NoError(t, err)

	// Create Gin middleware
	ginMiddleware := observableGateway.CreateGinMiddleware()
	assert.NotNil(t, ginMiddleware)

	// Test with Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(ginMiddleware)

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Make test request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Should have request ID header from logging middleware
	assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
}

func TestObservableGateway_ProcessRequest_WithObservers(t *testing.T) {
	factory := gateway.NewMiddlewareFactory("http://localhost:8001", nil)
	observableGateway := gateway.NewObservableGateway(factory)

	// Setup minimal middleware (just logging)
	configs := []gateway.MiddlewareConfig{
		gateway.NewBasicMiddlewareConfig("logging", true, nil),
	}

	err := observableGateway.SetupMiddleware(configs)
	require.NoError(t, err)

	// Create test gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest("GET", "/api/test", nil)
	c.Request = req
	c.Status(http.StatusOK)

	// Create context with gin context
	ctx := context.WithValue(context.Background(), "gin_context", c)

	// Create test request
	request := &gateway.HTTPRequest{
		Method:    "GET",
		Path:      "/api/test",
		Headers:   map[string]string{"User-Agent": "test"},
		StartTime: time.Now(),
		ClientIP:  "127.0.0.1",
	}

	// Process request
	response, err := observableGateway.ProcessRequest(ctx, request)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusOK, response.StatusCode)
}

func TestObservableGateway_Integration(t *testing.T) {
	// Test without auth cache to avoid Redis connection issues in tests
	factory := gateway.NewMiddlewareFactory("http://localhost:8001", nil)
	observableGateway := gateway.NewObservableGateway(factory)

	// Setup comprehensive middleware stack
	configs := []gateway.MiddlewareConfig{
		gateway.NewBasicMiddlewareConfig("logging", true, nil),
		gateway.NewBasicMiddlewareConfig("rate-limit", true, map[string]interface{}{
			"requests_per_second": 1000, // High limit for test
			"burst":               100,
		}),
	}

	err := observableGateway.SetupMiddleware(configs)
	require.NoError(t, err)

	// Create Gin router with the gateway
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(observableGateway.CreateGinMiddleware())

	// Add test routes
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"time":   time.Now().Unix(),
		})
	})

	router.GET("/api/users", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"users": []string{"user1", "user2"},
		})
	})

	// Test health endpoint
	t.Run("health check", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/health", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
		assert.Contains(t, w.Body.String(), "healthy")
	})

	// Test API endpoint
	t.Run("api endpoint", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/users", nil)
		req.Header.Set("User-Agent", "test-client/1.0")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
		assert.Contains(t, w.Body.String(), "users")
	})

	// Test multiple requests (rate limiting)
	t.Run("rate limiting", func(t *testing.T) {
		// Make several requests rapidly
		for i := 0; i < 5; i++ {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/health", nil)
			router.ServeHTTP(w, req)

			// All should succeed due to high rate limit
			assert.Equal(t, http.StatusOK, w.Code)
		}
	})

	// Verify gateway health
	assert.NoError(t, observableGateway.HealthCheck())
}

func TestObservableGateway_ConcurrentRequests(t *testing.T) {
	factory := gateway.NewMiddlewareFactory("http://localhost:8001", nil)
	observableGateway := gateway.NewObservableGateway(factory)

	configs := []gateway.MiddlewareConfig{
		gateway.NewBasicMiddlewareConfig("logging", true, nil),
	}

	err := observableGateway.SetupMiddleware(configs)
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(observableGateway.CreateGinMiddleware())

	router.GET("/test", func(c *gin.Context) {
		time.Sleep(10 * time.Millisecond) // Simulate processing time
		c.JSON(http.StatusOK, gin.H{"id": c.GetString("request_id")})
	})

	// Make concurrent requests
	const numRequests = 10
	results := make(chan int, numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			router.ServeHTTP(w, req)
			results <- w.Code
		}()
	}

	// Collect results
	for i := 0; i < numRequests; i++ {
		select {
		case code := <-results:
			assert.Equal(t, http.StatusOK, code)
		case <-time.After(time.Second):
			t.Fatal("Request timed out")
		}
	}
}
