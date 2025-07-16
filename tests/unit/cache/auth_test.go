package cache

import (
	"context"
	"testing"
	"time"

	"github.com/EliasRanz/ai-code-gen/internal/auth"
	"github.com/EliasRanz/ai-code-gen/internal/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAuthCacheIntegration tests auth service cache integration with core infrastructure
func TestAuthCacheIntegration(t *testing.T) {
	ctx := context.Background()

	// Create cache service for auth
	serviceConfig := cache.DefaultServiceConfig()
	service, err := cache.NewService(serviceConfig)
	require.NoError(t, err)
	defer service.Close()

	// Create auth cache manager using the service provider
	authConfig := auth.DefaultCacheConfig()
	provider := service.GetProvider()
	cm := auth.NewCacheManager(provider, authConfig)

	t.Run("Auth Cache Manager Integration", func(t *testing.T) {
		tokenHash := "test_integration_token"
		userContext := &auth.UserContext{
			UserID: "integration_user",
			Email:  "integration@test.com",
			Role:   "admin",
		}

		// Test auth-specific caching operations
		err := cm.SetUserContext(ctx, tokenHash, userContext)
		assert.NoError(t, err)

		cached, err := cm.GetUserContext(ctx, tokenHash)
		assert.NoError(t, err)
		require.NotNil(t, cached)
		assert.Equal(t, userContext.UserID, cached.UserID)
		assert.Equal(t, userContext.Email, cached.Email)
		assert.Equal(t, userContext.Role, cached.Role)

		// Test invalidation
		err = cm.InvalidateUserContext(ctx, tokenHash)
		assert.NoError(t, err)

		cached, err = cm.GetUserContext(ctx, tokenHash)
		assert.NoError(t, err)
		assert.Nil(t, cached)
	})

	t.Run("Auth Session Caching", func(t *testing.T) {
		sessionID := "test_session_123"
		userID := "user_123"
		sessionData := &auth.SessionData{
			UserID:    userID,
			Email:     "test@example.com",
			Role:      "user",
			ExpiresAt: time.Now().Add(time.Hour),
		}

		// Test session caching
		err := cm.CacheSession(ctx, sessionID, userID, sessionData)
		assert.NoError(t, err)

		cachedSession, err := cm.GetSession(ctx, sessionID)
		assert.NoError(t, err)
		require.NotNil(t, cachedSession)
		assert.Equal(t, sessionData.UserID, cachedSession.UserID)
		assert.Equal(t, sessionData.Email, cachedSession.Email)
		assert.Equal(t, sessionData.Role, cachedSession.Role)

		// Test session invalidation
		err = cm.InvalidateSession(ctx, sessionID)
		assert.NoError(t, err)

		cachedSession, err = cm.GetSession(ctx, sessionID)
		assert.NoError(t, err)
		assert.Nil(t, cachedSession)
	})

	t.Run("Auth Cache Health Check", func(t *testing.T) {
		// Test that auth cache works with infrastructure health checks
		err := service.HealthCheck(ctx)
		assert.NoError(t, err)

		// Test provider is accessible through cache manager
		assert.NotNil(t, provider)
		err = provider.HealthCheck(ctx)
		assert.NoError(t, err)
	})
}

// TestAuthCacheConfiguration tests auth cache configuration integration
func TestAuthCacheConfiguration(t *testing.T) {
	t.Run("Default Auth Configuration", func(t *testing.T) {
		config := auth.DefaultCacheConfig()
		assert.Equal(t, 5*time.Minute, config.TTL)
		assert.Equal(t, "auth:token:", config.TokenKeyPrefix)
		assert.Equal(t, "auth:session:", config.SessionKeyPrefix)
	})

	t.Run("Custom Auth Configuration", func(t *testing.T) {
		serviceConfig := cache.DefaultServiceConfig()
		service, err := cache.NewService(serviceConfig)
		require.NoError(t, err)
		defer service.Close()

		customConfig := auth.CacheConfig{
			TTL:              10 * time.Minute,
			TokenKeyPrefix:   "custom:auth:token:",
			SessionKeyPrefix: "custom:auth:session:",
		}

		provider := service.GetProvider()
		cm := auth.NewCacheManager(provider, customConfig)

		// Test that custom configuration is used
		ctx := context.Background()
		tokenHash := "custom_test_token"
		userContext := &auth.UserContext{
			UserID: "custom_user",
			Email:  "custom@test.com",
			Role:   "user",
		}

		err = cm.SetUserContext(ctx, tokenHash, userContext)
		assert.NoError(t, err)

		cached, err := cm.GetUserContext(ctx, tokenHash)
		assert.NoError(t, err)
		assert.NotNil(t, cached)
		assert.Equal(t, userContext.UserID, cached.UserID)
	})
}

// BenchmarkAuthCacheOperations benchmarks auth-specific cache operations
func BenchmarkAuthCacheOperations(b *testing.B) {
	ctx := context.Background()
	serviceConfig := cache.DefaultServiceConfig()
	service, _ := cache.NewService(serviceConfig)
	defer service.Close()

	authConfig := auth.DefaultCacheConfig()
	provider := service.GetProvider()
	cm := auth.NewCacheManager(provider, authConfig)

	userContext := &auth.UserContext{
		UserID: "bench_user",
		Email:  "bench@test.com",
		Role:   "user",
	}

	b.Run("AuthSetUserContext", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			tokenHash := "bench_token_" + string(rune(i))
			cm.SetUserContext(ctx, tokenHash, userContext)
		}
	})

	b.Run("AuthGetUserContext", func(b *testing.B) {
		tokenHash := "bench_get_token"
		cm.SetUserContext(ctx, tokenHash, userContext)
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			cm.GetUserContext(ctx, tokenHash)
		}
	})

	b.Run("AuthCacheSession", func(b *testing.B) {
		sessionData := &auth.SessionData{
			UserID:    "bench_user",
			Email:     "bench@test.com",
			Role:      "user",
			ExpiresAt: time.Now().Add(time.Hour),
		}

		for i := 0; i < b.N; i++ {
			sessionID := "bench_session_" + string(rune(i))
			cm.CacheSession(ctx, sessionID, "user123", sessionData)
		}
	})
}
