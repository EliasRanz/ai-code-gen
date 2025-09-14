package utils

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/EliasRanz/ai-code-gen/internal/cache"
)

// TryExistingRedis attempts to connect to an existing Redis instance
// This is useful when running tests with 'make dev' already running
func TryExistingRedis() (*cache.AuthCache, bool) {
	// Check common Redis URLs
	redisURLs := []string{
		"redis://localhost:6380", // Docker compose Redis port
		"redis://localhost:6379", // Default Redis port
		os.Getenv("REDIS_URL"),   // Environment variable
	}

	for _, redisURL := range redisURLs {
		if redisURL == "" {
			continue
		}

		// Try to connect to Redis
		authCache, err := cache.NewAuthCache(redisURL, 1*time.Minute)
		if err != nil {
			continue
		}

		// Test the connection
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		if err := authCache.HealthCheck(ctx); err == nil {
			cancel()
			return authCache, true
		}
		cancel()
		authCache.Close()
	}

	return nil, false
}

// SetupRedisWithFallback tries existing Redis first, then creates container
func SetupRedisWithFallback(testingInterface interface{}) (*cache.AuthCache, func(), error) {
	// First try to use existing Redis
	if authCache, ok := TryExistingRedis(); ok {
		return authCache, func() { authCache.Close() }, nil
	}

	// Fall back to container setup
	switch t := testingInterface.(type) {
	case *testing.T:
		container, cleanup := SetupRealRedisForPerformanceTesting(t)
		return container.AuthCache, cleanup, nil
	case *testing.B:
		authCache, cleanup := SetupFastRedisForTesting(t)
		return authCache, cleanup, nil
	default:
		return nil, nil, fmt.Errorf("unsupported testing interface")
	}
}
