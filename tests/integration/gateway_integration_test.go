package tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/EliasRanz/ai-code-gen/internal/cache"
	"github.com/EliasRanz/ai-code-gen/internal/gateway"
)

func TestGatewayMiddlewareChainIntegration(t *testing.T) {
	chain := gateway.NewMiddlewareChain()
	assert.NotNil(t, chain)

	// Test that we can get empty middleware list
	middlewares := chain.GetMiddleware()
	assert.NotNil(t, middlewares)
	assert.Len(t, middlewares, 0)
}

func TestGatewayMiddlewareFactoryIntegration(t *testing.T) {
	// Create auth cache - using redis:// URL for testing
	authCache, err := cache.NewAuthCache("redis://localhost:6379", 5*time.Minute)
	if err != nil {
		// Skip test if Redis is not available
		t.Skip("Redis not available, skipping integration test")
	}

	factory := gateway.NewMiddlewareFactory("http://auth-service:8081", authCache)
	assert.NotNil(t, factory)

	// Test listing available middleware
	available := factory.ListAvailableMiddleware()
	assert.NotNil(t, available)
	assert.Contains(t, available, "auth-proxy")
	assert.Contains(t, available, "rate-limit")
	assert.Contains(t, available, "logging")
}

func TestGatewayMiddlewareConfigIntegration(t *testing.T) {
	params := map[string]interface{}{
		"timeout": 30 * time.Second,
		"retries": 3,
	}

	config := gateway.NewBasicMiddlewareConfig("test-middleware", true, params)
	assert.NotNil(t, config)

	assert.Equal(t, "test-middleware", config.GetName())
	assert.True(t, config.IsEnabled())

	configParams := config.GetParameters()
	assert.NotNil(t, configParams)
	assert.Equal(t, 30*time.Second, configParams["timeout"])
	assert.Equal(t, 3, configParams["retries"])
}

func TestObservableGatewayIntegration(t *testing.T) {
	authCache, err := cache.NewAuthCache("redis://localhost:6379", 5*time.Minute)
	if err != nil {
		t.Skip("Redis not available, skipping integration test")
	}

	factory := gateway.NewMiddlewareFactory("http://auth-service:8081", authCache)
	gw := gateway.NewObservableGateway(factory)

	assert.NotNil(t, gw)

	// Test health check
	err = gw.HealthCheck()
	assert.NoError(t, err)
}

func TestGatewayEventNotifierIntegration(t *testing.T) {
	notifier := gateway.NewGatewayEventNotifier()
	assert.NotNil(t, notifier)

	// Create and subscribe an observer
	observer := gateway.NewMetricsObserver()
	err := notifier.Subscribe(observer)
	assert.NoError(t, err)

	// Test unsubscribe
	err = notifier.Unsubscribe(observer)
	assert.NoError(t, err)
}

func TestGatewayMetricsObserverIntegration(t *testing.T) {
	observer := gateway.NewMetricsObserver()
	assert.NotNil(t, observer)

	ctx := context.Background()
	request := &gateway.HTTPRequest{
		Method: "GET",
		Path:   "/test",
	}

	// Test request received
	err := observer.OnRequestReceived(ctx, request)
	assert.NoError(t, err)

	response := &gateway.HTTPResponse{
		StatusCode: 200,
	}

	// Test request processed
	err = observer.OnRequestProcessed(ctx, request, response)
	assert.NoError(t, err)
}

func TestGatewaySecurityObserverIntegration(t *testing.T) {
	observer := gateway.NewSecurityObserver()
	assert.NotNil(t, observer)

	ctx := context.Background()
	request := &gateway.HTTPRequest{
		Method: "GET",
		Path:   "/test",
	}

	// Test request received
	err := observer.OnRequestReceived(ctx, request)
	assert.NoError(t, err)
}

func TestGatewayMetricsCollectorIntegration(t *testing.T) {
	collector := gateway.NewMetricsCollector()
	assert.NotNil(t, collector)

	// Test incrementing request count
	collector.IncrementRequestCount("GET", "/test")

	// Test recording latency
	collector.RecordLatency("/test", 100*time.Millisecond)

	// Test incrementing response code
	collector.IncrementResponseCode(200)

	// Test getting request count (returns int, not float64)
	count := collector.GetRequestCount()
	assert.GreaterOrEqual(t, count, 0)
}

func TestGatewayMiddlewareChainOperationsIntegration(t *testing.T) {
	chain := gateway.NewMiddlewareChain()

	// Create logging middleware (which doesn't require auth cache)
	middleware := gateway.NewLoggingMiddleware()

	// Test adding middleware to chain
	returnedChain := chain.Add(middleware)
	assert.NotNil(t, returnedChain)

	// Test getting middleware list
	middlewares := chain.GetMiddleware()
	assert.Len(t, middlewares, 1)
	assert.Equal(t, "logging", middlewares[0].GetName())
}

func TestGatewayLoggingMiddlewareIntegration(t *testing.T) {
	middleware := gateway.NewLoggingMiddleware()
	assert.NotNil(t, middleware)

	// Test basic properties
	assert.Equal(t, "logging", middleware.GetName())
	assert.Equal(t, 10, middleware.GetOrder()) // Logging runs early in chain

	// Test health check
	err := middleware.HealthCheck()
	assert.NoError(t, err)

	// Test config validation
	err = middleware.ValidateConfig()
	assert.NoError(t, err)
}

func TestGatewayFullStackIntegration(t *testing.T) {
	// This test integrates multiple gateway components together
	authCache, err := cache.NewAuthCache("redis://localhost:6379", 5*time.Minute)
	if err != nil {
		t.Skip("Redis not available, skipping full stack integration test")
	}

	// Create factory
	factory := gateway.NewMiddlewareFactory("http://auth-service:8081", authCache)
	assert.NotNil(t, factory)

	// Create observable gateway
	gw := gateway.NewObservableGateway(factory)
	assert.NotNil(t, gw)

	// Setup middleware configuration
	configs := []gateway.MiddlewareConfig{
		gateway.NewBasicMiddlewareConfig("logging", true, map[string]interface{}{
			"level": "info",
		}),
	}

	err = gw.SetupMiddleware(configs)
	assert.NoError(t, err)

	// Test health check
	err = gw.HealthCheck()
	assert.NoError(t, err)
}
