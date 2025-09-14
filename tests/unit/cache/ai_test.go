package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/EliasRanz/ai-code-gen/internal/ai"
	"github.com/EliasRanz/ai-code-gen/internal/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAICacheIntegration tests AI service cache integration with core infrastructure
func TestAICacheIntegration(t *testing.T) {
	ctx := context.Background()

	// Create cache service for AI
	serviceConfig := cache.DefaultServiceConfig()
	service, err := cache.NewService(serviceConfig)
	require.NoError(t, err)
	defer service.Close()

	// Create AI cache manager using the service provider
	aiConfig := ai.DefaultCacheConfig()
	provider := service.GetProvider()
	cm := ai.NewCacheManager(provider, aiConfig)

	t.Run("AI Generation Caching", func(t *testing.T) {
		generationID := "test_generation_123"
		generation := &ai.CachedGenerationResult{
			GenerationID: generationID,
			UserID:       "user123",
			ProjectID:    "proj456",
			Prompt:       "Generate a hello world function",
			Response:     "Generated code content",
			Status:       "completed",
			Model:        "gpt-3.5-turbo",
			TokensUsed:   150,
		}

		// Test generation caching
		err := cm.CacheGeneration(ctx, generationID, generation)
		assert.NoError(t, err)

		cachedGeneration, err := cm.GetGeneration(ctx, generationID)
		assert.NoError(t, err)
		require.NotNil(t, cachedGeneration)
		assert.Equal(t, generation.GenerationID, cachedGeneration.GenerationID)
		assert.Equal(t, generation.UserID, cachedGeneration.UserID)
		assert.Equal(t, generation.Response, cachedGeneration.Response)
		assert.Equal(t, generation.Status, cachedGeneration.Status)
		assert.Equal(t, generation.Model, cachedGeneration.Model)
		assert.Equal(t, generation.TokensUsed, cachedGeneration.TokensUsed)

		// Test generation invalidation
		err = cm.InvalidateGeneration(ctx, generationID)
		assert.NoError(t, err)

		cachedGeneration, err = cm.GetGeneration(ctx, generationID)
		assert.NoError(t, err)
		assert.Nil(t, cachedGeneration)
	})

	t.Run("AI Rate Limit Caching", func(t *testing.T) {
		userID := "rate_limit_user"
		rateLimitData := &ai.RateLimitData{
			UserID:       userID,
			RequestCount: 10,
			TokensUsed:   500,
			WindowStart:  time.Now().Add(-time.Hour),
			WindowEnd:    time.Now(),
			Tier:         "free",
			IsBlocked:    false,
		}

		// Test rate limit caching
		err := cm.CacheRateLimit(ctx, userID, rateLimitData)
		assert.NoError(t, err)

		cachedRateLimit, err := cm.GetRateLimit(ctx, userID)
		assert.NoError(t, err)
		require.NotNil(t, cachedRateLimit)
		assert.Equal(t, rateLimitData.UserID, cachedRateLimit.UserID)
		assert.Equal(t, rateLimitData.RequestCount, cachedRateLimit.RequestCount)
		assert.Equal(t, rateLimitData.TokensUsed, cachedRateLimit.TokensUsed)
		assert.Equal(t, rateLimitData.Tier, cachedRateLimit.Tier)
		assert.Equal(t, rateLimitData.IsBlocked, cachedRateLimit.IsBlocked)

		// Test rate limit invalidation
		err = cm.InvalidateRateLimit(ctx, userID)
		assert.NoError(t, err)

		cachedRateLimit, err = cm.GetRateLimit(ctx, userID)
		assert.NoError(t, err)
		assert.Nil(t, cachedRateLimit)
	})

	t.Run("AI Model Response Caching", func(t *testing.T) {
		requestHash := "model_response_123"
		modelResponse := &ai.ModelResponse{
			RequestHash: requestHash,
			Model:       "gpt-3.5-turbo",
			Response:    "Model response content",
			TokensUsed:  100,
		}

		// Test model response caching
		err := cm.CacheModelResponse(ctx, requestHash, modelResponse)
		assert.NoError(t, err)

		cachedResponse, err := cm.GetModelResponse(ctx, requestHash)
		assert.NoError(t, err)
		require.NotNil(t, cachedResponse)
		assert.Equal(t, modelResponse.RequestHash, cachedResponse.RequestHash)
		assert.Equal(t, modelResponse.Model, cachedResponse.Model)
		assert.Equal(t, modelResponse.Response, cachedResponse.Response)
		assert.Equal(t, modelResponse.TokensUsed, cachedResponse.TokensUsed)

		// Note: AI cache doesn't have InvalidateModelResponse method
		// Model responses are typically cached with TTL and expire automatically
	})

	t.Run("AI Cache Health Check", func(t *testing.T) {
		// Test that AI cache works with infrastructure health checks
		err := service.HealthCheck(ctx)
		assert.NoError(t, err)

		// Test provider is accessible through cache manager
		assert.NotNil(t, provider)
		err = provider.HealthCheck(ctx)
		assert.NoError(t, err)
	})
}

// TestAICacheConfiguration tests AI cache configuration integration
func TestAICacheConfiguration(t *testing.T) {
	t.Run("Default AI Configuration", func(t *testing.T) {
		config := ai.DefaultCacheConfig()
		assert.Equal(t, 30*time.Minute, config.TTL)
		assert.Equal(t, "ai:generation:", config.GenerationKeyPrefix)
		assert.Equal(t, "ai:rate_limit:", config.RateLimitKeyPrefix)
		assert.Equal(t, "ai:model_response:", config.ModelResponseKeyPrefix)
	})

	t.Run("Custom AI Configuration", func(t *testing.T) {
		serviceConfig := cache.DefaultServiceConfig()
		service, err := cache.NewService(serviceConfig)
		require.NoError(t, err)
		defer service.Close()

		customConfig := ai.CacheConfig{
			TTL:                    45 * time.Minute,
			GenerationKeyPrefix:    "custom:ai:generation:",
			RateLimitKeyPrefix:     "custom:ai:rate_limit:",
			ModelResponseKeyPrefix: "custom:ai:model_response:",
		}

		provider := service.GetProvider()
		cm := ai.NewCacheManager(provider, customConfig)

		// Test that custom configuration is used
		ctx := context.Background()
		generationID := "custom_test_generation"
		generation := &ai.CachedGenerationResult{
			GenerationID: generationID,
			UserID:       "custom_user",
			ProjectID:    "custom_proj",
			Prompt:       "Generate custom code",
			Response:     "Custom generated content",
			Status:       "completed",
			Model:        "gpt-4",
		}

		err = cm.CacheGeneration(ctx, generationID, generation)
		assert.NoError(t, err)

		cached, err := cm.GetGeneration(ctx, generationID)
		assert.NoError(t, err)
		assert.NotNil(t, cached)
		assert.Equal(t, generation.Response, cached.Response)
	})
}

// BenchmarkAICacheOperations benchmarks AI-specific cache operations
func BenchmarkAICacheOperations(b *testing.B) {
	ctx := context.Background()
	serviceConfig := cache.DefaultServiceConfig()
	service, _ := cache.NewService(serviceConfig)
	defer service.Close()

	aiConfig := ai.DefaultCacheConfig()
	provider := service.GetProvider()
	cm := ai.NewCacheManager(provider, aiConfig)

	generation := &ai.CachedGenerationResult{
		GenerationID: "bench_generation",
		UserID:       "bench_user",
		ProjectID:    "bench_proj",
		Prompt:       "Generate benchmark code",
		Response:     "Benchmark generated content",
		Status:       "completed",
		Model:        "gpt-3.5-turbo",
		TokensUsed:   100,
	}

	b.Run("AICacheGeneration", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			generationID := "bench_generation_" + string(rune(i))
			cm.CacheGeneration(ctx, generationID, generation)
		}
	})

	b.Run("AIGetGeneration", func(b *testing.B) {
		generationID := "bench_get_generation"
		cm.CacheGeneration(ctx, generationID, generation)
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			cm.GetGeneration(ctx, generationID)
		}
	})

	b.Run("AICacheRateLimit", func(b *testing.B) {
		rateLimitData := &ai.RateLimitData{
			UserID:       "bench_user",
			RequestCount: 5,
			TokensUsed:   250,
			WindowStart:  time.Now().Add(-30 * time.Minute),
			WindowEnd:    time.Now(),
			Tier:         "free",
			IsBlocked:    false,
		}

		for i := 0; i < b.N; i++ {
			userID := "bench_rate_limit_user_" + string(rune(i))
			cm.CacheRateLimit(ctx, userID, rateLimitData)
		}
	})
}
