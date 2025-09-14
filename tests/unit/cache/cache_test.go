package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/EliasRanz/ai-code-gen/internal/cache"
	"github.com/EliasRanz/ai-code-gen/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCacheInfrastructure tests the core cache infrastructure including
// factory patterns, providers, and circuit breaker functionality
func TestCacheInfrastructure(t *testing.T) {
	t.Run("Factory Pattern", func(t *testing.T) {
		factory := cache.NewCacheFactory()

		// Test available providers
		providers := factory.ListAvailableProviders()
		assert.Contains(t, providers, cache.ProviderTypeRedis)
		assert.Contains(t, providers, cache.ProviderTypeMemory)
		assert.Contains(t, providers, cache.ProviderTypeMulti)

		// Test memory provider creation
		config := cache.DefaultCacheConfig()
		provider, err := factory.CreateProvider(cache.ProviderTypeMemory, config)
		require.NoError(t, err)
		assert.NotNil(t, provider)
		defer provider.Close()

		// Test invalid provider type
		_, err = factory.CreateProvider("invalid", config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported cache provider type")
	})

	t.Run("Cache Configuration Validation", func(t *testing.T) {
		// Valid config
		config := cache.DefaultCacheConfig()
		err := cache.ValidateCacheConfig(config)
		assert.NoError(t, err)

		// Invalid host
		config.Host = ""
		err = cache.ValidateCacheConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "host is required")

		// Invalid port
		config = cache.DefaultCacheConfig()
		config.Port = 0
		err = cache.ValidateCacheConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "port must be between")

		// Invalid max connections
		config = cache.DefaultCacheConfig()
		config.MaxConnections = 0
		err = cache.ValidateCacheConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "max connections must be positive")

		// Invalid idle connections
		config = cache.DefaultCacheConfig()
		config.MaxIdleConnections = config.MaxConnections + 1
		err = cache.ValidateCacheConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "max idle connections cannot exceed")
	})
}

// TestMemoryProvider tests the in-memory cache provider implementation
func TestMemoryProvider(t *testing.T) {
	ctx := context.Background()
	config := cache.DefaultCacheConfig()

	provider, err := cache.NewMemoryProvider(config)
	require.NoError(t, err)
	defer provider.Close()

	t.Run("Basic Operations", func(t *testing.T) {
		// Test Set and Get
		err := provider.Set(ctx, "test:key", "test:value", time.Minute)
		assert.NoError(t, err)

		value, err := provider.Get(ctx, "test:key")
		assert.NoError(t, err)
		assert.Equal(t, "test:value", value)

		// Test Exists
		exists, err := provider.Exists(ctx, "test:key")
		assert.NoError(t, err)
		assert.True(t, exists)

		// Test Delete
		err = provider.Delete(ctx, "test:key")
		assert.NoError(t, err)

		value, err = provider.Get(ctx, "test:key")
		assert.NoError(t, err)
		assert.Empty(t, value)
	})

	t.Run("Batch Operations", func(t *testing.T) {
		// Test MSet
		pairs := map[string]string{
			"batch:key1": "value1",
			"batch:key2": "value2",
			"batch:key3": "value3",
		}
		err := provider.MSet(ctx, pairs, time.Minute)
		assert.NoError(t, err)

		// Test MGet
		keys := []string{"batch:key1", "batch:key2", "batch:key3", "batch:missing"}
		values, err := provider.MGet(ctx, keys)
		assert.NoError(t, err)
		assert.Len(t, values, 4)
		assert.Equal(t, "value1", values[0])
		assert.Equal(t, "value2", values[1])
		assert.Equal(t, "value3", values[2])
		assert.Empty(t, values[3]) // Missing key

		// Test MDelete
		err = provider.MDelete(ctx, keys[:3])
		assert.NoError(t, err)

		// Verify deletion
		values, err = provider.MGet(ctx, keys[:3])
		assert.NoError(t, err)
		for _, value := range values {
			assert.Empty(t, value)
		}
	})
}

// TestCircuitBreaker tests the circuit breaker implementation
// for cache resilience
func TestCacheCircuitBreaker(t *testing.T) {
	ctx := context.Background()

	t.Run("Circuit Breaker States", func(t *testing.T) {
		config := cache.CircuitBreakerConfig{
			FailureThreshold:       3,
			RequestVolumeThreshold: 5,
			RecoveryTimeout:        100 * time.Millisecond,
			MaxConcurrentRequests:  10,
		}

		cb := cache.NewCircuitBreaker(config)

		// Initial state should be closed
		assert.Equal(t, cache.StateClosed, cb.State())

		// Successful operations
		for i := 0; i < 5; i++ {
			result, err := cb.Execute(ctx, func() (interface{}, error) {
				return "success", nil
			})
			assert.NoError(t, err)
			assert.Equal(t, "success", result)
		}

		// Should still be closed
		assert.Equal(t, cache.StateClosed, cb.State())

		// Now cause failures
		for i := 0; i < 3; i++ {
			_, err := cb.Execute(ctx, func() (interface{}, error) {
				return nil, assert.AnError
			})
			assert.Error(t, err)
		}

		// Should now be open
		assert.Equal(t, cache.StateOpen, cb.State())

		// Further requests should fail fast
		_, err := cb.Execute(ctx, func() (interface{}, error) {
			return "should not execute", nil
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "circuit breaker is open")

		// Wait for recovery timeout
		time.Sleep(150 * time.Millisecond)

		// Should transition to half-open and allow one request
		result, err := cb.Execute(ctx, func() (interface{}, error) {
			return "recovery", nil
		})
		assert.NoError(t, err)
		assert.Equal(t, "recovery", result)

		// Should be closed again
		assert.Equal(t, cache.StateClosed, cb.State())
	})
}

// TestCacheService tests the unified cache service infrastructure
func TestCacheService(t *testing.T) {
	t.Run("Service Creation", func(t *testing.T) {
		// Test with default config
		config := cache.DefaultServiceConfig()
		service, err := cache.NewService(config)
		require.NoError(t, err)
		assert.NotNil(t, service)
		defer service.Close()

		// Test health check
		err = service.HealthCheck(context.Background())
		assert.NoError(t, err)

		// Test provider access
		provider := service.GetProvider()
		assert.NotNil(t, provider)
	})

	t.Run("Service Configuration", func(t *testing.T) {
		config := cache.DefaultServiceConfig()

		// Test with invalid provider
		config.ProviderType = "invalid"
		_, err := cache.NewService(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported cache provider type")

		// Test with valid memory provider
		config = cache.DefaultServiceConfig()
		config.ProviderType = cache.ProviderTypeMemory
		service, err := cache.NewService(config)
		require.NoError(t, err)
		assert.NotNil(t, service)
		defer service.Close()
	})
}

// TestGlobalConfigIntegration tests cache service creation from global config
func TestGlobalConfigIntegration(t *testing.T) {
	// Create global config
	globalConfig := &config.Config{
		Redis: config.RedisConfig{
			Host:                   "test-host",
			Port:                   6380,
			Password:               "test-password",
			DB:                     1,
			MaxConnections:         50,
			MaxIdleConnections:     5,
			ConnectionTimeout:      10 * time.Second,
			FailureThreshold:       3,
			RequestVolumeThreshold: 5,
			RecoveryTimeout:        60 * time.Second,
		},
	}

	service, err := cache.NewServiceFromConfig(globalConfig)
	require.NoError(t, err)
	defer service.Close()

	serviceConfig := service.GetConfig()
	assert.Equal(t, "test-host", serviceConfig.Host)
	assert.Equal(t, 6380, serviceConfig.Port)
	assert.Equal(t, "test-password", serviceConfig.Password)
	assert.Equal(t, 1, serviceConfig.DB)
}

// BenchmarkCacheInfrastructure benchmarks core cache infrastructure
func BenchmarkCacheInfrastructure(b *testing.B) {
	ctx := context.Background()
	config := cache.DefaultCacheConfig()
	provider, _ := cache.NewMemoryProvider(config)
	defer provider.Close()

	b.Run("ProviderSet", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			provider.Set(ctx, "bench:key", "bench:value", time.Minute)
		}
	})

	b.Run("ProviderGet", func(b *testing.B) {
		provider.Set(ctx, "bench:key", "bench:value", time.Minute)
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			provider.Get(ctx, "bench:key")
		}
	})
}
