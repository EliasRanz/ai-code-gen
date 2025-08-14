package gateway_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/EliasRanz/ai-code-gen/internal/config"
	"github.com/EliasRanz/ai-code-gen/internal/gateway"
)

// MockConfigProvider implements config.ConfigProvider for testing
type MockConfigProvider struct {
	mock.Mock
	data map[string]interface{}
}

func NewMockConfigProvider() *MockConfigProvider {
	return &MockConfigProvider{
		data: make(map[string]interface{}),
	}
}

func (m *MockConfigProvider) Load(ctx context.Context) (config.ConfigData, error) {
	args := m.Called(ctx)
	return config.ConfigData(m.data), args.Error(0)
}

func (m *MockConfigProvider) Watch(ctx context.Context, callback func(config.ConfigData)) error {
	args := m.Called(ctx, callback)
	return args.Error(0)
}

func (m *MockConfigProvider) Get(ctx context.Context, key string) (interface{}, error) {
	args := m.Called(ctx, key)
	return m.data[key], args.Error(0)
}

func (m *MockConfigProvider) Validate(ctx context.Context, data config.ConfigData) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

func (m *MockConfigProvider) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockConfigProvider) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockConfigProvider) Set(key string, value interface{}) {
	m.data[key] = value
}

func TestGatewayServiceConfig_Structure(t *testing.T) {
	config := &gateway.GatewayServiceConfig{}

	// Test default structure
	assert.NotNil(t, &config.Service)
	assert.NotNil(t, &config.Routing)
	assert.NotNil(t, &config.RateLimit)
	assert.NotNil(t, &config.CORS)
	assert.NotNil(t, &config.Auth)
	assert.NotNil(t, &config.LoadBalancer)
	assert.NotNil(t, &config.Logging)
	assert.NotNil(t, &config.Observability)
}

func TestGatewayConfigManagerSetup(t *testing.T) {
	provider := NewMockConfigProvider()
	manager := gateway.NewGatewayConfigManager(provider)

	assert.NotNil(t, manager)
}

func TestGatewayConfigManager_ApplyDefaults(t *testing.T) {
	provider := NewMockConfigProvider()
	provider.On("Load", mock.Anything).Return(nil)

	manager := gateway.NewGatewayConfigManager(provider)

	ctx := context.Background()
	err := manager.LoadConfig(ctx)
	require.NoError(t, err)

	config := manager.GetConfig()
	assert.NotNil(t, config)

	// Check that defaults are applied
	assert.Equal(t, "api-gateway", config.Service.Name)
	assert.Equal(t, "0.0.0.0", config.Service.Host)
	assert.Equal(t, 8080, config.Service.Port)
	assert.Equal(t, "development", config.Service.Environment)

	assert.Equal(t, 100, config.RateLimit.RequestsPerSec)
	assert.Equal(t, 20, config.RateLimit.BurstSize)
	assert.Equal(t, 5*time.Minute, config.RateLimit.CleanupInterval)

	assert.Contains(t, config.CORS.AllowedOrigins, "*")
	assert.Contains(t, config.CORS.AllowedMethods, "GET")
	assert.Contains(t, config.CORS.AllowedHeaders, "*")
	assert.Equal(t, 86400, config.CORS.MaxAge)

	assert.Equal(t, "Authorization", config.Auth.TokenHeader)
	assert.Equal(t, 5*time.Second, config.Auth.Timeout)

	provider.AssertExpectations(t)
}

func TestRoutingConfig_Structure(t *testing.T) {
	config := gateway.RoutingConfig{
		Routes: []gateway.RouteConfig{
			{
				Path:        "/api/v1/users",
				Target:      "http://user-service:8080",
				Methods:     []string{"GET", "POST"},
				StripPrefix: true,
			},
		},
		DefaultRoute:  "http://default-service:8080",
		StripPrefixes: []string{"/api/v1"},
	}

	assert.Len(t, config.Routes, 1)
	assert.Equal(t, "/api/v1/users", config.Routes[0].Path)
	assert.Equal(t, "http://user-service:8080", config.Routes[0].Target)
	assert.Contains(t, config.Routes[0].Methods, "GET")
	assert.True(t, config.Routes[0].StripPrefix)
	assert.Equal(t, "http://default-service:8080", config.DefaultRoute)
	assert.Contains(t, config.StripPrefixes, "/api/v1")
}

func TestRateLimitConfig_Structure(t *testing.T) {
	config := gateway.RateLimitConfig{
		Enabled:         true,
		RequestsPerSec:  100,
		BurstSize:       20,
		CleanupInterval: 5 * time.Minute,
		IPWhitelist:     []string{"127.0.0.1", "192.168.1.0/24"},
	}

	assert.True(t, config.Enabled)
	assert.Equal(t, 100, config.RequestsPerSec)
	assert.Equal(t, 20, config.BurstSize)
	assert.Equal(t, 5*time.Minute, config.CleanupInterval)
	assert.Contains(t, config.IPWhitelist, "127.0.0.1")
	assert.Contains(t, config.IPWhitelist, "192.168.1.0/24")
}

func TestCORSConfig_Structure(t *testing.T) {
	config := gateway.CORSConfig{
		Enabled:          true,
		AllowedOrigins:   []string{"http://localhost:3000", "https://example.com"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		ExposedHeaders:   []string{"X-Total-Count"},
		AllowCredentials: true,
		MaxAge:           3600,
	}

	assert.True(t, config.Enabled)
	assert.Contains(t, config.AllowedOrigins, "http://localhost:3000")
	assert.Contains(t, config.AllowedMethods, "GET")
	assert.Contains(t, config.AllowedHeaders, "Content-Type")
	assert.Contains(t, config.ExposedHeaders, "X-Total-Count")
	assert.True(t, config.AllowCredentials)
	assert.Equal(t, 3600, config.MaxAge)
}

func TestAuthProxyConfig_Structure(t *testing.T) {
	config := gateway.AuthProxyConfig{
		Enabled:        true,
		AuthServiceURL: "http://auth-service:8081",
		ExcludePaths:   []string{"/health", "/metrics"},
		TokenHeader:    "Authorization",
		Timeout:        10 * time.Second,
	}

	assert.True(t, config.Enabled)
	assert.Equal(t, "http://auth-service:8081", config.AuthServiceURL)
	assert.Contains(t, config.ExcludePaths, "/health")
	assert.Contains(t, config.ExcludePaths, "/metrics")
	assert.Equal(t, "Authorization", config.TokenHeader)
	assert.Equal(t, 10*time.Second, config.Timeout)
}

func TestLoadBalancerConfig_Structure(t *testing.T) {
	config := gateway.LoadBalancerConfig{
		Strategy: "round_robin",
		HealthCheck: gateway.HealthCheck{
			Enabled:  true,
			Interval: 30 * time.Second,
			Timeout:  5 * time.Second,
			Path:     "/health",
		},
		RetryPolicy: gateway.RetryPolicy{
			MaxRetries: 3,
			Delay:      100 * time.Millisecond,
			MaxDelay:   1 * time.Second,
		},
		CircuitBreaker: gateway.CircuitBreaker{
			Enabled:                true,
			FailureThreshold:       5,
			RecoveryTimeout:        30 * time.Second,
			RequestVolumeThreshold: 10,
		},
	}

	assert.Equal(t, "round_robin", config.Strategy)

	// Health check
	assert.True(t, config.HealthCheck.Enabled)
	assert.Equal(t, 30*time.Second, config.HealthCheck.Interval)
	assert.Equal(t, "/health", config.HealthCheck.Path)

	// Retry policy
	assert.Equal(t, 3, config.RetryPolicy.MaxRetries)
	assert.Equal(t, 100*time.Millisecond, config.RetryPolicy.Delay)

	// Circuit breaker
	assert.True(t, config.CircuitBreaker.Enabled)
	assert.Equal(t, 5, config.CircuitBreaker.FailureThreshold)
}

func TestObservabilityConfig_Structure(t *testing.T) {
	config := gateway.ObservabilityConfig{
		MetricsEnabled: true,
		TracingEnabled: true,
		JaegerEndpoint: "http://jaeger:14268",
	}

	assert.True(t, config.MetricsEnabled)
	assert.True(t, config.TracingEnabled)
	assert.Equal(t, "http://jaeger:14268", config.JaegerEndpoint)
}

func TestGatewayConfigManager_Reload(t *testing.T) {
	provider := NewMockConfigProvider()
	provider.On("Load", mock.Anything).Return(nil)

	manager := gateway.NewGatewayConfigManager(provider)

	ctx := context.Background()

	// Initial load
	err := manager.LoadConfig(ctx)
	require.NoError(t, err)

	// Reload
	err = manager.Reload(ctx)
	require.NoError(t, err)

	provider.AssertExpectations(t)
}
