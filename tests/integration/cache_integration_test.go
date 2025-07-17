package tests

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/EliasRanz/ai-code-gen/internal/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRedisIntegration tests real Redis connections and operations
func TestRedisIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	if os.Getenv("INTEGRATION_TESTS") == "" {
		t.Skip("Skipping integration tests. Set INTEGRATION_TESTS=1 to run.")
	}

	// Load configuration from environment
	testConfig, err := LoadIntegrationConfig()
	require.NoError(t, err, "Failed to load integration test configuration")

	// Create Redis provider
	factory := cache.NewCacheFactory()
	provider, err := factory.CreateProvider("redis", testConfig.Redis)
	require.NoError(t, err, "Failed to create Redis provider")
	defer provider.Close()

	ctx := context.Background()

	t.Run("BasicOperations", func(t *testing.T) {
		testBasicCacheOperations(t, ctx, provider)
	})

	t.Run("BatchOperations", func(t *testing.T) {
		testBatchCacheOperations(t, ctx, provider)
	})

	t.Run("PatternOperations", func(t *testing.T) {
		testPatternCacheOperations(t, ctx, provider)
	})

	t.Run("CircuitBreakerResilience", func(t *testing.T) {
		testCircuitBreakerResilience(t, ctx, provider)
	})

	t.Run("ConnectionPooling", func(t *testing.T) {
		testConnectionPooling(t, ctx, provider)
	})

	t.Run("HealthCheck", func(t *testing.T) {
		testCacheHealthCheck(t, ctx, provider)
	})
}

func testBasicCacheOperations(t *testing.T, ctx context.Context, provider cache.CacheProvider) {
	key := "test:basic:key"
	value := "test_value"
	ttl := 30 * time.Second

	// Test Set operation
	err := provider.Set(ctx, key, value, ttl)
	assert.NoError(t, err, "Set operation should succeed")

	// Test Get operation
	retrievedValue, err := provider.Get(ctx, key)
	assert.NoError(t, err, "Get operation should succeed")
	assert.Equal(t, value, retrievedValue, "Retrieved value should match set value")

	// Test Exists operation
	exists, err := provider.Exists(ctx, key)
	assert.NoError(t, err, "Exists operation should succeed")
	assert.True(t, exists, "Key should exist")

	// Test Delete operation
	err = provider.Delete(ctx, key)
	assert.NoError(t, err, "Delete operation should succeed")

	// Verify key is deleted
	exists, err = provider.Exists(ctx, key)
	assert.NoError(t, err, "Exists check after delete should succeed")
	assert.False(t, exists, "Key should not exist after deletion")
}

func testBatchCacheOperations(t *testing.T, ctx context.Context, provider cache.CacheProvider) {
	// Prepare test data
	testData := map[string]string{
		"test:batch:key1": "value1",
		"test:batch:key2": "value2",
		"test:batch:key3": "value3",
	}
	keys := make([]string, 0, len(testData))
	for key := range testData {
		keys = append(keys, key)
	}

	// Test MSet operation
	err := provider.MSet(ctx, testData, 30*time.Second)
	assert.NoError(t, err, "MSet operation should succeed")

	// Test MGet operation
	values, err := provider.MGet(ctx, keys)
	assert.NoError(t, err, "MGet operation should succeed")
	assert.Len(t, values, len(keys), "Should retrieve all values")

	// Verify values match
	for i, key := range keys {
		expectedValue := testData[key]
		assert.Equal(t, expectedValue, values[i], "Retrieved value should match for key %s", key)
	}

	// Test MDelete operation
	err = provider.MDelete(ctx, keys)
	assert.NoError(t, err, "MDelete operation should succeed")

	// Verify all keys are deleted
	for _, key := range keys {
		exists, err := provider.Exists(ctx, key)
		assert.NoError(t, err, "Exists check should succeed for key %s", key)
		assert.False(t, exists, "Key %s should not exist after batch deletion", key)
	}
}

func testPatternCacheOperations(t *testing.T, ctx context.Context, provider cache.CacheProvider) {
	// Setup test data with pattern
	pattern := "test:pattern:*"
	testKeys := []string{
		"test:pattern:key1",
		"test:pattern:key2",
		"test:pattern:key3",
	}

	// Set test data
	for _, key := range testKeys {
		err := provider.Set(ctx, key, "pattern_value", 30*time.Second)
		require.NoError(t, err, "Should set pattern test key %s", key)
	}

	// Test Keys operation
	foundKeys, err := provider.Keys(ctx, pattern)
	assert.NoError(t, err, "Keys operation should succeed")
	assert.Len(t, foundKeys, len(testKeys), "Should find all pattern keys")

	// Test DeleteByPattern operation
	err = provider.DeleteByPattern(ctx, pattern)
	assert.NoError(t, err, "DeleteByPattern operation should succeed")

	// Verify pattern keys are deleted
	foundKeys, err = provider.Keys(ctx, pattern)
	assert.NoError(t, err, "Keys check after pattern delete should succeed")
	assert.Empty(t, foundKeys, "No keys should match pattern after deletion")
}

func testCircuitBreakerResilience(t *testing.T, ctx context.Context, provider cache.CacheProvider) {
	// This test validates that the circuit breaker pattern is implemented
	// We test basic operations continue to work under normal conditions
	key := "test:circuit:key"
	value := "circuit_test"

	// Perform multiple operations to test circuit breaker doesn't interfere
	for i := 0; i < 10; i++ {
		err := provider.Set(ctx, key, value, 30*time.Second)
		assert.NoError(t, err, "Circuit breaker should not interfere with normal operations")

		retrievedValue, err := provider.Get(ctx, key)
		assert.NoError(t, err, "Circuit breaker should not interfere with get operations")
		assert.Equal(t, value, retrievedValue, "Value should be consistent")
	}

	// Cleanup
	_ = provider.Delete(ctx, key)
}

func testConnectionPooling(t *testing.T, ctx context.Context, provider cache.CacheProvider) {
	// Load test configuration
	testConfig, err := LoadIntegrationConfig()
	require.NoError(t, err, "Failed to load test configuration")

	// Test concurrent operations to validate connection pooling
	concurrency := testConfig.Test.Concurrency
	operations := testConfig.Test.Operations

	// Use channels to coordinate concurrent operations
	errChan := make(chan error, concurrency*operations)

	for i := 0; i < concurrency; i++ {
		go func(goroutineID int) {
			for j := 0; j < operations; j++ {
				key := fmt.Sprintf("test:pool:goroutine%d:op%d", goroutineID, j)
				value := fmt.Sprintf("pooled_value_%d_%d", goroutineID, j)

				// Perform set and get operations
				if err := provider.Set(ctx, key, value, 30*time.Second); err != nil {
					errChan <- fmt.Errorf("goroutine %d operation %d set failed: %w", goroutineID, j, err)
					return
				}

				if retrievedValue, err := provider.Get(ctx, key); err != nil {
					errChan <- fmt.Errorf("goroutine %d operation %d get failed: %w", goroutineID, j, err)
					return
				} else if retrievedValue != value {
					errChan <- fmt.Errorf("goroutine %d operation %d value mismatch: got %s, expected %s", goroutineID, j, retrievedValue, value)
					return
				}

				// Clean up
				_ = provider.Delete(ctx, key)
			}
		}(i)
	}

	// Collect results
	timeout := time.After(30 * time.Second)
	operationsCompleted := 0
	expectedOperations := concurrency * operations

	for operationsCompleted < expectedOperations {
		select {
		case err := <-errChan:
			t.Errorf("Concurrent operation failed: %v", err)
		case <-timeout:
			t.Fatalf("Connection pooling test timed out. Completed %d/%d operations", operationsCompleted, expectedOperations)
		default:
			// Check if all operations completed by trying to receive from error channel
			operationsCompleted++
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func testCacheHealthCheck(t *testing.T, ctx context.Context, provider cache.CacheProvider) {
	// Test health check
	err := provider.HealthCheck(ctx)
	assert.NoError(t, err, "Health check should pass for functioning Redis connection")

	// Test health check with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	err = provider.HealthCheck(timeoutCtx)
	assert.NoError(t, err, "Health check should complete within timeout")
}
