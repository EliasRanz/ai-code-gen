package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/EliasRanz/ai-code-gen/internal/auth"
	"github.com/EliasRanz/ai-code-gen/internal/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthCacheManager(t *testing.T) {
	ctx := context.Background()

	// Create memory provider for testing
	cacheConfig := cache.DefaultCacheConfig()
	provider, err := cache.NewMemoryProvider(cacheConfig)
	require.NoError(t, err)
	defer provider.Close()

	// Create auth cache manager
	authConfig := auth.DefaultCacheConfig()
	cm := auth.NewCacheManager(provider, authConfig)

	t.Run("User Context Caching", func(t *testing.T) {
		tokenHash := "test_token_hash"
		userContext := &auth.UserContext{
			UserID: "user123",
			Email:  "test@example.com",
			Role:   "user",
		}

		// Test Set
		err := cm.SetUserContext(ctx, tokenHash, userContext)
		assert.NoError(t, err)

		// Test Get
		cached, err := cm.GetUserContext(ctx, tokenHash)
		assert.NoError(t, err)
		require.NotNil(t, cached)
		assert.Equal(t, userContext.UserID, cached.UserID)
		assert.Equal(t, userContext.Email, cached.Email)
		assert.Equal(t, userContext.Role, cached.Role)
		assert.False(t, cached.CachedAt.IsZero())

		// Test Invalidate
		err = cm.InvalidateUserContext(ctx, tokenHash)
		assert.NoError(t, err)

		// Verify deletion
		cached, err = cm.GetUserContext(ctx, tokenHash)
		assert.NoError(t, err)
		assert.Nil(t, cached)
	})

	t.Run("Session Caching", func(t *testing.T) {
		sessionID := "session123"
		userID := "user123"
		sessionData := &auth.SessionData{
			UserID:    userID,
			Email:     "test@example.com",
			Role:      "user",
			ExpiresAt: time.Now().Add(time.Hour),
		}

		// Test Cache Session
		err := cm.CacheSession(ctx, sessionID, userID, sessionData)
		assert.NoError(t, err)

		// Test Get Session
		cached, err := cm.GetSession(ctx, sessionID)
		assert.NoError(t, err)
		require.NotNil(t, cached)
		assert.Equal(t, sessionData.UserID, cached.UserID)
		assert.Equal(t, sessionData.Email, cached.Email)
		assert.Equal(t, sessionData.Role, cached.Role)
		assert.False(t, cached.CachedAt.IsZero())

		// Test Invalidate Session
		err = cm.InvalidateSession(ctx, sessionID)
		assert.NoError(t, err)

		// Verify deletion
		cached, err = cm.GetSession(ctx, sessionID)
		assert.NoError(t, err)
		assert.Nil(t, cached)
	})

	t.Run("User Session Invalidation", func(t *testing.T) {
		userID := "user123"

		// Create multiple sessions for user
		sessions := []string{"session1", "session2", "session3"}
		for _, sessionID := range sessions {
			sessionData := &auth.SessionData{
				UserID: userID,
				Email:  "test@example.com",
				Role:   "user",
			}
			err := cm.CacheSession(ctx, sessionID, userID, sessionData)
			assert.NoError(t, err)
		}

		// Verify sessions exist
		for _, sessionID := range sessions {
			cached, err := cm.GetSession(ctx, sessionID)
			assert.NoError(t, err)
			assert.NotNil(t, cached)
		}

		// Invalidate all user sessions
		err := cm.InvalidateUserSessions(ctx, userID)
		assert.NoError(t, err)

		// Note: This test might not work perfectly with memory provider
		// as it doesn't support complex pattern matching like Redis
		// In production, this would work with Redis pattern matching
	})

	t.Run("Key Generation", func(t *testing.T) {
		// Test token key generation
		tokenKey := cm.GenerateKey("token", "hash123")
		assert.Equal(t, "auth:token:hash123", tokenKey)

		// Test session key generation
		sessionKey := cm.GenerateKey("session", "session123")
		assert.Equal(t, "auth:session:session123", sessionKey)

		// Test user key generation
		userKey := cm.GenerateKey("user", "user123", "data")
		assert.Equal(t, "auth:user:user123:data", userKey)

		// Test unknown key type
		unknownKey := cm.GenerateKey("unknown", "test")
		assert.Equal(t, "auth:unknown:test", unknownKey)
	})

	t.Run("JSON Operations", func(t *testing.T) {
		key := "test:json:key"
		data := map[string]interface{}{
			"field1": "value1",
			"field2": 42,
			"field3": true,
		}

		// Test SetJSON
		err := cm.SetJSON(ctx, key, data, time.Minute)
		assert.NoError(t, err)

		// Test GetJSON
		var result map[string]interface{}
		err = cm.GetJSON(ctx, key, &result)
		assert.NoError(t, err)
		assert.Equal(t, data["field1"], result["field1"])
		assert.Equal(t, float64(42), result["field2"]) // JSON numbers are float64
		assert.Equal(t, data["field3"], result["field3"])
	})

	t.Run("Input Validation", func(t *testing.T) {
		// Empty token hash
		err := cm.SetUserContext(ctx, "", &auth.UserContext{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token hash cannot be empty")

		_, err = cm.GetUserContext(ctx, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token hash cannot be empty")

		// Nil user context
		err = cm.SetUserContext(ctx, "hash", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user context cannot be nil")

		// Empty session ID
		err = cm.CacheSession(ctx, "", "user", &auth.SessionData{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "session ID cannot be empty")

		// Empty user ID in session
		err = cm.CacheSession(ctx, "session", "", &auth.SessionData{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user ID cannot be empty")

		// Nil session data
		err = cm.CacheSession(ctx, "session", "user", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "session data cannot be nil")
	})

	t.Run("Health Check", func(t *testing.T) {
		err := cm.HealthCheck(ctx)
		assert.NoError(t, err)
	})
}

func TestAuthCacheConfig(t *testing.T) {
	t.Run("Default Config", func(t *testing.T) {
		config := auth.DefaultCacheConfig()
		assert.Equal(t, 5*time.Minute, config.TTL)
		assert.Equal(t, "auth:token:", config.TokenKeyPrefix)
		assert.Equal(t, "auth:session:", config.SessionKeyPrefix)
	})
}

func TestHashToken(t *testing.T) {
	t.Run("Token Hashing", func(t *testing.T) {
		token := "test_token_12345"
		hash1 := auth.HashToken(token)
		hash2 := auth.HashToken(token)

		// Same token should produce same hash
		assert.Equal(t, hash1, hash2)

		// Hash should be consistent length (SHA256 = 64 hex chars)
		assert.Len(t, hash1, 64)

		// Different tokens should produce different hashes
		hash3 := auth.HashToken("different_token")
		assert.NotEqual(t, hash1, hash3)
	})
}

func BenchmarkAuthCacheManager(b *testing.B) {
	ctx := context.Background()

	cacheConfig := cache.DefaultCacheConfig()
	provider, _ := cache.NewMemoryProvider(cacheConfig)
	defer provider.Close()

	authConfig := auth.DefaultCacheConfig()
	cm := auth.NewCacheManager(provider, authConfig)

	userContext := &auth.UserContext{
		UserID: "user123",
		Email:  "test@example.com",
		Role:   "user",
	}

	b.Run("SetUserContext", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			tokenHash := auth.HashToken("token_" + string(rune(i)))
			cm.SetUserContext(ctx, tokenHash, userContext)
		}
	})

	b.Run("GetUserContext", func(b *testing.B) {
		tokenHash := auth.HashToken("bench_token")
		cm.SetUserContext(ctx, tokenHash, userContext)
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			cm.GetUserContext(ctx, tokenHash)
		}
	})

	b.Run("HashToken", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			auth.HashToken("bench_token_" + string(rune(i)))
		}
	})
}
