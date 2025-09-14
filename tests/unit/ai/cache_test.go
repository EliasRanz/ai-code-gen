package ai_test

import (
	"context"
	"testing"
	"time"

	"github.com/EliasRanz/ai-code-gen/internal/ai"
	"github.com/EliasRanz/ai-code-gen/internal/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAICacheManager(t *testing.T) {
	ctx := context.Background()

	// Create memory provider for testing
	cacheConfig := cache.DefaultCacheConfig()
	provider, err := cache.NewMemoryProvider(cacheConfig)
	require.NoError(t, err)
	defer provider.Close()

	// Create AI cache manager
	aiConfig := ai.DefaultCacheConfig()
	cm := ai.NewCacheManager(provider, aiConfig)

	t.Run("Generation Caching", func(t *testing.T) {
		generationID := "gen123"
		generation := &ai.CachedGenerationResult{
			GenerationID: generationID,
			UserID:       "user123",
			ProjectID:    "proj123",
			Prompt:       "Generate a React component",
			Response:     "import React from 'react';...",
			Status:       "completed",
			Model:        "gpt-4",
			TokensUsed:   150,
			CreatedAt:    time.Now(),
		}

		// Test Cache Generation
		err := cm.CacheGeneration(ctx, generationID, generation)
		assert.NoError(t, err)

		// Test Get Generation
		cached, err := cm.GetGeneration(ctx, generationID)
		assert.NoError(t, err)
		require.NotNil(t, cached)
		assert.Equal(t, generation.GenerationID, cached.GenerationID)
		assert.Equal(t, generation.UserID, cached.UserID)
		assert.Equal(t, generation.Prompt, cached.Prompt)
		assert.Equal(t, generation.Response, cached.Response)
		assert.Equal(t, generation.TokensUsed, cached.TokensUsed)
		assert.False(t, cached.CachedAt.IsZero())

		// Test Invalidate Generation
		err = cm.InvalidateGeneration(ctx, generationID)
		assert.NoError(t, err)

		// Verify deletion
		cached, err = cm.GetGeneration(ctx, generationID)
		assert.NoError(t, err)
		assert.Nil(t, cached)
	})

	t.Run("User Generations Caching", func(t *testing.T) {
		userID := "user123"
		generations := []ai.GenerationSummary{
			{
				GenerationID: "gen1",
				ProjectID:    "proj1",
				Prompt:       "Create a button component",
				Status:       "completed",
				TokensUsed:   100,
				CreatedAt:    time.Now().Add(-2 * time.Hour),
			},
			{
				GenerationID: "gen2",
				ProjectID:    "proj2",
				Prompt:       "Generate API endpoints",
				Status:       "processing",
				TokensUsed:   250,
				CreatedAt:    time.Now().Add(-1 * time.Hour),
			},
		}

		// Test Cache User Generations
		err := cm.CacheUserGenerations(ctx, userID, generations)
		assert.NoError(t, err)

		// Test Get User Generations
		cached, err := cm.GetUserGenerations(ctx, userID)
		assert.NoError(t, err)
		require.NotNil(t, cached)
		assert.Len(t, cached, 2)
		assert.Equal(t, generations[0].GenerationID, cached[0].GenerationID)
		assert.Equal(t, generations[1].Prompt, cached[1].Prompt)

		// Test Invalidate User Generations
		err = cm.InvalidateUserGenerations(ctx, userID)
		assert.NoError(t, err)

		// Verify deletion
		cached, err = cm.GetUserGenerations(ctx, userID)
		assert.NoError(t, err)
		assert.Nil(t, cached)
	})

	t.Run("Model Response Caching", func(t *testing.T) {
		requestHash := "hash123"
		response := &ai.ModelResponse{
			RequestHash: requestHash,
			Model:       "gpt-4",
			Response:    "Generated code response",
			TokensUsed:  200,
		}

		// Test Cache Model Response
		err := cm.CacheModelResponse(ctx, requestHash, response)
		assert.NoError(t, err)

		// Test Get Model Response
		cached, err := cm.GetModelResponse(ctx, requestHash)
		assert.NoError(t, err)
		require.NotNil(t, cached)
		assert.Equal(t, response.RequestHash, cached.RequestHash)
		assert.Equal(t, response.Model, cached.Model)
		assert.Equal(t, response.Response, cached.Response)
		assert.Equal(t, response.TokensUsed, cached.TokensUsed)
		assert.False(t, cached.CachedAt.IsZero())
	})

	t.Run("Rate Limit Caching", func(t *testing.T) {
		userID := "user123"
		rateLimitData := &ai.RateLimitData{
			UserID:       userID,
			RequestCount: 5,
			TokensUsed:   1000,
			WindowStart:  time.Now().Add(-time.Hour),
			WindowEnd:    time.Now(),
			Tier:         "free",
			IsBlocked:    false,
		}

		// Test Cache Rate Limit
		err := cm.CacheRateLimit(ctx, userID, rateLimitData)
		assert.NoError(t, err)

		// Test Get Rate Limit
		cached, err := cm.GetRateLimit(ctx, userID)
		assert.NoError(t, err)
		require.NotNil(t, cached)
		assert.Equal(t, rateLimitData.UserID, cached.UserID)
		assert.Equal(t, rateLimitData.RequestCount, cached.RequestCount)
		assert.Equal(t, rateLimitData.TokensUsed, cached.TokensUsed)
		assert.Equal(t, rateLimitData.Tier, cached.Tier)
		assert.Equal(t, rateLimitData.IsBlocked, cached.IsBlocked)
		assert.False(t, cached.CachedAt.IsZero())

		// Test Invalidate Rate Limit
		err = cm.InvalidateRateLimit(ctx, userID)
		assert.NoError(t, err)

		// Verify deletion
		cached, err = cm.GetRateLimit(ctx, userID)
		assert.NoError(t, err)
		assert.Nil(t, cached)
	})

	t.Run("Key Generation", func(t *testing.T) {
		// Test generation key
		genKey := cm.GenerateKey("generation", "gen123")
		assert.Equal(t, "ai:generation:gen123", genKey)

		// Test user generations key
		userGenKey := cm.GenerateKey("user_generations", "user123")
		assert.Equal(t, "ai:user_generations:user123", userGenKey)

		// Test model response key
		modelKey := cm.GenerateKey("model_response", "hash123")
		assert.Equal(t, "ai:model_response:hash123", modelKey)

		// Test rate limit key
		rateLimitKey := cm.GenerateKey("rate_limit", "user123")
		assert.Equal(t, "ai:rate_limit:user123", rateLimitKey)

		// Test unknown key type
		unknownKey := cm.GenerateKey("unknown", "test")
		assert.Equal(t, "ai:unknown:test", unknownKey)
	})

	t.Run("JSON Operations", func(t *testing.T) {
		key := "test:json:key"
		data := map[string]interface{}{
			"model":       "gpt-4",
			"tokens_used": 150,
			"completed":   true,
		}

		// Test SetJSON
		err := cm.SetJSON(ctx, key, data, time.Minute)
		assert.NoError(t, err)

		// Test GetJSON
		var result map[string]interface{}
		err = cm.GetJSON(ctx, key, &result)
		assert.NoError(t, err)
		assert.Equal(t, data["model"], result["model"])
		assert.Equal(t, float64(150), result["tokens_used"]) // JSON numbers are float64
		assert.Equal(t, data["completed"], result["completed"])
	})

	t.Run("Input Validation", func(t *testing.T) {
		// Empty generation ID
		err := cm.CacheGeneration(ctx, "", &ai.CachedGenerationResult{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "generation ID cannot be empty")

		_, err = cm.GetGeneration(ctx, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "generation ID cannot be empty")

		// Nil generation result
		err = cm.CacheGeneration(ctx, "gen", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "generation result cannot be nil")

		// Empty user ID
		err = cm.CacheUserGenerations(ctx, "", []ai.GenerationSummary{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user ID cannot be empty")

		// Empty request hash
		err = cm.CacheModelResponse(ctx, "", &ai.ModelResponse{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "request hash cannot be empty")

		// Nil model response
		err = cm.CacheModelResponse(ctx, "hash", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "model response cannot be nil")

		// Empty user ID for rate limit
		err = cm.CacheRateLimit(ctx, "", &ai.RateLimitData{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user ID cannot be empty")

		// Nil rate limit data
		err = cm.CacheRateLimit(ctx, "user", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rate limit data cannot be nil")
	})

	t.Run("TTL Configuration", func(t *testing.T) {
		// Model response should have shorter TTL (15 minutes)
		requestHash := "short_ttl_hash"
		response := &ai.ModelResponse{
			RequestHash: requestHash,
			Model:       "gpt-4",
			Response:    "test response",
			TokensUsed:  100,
		}

		err := cm.CacheModelResponse(ctx, requestHash, response)
		assert.NoError(t, err)

		// Rate limit should have 1 hour TTL
		userID := "rate_limit_user"
		rateLimitData := &ai.RateLimitData{
			UserID:       userID,
			RequestCount: 1,
			TokensUsed:   100,
			Tier:         "free",
		}

		err = cm.CacheRateLimit(ctx, userID, rateLimitData)
		assert.NoError(t, err)

		// Both should be retrievable immediately
		cached, err := cm.GetModelResponse(ctx, requestHash)
		assert.NoError(t, err)
		assert.NotNil(t, cached)

		cachedRL, err := cm.GetRateLimit(ctx, userID)
		assert.NoError(t, err)
		assert.NotNil(t, cachedRL)
	})

	t.Run("Health Check", func(t *testing.T) {
		err := cm.HealthCheck(ctx)
		assert.NoError(t, err)
	})
}

func TestAICacheConfig(t *testing.T) {
	t.Run("Default Config", func(t *testing.T) {
		config := ai.DefaultCacheConfig()
		assert.Equal(t, 30*time.Minute, config.TTL)
		assert.Equal(t, "ai:generation:", config.GenerationKeyPrefix)
		assert.Equal(t, "ai:rate_limit:", config.RateLimitKeyPrefix)
		assert.Equal(t, "ai:model_response:", config.ModelResponseKeyPrefix)
	})
}

func BenchmarkAICacheManager(b *testing.B) {
	ctx := context.Background()

	cacheConfig := cache.DefaultCacheConfig()
	provider, _ := cache.NewMemoryProvider(cacheConfig)
	defer provider.Close()

	aiConfig := ai.DefaultCacheConfig()
	cm := ai.NewCacheManager(provider, aiConfig)

	generation := &ai.CachedGenerationResult{
		GenerationID: "gen123",
		UserID:       "user123",
		ProjectID:    "proj123",
		Prompt:       "Generate code",
		Response:     "Generated code response",
		Status:       "completed",
		Model:        "gpt-4",
		TokensUsed:   150,
	}

	b.Run("CacheGeneration", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			genID := "gen_" + string(rune(i))
			cm.CacheGeneration(ctx, genID, generation)
		}
	})

	b.Run("GetGeneration", func(b *testing.B) {
		genID := "bench_gen"
		cm.CacheGeneration(ctx, genID, generation)
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			cm.GetGeneration(ctx, genID)
		}
	})

	modelResponse := &ai.ModelResponse{
		RequestHash: "hash123",
		Model:       "gpt-4",
		Response:    "response",
		TokensUsed:  100,
	}

	b.Run("CacheModelResponse", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			hash := "hash_" + string(rune(i))
			cm.CacheModelResponse(ctx, hash, modelResponse)
		}
	})

	rateLimitData := &ai.RateLimitData{
		UserID:       "user123",
		RequestCount: 1,
		TokensUsed:   100,
		Tier:         "free",
	}

	b.Run("CacheRateLimit", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			userID := "user_" + string(rune(i))
			cm.CacheRateLimit(ctx, userID, rateLimitData)
		}
	})
}
