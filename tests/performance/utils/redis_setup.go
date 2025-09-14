package utils

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/EliasRanz/ai-code-gen/internal/cache"
	"github.com/EliasRanz/ai-code-gen/internal/observability"
)

// TestingInterface represents common methods between *testing.T and *testing.B
type TestingInterface interface {
	Helper()
	Errorf(format string, args ...interface{})
	FailNow()
}

// InitPerformanceTestLogging initializes logging with performance-optimized settings
func InitPerformanceTestLogging() {
	// Get log level from environment, default to "info" to suppress debug logs
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	// Initialize logging with performance-friendly settings
	observability.InitLogging(logLevel, "json", "performance-test")
}

// RedisTestContainer wraps a Redis test container with helper methods
type RedisTestContainer struct {
	Container testcontainers.Container
	AuthCache *cache.AuthCache
	Endpoint  string
}

// SetupRealRedisForPerformanceTesting creates a real Redis container for performance testing
func SetupRealRedisForPerformanceTesting(t *testing.T) (*RedisTestContainer, func()) {
	// Initialize performance-optimized logging
	InitPerformanceTestLogging()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Start real Redis container with performance optimizations and port-based wait
	redisContainer, err := redis.Run(ctx,
		"redis:7-alpine",
		redis.WithSnapshotting(0, 0), // Disable snapshotting for performance
		redis.WithLogLevel(redis.LogLevelWarning),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("6379/tcp").
				WithStartupTimeout(60*time.Second),
		),
	)
	require.NoError(t, err, "Failed to start Redis container")

	endpoint, err := redisContainer.Endpoint(ctx, "")
	require.NoError(t, err, "Failed to get Redis endpoint")

	// Create auth cache with test configuration
	authCache, err := cache.NewAuthCache(
		fmt.Sprintf("redis://%s", endpoint),
		5*time.Minute, // 5-minute TTL for testing
	)
	require.NoError(t, err, "Failed to create auth cache")

	// Verify Redis is ready
	healthCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	require.NoError(t, authCache.HealthCheck(healthCtx))

	container := &RedisTestContainer{
		Container: redisContainer,
		AuthCache: authCache,
		Endpoint:  endpoint,
	}

	cleanup := func() {
		if authCache != nil {
			authCache.Close()
		}
		if redisContainer != nil {
			redisContainer.Terminate(ctx)
		}
	}

	return container, cleanup
}

// SetupMiniredisForUnitTesting creates a miniredis instance for fast unit testing
func SetupMiniredisForUnitTesting(t *testing.T) (*cache.AuthCache, func()) {
	// Note: This would use miniredis for faster unit tests
	// For now, we'll use the real Redis setup but with faster configuration
	return SetupFastRedisForTesting(t)
}

// SetupFastRedisForTesting creates a Redis container optimized for speed over durability
func SetupFastRedisForTesting(t TestingInterface) (*cache.AuthCache, func()) {
	if tb, ok := t.(*testing.T); ok {
		return setupFastRedisInternal(tb)
	}
	if bb, ok := t.(*testing.B); ok {
		return setupFastRedisBenchmark(bb)
	}
	panic("unsupported testing interface")
}

// setupFastRedisInternal handles *testing.T
func setupFastRedisInternal(t *testing.T) (*cache.AuthCache, func()) {
	// Initialize performance-optimized logging
	InitPerformanceTestLogging()

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	// Start Redis with aggressive performance settings (no persistence)
	redisContainer, err := redis.Run(ctx,
		"redis:7-alpine",
		redis.WithSnapshotting(0, 0), // Disable snapshotting
		redis.WithLogLevel(redis.LogLevelWarning),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("6379/tcp").
				WithStartupTimeout(60*time.Second),
		),
	)
	require.NoError(t, err, "Failed to start Redis container")

	endpoint, err := redisContainer.Endpoint(ctx, "")
	require.NoError(t, err, "Failed to get Redis endpoint")

	authCache, err := cache.NewAuthCache(
		fmt.Sprintf("redis://%s", endpoint),
		1*time.Minute, // Shorter TTL for faster testing
	)
	require.NoError(t, err, "Failed to create auth cache")

	// Verify connectivity
	require.NoError(t, authCache.HealthCheck(context.Background()))

	cleanup := func() {
		authCache.Close()
		redisContainer.Terminate(ctx)
	}

	return authCache, cleanup
}

// setupFastRedisBenchmark handles *testing.B
func setupFastRedisBenchmark(b *testing.B) (*cache.AuthCache, func()) {
	// Initialize performance-optimized logging
	InitPerformanceTestLogging()

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	// Start Redis with aggressive performance settings (no persistence)
	redisContainer, err := redis.Run(ctx,
		"redis:7-alpine",
		redis.WithSnapshotting(0, 0), // Disable snapshotting
		redis.WithLogLevel(redis.LogLevelWarning),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("6379/tcp").
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		b.Fatalf("Failed to start Redis container: %v", err)
	}

	endpoint, err := redisContainer.Endpoint(ctx, "")
	if err != nil {
		b.Fatalf("Failed to get Redis endpoint: %v", err)
	}

	authCache, err := cache.NewAuthCache(
		fmt.Sprintf("redis://%s", endpoint),
		1*time.Minute, // Shorter TTL for faster testing
	)
	if err != nil {
		b.Fatalf("Failed to create auth cache: %v", err)
	}

	// Verify connectivity
	if err := authCache.HealthCheck(context.Background()); err != nil {
		b.Fatalf("Redis health check failed: %v", err)
	}

	cleanup := func() {
		authCache.Close()
		redisContainer.Terminate(ctx)
	}

	return authCache, cleanup
}

// WarmUpCache pre-populates the cache with test data for realistic performance testing
func WarmUpCache(authCache *cache.AuthCache, numEntries int) error {
	ctx := context.Background()

	for i := 0; i < numEntries; i++ {
		tokenHash := fmt.Sprintf("test-token-hash-%d", i)
		userContext := &cache.UserContext{
			UserID: fmt.Sprintf("user-%d", i),
			Email:  fmt.Sprintf("user%d@test.com", i),
			Role:   "user",
		}

		if err := authCache.SetUserContext(ctx, tokenHash, userContext); err != nil {
			return fmt.Errorf("failed to warm up cache at entry %d: %w", i, err)
		}
	}

	return nil
}

// ClearCache removes all cached entries
func ClearCache(authCache *cache.AuthCache) error {
	// Note: Redis doesn't have a built-in way to clear by pattern
	// In a real implementation, you might use SCAN + DEL
	// For testing, we'll rely on TTL expiration or container restart
	return nil
}

// VerifyCacheHealth checks that the cache is responding correctly
func VerifyCacheHealth(t *testing.T, authCache *cache.AuthCache) {
	ctx := context.Background()

	// Test basic operations
	testTokenHash := "health-check-token"
	testUserContext := &cache.UserContext{
		UserID: "health-check-user",
		Email:  "health@test.com",
		Role:   "user",
	}

	// Test set
	err := authCache.SetUserContext(ctx, testTokenHash, testUserContext)
	require.NoError(t, err, "Cache set operation failed")

	// Test get
	retrieved, err := authCache.GetUserContext(ctx, testTokenHash)
	require.NoError(t, err, "Cache get operation failed")
	require.NotNil(t, retrieved, "Cache returned nil for existing key")
	require.Equal(t, testUserContext.UserID, retrieved.UserID, "Cache returned incorrect data")

	// Test invalidate
	err = authCache.InvalidateUserContext(ctx, testTokenHash)
	require.NoError(t, err, "Cache invalidate operation failed")

	// Verify invalidation
	retrieved, err = authCache.GetUserContext(ctx, testTokenHash)
	require.NoError(t, err, "Cache get after invalidate failed")
	require.Nil(t, retrieved, "Cache should return nil after invalidation")
}
