package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/EliasRanz/ai-code-gen/internal/cache"
)

func TestNewAuthCache(t *testing.T) {
	// Setup mini Redis server for testing
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	redisURL := "redis://" + mr.Addr()
	ttl := 5 * time.Minute

	authCache, err := cache.NewAuthCache(redisURL, ttl)
	require.NoError(t, err)
	require.NotNil(t, authCache)
	defer authCache.Close()

	// Test health check
	err = authCache.HealthCheck(context.Background())
	assert.NoError(t, err)
}

func TestAuthCache_SetAndGetUserContext(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	authCache, err := cache.NewAuthCache("redis://"+mr.Addr(), 5*time.Minute)
	require.NoError(t, err)
	defer authCache.Close()

	ctx := context.Background()
	token := "test-token-123"
	tokenHash := cache.HashToken(token)

	userContext := &cache.UserContext{
		UserID: "user-123",
		Email:  "test@example.com",
		Role:   "user",
	}

	// Test cache miss
	result, err := authCache.GetUserContext(ctx, tokenHash)
	require.NoError(t, err)
	assert.Nil(t, result)

	// Test cache set
	err = authCache.SetUserContext(ctx, tokenHash, userContext)
	require.NoError(t, err)

	// Test cache hit
	result, err = authCache.GetUserContext(ctx, tokenHash)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, userContext.UserID, result.UserID)
	assert.Equal(t, userContext.Email, result.Email)
	assert.Equal(t, userContext.Role, result.Role)
	assert.False(t, result.CachedAt.IsZero())
}

func TestAuthCache_InvalidateUserContext(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	authCache, err := cache.NewAuthCache("redis://"+mr.Addr(), 5*time.Minute)
	require.NoError(t, err)
	defer authCache.Close()

	ctx := context.Background()
	token := "test-token-456"
	tokenHash := cache.HashToken(token)

	userContext := &cache.UserContext{
		UserID: "user-456",
		Email:  "test2@example.com",
		Role:   "admin",
	}

	// Set cache
	err = authCache.SetUserContext(ctx, tokenHash, userContext)
	require.NoError(t, err)

	// Verify cache exists
	result, err := authCache.GetUserContext(ctx, tokenHash)
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Invalidate cache
	err = authCache.InvalidateUserContext(ctx, tokenHash)
	require.NoError(t, err)

	// Verify cache is gone
	result, err = authCache.GetUserContext(ctx, tokenHash)
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestHashToken(t *testing.T) {
	token1 := "test-token-123"
	token2 := "test-token-456"
	token3 := "test-token-123" // Same as token1

	hash1 := cache.HashToken(token1)
	hash2 := cache.HashToken(token2)
	hash3 := cache.HashToken(token3)

	// Different tokens should have different hashes
	assert.NotEqual(t, hash1, hash2)
	
	// Same tokens should have same hashes
	assert.Equal(t, hash1, hash3)
	
	// Hashes should be consistent length (64 chars for SHA256)
	assert.Len(t, hash1, 64)
	assert.Len(t, hash2, 64)
}

func TestAuthCache_TTLExpiration(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	// Use very short TTL for testing
	authCache, err := cache.NewAuthCache("redis://"+mr.Addr(), 100*time.Millisecond)
	require.NoError(t, err)
	defer authCache.Close()

	ctx := context.Background()
	token := "test-token-ttl"
	tokenHash := cache.HashToken(token)

	userContext := &cache.UserContext{
		UserID: "user-ttl",
		Email:  "ttl@example.com",
		Role:   "user",
	}

	// Set cache
	err = authCache.SetUserContext(ctx, tokenHash, userContext)
	require.NoError(t, err)

	// Should be available immediately
	result, err := authCache.GetUserContext(ctx, tokenHash)
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Fast-forward time in mini Redis
	mr.FastForward(200 * time.Millisecond)

	// Should be expired now
	result, err = authCache.GetUserContext(ctx, tokenHash)
	require.NoError(t, err)
	assert.Nil(t, result)
}
