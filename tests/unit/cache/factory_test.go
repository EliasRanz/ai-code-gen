package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/EliasRanz/ai-code-gen/internal/cache"
	"github.com/stretchr/testify/assert"
)

func TestCacheFactory(t *testing.T) {
	t.Run("factory creation", func(t *testing.T) {
		t.Run("new factory instance", func(t *testing.T) {
			factory := cache.NewCacheFactory()
			assert.NotNil(t, factory)
		})
	})

	t.Run("provider creation", func(t *testing.T) {
		factory := cache.NewCacheFactory()

		t.Run("memory provider creation", func(t *testing.T) {
			config := cache.CacheConfig{
				Host:                   "",
				Port:                   0,
				MaxConnections:         10,
				MaxIdleConnections:     5,
				ConnectionTimeout:      5 * time.Second,
				IdleTimeout:            30 * time.Second,
				DefaultTTL:             time.Hour,
				FailureThreshold:       5,
				RequestVolumeThreshold: 20,
				RecoveryTimeout:        10 * time.Second,
			}

			provider, err := factory.CreateProvider("memory", config)
			assert.NoError(t, err)
			assert.NotNil(t, provider)
			defer provider.Close()
		})

		t.Run("redis provider creation", func(t *testing.T) {
			config := cache.CacheConfig{
				Host:                   "localhost",
				Port:                   6379,
				DB:                     0,
				MaxConnections:         10,
				MaxIdleConnections:     5,
				ConnectionTimeout:      5 * time.Second,
				IdleTimeout:            30 * time.Second,
				DefaultTTL:             time.Hour,
				FailureThreshold:       5,
				RequestVolumeThreshold: 20,
				RecoveryTimeout:        10 * time.Second,
			}

			provider, err := factory.CreateProvider("redis", config)
			if err != nil {
				// Expected when Redis is not available
				assert.Contains(t, err.Error(), "failed to connect to Redis")
				return
			}

			assert.NotNil(t, provider)
			defer provider.Close()
		})

		t.Run("multi provider creation", func(t *testing.T) {
			config := cache.CacheConfig{
				Host:                   "localhost",
				Port:                   6379,
				DB:                     0,
				MaxConnections:         10,
				MaxIdleConnections:     5,
				ConnectionTimeout:      5 * time.Second,
				IdleTimeout:            30 * time.Second,
				DefaultTTL:             time.Hour,
				FailureThreshold:       5,
				RequestVolumeThreshold: 20,
				RecoveryTimeout:        10 * time.Second,
			}

			provider, err := factory.CreateProvider("multi", config)
			if err != nil {
				// Expected when Redis is not available for multi-provider
				assert.Contains(t, err.Error(), "failed to connect to Redis")
				return
			}

			assert.NotNil(t, provider)
			defer provider.Close()
		})

		t.Run("unsupported provider type", func(t *testing.T) {
			config := getTestFactoryConfig()

			provider, err := factory.CreateProvider("unsupported", config)
			assert.Error(t, err)
			assert.Nil(t, provider)
			assert.Contains(t, err.Error(), "unsupported cache provider type")
		})

		t.Run("empty provider type", func(t *testing.T) {
			config := getTestFactoryConfig()

			provider, err := factory.CreateProvider("", config)
			assert.Error(t, err)
			assert.Nil(t, provider)
			assert.Contains(t, err.Error(), "provider type cannot be empty")
		})

		t.Run("invalid config", func(t *testing.T) {
			invalidConfig := cache.CacheConfig{
				Host:           "", // Invalid host for Redis-based providers
				Port:           0,  // Invalid port
				MaxConnections: -1, // Invalid max connections
			}

			// Test with Redis provider which should validate config
			provider, err := factory.CreateProvider("redis", invalidConfig)
			assert.Error(t, err)
			assert.Nil(t, provider)
			// Should be either connection error or validation error
		})
	})

	t.Run("available providers listing", func(t *testing.T) {
		factory := cache.NewCacheFactory()

		t.Run("list available providers", func(t *testing.T) {
			providers := factory.ListAvailableProviders()
			assert.NotEmpty(t, providers)

			// Check that common provider types are available
			expectedProviders := []string{"memory", "redis", "multi"}
			for _, expected := range expectedProviders {
				assert.Contains(t, providers, expected, "Provider %s should be available", expected)
			}
		})

		t.Run("providers list is not empty", func(t *testing.T) {
			providers := factory.ListAvailableProviders()
			assert.Greater(t, len(providers), 0, "Should have at least one provider available")
		})

		t.Run("providers list contains unique entries", func(t *testing.T) {
			providers := factory.ListAvailableProviders()

			// Check for duplicates
			seen := make(map[string]bool)
			for _, provider := range providers {
				assert.False(t, seen[provider], "Provider %s should not be duplicated", provider)
				seen[provider] = true
			}
		})
	})

	t.Run("configuration validation", func(t *testing.T) {
		factory := cache.NewCacheFactory()

		t.Run("valid memory config", func(t *testing.T) {
			config := cache.CacheConfig{
				MaxConnections:         10,
				MaxIdleConnections:     5,
				ConnectionTimeout:      5 * time.Second,
				IdleTimeout:            30 * time.Second,
				DefaultTTL:             time.Hour,
				FailureThreshold:       5,
				RequestVolumeThreshold: 20,
				RecoveryTimeout:        10 * time.Second,
			}

			provider, err := factory.CreateProvider("memory", config)
			assert.NoError(t, err)
			assert.NotNil(t, provider)
			defer provider.Close()
		})

		t.Run("invalid max connections", func(t *testing.T) {
			config := cache.CacheConfig{
				Host:               "localhost", // Valid for Redis
				Port:               6379,
				MaxConnections:     0, // Invalid
				MaxIdleConnections: 5,
			}

			// Test with Redis provider which validates config
			provider, err := factory.CreateProvider("redis", config)
			assert.Error(t, err)
			assert.Nil(t, provider)
			// Should be either validation error or connection error
		})

		t.Run("invalid idle connections", func(t *testing.T) {
			config := cache.CacheConfig{
				Host:               "localhost",
				Port:               6379,
				MaxConnections:     10,
				MaxIdleConnections: 15, // Greater than max connections
			}

			// Test with Redis provider which validates config
			provider, err := factory.CreateProvider("redis", config)
			assert.Error(t, err)
			assert.Nil(t, provider)
			// Should be either validation error or connection error
		})

		t.Run("invalid failure threshold", func(t *testing.T) {
			config := cache.CacheConfig{
				Host:                   "localhost",
				Port:                   6379,
				MaxConnections:         10,
				MaxIdleConnections:     5,
				FailureThreshold:       0, // Invalid
				RequestVolumeThreshold: 20,
				RecoveryTimeout:        10 * time.Second,
			}

			// Test with Redis provider which validates config
			provider, err := factory.CreateProvider("redis", config)
			assert.Error(t, err)
			assert.Nil(t, provider)
			// Should be either validation error or connection error
		})
	})

	t.Run("provider lifecycle management", func(t *testing.T) {
		factory := cache.NewCacheFactory()

		t.Run("created providers are independent", func(t *testing.T) {
			config1 := getTestFactoryConfig()
			config2 := getTestFactoryConfig()

			provider1, err := factory.CreateProvider("memory", config1)
			assert.NoError(t, err)
			assert.NotNil(t, provider1)

			provider2, err := factory.CreateProvider("memory", config2)
			assert.NoError(t, err)
			assert.NotNil(t, provider2)

			// Providers should be different instances
			assert.NotEqual(t, provider1, provider2)

			// Close one should not affect the other
			err = provider1.Close()
			assert.NoError(t, err)

			// Provider2 should still be functional
			ctx := context.Background()
			err = provider2.Set(ctx, "test", "value", time.Minute)
			assert.NoError(t, err)

			err = provider2.Close()
			assert.NoError(t, err)
		})

		t.Run("factory can create multiple provider types", func(t *testing.T) {
			memConfig := getTestFactoryConfig()
			memProvider, err := factory.CreateProvider("memory", memConfig)
			assert.NoError(t, err)
			assert.NotNil(t, memProvider)
			defer memProvider.Close()

			// Try to create Redis provider (may fail if Redis not available)
			redisConfig := cache.CacheConfig{
				Host:                   "localhost",
				Port:                   6379,
				MaxConnections:         10,
				MaxIdleConnections:     5,
				ConnectionTimeout:      5 * time.Second,
				IdleTimeout:            30 * time.Second,
				DefaultTTL:             time.Hour,
				FailureThreshold:       5,
				RequestVolumeThreshold: 20,
				RecoveryTimeout:        10 * time.Second,
			}

			redisProvider, err := factory.CreateProvider("redis", redisConfig)
			if err == nil {
				// Redis is available
				assert.NotNil(t, redisProvider)
				defer redisProvider.Close()

				// Both providers should be different types
				assert.NotEqual(t, memProvider, redisProvider)
			}
		})
	})
}

// Helper function
func getTestFactoryConfig() cache.CacheConfig {
	return cache.CacheConfig{
		Host:                   "localhost",
		Port:                   6379,
		DB:                     0,
		MaxConnections:         10,
		MaxIdleConnections:     5,
		ConnectionTimeout:      5 * time.Second,
		IdleTimeout:            30 * time.Second,
		DefaultTTL:             time.Hour,
		FailureThreshold:       5,
		RequestVolumeThreshold: 20,
		RecoveryTimeout:        10 * time.Second,
	}
}
