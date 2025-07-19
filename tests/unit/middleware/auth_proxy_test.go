package middleware_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/EliasRanz/ai-code-gen/internal/cache"
	"github.com/EliasRanz/ai-code-gen/internal/gateway"
)

func TestAuthServiceProxy_NoAuthHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Mock auth service - not used in this test
	authServiceURL := "http://localhost:8001"
	router.Use(gateway.AuthServiceProxy(authServiceURL, nil))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Authorization header required", response["error"])
}

func TestAuthServiceProxy_InvalidAuthHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	authServiceURL := "http://localhost:8001"
	router.Use(gateway.AuthServiceProxy(authServiceURL, nil))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "InvalidFormat token123")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid authorization header format", response["error"])
}

func TestAuthServiceProxy_EmptyToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	authServiceURL := "http://localhost:8001"
	router.Use(gateway.AuthServiceProxy(authServiceURL, nil))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer ")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Token not provided", response["error"])
}

func TestAuthServiceProxy_ValidToken_MockAuthService(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create mock auth service
	mockAuthServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/auth/validate" && r.Method == "POST" {
			// Verify request
			var reqBody map[string]string
			json.NewDecoder(r.Body).Decode(&reqBody)

			if reqBody["access_token"] == "valid-token" {
				// Use the UserContext struct from middleware package
				response := map[string]interface{}{
					"user_id": "123",
					"email":   "test@example.com",
					"role":    "user",
					"active":  true,
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			} else {
				w.WriteHeader(http.StatusUnauthorized)
			}
		}
	}))
	defer mockAuthServer.Close()

	// Create test router
	router := gin.New()
	router.Use(gateway.AuthServiceProxy(mockAuthServer.URL, nil))
	router.GET("/test", func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		userEmail, _ := c.Get("user_email")
		userRole, _ := c.Get("user_role")
		authenticated, _ := c.Get("authenticated")

		c.JSON(200, gin.H{
			"message":       "success",
			"user_id":       userID,
			"user_email":    userEmail,
			"user_role":     userRole,
			"authenticated": authenticated,
		})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["message"])
	assert.Equal(t, "123", response["user_id"])
	assert.Equal(t, "test@example.com", response["user_email"])
	assert.Equal(t, "user", response["user_role"])
	assert.Equal(t, true, response["authenticated"])
}

func TestAuthServiceRoleProxy_ValidToken_ValidRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create mock auth service
	mockAuthServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/auth/check-role" && r.Method == "POST" {
			var reqBody map[string]string
			json.NewDecoder(r.Body).Decode(&reqBody)

			if reqBody["access_token"] == "admin-token" && reqBody["required_role"] == "admin" {
				w.WriteHeader(http.StatusOK)
			} else {
				w.WriteHeader(http.StatusForbidden)
			}
		} else if r.URL.Path == "/api/auth/validate" && r.Method == "POST" {
			var reqBody map[string]string
			json.NewDecoder(r.Body).Decode(&reqBody)

			if reqBody["access_token"] == "admin-token" {
				response := map[string]interface{}{
					"user_id": "123",
					"email":   "admin@example.com",
					"role":    "admin",
					"active":  true,
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			}
		}
	}))
	defer mockAuthServer.Close()

	// Create test router
	router := gin.New()
	router.Use(gateway.AuthServiceRoleProxy(mockAuthServer.URL, nil, "admin"))
	router.GET("/admin-test", func(c *gin.Context) {
		userRole, _ := c.Get("user_role")
		c.JSON(200, gin.H{
			"message":   "admin access granted",
			"user_role": userRole,
		})
	})

	req, _ := http.NewRequest("GET", "/admin-test", nil)
	req.Header.Set("Authorization", "Bearer admin-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "admin access granted", response["message"])
	assert.Equal(t, "admin", response["user_role"])
}

func TestAuthServiceRoleProxy_InvalidRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create mock auth service
	mockAuthServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/auth/check-role" && r.Method == "POST" {
			var reqBody map[string]string
			json.NewDecoder(r.Body).Decode(&reqBody)

			// Always deny role check for this test
			w.WriteHeader(http.StatusForbidden)
		} else if r.URL.Path == "/api/auth/validate" && r.Method == "POST" {
			var reqBody map[string]string
			json.NewDecoder(r.Body).Decode(&reqBody)

			if reqBody["access_token"] == "user-token" {
				response := map[string]interface{}{
					"user_id": "456",
					"email":   "user@example.com",
					"role":    "user",
					"active":  true,
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			}
		}
	}))
	defer mockAuthServer.Close()

	// Create test router
	router := gin.New()
	router.Use(gateway.AuthServiceRoleProxy(mockAuthServer.URL, nil, "admin"))
	router.GET("/admin-test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "should not reach here"})
	})

	req, _ := http.NewRequest("GET", "/admin-test", nil)
	req.Header.Set("Authorization", "Bearer user-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

// =============================================================================
// CACHING TESTS
// =============================================================================

func TestAuthServiceProxyWithCache_HitAndMiss(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup mini Redis server for testing
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	// Initialize auth cache
	authCache, err := cache.NewAuthCache("redis://"+mr.Addr(), 5*time.Minute)
	require.NoError(t, err)
	defer authCache.Close()

	// Setup mock auth service
	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/auth/validate" {
			userContext := gateway.UserContext{
				UserID: "user-123",
				Email:  "test@example.com",
				Role:   "user",
				Active: true,
			}
			json.NewEncoder(w).Encode(userContext)
		}
	}))
	defer authServer.Close()

	// Setup test router
	router := gin.New()
	router.Use(gateway.AuthServiceProxy(authServer.URL, authCache))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	// First request - should hit auth service and cache result
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.Header.Set("Authorization", "Bearer valid-token")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	assert.Equal(t, 200, w1.Code)

	// Second request - should use cache (verify by checking cache directly)
	tokenHash := cache.HashToken("valid-token")
	cachedContext, err := authCache.GetUserContext(context.Background(), tokenHash)
	require.NoError(t, err)
	require.NotNil(t, cachedContext)
	assert.Equal(t, "user-123", cachedContext.UserID)

	// Second request should also succeed
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.Header.Set("Authorization", "Bearer valid-token")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	assert.Equal(t, 200, w2.Code)
}

func TestAuthServiceProxyWithCache_GracefulFailover(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup mock auth service
	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/auth/validate" {
			userContext := gateway.UserContext{
				UserID: "user-456",
				Email:  "test2@example.com",
				Role:   "admin",
				Active: true,
			}
			json.NewEncoder(w).Encode(userContext)
		}
	}))
	defer authServer.Close()

	// Test with nil cache (cache not available)
	router := gin.New()
	router.Use(gateway.AuthServiceProxy(authServer.URL, nil))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should still work without cache
	assert.Equal(t, 200, w.Code)
}

func TestAuthServiceRoleProxyWithCache_AdminRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup mini Redis server for testing
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	// Initialize auth cache
	authCache, err := cache.NewAuthCache("redis://"+mr.Addr(), 5*time.Minute)
	require.NoError(t, err)
	defer authCache.Close()

	// Setup mock auth service
	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/auth/check-role" {
			w.WriteHeader(200) // Role check passes
		} else if r.URL.Path == "/api/auth/validate" {
			userContext := gateway.UserContext{
				UserID: "admin-123",
				Email:  "admin@example.com",
				Role:   "admin",
				Active: true,
			}
			json.NewEncoder(w).Encode(userContext)
		}
	}))
	defer authServer.Close()

	// Setup test router
	router := gin.New()
	router.Use(gateway.AuthServiceRoleProxy(authServer.URL, authCache, "admin"))
	router.GET("/admin", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "admin access granted"})
	})

	// First request - should validate role and cache result
	req1 := httptest.NewRequest("GET", "/admin", nil)
	req1.Header.Set("Authorization", "Bearer admin-token")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	assert.Equal(t, 200, w1.Code)

	// Verify cached
	tokenHash := cache.HashToken("admin-token")
	cachedContext, err := authCache.GetUserContext(context.Background(), tokenHash)
	require.NoError(t, err)
	require.NotNil(t, cachedContext)
	assert.Equal(t, "admin", cachedContext.Role)
}

func TestCacheInvalidation_UserTokenExpiry(t *testing.T) {
	// Setup mini Redis server for testing
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	// Initialize auth cache
	authCache, err := cache.NewAuthCache("redis://"+mr.Addr(), 5*time.Minute)
	require.NoError(t, err)
	defer authCache.Close()

	// Cache a user context
	token := "test-token"
	userContext := &cache.UserContext{
		UserID: "user-123",
		Email:  "test@example.com",
		Role:   "user",
	}

	tokenHash := cache.HashToken(token)
	err = authCache.SetUserContext(context.Background(), tokenHash, userContext)
	require.NoError(t, err)

	// Verify cached
	cached, err := authCache.GetUserContext(context.Background(), tokenHash)
	require.NoError(t, err)
	require.NotNil(t, cached)

	// Invalidate cache
	err = gateway.InvalidateUserCache(context.Background(), authCache, token)
	require.NoError(t, err)

	// Verify cache is cleared
	cached, err = authCache.GetUserContext(context.Background(), tokenHash)
	require.NoError(t, err)
	assert.Nil(t, cached)
}
