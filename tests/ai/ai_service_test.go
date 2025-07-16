package ai_test

import (
	"context"
	"testing"

	"github.com/EliasRanz/ai-code-gen/internal/ai"
	"github.com/EliasRanz/ai-code-gen/internal/ai/llm"
	"github.com/EliasRanz/ai-code-gen/internal/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAIService(t *testing.T) {
	// Setup dependencies
	rateLimiter := ai.NewRateLimiter(10, 2) // 10 requests per second, burst of 2
	quotaManager := ai.NewQuotaManager()
	cacheProvider, err := cache.NewMemoryProvider(cache.CacheConfig{})
	require.NoError(t, err)
	cacheManager := ai.NewCacheManager(cacheProvider, ai.DefaultCacheConfig())

	service := ai.NewAIService(rateLimiter, quotaManager, cacheManager)
	require.NotNil(t, service)

	t.Run("generate with builder", func(t *testing.T) {
		ctx := context.Background()
		userID := "test-user"
		prompt := "Create a hello world function in Go"

		resp, err := service.GenerateWithBuilder(ctx, userID, prompt)
		require.NoError(t, err)
		require.NotNil(t, resp)

		assert.NotEmpty(t, resp.Content)
		assert.Equal(t, "openai", resp.Provider)
		assert.Greater(t, resp.TokensUsed, 0)
		assert.NotEmpty(t, resp.RequestID)
		assert.Contains(t, resp.Content, "package main")
	})

	t.Run("generate code legacy interface", func(t *testing.T) {
		ctx := context.Background()
		req := ai.GenerationRequest{
			UserID:   "test-user",
			Prompt:   "Create a simple function",
			Language: "javascript",
		}

		result, err := service.GenerateCode(ctx, req)
		require.NoError(t, err)

		assert.NotEmpty(t, result.Code)
		assert.NotEmpty(t, result.Model)
		assert.Greater(t, result.UsedTokens, 0)
	})

	t.Run("get available providers", func(t *testing.T) {
		providers := service.GetAvailableProviders()
		assert.Contains(t, providers, "openai")
		assert.Contains(t, providers, "vllm")
		assert.GreaterOrEqual(t, len(providers), 2)
	})

	t.Run("get provider info", func(t *testing.T) {
		info, err := service.GetProviderInfo("openai")
		require.NoError(t, err)
		assert.Equal(t, "OpenAI", info.Name)
		assert.True(t, info.FreeTier)

		info, err = service.GetProviderInfo("vllm")
		require.NoError(t, err)
		assert.Equal(t, "vLLM", info.Name)
		assert.True(t, info.FreeTier)

		_, err = service.GetProviderInfo("invalid")
		require.Error(t, err)
	})

	t.Run("health check", func(t *testing.T) {
		ctx := context.Background()
		results := service.HealthCheck(ctx)

		// Results should be empty initially since no providers are instantiated
		// Providers are created lazily on first use
		assert.NotNil(t, results)
	})

	t.Run("rate limiting integration", func(t *testing.T) {
		ctx := context.Background()
		userID := "rate-limited-user"

		// Exhaust rate limit
		for i := 0; i < 5; i++ {
			_, err := service.GenerateWithBuilder(ctx, userID, "test prompt")
			if err != nil {
				assert.Equal(t, llm.ErrRateLimitExceeded, err)
				break
			}
		}
	})

	t.Run("close service", func(t *testing.T) {
		err := service.Close()
		assert.NoError(t, err)
	})
}

func TestAIServiceRateLimitingAdapter(t *testing.T) {
	rateLimiter := ai.NewRateLimiter(1, 1) // Very restrictive: 1 req/sec, burst of 1
	quotaManager := ai.NewQuotaManager()
	cacheProvider, err := cache.NewMemoryProvider(cache.CacheConfig{})
	require.NoError(t, err)
	cacheManager := ai.NewCacheManager(cacheProvider, ai.DefaultCacheConfig())

	service := ai.NewAIService(rateLimiter, quotaManager, cacheManager)
	ctx := context.Background()

	t.Run("first request should succeed", func(t *testing.T) {
		resp, err := service.GenerateWithBuilder(ctx, "user1", "test prompt")
		require.NoError(t, err)
		require.NotNil(t, resp)
	})

	t.Run("immediate second request should be rate limited", func(t *testing.T) {
		resp, err := service.GenerateWithBuilder(ctx, "user1", "test prompt")
		require.Error(t, err)
		require.Nil(t, resp)
		assert.Equal(t, llm.ErrRateLimitExceeded, err)
	})

	// Cleanup
	service.Close()
}

func TestAIServiceQuotaAdapter(t *testing.T) {
	rateLimiter := ai.NewRateLimiter(100, 10) // Allow rate limiting
	quotaManager := ai.NewQuotaManager()

	// Initialize quota and then exhaust it for test user
	quotaManager.CheckQuota("quota-exceeded-user", 100) // Initialize quota
	for i := 0; i < 100; i++ {
		quotaManager.UseQuota("quota-exceeded-user")
	}

	cacheProvider, err := cache.NewMemoryProvider(cache.CacheConfig{})
	require.NoError(t, err)
	cacheManager := ai.NewCacheManager(cacheProvider, ai.DefaultCacheConfig())

	service := ai.NewAIService(rateLimiter, quotaManager, cacheManager)
	ctx := context.Background()

	t.Run("quota exceeded user should be blocked", func(t *testing.T) {
		resp, err := service.GenerateWithBuilder(ctx, "quota-exceeded-user", "test prompt")
		require.Error(t, err)
		require.Nil(t, resp)
		assert.Equal(t, llm.ErrQuotaExceeded, err)
	})

	t.Run("normal user should work", func(t *testing.T) {
		resp, err := service.GenerateWithBuilder(ctx, "normal-user", "test prompt")
		require.NoError(t, err)
		require.NotNil(t, resp)
	})

	// Cleanup
	service.Close()
}
