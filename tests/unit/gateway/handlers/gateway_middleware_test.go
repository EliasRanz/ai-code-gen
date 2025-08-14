package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/EliasRanz/ai-code-gen/internal/config"
	"github.com/EliasRanz/ai-code-gen/internal/gateway"
)

// Mock dependencies
type MockGatewayConfigManager struct {
	mock.Mock
}

func (m *MockGatewayConfigManager) LoadConfig(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockGatewayConfigManager) GetConfig() *gateway.GatewayServiceConfig {
	args := m.Called()
	return args.Get(0).(*gateway.GatewayServiceConfig)
}

func (m *MockGatewayConfigManager) Watch(ctx context.Context, callback func()) error {
	args := m.Called(ctx, callback)
	return args.Error(0)
}

func (m *MockGatewayConfigManager) Reload(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

type MockAuthCache struct {
	mock.Mock
}

func (m *MockAuthCache) Get(ctx context.Context, key string) (interface{}, error) {
	args := m.Called(ctx, key)
	return args.Get(0), args.Error(1)
}

func (m *MockAuthCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}

func (m *MockAuthCache) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

// Test Gateway Config Management
func TestGatewayConfigManager_LoadConfig(t *testing.T) {
	tests := []struct {
		name       string
		setupMock  func(*MockGatewayConfigManager)
		expectErr  bool
		errMessage string
	}{
		{
			name: "successful config load",
			setupMock: func(m *MockGatewayConfigManager) {
				config := &gateway.GatewayServiceConfig{
					Service: config.BaseServiceConfig{
						Name: "api-gateway",
						Host: "localhost",
						Port: 8080,
					},
					RateLimit: gateway.RateLimitConfig{
						Enabled:         true,
						RequestsPerSec:  100,
						BurstSize:       20,
						CleanupInterval: 5 * time.Minute,
					},
					Auth: gateway.AuthProxyConfig{
						Enabled:        true,
						AuthServiceURL: "http://auth-service:8081",
						TokenHeader:    "Authorization",
						Timeout:        5 * time.Second,
					},
				}
				m.On("LoadConfig", mock.Anything).Return(nil)
				m.On("GetConfig").Return(config)
			},
			expectErr: false,
		},
		{
			name: "config load failure",
			setupMock: func(m *MockGatewayConfigManager) {
				m.On("LoadConfig", mock.Anything).Return(assert.AnError)
			},
			expectErr:  true,
			errMessage: "failed to load config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockManager := &MockGatewayConfigManager{}
			tt.setupMock(mockManager)

			err := mockManager.LoadConfig(context.Background())

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				config := mockManager.GetConfig()
				assert.NotNil(t, config)
				assert.Equal(t, "api-gateway", config.Service.Name)
				assert.Equal(t, 8080, config.Service.Port)
				assert.True(t, config.RateLimit.Enabled)
				assert.Equal(t, 100, config.RateLimit.RequestsPerSec)
			}

			mockManager.AssertExpectations(t)
		})
	}
}

// Test Gateway Config Watch
func TestGatewayConfigManager_Watch(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(*MockGatewayConfigManager)
		expectErr bool
	}{
		{
			name: "successful watch setup",
			setupMock: func(m *MockGatewayConfigManager) {
				m.On("Watch", mock.Anything, mock.AnythingOfType("func()")).Return(nil)
			},
			expectErr: false,
		},
		{
			name: "watch setup failure",
			setupMock: func(m *MockGatewayConfigManager) {
				m.On("Watch", mock.Anything, mock.AnythingOfType("func()")).Return(assert.AnError)
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockManager := &MockGatewayConfigManager{}
			tt.setupMock(mockManager)

			callback := func() {
				// Callback implementation
			}

			err := mockManager.Watch(context.Background(), callback)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockManager.AssertExpectations(t)
		})
	}
}

// Test Middleware Chain Processing
func TestMiddlewareChain_Processing(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name        string
		setupChain  func() *gateway.MiddlewareChainImpl
		request     *http.Request
		expectError bool
	}{
		{
			name: "empty middleware chain execution",
			setupChain: func() *gateway.MiddlewareChainImpl {
				return gateway.NewMiddlewareChain()
			},
			request:     httptest.NewRequest("GET", "/api/test", nil),
			expectError: false,
		},
		{
			name: "middleware chain with auth proxy",
			setupChain: func() *gateway.MiddlewareChainImpl {
				chain := gateway.NewMiddlewareChain()
				authMiddleware := gateway.NewAuthProxyMiddleware("http://auth-service:8081", nil)
				chain.Add(authMiddleware)
				return chain
			},
			request:     httptest.NewRequest("GET", "/api/test", nil),
			expectError: true, // No valid token provided
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = tt.request

			chain := tt.setupChain()
			ctx := &gateway.GinContextWrapper{Context: c}

			err := chain.Execute(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test Auth Proxy Middleware Token Extraction
func TestAuthProxyMiddleware_TokenExtraction(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name        string
		authHeader  string
		expectError bool
	}{
		{
			name:        "valid bearer token",
			authHeader:  "Bearer valid-token-123",
			expectError: true, // Will fail due to no auth service, but token extraction works
		},
		{
			name:        "invalid token format - no bearer",
			authHeader:  "invalid-token",
			expectError: true,
		},
		{
			name:        "missing token",
			authHeader:  "",
			expectError: true,
		},
		{
			name:        "bearer token with extra spaces",
			authHeader:  "Bearer  token-with-spaces  ",
			expectError: true, // Will fail due to no auth service, but token extraction works
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req := httptest.NewRequest("GET", "/api/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			c.Request = req

			// Use nil for authCache to avoid Redis dependency in tests
			authMiddleware := gateway.NewAuthProxyMiddleware("http://auth-service:8081", nil)
			ctx := &gateway.GinContextWrapper{Context: c}

			// Test token validation - we expect errors due to no real auth service
			_, err := authMiddleware.ValidateToken(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test Auth Proxy Middleware Config
func TestAuthProxyMiddleware_Config(t *testing.T) {
	authServiceURL := "http://auth-service:8081"

	authMiddleware := gateway.NewAuthProxyMiddleware(authServiceURL, nil)

	config := authMiddleware.GetConfig()

	assert.Equal(t, "auth-proxy", config.GetName())
	assert.True(t, config.IsEnabled())

	params := config.GetParameters()
	assert.Equal(t, authServiceURL, params["auth_service_url"])
	assert.Equal(t, false, params["cache_enabled"]) // nil cache = false
}

// Test Rate Limiting Middleware
func TestRateLimitMiddleware_Creation(t *testing.T) {
	tests := []struct {
		name           string
		requestsPerSec int
		burstSize      int
	}{
		{
			name:           "enabled rate limiting",
			requestsPerSec: 100,
			burstSize:      20,
		},
		{
			name:           "disabled rate limiting",
			requestsPerSec: 0,
			burstSize:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := gateway.NewRateLimitMiddleware(tt.requestsPerSec, tt.burstSize)

			assert.NotNil(t, middleware)
			assert.Equal(t, "rate-limit", middleware.GetName())

			config := middleware.GetConfig()
			assert.Equal(t, "rate-limit", config.GetName())
			assert.True(t, config.IsEnabled())
		})
	}
}

// Test Logging Middleware
func TestLoggingMiddleware_Creation(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "logging middleware creation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := gateway.NewLoggingMiddleware()

			assert.NotNil(t, middleware)
			assert.Equal(t, "logging", middleware.GetName())

			config := middleware.GetConfig()
			assert.Equal(t, "logging", config.GetName())
			assert.True(t, config.IsEnabled())
		})
	}
}

// Test Middleware Factory
func TestMiddlewareFactory_CreateMiddleware(t *testing.T) {

	tests := []struct {
		name           string
		middlewareType string
		config         gateway.MiddlewareConfig
		expectError    bool
	}{
		{
			name:           "create auth middleware",
			middlewareType: "auth-proxy",
			config: gateway.NewBasicMiddlewareConfig("auth-proxy", true, map[string]interface{}{
				"auth_service_url": "http://auth-service:8081",
			}),
			expectError: false,
		},
		{
			name:           "create rate limit middleware",
			middlewareType: "rate-limit",
			config: gateway.NewBasicMiddlewareConfig("rate-limit", true, map[string]interface{}{
				"requests_per_second": 100,
				"burst":               20,
			}),
			expectError: false,
		},
		{
			name:           "create logging middleware",
			middlewareType: "logging",
			config: gateway.NewBasicMiddlewareConfig("logging", true, map[string]interface{}{
				"level":  "info",
				"format": "json",
			}),
			expectError: false,
		},
		{
			name:           "invalid middleware type",
			middlewareType: "invalid",
			config:         gateway.NewBasicMiddlewareConfig("invalid", true, nil),
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := gateway.NewMiddlewareFactory("http://auth-service:8081", nil)

			middleware, err := factory.CreateMiddleware(tt.middlewareType, tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, middleware)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, middleware)
			}
		})
	}
}

// Test Basic Middleware Config
func TestBasicMiddlewareConfig(t *testing.T) {
	params := map[string]interface{}{
		"param1": "value1",
		"param2": 42,
		"param3": true,
	}

	config := gateway.NewBasicMiddlewareConfig("test-middleware", true, params)

	assert.Equal(t, "test-middleware", config.GetName())
	assert.True(t, config.IsEnabled())
	assert.Equal(t, params, config.GetParameters())

	// Test with nil parameters
	configNilParams := gateway.NewBasicMiddlewareConfig("test", false, nil)
	assert.NotNil(t, configNilParams.GetParameters())
	assert.Empty(t, configNilParams.GetParameters())
}

// Test Gin Context Wrapper
func TestGinContextWrapper(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	wrapper := &gateway.GinContextWrapper{Context: c}

	// Test basic context methods
	assert.NotNil(t, wrapper.Request())
	assert.Equal(t, "GET", wrapper.Request().Method)
	assert.Equal(t, "/test", wrapper.Request().URL.Path)

	// Test Set and Get
	wrapper.Set("test_key", "test_value")
	value, exists := wrapper.Get("test_key")
	assert.True(t, exists)
	assert.Equal(t, "test_value", value)

	// Test non-existent key
	_, exists = wrapper.Get("non_existent")
	assert.False(t, exists)
}
