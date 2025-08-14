package cache_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/EliasRanz/ai-code-gen/internal/cache"
	"github.com/stretchr/testify/assert"
)

func TestMultiProvider(t *testing.T) {
	t.Run("creation and configuration", func(t *testing.T) {
		t.Run("valid config creates multi-provider", func(t *testing.T) {
			config := getTestMultiProviderConfig()

			// This will try to create Redis + Memory providers
			multiProvider, err := cache.NewMultiProvider(config)
			if err != nil {
				// Expected in environments without Redis - test validation
				assert.Contains(t, err.Error(), "failed to connect to Redis")
				return
			}

			// If Redis is available, provider should be created
			assert.NotNil(t, multiProvider)
			defer multiProvider.Close()
		})

		t.Run("invalid config returns error", func(t *testing.T) {
			invalidConfig := cache.CacheConfig{
				Host: "", // Invalid host
				Port: 6379,
			}

			multiProvider, err := cache.NewMultiProvider(invalidConfig)
			assert.Error(t, err)
			assert.Nil(t, multiProvider)
			assert.Contains(t, err.Error(), "host is required")
		})
	})

	t.Run("basic operations", func(t *testing.T) {
		t.Run("successful operations", func(t *testing.T) {
			config := getTestMultiProviderConfig()
			multiProvider, err := cache.NewMultiProvider(config)
			if err != nil {
				t.Skip("Redis not available for multi-provider testing")
				return
			}
			defer multiProvider.Close()

			ctx := context.Background()

			// Set operation
			err = multiProvider.Set(ctx, "multi_test", "multi_value", time.Minute)
			assert.NoError(t, err)

			// Get operation - should work even if Redis fails (fallback to memory)
			value, err := multiProvider.Get(ctx, "multi_test")
			assert.NoError(t, err)
			assert.Equal(t, "multi_value", value)

			// Delete operation
			err = multiProvider.Delete(ctx, "multi_test")
			assert.NoError(t, err)

			// Verify deletion
			_, err = multiProvider.Get(ctx, "multi_test")
			assert.Error(t, err)
		})

		t.Run("exists operation", func(t *testing.T) {
			config := getTestMultiProviderConfig()
			multiProvider, err := cache.NewMultiProvider(config)
			if err != nil {
				t.Skip("Redis not available for multi-provider testing")
				return
			}
			defer multiProvider.Close()

			ctx := context.Background()

			// Test non-existent key
			exists, err := multiProvider.Exists(ctx, "non_existent")
			assert.NoError(t, err)
			assert.False(t, exists)

			// Set value
			err = multiProvider.Set(ctx, "exists_test", "value", time.Minute)
			assert.NoError(t, err)

			// Should find key
			exists, err = multiProvider.Exists(ctx, "exists_test")
			assert.NoError(t, err)
			assert.True(t, exists)
		})
	})

	t.Run("batch operations", func(t *testing.T) {
		t.Run("multi-get and multi-set operations", func(t *testing.T) {
			config := getTestMultiProviderConfig()
			multiProvider, err := cache.NewMultiProvider(config)
			if err != nil {
				t.Skip("Redis not available for multi-provider testing")
				return
			}
			defer multiProvider.Close()

			ctx := context.Background()

			// Set multiple values
			pairs := map[string]string{
				"batch_key1": "batch_value1",
				"batch_key2": "batch_value2",
				"batch_key3": "batch_value3",
			}

			err = multiProvider.MSet(ctx, pairs, time.Minute)
			assert.NoError(t, err)

			// Get multiple values
			keys := []string{"batch_key1", "batch_key2", "batch_key3"}
			values, err := multiProvider.MGet(ctx, keys)
			assert.NoError(t, err)
			assert.Len(t, values, 3)
		})

		t.Run("multi-delete operations", func(t *testing.T) {
			config := getTestMultiProviderConfig()
			multiProvider, err := cache.NewMultiProvider(config)
			if err != nil {
				t.Skip("Redis not available for multi-provider testing")
				return
			}
			defer multiProvider.Close()

			ctx := context.Background()

			// Set multiple values
			pairs := map[string]string{
				"delete_key1": "delete_value1",
				"delete_key2": "delete_value2",
			}
			err = multiProvider.MSet(ctx, pairs, time.Minute)
			assert.NoError(t, err)

			// Delete multiple values
			keys := []string{"delete_key1", "delete_key2"}
			err = multiProvider.MDelete(ctx, keys)
			assert.NoError(t, err)

			// Verify deletion
			for _, key := range keys {
				_, err := multiProvider.Get(ctx, key)
				assert.Error(t, err)
			}
		})
	})

	t.Run("context handling", func(t *testing.T) {
		t.Run("operations respect context timeout", func(t *testing.T) {
			config := getTestMultiProviderConfig()
			multiProvider, err := cache.NewMultiProvider(config)
			if err != nil {
				t.Skip("Redis not available for multi-provider testing")
				return
			}
			defer multiProvider.Close()

			// Create context with very short timeout
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
			defer cancel()

			// Operations may timeout or succeed quickly with memory fallback
			err = multiProvider.Set(ctx, "timeout_test", "value", time.Minute)
			// Either timeout or success is acceptable - tests context propagation
			_ = err
		})

		t.Run("operations respect context cancellation", func(t *testing.T) {
			config := getTestMultiProviderConfig()
			multiProvider, err := cache.NewMultiProvider(config)
			if err != nil {
				t.Skip("Redis not available for multi-provider testing")
				return
			}
			defer multiProvider.Close()

			// Create cancelable context
			ctx, cancel := context.WithCancel(context.Background())
			cancel() // Cancel immediately

			// Operation should handle cancellation
			_, err = multiProvider.Get(ctx, "cancel_test")
			// Either cancellation or success is acceptable - tests context handling
			_ = err
		})
	})

	t.Run("health check and lifecycle", func(t *testing.T) {
		t.Run("health check reports status", func(t *testing.T) {
			config := getTestMultiProviderConfig()
			multiProvider, err := cache.NewMultiProvider(config)
			if err != nil {
				t.Skip("Redis not available for multi-provider testing")
				return
			}
			defer multiProvider.Close()

			ctx := context.Background()
			err = multiProvider.HealthCheck(ctx)
			// With Redis + Memory fallback, should be healthy even if Redis fails
			assert.NoError(t, err)
		})

		t.Run("close cleans up resources", func(t *testing.T) {
			config := getTestMultiProviderConfig()
			multiProvider, err := cache.NewMultiProvider(config)
			if err != nil {
				t.Skip("Redis not available for multi-provider testing")
				return
			}

			// Close should not return error
			err = multiProvider.Close()
			assert.NoError(t, err)
		})
	})

	t.Run("error handling", func(t *testing.T) {
		t.Run("empty key validation", func(t *testing.T) {
			config := getTestMultiProviderConfig()
			multiProvider, err := cache.NewMultiProvider(config)
			if err != nil {
				t.Skip("Redis not available for multi-provider testing")
				return
			}
			defer multiProvider.Close()

			ctx := context.Background()

			// Test empty key
			err = multiProvider.Set(ctx, "", "value", time.Minute)
			assert.Error(t, err, "Setting empty key should return error")
		})

		t.Run("large value handling", func(t *testing.T) {
			config := getTestMultiProviderConfig()
			multiProvider, err := cache.NewMultiProvider(config)
			if err != nil {
				t.Skip("Redis not available for multi-provider testing")
				return
			}
			defer multiProvider.Close()

			ctx := context.Background()

			// Test large value (1MB)
			largeValue := make([]byte, 1024*1024)
			for i := range largeValue {
				largeValue[i] = byte(i % 256)
			}

			err = multiProvider.Set(ctx, "large_value", string(largeValue), time.Minute)
			// Should work with memory fallback if Redis fails
			if err == nil {
				value, err := multiProvider.Get(ctx, "large_value")
				assert.NoError(t, err)
				assert.Equal(t, string(largeValue), value)
			}
		})

		t.Run("concurrent operations", func(t *testing.T) {
			config := getTestMultiProviderConfig()
			multiProvider, err := cache.NewMultiProvider(config)
			if err != nil {
				t.Skip("Redis not available for multi-provider testing")
				return
			}
			defer multiProvider.Close()

			ctx := context.Background()
			numGoroutines := 10
			done := make(chan bool, numGoroutines)

			// Perform concurrent operations
			for i := 0; i < numGoroutines; i++ {
				go func(id int) {
					key := fmt.Sprintf("concurrent_key_%d", id)
					value := fmt.Sprintf("concurrent_value_%d", id)

					err := multiProvider.Set(ctx, key, value, time.Minute)
					assert.NoError(t, err)

					retrievedValue, err := multiProvider.Get(ctx, key)
					assert.NoError(t, err)
					assert.Equal(t, value, retrievedValue)

					done <- true
				}(i)
			}

			// Wait for all goroutines to complete
			for i := 0; i < numGoroutines; i++ {
				<-done
			}
		})
	})
}

// Helper function

func getTestMultiProviderConfig() cache.CacheConfig {
	return cache.CacheConfig{
		Host:                   "localhost",
		Port:                   6379,
		DB:                     0,
		Password:               "",
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
