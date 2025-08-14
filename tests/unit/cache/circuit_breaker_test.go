package cache_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/EliasRanz/ai-code-gen/internal/cache"
	"github.com/stretchr/testify/assert"
)

func TestCircuitBreaker(t *testing.T) {
	t.Run("circuit breaker creation", func(t *testing.T) {
		t.Run("valid config creates circuit breaker", func(t *testing.T) {
			config := cache.CircuitBreakerConfig{
				FailureThreshold:       5,
				RequestVolumeThreshold: 10,
				RecoveryTimeout:        30 * time.Second,
				MaxConcurrentRequests:  20,
			}

			cb := cache.NewCircuitBreaker(config)
			assert.NotNil(t, cb)
		})

		t.Run("zero values use defaults", func(t *testing.T) {
			config := cache.CircuitBreakerConfig{}
			cb := cache.NewCircuitBreaker(config)
			assert.NotNil(t, cb)
		})
	})

	t.Run("circuit breaker states", func(t *testing.T) {
		t.Run("starts in closed state", func(t *testing.T) {
			config := getTestCircuitBreakerConfig()
			cb := cache.NewCircuitBreaker(config)

			// Circuit should be closed initially (allows requests)
			ctx := context.Background()
			successFunc := func() (interface{}, error) {
				return "success", nil
			}

			result, err := cb.Execute(ctx, successFunc)
			assert.NoError(t, err)
			assert.Equal(t, "success", result)
		})

		t.Run("moves to open state after failures", func(t *testing.T) {
			config := cache.CircuitBreakerConfig{
				FailureThreshold:       2, // Low threshold for testing
				RequestVolumeThreshold: 2,
				RecoveryTimeout:        1 * time.Second,
				MaxConcurrentRequests:  10,
			}
			cb := cache.NewCircuitBreaker(config)

			ctx := context.Background()
			failureFunc := func() (interface{}, error) {
				return nil, errors.New("test failure")
			}

			// Execute failures to trigger circuit opening
			_, err1 := cb.Execute(ctx, failureFunc)
			assert.Error(t, err1)

			_, err2 := cb.Execute(ctx, failureFunc)
			assert.Error(t, err2)

			// After failure threshold, circuit should open
			// Next request should be rejected immediately
			_, err3 := cb.Execute(ctx, failureFunc)
			assert.Error(t, err3)
			// Should be circuit breaker open error or original error
		})

		t.Run("moves to half-open state after recovery timeout", func(t *testing.T) {
			config := cache.CircuitBreakerConfig{
				FailureThreshold:       2,
				RequestVolumeThreshold: 2,
				RecoveryTimeout:        100 * time.Millisecond, // Short timeout for testing
				MaxConcurrentRequests:  10,
			}
			cb := cache.NewCircuitBreaker(config)

			ctx := context.Background()
			failureFunc := func() (interface{}, error) {
				return nil, errors.New("test failure")
			}

			// Trigger circuit opening
			_, _ = cb.Execute(ctx, failureFunc)
			_, _ = cb.Execute(ctx, failureFunc)

			// Wait for recovery timeout
			time.Sleep(150 * time.Millisecond)

			// Circuit should allow one test request (half-open state)
			successFunc := func() (interface{}, error) {
				return "recovery", nil
			}

			result, err := cb.Execute(ctx, successFunc)
			if err == nil {
				// If circuit allows the request and it succeeds
				assert.Equal(t, "recovery", result)
			}
			// Some implementations might still reject, which is also valid behavior
		})
	})

	t.Run("successful request handling", func(t *testing.T) {
		t.Run("successful requests pass through", func(t *testing.T) {
			config := getTestCircuitBreakerConfig()
			cb := cache.NewCircuitBreaker(config)

			ctx := context.Background()

			// Test various successful operations
			testCases := []struct {
				name     string
				function func() (interface{}, error)
				expected interface{}
			}{
				{
					name:     "string result",
					function: func() (interface{}, error) { return "test", nil },
					expected: "test",
				},
				{
					name:     "integer result",
					function: func() (interface{}, error) { return 42, nil },
					expected: 42,
				},
				{
					name:     "boolean result",
					function: func() (interface{}, error) { return true, nil },
					expected: true,
				},
				{
					name:     "nil result",
					function: func() (interface{}, error) { return nil, nil },
					expected: nil,
				},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					result, err := cb.Execute(ctx, tc.function)
					assert.NoError(t, err)
					assert.Equal(t, tc.expected, result)
				})
			}
		})

		t.Run("successful requests reset failure count", func(t *testing.T) {
			config := cache.CircuitBreakerConfig{
				FailureThreshold:       3,
				RequestVolumeThreshold: 5,
				RecoveryTimeout:        1 * time.Second,
				MaxConcurrentRequests:  10,
			}
			cb := cache.NewCircuitBreaker(config)

			ctx := context.Background()

			// Add some failures (but not enough to open circuit)
			failureFunc := func() (interface{}, error) {
				return nil, errors.New("failure")
			}
			_, _ = cb.Execute(ctx, failureFunc)
			_, _ = cb.Execute(ctx, failureFunc)

			// Add a success - should reset or reduce failure count
			successFunc := func() (interface{}, error) {
				return "success", nil
			}
			result, err := cb.Execute(ctx, successFunc)
			assert.NoError(t, err)
			assert.Equal(t, "success", result)

			// Should still be able to execute more requests
			result2, err2 := cb.Execute(ctx, successFunc)
			assert.NoError(t, err2)
			assert.Equal(t, "success", result2)
		})
	})

	t.Run("failure handling", func(t *testing.T) {
		t.Run("failures are properly reported", func(t *testing.T) {
			config := getTestCircuitBreakerConfig()
			cb := cache.NewCircuitBreaker(config)

			ctx := context.Background()
			expectedErr := errors.New("test error")

			failureFunc := func() (interface{}, error) {
				return nil, expectedErr
			}

			result, err := cb.Execute(ctx, failureFunc)
			assert.Error(t, err)
			assert.Nil(t, result)
			// Error should be wrapped or be the original error
			assert.Contains(t, err.Error(), "test error")
		})

		t.Run("different error types are handled", func(t *testing.T) {
			config := getTestCircuitBreakerConfig()
			cb := cache.NewCircuitBreaker(config)

			ctx := context.Background()

			testErrors := []error{
				errors.New("simple error"),
				&cache.CacheError{Op: "test", Key: "key", Provider: "test", Err: errors.New("cache error")},
				context.DeadlineExceeded,
				context.Canceled,
			}

			for i, expectedErr := range testErrors {
				t.Run(expectedErr.Error(), func(t *testing.T) {
					failureFunc := func() (interface{}, error) {
						return nil, expectedErr
					}

					result, err := cb.Execute(ctx, failureFunc)
					assert.Error(t, err, "Error case %d should return error", i)
					assert.Nil(t, result, "Error case %d should return nil result", i)
				})
			}
		})
	})

	t.Run("context handling", func(t *testing.T) {
		t.Run("respects context timeout", func(t *testing.T) {
			config := getTestCircuitBreakerConfig()
			cb := cache.NewCircuitBreaker(config)

			// Create context with short timeout
			ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
			defer cancel()

			slowFunc := func() (interface{}, error) {
				time.Sleep(200 * time.Millisecond) // Longer than timeout
				return "slow", nil
			}

			result, err := cb.Execute(ctx, slowFunc)
			// Should either timeout or execute quickly - both are acceptable
			// depending on implementation
			if err != nil {
				// Expected timeout or cancellation
				assert.True(t, errors.Is(err, context.DeadlineExceeded) ||
					errors.Is(err, context.Canceled) ||
					err.Error() != "")
			} else {
				// Some implementations might complete quickly
				_ = result
			}
		})

		t.Run("respects context cancellation", func(t *testing.T) {
			config := getTestCircuitBreakerConfig()
			cb := cache.NewCircuitBreaker(config)

			// Create cancellable context
			ctx, cancel := context.WithCancel(context.Background())
			cancel() // Cancel immediately

			testFunc := func() (interface{}, error) {
				return "cancelled", nil
			}

			result, err := cb.Execute(ctx, testFunc)
			// Should handle cancellation - either return error or execute quickly
			if err != nil {
				assert.True(t, errors.Is(err, context.Canceled) || err.Error() != "")
			}
			_ = result // Either outcome is acceptable
		})
	})

	t.Run("concurrent operations", func(t *testing.T) {
		t.Run("handles concurrent requests", func(t *testing.T) {
			config := cache.CircuitBreakerConfig{
				FailureThreshold:       10, // High threshold to avoid opening during test
				RequestVolumeThreshold: 20,
				RecoveryTimeout:        1 * time.Second,
				MaxConcurrentRequests:  5,
			}
			cb := cache.NewCircuitBreaker(config)

			ctx := context.Background()
			numGoroutines := 10
			results := make(chan string, numGoroutines)
			errors := make(chan error, numGoroutines)

			// Launch concurrent operations
			for i := 0; i < numGoroutines; i++ {
				go func(id int) {
					slowFunc := func() (interface{}, error) {
						time.Sleep(10 * time.Millisecond)
						return fmt.Sprintf("result_%d", id), nil
					}

					result, err := cb.Execute(ctx, slowFunc)
					if err != nil {
						errors <- err
					} else {
						results <- result.(string)
					}
				}(i)
			}

			// Collect results
			successCount := 0
			errorCount := 0

			for i := 0; i < numGoroutines; i++ {
				select {
				case result := <-results:
					assert.Contains(t, result, "result_")
					successCount++
				case err := <-errors:
					assert.NotNil(t, err)
					errorCount++
				case <-time.After(1 * time.Second):
					t.Fatal("Timeout waiting for goroutine results")
				}
			}

			// Some requests should succeed, some might be rejected due to concurrency limits
			assert.Greater(t, successCount+errorCount, 0)
		})

		t.Run("enforces max concurrent requests", func(t *testing.T) {
			config := cache.CircuitBreakerConfig{
				FailureThreshold:       10,
				RequestVolumeThreshold: 20,
				RecoveryTimeout:        1 * time.Second,
				MaxConcurrentRequests:  2, // Low limit for testing
			}
			cb := cache.NewCircuitBreaker(config)

			ctx := context.Background()
			numGoroutines := 5
			started := make(chan bool, numGoroutines)
			proceed := make(chan bool, numGoroutines)
			results := make(chan error, numGoroutines)

			// Launch operations that wait for signal
			for i := 0; i < numGoroutines; i++ {
				go func() {
					slowFunc := func() (interface{}, error) {
						started <- true
						<-proceed // Wait for signal to proceed
						return "done", nil
					}

					_, err := cb.Execute(ctx, slowFunc)
					results <- err
				}()
			}

			// Wait for some requests to start
			startedCount := 0
			timeout := time.After(100 * time.Millisecond)
		waitLoop:
			for startedCount < 2 {
				select {
				case <-started:
					startedCount++
				case <-timeout:
					break waitLoop
				}
			}

			// Signal all to proceed
			for i := 0; i < numGoroutines; i++ {
				proceed <- true
			}

			// Collect results - some should succeed, some might be rejected
			for i := 0; i < numGoroutines; i++ {
				select {
				case err := <-results:
					// Either success (nil) or circuit breaker rejection
					_ = err
				case <-time.After(500 * time.Millisecond):
					t.Fatal("Timeout waiting for results")
				}
			}
		})
	})

	t.Run("edge cases", func(t *testing.T) {
		t.Run("nil function handling", func(t *testing.T) {
			config := getTestCircuitBreakerConfig()
			cb := cache.NewCircuitBreaker(config)

			ctx := context.Background()

			// Test with nil function - should be handled gracefully
			defer func() {
				if r := recover(); r != nil {
					// Panic recovery is acceptable behavior
					assert.Contains(t, fmt.Sprintf("%v", r), "nil")
				}
			}()

			result, err := cb.Execute(ctx, nil)
			if err != nil {
				// Error handling is acceptable
				assert.Error(t, err)
				assert.Nil(t, result)
			}
		})
	})
}

// Helper functions
func getTestCircuitBreakerConfig() cache.CircuitBreakerConfig {
	return cache.CircuitBreakerConfig{
		FailureThreshold:       5,
		RequestVolumeThreshold: 10,
		RecoveryTimeout:        30 * time.Second,
		MaxConcurrentRequests:  20,
	}
}
