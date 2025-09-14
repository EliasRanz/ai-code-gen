package auth_cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/EliasRanz/ai-code-gen/internal/cache"
	"github.com/EliasRanz/ai-code-gen/tests/performance/utils"
)

// BenchmarkCacheGet tests individual cache get operations
func BenchmarkCacheGet(b *testing.B) {
	authCache, cleanup := utils.SetupFastRedisForTesting(b)
	defer cleanup()

	// Pre-populate cache with test data
	ctx := context.Background()
	userContext := &cache.UserContext{
		UserID: "benchmark-user-123",
		Email:  "benchmark@test.com",
		Role:   "user",
	}
	tokenHash := "benchmark-token-hash"
	err := authCache.SetUserContext(ctx, tokenHash, userContext)
	require.NoError(b, err)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := authCache.GetUserContext(ctx, tokenHash)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkCacheSet tests individual cache set operations
func BenchmarkCacheSet(b *testing.B) {
	authCache, cleanup := utils.SetupFastRedisForTesting(b)
	defer cleanup()

	ctx := context.Background()
	userContext := &cache.UserContext{
		UserID: "benchmark-user-123",
		Email:  "benchmark@test.com",
		Role:   "user",
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			tokenHash := cache.HashToken(b.Name() + string(rune(i)))
			err := authCache.SetUserContext(ctx, tokenHash, userContext)
			if err != nil {
				b.Fatal(err)
			}
			i++
		}
	})
}

// BenchmarkCacheGetMiss tests cache miss performance (empty cache)
func BenchmarkCacheGetMiss(b *testing.B) {
	authCache, cleanup := utils.SetupFastRedisForTesting(b)
	defer cleanup()

	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			tokenHash := cache.HashToken(b.Name() + string(rune(i)))
			_, err := authCache.GetUserContext(ctx, tokenHash)
			if err != nil {
				b.Fatal(err)
			}
			i++
		}
	})
}

// BenchmarkCacheHitRatio tests performance under different cache hit ratios
func BenchmarkCacheHitRatio(b *testing.B) {
	hitRatios := []float64{0.0, 0.25, 0.50, 0.75, 0.90, 0.95}

	for _, hitRatio := range hitRatios {
		b.Run(b.Name()+string(rune(int(hitRatio*100))), func(b *testing.B) {
			benchmarkWithHitRatio(b, hitRatio)
		})
	}
}

// benchmarkWithHitRatio tests cache performance with specified hit ratio
func benchmarkWithHitRatio(b *testing.B, hitRatio float64) {
	authCache, cleanup := utils.SetupFastRedisForTesting(b)
	defer cleanup()

	// Generate test data with specified hit ratio
	generator := utils.NewTestDataGenerator(100, 1)
	tokens := generator.GenerateTestTokens()

	// Pre-populate cache based on hit ratio
	ctx := context.Background()
	numToCache := int(float64(len(tokens)) * hitRatio)
	for i := 0; i < numToCache; i++ {
		err := authCache.SetUserContext(ctx, tokens[i].Hash, tokens[i].UserContext)
		require.NoError(b, err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			token := tokens[i%len(tokens)]
			_, err := authCache.GetUserContext(ctx, token.Hash)
			if err != nil {
				b.Fatal(err)
			}
			i++
		}
	})
}

// BenchmarkCacheInvalidation tests cache invalidation performance
func BenchmarkCacheInvalidation(b *testing.B) {
	authCache, cleanup := utils.SetupFastRedisForTesting(b)
	defer cleanup()

	ctx := context.Background()
	userContext := &cache.UserContext{
		UserID: "benchmark-user-123",
		Email:  "benchmark@test.com",
		Role:   "user",
	}

	// Pre-populate with tokens to invalidate
	tokens := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		tokens[i] = cache.HashToken(b.Name() + string(rune(i)))
		err := authCache.SetUserContext(ctx, tokens[i], userContext)
		require.NoError(b, err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := authCache.InvalidateUserContext(ctx, tokens[i])
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkConcurrentAccess tests performance under concurrent access
func BenchmarkConcurrentAccess(b *testing.B) {
	concurrencyLevels := []int{1, 5, 10, 25, 50, 100}

	for _, concurrency := range concurrencyLevels {
		b.Run(b.Name()+string(rune(concurrency)), func(b *testing.B) {
			benchmarkConcurrentAccess(b, concurrency)
		})
	}
}

// benchmarkConcurrentAccess tests cache with specified concurrency level
func benchmarkConcurrentAccess(b *testing.B, concurrency int) {
	authCache, cleanup := utils.SetupFastRedisForTesting(b)
	defer cleanup()

	// Pre-populate cache
	ctx := context.Background()
	tokens := make([]string, 1000)
	userContext := &cache.UserContext{
		UserID: "concurrent-user",
		Email:  "concurrent@test.com",
		Role:   "user",
	}

	for i := 0; i < len(tokens); i++ {
		tokens[i] = cache.HashToken(b.Name() + string(rune(i)))
		err := authCache.SetUserContext(ctx, tokens[i], userContext)
		require.NoError(b, err)
	}

	b.SetParallelism(concurrency)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			token := tokens[i%len(tokens)]
			_, err := authCache.GetUserContext(ctx, token)
			if err != nil {
				b.Fatal(err)
			}
			i++
		}
	})
}

// TestCachePerformanceRegression validates that cache performance meets SLA requirements
func TestCachePerformanceRegression(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance regression test in short mode")
	}

	// Skip performance tests by default unless explicitly enabled
	if !isPerformanceTestingEnabled() {
		t.Skip("Skipping performance test - set PERFORMANCE_TESTS=1 to enable")
	}

	authCache, cleanup := utils.SetupRealRedisForPerformanceTesting(t)
	defer cleanup()

	metrics := utils.NewPerformanceMetrics()
	ctx := context.Background()

	// Generate realistic test data
	generator := utils.NewTestDataGenerator(1000, 2)
	tokens := generator.GenerateRealisticTokens()

	// Warm up cache with 80% of tokens
	numToCache := int(0.8 * float64(len(tokens)))
	for i := 0; i < numToCache; i++ {
		err := authCache.AuthCache.SetUserContext(ctx, tokens[i].Hash, tokens[i].UserContext)
		require.NoError(t, err)
	}

	// Performance test: 1000 requests
	start := time.Now()
	for i := 0; i < 1000; i++ {
		token := tokens[i%len(tokens)]
		requestStart := time.Now()

		result, err := authCache.AuthCache.GetUserContext(ctx, token.Hash)
		duration := time.Since(requestStart)

		if err != nil {
			metrics.RecordCacheError(duration)
		} else if result != nil {
			metrics.RecordCacheHit(duration)
		} else {
			metrics.RecordCacheMiss(duration)
		}
	}
	totalDuration := time.Since(start)

	// Generate performance report
	report := metrics.GenerateReport("Performance Regression Test")

	// Validate SLA requirements
	assert.Less(t, report.Percentiles.P95, 5*time.Millisecond,
		"P95 latency %v exceeds SLA of 5ms", report.Percentiles.P95)
	assert.Less(t, report.ErrorRate, 0.01,
		"Error rate %.2f%% exceeds SLA of 1%%", report.ErrorRate*100)
	assert.Greater(t, report.CacheHitRate, 0.75,
		"Cache hit rate %.2f%% below expected 75%%", report.CacheHitRate*100)
	assert.Greater(t, report.ThroughputRPS, 500.0,
		"Throughput %.2f req/sec below expected 500 req/sec", report.ThroughputRPS)

	// Print detailed results
	t.Logf("=== Performance Regression Test Results ===")
	t.Logf("Total Duration: %v", totalDuration)
	t.Logf("Total Requests: %d", report.TotalRequests)
	t.Logf("Throughput: %.2f req/sec", report.ThroughputRPS)
	t.Logf("Cache Hit Rate: %.2f%%", report.CacheHitRate*100)
	t.Logf("Error Rate: %.2f%%", report.ErrorRate*100)
	t.Logf("P50 Latency: %v", report.Percentiles.P50)
	t.Logf("P95 Latency: %v", report.Percentiles.P95)
	t.Logf("P99 Latency: %v", report.Percentiles.P99)
	t.Logf("Memory Usage: %.2f MB", report.MemoryUsageMB)
	t.Logf("=========================================")

	metrics.PrintSummary("Performance Regression Test")
}
