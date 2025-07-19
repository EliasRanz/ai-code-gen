package gateway_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/EliasRanz/ai-code-gen/internal/gateway"
)

func TestMiddlewareFactory_CreateMiddleware(t *testing.T) {
	factory := gateway.NewMiddlewareFactory("http://localhost:8001", nil)

	tests := []struct {
		name           string
		middlewareType string
		expectError    bool
	}{
		{
			name:           "create auth-proxy middleware",
			middlewareType: "auth-proxy",
			expectError:    false,
		},
		{
			name:           "create logging middleware",
			middlewareType: "logging",
			expectError:    false,
		},
		{
			name:           "create rate-limit middleware",
			middlewareType: "rate-limit",
			expectError:    false,
		},
		{
			name:           "unknown middleware type",
			middlewareType: "unknown",
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := gateway.NewBasicMiddlewareConfig(tt.middlewareType, true, map[string]interface{}{
				"requests_per_second": 10,
				"burst":               5,
			})

			middleware, err := factory.CreateMiddleware(tt.middlewareType, config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, middleware)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, middleware)
				assert.Equal(t, tt.middlewareType, middleware.GetName())
			}
		})
	}
}

func TestMiddlewareFactory_CreateChain(t *testing.T) {
	factory := gateway.NewMiddlewareFactory("http://localhost:8001", nil)

	// Create middleware instances
	authConfig := gateway.NewBasicMiddlewareConfig("auth-proxy", true, nil)
	authMiddleware, err := factory.CreateMiddleware("auth-proxy", authConfig)
	require.NoError(t, err)

	loggingConfig := gateway.NewBasicMiddlewareConfig("logging", true, nil)
	loggingMiddleware, err := factory.CreateMiddleware("logging", loggingConfig)
	require.NoError(t, err)

	middlewares := []gateway.Middleware{authMiddleware, loggingMiddleware}

	// Create chain
	chain := factory.CreateChain(middlewares)
	assert.NotNil(t, chain)

	// Verify chain contains both middleware
	chainMiddlewares := chain.GetMiddleware()
	assert.Len(t, chainMiddlewares, 2)

	// Should be ordered by priority (logging first with order 10, auth second with order 100)
	assert.Equal(t, "logging", chainMiddlewares[0].GetName())
	assert.Equal(t, "auth-proxy", chainMiddlewares[1].GetName())
}

func TestMiddlewareFactory_ListAvailableMiddleware(t *testing.T) {
	factory := gateway.NewMiddlewareFactory("http://localhost:8001", nil)

	available := factory.ListAvailableMiddleware()

	expectedTypes := []string{"auth-proxy", "logging", "rate-limit"}
	assert.ElementsMatch(t, expectedTypes, available)
}

func TestMiddlewareFactory_WithAuthCache(t *testing.T) {
	// Test factory without auth cache to avoid Redis connection in tests
	factory := gateway.NewMiddlewareFactory("http://localhost:8001", nil)

	config := gateway.NewBasicMiddlewareConfig("auth-proxy", true, nil)
	middleware, err := factory.CreateMiddleware("auth-proxy", config)

	require.NoError(t, err)
	assert.NotNil(t, middleware)

	// Verify configuration without cache
	middlewareConfig := middleware.GetConfig()
	params := middlewareConfig.GetParameters()
	assert.Equal(t, false, params["cache_enabled"])
}

func TestBasicMiddlewareConfig(t *testing.T) {
	params := map[string]interface{}{
		"requests_per_second": 100,
		"burst":               10,
		"timeout":             "5s",
	}

	config := gateway.NewBasicMiddlewareConfig("test-middleware", true, params)

	assert.Equal(t, "test-middleware", config.GetName())
	assert.True(t, config.IsEnabled())
	assert.Equal(t, params, config.GetParameters())
}

func TestBasicMiddlewareConfig_NilParameters(t *testing.T) {
	config := gateway.NewBasicMiddlewareConfig("test", false, nil)

	assert.Equal(t, "test", config.GetName())
	assert.False(t, config.IsEnabled())
	assert.NotNil(t, config.GetParameters())
	assert.Empty(t, config.GetParameters())
}

func TestMiddlewareChain_Add(t *testing.T) {
	chain := gateway.NewMiddlewareChain()

	// Create test middleware with different orders
	logging := gateway.NewLoggingMiddleware()                            // order 10
	rateLimit := gateway.NewRateLimitMiddleware(10, 5)                   // order 20
	auth := gateway.NewAuthProxyMiddleware("http://localhost:8001", nil) // order 100

	// Add in random order
	chain.Add(auth)
	chain.Add(logging)
	chain.Add(rateLimit)

	middlewares := chain.GetMiddleware()
	assert.Len(t, middlewares, 3)

	// Should be sorted by order
	assert.Equal(t, "logging", middlewares[0].GetName())    // order 10
	assert.Equal(t, "rate-limit", middlewares[1].GetName()) // order 20
	assert.Equal(t, "auth-proxy", middlewares[2].GetName()) // order 100
}

func TestMiddlewareChain_Execute(t *testing.T) {
	chain := gateway.NewMiddlewareChain()

	// Create test middleware
	logging := gateway.NewLoggingMiddleware()
	chain.Add(logging)

	// Create test context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest("GET", "/test", nil)
	c.Request = req

	ctx := gateway.WrapGinContext(c)

	// Execute chain
	err := chain.Execute(ctx)

	// Should complete without error (logging middleware doesn't fail)
	assert.NoError(t, err)
}

func TestMiddlewareChain_ExecuteEmpty(t *testing.T) {
	chain := gateway.NewMiddlewareChain()

	// Create test context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest("GET", "/test", nil)
	c.Request = req

	ctx := gateway.WrapGinContext(c)

	// Execute empty chain
	err := chain.Execute(ctx)

	// Should complete without error
	assert.NoError(t, err)
}

func TestMiddlewareChain_ConcurrentAccess(t *testing.T) {
	chain := gateway.NewMiddlewareChain()

	// Test concurrent add operations
	logging := gateway.NewLoggingMiddleware()
	rateLimit := gateway.NewRateLimitMiddleware(10, 5)

	// Add middleware concurrently
	go func() {
		chain.Add(logging)
	}()
	go func() {
		chain.Add(rateLimit)
	}()

	// Give goroutines time to complete
	time.Sleep(10 * time.Millisecond)

	middlewares := chain.GetMiddleware()
	assert.Len(t, middlewares, 2)
}
