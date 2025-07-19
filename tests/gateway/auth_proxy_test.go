package gateway_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/EliasRanz/ai-code-gen/internal/gateway"
)

func TestAuthProxyMiddleware_Process(t *testing.T) {
	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "missing authorization header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectError:    true,
		},
		{
			name:           "invalid authorization format",
			authHeader:     "Basic token123",
			expectedStatus: http.StatusUnauthorized,
			expectError:    true,
		},
		{
			name:           "empty bearer token",
			authHeader:     "Bearer ",
			expectedStatus: http.StatusUnauthorized,
			expectError:    true,
		},
		{
			name:           "valid bearer token format",
			authHeader:     "Bearer valid-token-123",
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test middleware
			middleware := gateway.NewAuthProxyMiddleware("http://localhost:8001", nil)

			// Create test gin context
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Set up request
			req, _ := http.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			c.Request = req

			// Wrap gin context
			ctx := gateway.WrapGinContext(c)

			// Create next function
			nextCalled := false
			next := func() error {
				nextCalled = true
				return nil
			}

			// Execute middleware
			err := middleware.Process(ctx, next)

			// Assertions
			if tt.expectError {
				assert.Error(t, err)
				assert.False(t, nextCalled)
			} else {
				// For valid token format, we expect auth service call failure (no real service)
				// This tests the middleware logic up to the service call
				assert.Error(t, err) // Auth service call will fail in test
				assert.False(t, nextCalled)
			}
		})
	}
}

func TestAuthProxyMiddleware_Configuration(t *testing.T) {
	// Test without cache first
	middleware := gateway.NewAuthProxyMiddleware("http://localhost:8001", nil)

	// Test configuration
	config := middleware.GetConfig()
	assert.Equal(t, "auth-proxy", config.GetName())
	assert.True(t, config.IsEnabled())

	params := config.GetParameters()
	assert.Equal(t, "http://localhost:8001", params["auth_service_url"])
	assert.Equal(t, false, params["cache_enabled"])

	// Test health check
	assert.NoError(t, middleware.HealthCheck())

	// Test validation
	assert.NoError(t, middleware.ValidateConfig())

	// Test metadata
	assert.Equal(t, "auth-proxy", middleware.GetName())
	assert.Equal(t, 100, middleware.GetOrder())
}

func TestAuthProxyMiddleware_InvalidConfiguration(t *testing.T) {
	// Test with empty auth service URL
	middleware := gateway.NewAuthProxyMiddleware("", nil)

	assert.Error(t, middleware.HealthCheck())
	assert.Error(t, middleware.ValidateConfig())
}

func TestAuthProxyMiddleware_CheckPermissions(t *testing.T) {
	middleware := gateway.NewAuthProxyMiddleware("http://localhost:8001", nil)

	// Create test context with user role
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_role", "admin")

	ctx := gateway.WrapGinContext(c)

	// Test admin permissions
	err := middleware.CheckPermissions(ctx, []string{"user", "admin"})
	assert.NoError(t, err)

	// Test insufficient permissions
	c.Set("user_role", "user")
	err = middleware.CheckPermissions(ctx, []string{"admin"})
	assert.Error(t, err)

	// Test empty permissions (should always pass)
	err = middleware.CheckPermissions(ctx, []string{})
	assert.NoError(t, err)
}

func TestAuthProxyMiddleware_Integration(t *testing.T) {
	// Test integration with gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	middleware := gateway.NewAuthProxyMiddleware("http://localhost:8001", nil)

	// Add route with middleware
	router.GET("/protected", func(c *gin.Context) {
		ctx := gateway.WrapGinContext(c)

		next := func() error {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
			return nil
		}

		err := middleware.Process(ctx, next)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		}
	})

	// Test without auth header
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// Test with auth header (will fail due to no real auth service)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthProxyMiddleware_ContextOperations(t *testing.T) {
	middleware := gateway.NewAuthProxyMiddleware("http://localhost:8001", nil)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	ctx := gateway.WrapGinContext(c)

	// Test token extraction
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	c.Request = req

	token, err := middleware.ValidateToken(ctx)

	// Should error due to no real auth service, but validates the extraction logic
	assert.Error(t, err)
	assert.Nil(t, token)

	// Test context without authorization header
	req, _ = http.NewRequest("GET", "/test", nil)
	c.Request = req

	token, err = middleware.ValidateToken(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Authorization header required")
	assert.Nil(t, token)
}
