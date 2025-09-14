package gateway_auth_proxy_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/EliasRanz/ai-code-gen/internal/gateway"
)

// Test server helper for auth service mocking
func setupAuthServiceMock(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	server := httptest.NewServer(handler)
	t.Cleanup(func() {
		server.Close()
	})
	return server
}

// Tests for NewAuthProxyMiddleware
func TestNewAuthProxyMiddleware(t *testing.T) {
	t.Run("create without cache", func(t *testing.T) {
		authServiceURL := "http://auth-service:8080"

		middleware := gateway.NewAuthProxyMiddleware(authServiceURL, nil)

		assert.NotNil(t, middleware)
		assert.Equal(t, "auth-proxy", middleware.GetName())
		assert.Equal(t, 100, middleware.GetOrder())
	})

	t.Run("get configuration", func(t *testing.T) {
		middleware := gateway.NewAuthProxyMiddleware("http://auth-service:8080", nil)

		config := middleware.GetConfig()

		assert.NotNil(t, config)
		assert.True(t, config.IsEnabled())
		assert.Equal(t, "auth-proxy", config.GetName())

		params := config.GetParameters()
		assert.Equal(t, "http://auth-service:8080", params["auth_service_url"])
		assert.Equal(t, false, params["cache_enabled"])
	})
}

// Tests for configuration and health checks
func TestAuthProxyMiddleware_Configuration(t *testing.T) {
	t.Run("health check success", func(t *testing.T) {
		middleware := gateway.NewAuthProxyMiddleware("http://auth-service:8080", nil)

		err := middleware.HealthCheck()

		assert.NoError(t, err)
	})

	t.Run("health check failure - empty URL", func(t *testing.T) {
		middleware := gateway.NewAuthProxyMiddleware("", nil)

		err := middleware.HealthCheck()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "auth service URL not configured")
	})

	t.Run("validate configuration success", func(t *testing.T) {
		middleware := gateway.NewAuthProxyMiddleware("http://auth-service:8080", nil)

		err := middleware.ValidateConfig()

		assert.NoError(t, err)
	})

	t.Run("validate configuration failure", func(t *testing.T) {
		middleware := gateway.NewAuthProxyMiddleware("", nil)

		err := middleware.ValidateConfig()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "auth service URL is required")
	})
}

// Tests for ValidateToken with real gin context
func TestAuthProxyMiddleware_ValidateToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful validation with auth service", func(t *testing.T) {
		authServer := setupAuthServiceMock(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/api/auth/validate", r.URL.Path)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			var requestBody map[string]string
			err := json.NewDecoder(r.Body).Decode(&requestBody)
			assert.NoError(t, err)
			assert.Equal(t, "valid-token", requestBody["access_token"])

			userContext := gateway.UserContext{
				UserID: "user123",
				Email:  "test@example.com",
				Role:   "user",
				Active: true,
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(userContext)
		})

		middleware := gateway.NewAuthProxyMiddleware(authServer.URL, nil)

		// Create gin context with authorization header
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer valid-token")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		ctx := gateway.WrapGinContext(c)

		userContext, err := middleware.ValidateToken(ctx)

		assert.NoError(t, err)
		assert.NotNil(t, userContext)
		assert.Equal(t, "user123", userContext.UserID)
		assert.Equal(t, "test@example.com", userContext.Email)
		assert.Equal(t, "user", userContext.Role)
		assert.True(t, userContext.Active)
	})

	t.Run("missing authorization header", func(t *testing.T) {
		middleware := gateway.NewAuthProxyMiddleware("http://auth-service:8080", nil)

		req, _ := http.NewRequest("GET", "/test", nil)
		// No authorization header

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		ctx := gateway.WrapGinContext(c)

		userContext, err := middleware.ValidateToken(ctx)

		assert.Error(t, err)
		assert.Nil(t, userContext)
		assert.Contains(t, err.Error(), "Authorization header required")
	})

	t.Run("invalid authorization header format", func(t *testing.T) {
		middleware := gateway.NewAuthProxyMiddleware("http://auth-service:8080", nil)

		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Basic invalid-format")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		ctx := gateway.WrapGinContext(c)

		userContext, err := middleware.ValidateToken(ctx)

		assert.Error(t, err)
		assert.Nil(t, userContext)
		assert.Contains(t, err.Error(), "Invalid authorization header format")
	})

	t.Run("empty token", func(t *testing.T) {
		middleware := gateway.NewAuthProxyMiddleware("http://auth-service:8080", nil)

		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer ")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		ctx := gateway.WrapGinContext(c)

		userContext, err := middleware.ValidateToken(ctx)

		assert.Error(t, err)
		assert.Nil(t, userContext)
		assert.Contains(t, err.Error(), "Token not provided")
	})

	t.Run("auth service error", func(t *testing.T) {
		authServer := setupAuthServiceMock(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Invalid token"))
		})

		middleware := gateway.NewAuthProxyMiddleware(authServer.URL, nil)

		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		ctx := gateway.WrapGinContext(c)

		userContext, err := middleware.ValidateToken(ctx)

		assert.Error(t, err)
		assert.Nil(t, userContext)
		assert.Contains(t, err.Error(), "auth service returned status 401")
	})

	t.Run("malformed auth service response", func(t *testing.T) {
		authServer := setupAuthServiceMock(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"invalid": json`)) // Malformed JSON
		})

		middleware := gateway.NewAuthProxyMiddleware(authServer.URL, nil)

		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer malformed-token")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		ctx := gateway.WrapGinContext(c)

		userContext, err := middleware.ValidateToken(ctx)

		assert.Error(t, err)
		assert.Nil(t, userContext)
		assert.Contains(t, err.Error(), "failed to decode auth service response")
	})
}

// Tests for CheckPermissions with real gin context
func TestAuthProxyMiddleware_CheckPermissions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("no permissions required", func(t *testing.T) {
		middleware := gateway.NewAuthProxyMiddleware("http://auth-service:8080", nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		ctx := gateway.WrapGinContext(c)

		err := middleware.CheckPermissions(ctx, []string{})

		assert.NoError(t, err)
	})

	t.Run("admin has all permissions", func(t *testing.T) {
		middleware := gateway.NewAuthProxyMiddleware("http://auth-service:8080", nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user_role", "admin")
		ctx := gateway.WrapGinContext(c)

		err := middleware.CheckPermissions(ctx, []string{"read", "write", "delete"})

		assert.NoError(t, err)
	})

	t.Run("user has matching permission", func(t *testing.T) {
		middleware := gateway.NewAuthProxyMiddleware("http://auth-service:8080", nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user_role", "user")
		ctx := gateway.WrapGinContext(c)

		err := middleware.CheckPermissions(ctx, []string{"read", "user", "write"})

		assert.NoError(t, err)
	})

	t.Run("insufficient permissions", func(t *testing.T) {
		middleware := gateway.NewAuthProxyMiddleware("http://auth-service:8080", nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user_role", "guest")
		ctx := gateway.WrapGinContext(c)

		err := middleware.CheckPermissions(ctx, []string{"admin", "write"})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient permissions")
		assert.Contains(t, err.Error(), "guest")
	})

	t.Run("user role not found", func(t *testing.T) {
		middleware := gateway.NewAuthProxyMiddleware("http://auth-service:8080", nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		ctx := gateway.WrapGinContext(c)

		err := middleware.CheckPermissions(ctx, []string{"read"})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user role not found in context")
	})

	t.Run("invalid role type", func(t *testing.T) {
		middleware := gateway.NewAuthProxyMiddleware("http://auth-service:8080", nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user_role", 123) // Invalid type
		ctx := gateway.WrapGinContext(c)

		err := middleware.CheckPermissions(ctx, []string{"read"})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid user role type")
	})
}

// Tests for Process method with full middleware workflow
func TestAuthProxyMiddleware_Process(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful authentication flow", func(t *testing.T) {
		authServer := setupAuthServiceMock(t, func(w http.ResponseWriter, r *http.Request) {
			userContext := gateway.UserContext{
				UserID: "flow-user",
				Email:  "flow@example.com",
				Role:   "user",
				Active: true,
			}
			json.NewEncoder(w).Encode(userContext)
		})

		middleware := gateway.NewAuthProxyMiddleware(authServer.URL, nil)

		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer flow-token")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		ctx := gateway.WrapGinContext(c)

		nextCalled := false
		next := func() error {
			nextCalled = true
			return nil
		}

		err := middleware.Process(ctx, next)

		assert.NoError(t, err)
		assert.True(t, nextCalled)

		// Verify user context was set
		userID, exists := c.Get("user_id")
		assert.True(t, exists)
		assert.Equal(t, "flow-user", userID)

		userEmail, exists := c.Get("user_email")
		assert.True(t, exists)
		assert.Equal(t, "flow@example.com", userEmail)

		userRole, exists := c.Get("user_role")
		assert.True(t, exists)
		assert.Equal(t, "user", userRole)

		authenticated, exists := c.Get("authenticated")
		assert.True(t, exists)
		assert.True(t, authenticated.(bool))
	})

	t.Run("authentication failure - no authorization header", func(t *testing.T) {
		middleware := gateway.NewAuthProxyMiddleware("http://auth-service:8080", nil)

		req, _ := http.NewRequest("GET", "/protected", nil)
		// No authorization header

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		ctx := &gateway.GinContextWrapper{Context: c}

		nextCalled := false
		next := func() error {
			nextCalled = true
			return nil
		}

		err := middleware.Process(ctx, next)

		assert.Error(t, err)
		assert.False(t, nextCalled)
		assert.Contains(t, err.Error(), "auth error")

		// Verify gin context was aborted and error response sent
		assert.True(t, c.IsAborted())
	})

	t.Run("authentication failure - invalid token", func(t *testing.T) {
		authServer := setupAuthServiceMock(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		})

		middleware := gateway.NewAuthProxyMiddleware(authServer.URL, nil)

		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		ctx := &gateway.GinContextWrapper{Context: c}

		nextCalled := false
		next := func() error {
			nextCalled = true
			return nil
		}

		err := middleware.Process(ctx, next)

		assert.Error(t, err)
		assert.False(t, nextCalled)
		assert.Contains(t, err.Error(), "auth error")

		// Verify gin context was aborted
		assert.True(t, c.IsAborted())
	})
}

// Integration test with complete gin router
func TestAuthProxyMiddleware_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("complete authentication flow with gin router", func(t *testing.T) {
		authServer := setupAuthServiceMock(t, func(w http.ResponseWriter, r *http.Request) {
			userContext := gateway.UserContext{
				UserID: "integration-user",
				Email:  "integration@example.com",
				Role:   "user",
				Active: true,
			}
			json.NewEncoder(w).Encode(userContext)
		})

		router := gin.New()
		middleware := gateway.NewAuthProxyMiddleware(authServer.URL, nil)

		router.GET("/protected", func(c *gin.Context) {
			ctx := gateway.WrapGinContext(c)

			err := middleware.Process(ctx, func() error {
				// Simulate protected endpoint logic
				userID, _ := c.Get("user_id")
				c.JSON(200, gin.H{
					"message": "success",
					"user_id": userID,
				})
				return nil
			})

			if err != nil {
				// Error already handled by middleware
				return
			}
		})

		// Test with valid token
		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer integration-token")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "success", response["message"])
		assert.Equal(t, "integration-user", response["user_id"])
	})

	t.Run("authentication failure returns 401", func(t *testing.T) {
		router := gin.New()
		middleware := gateway.NewAuthProxyMiddleware("http://nonexistent:8080", nil)

		router.GET("/protected", func(c *gin.Context) {
			ctx := &gateway.GinContextWrapper{Context: c}

			middleware.Process(ctx, func() error {
				c.JSON(200, gin.H{"message": "should not reach here"})
				return nil
			})
		})

		// Test with invalid token
		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, 401, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Authentication failed", response["error"])
	})
}
