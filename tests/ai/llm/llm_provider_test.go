package ai_test

import (
	"context"
	"testing"

	"github.com/EliasRanz/ai-code-gen/internal/ai/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLLMFactory(t *testing.T) {
	factory := llm.NewLLMFactory()

	t.Run("list available providers", func(t *testing.T) {
		providers := factory.ListAvailableProviders()
		assert.Contains(t, providers, "openai")
		assert.Contains(t, providers, "vllm")
		assert.Len(t, providers, 2)
	})

	t.Run("create OpenAI provider", func(t *testing.T) {
		config := llm.ProviderConfig{
			FreeTierOnly: true,
			Model:        "gpt-3.5-turbo",
		}

		provider, err := factory.CreateProvider("openai", config)
		require.NoError(t, err)
		require.NotNil(t, provider)

		info := provider.GetProviderInfo()
		assert.Equal(t, "OpenAI", info.Name)
		assert.True(t, info.FreeTier)
	})

	t.Run("create vLLM provider", func(t *testing.T) {
		config := llm.ProviderConfig{
			FreeTierOnly: true,
			BaseURL:      "http://localhost:8000",
		}

		provider, err := factory.CreateProvider("vllm", config)
		require.NoError(t, err)
		require.NotNil(t, provider)

		info := provider.GetProviderInfo()
		assert.Equal(t, "vLLM", info.Name)
		assert.True(t, info.FreeTier)
	})

	t.Run("unknown provider should fail", func(t *testing.T) {
		config := llm.ProviderConfig{FreeTierOnly: true}

		provider, err := factory.CreateProvider("unknown", config)
		require.Error(t, err)
		require.Nil(t, provider)
		assert.Contains(t, err.Error(), "provider not found")
	})

	t.Run("non-free tier config should fail", func(t *testing.T) {
		config := llm.ProviderConfig{FreeTierOnly: false}

		provider, err := factory.CreateProvider("openai", config)
		require.Error(t, err)
		require.Nil(t, provider)
		assert.Contains(t, err.Error(), "free tier configuration required")
	})
}

func TestOpenAIClient(t *testing.T) {
	config := llm.ProviderConfig{
		FreeTierOnly: true,
		Model:        "gpt-3.5-turbo",
	}

	client, err := llm.NewOpenAIClient(config)
	require.NoError(t, err)
	require.NotNil(t, client)

	t.Run("provider info", func(t *testing.T) {
		info := client.GetProviderInfo()
		assert.Equal(t, "OpenAI", info.Name)
		assert.True(t, info.FreeTier)
		assert.Contains(t, info.Models, "gpt-3.5-turbo")
		assert.Contains(t, info.Capabilities, "code_generation")
	})

	t.Run("provider limits", func(t *testing.T) {
		limits := client.GetLimits()
		assert.Equal(t, 3, limits.RequestsPerMinute)
		assert.Equal(t, 200, limits.TokensPerMinute)
		assert.Equal(t, 100, limits.DailyQuota)
		assert.Equal(t, 2000, limits.MaxTokensPerRequest)
	})

	t.Run("health check", func(t *testing.T) {
		ctx := context.Background()
		err := client.HealthCheck(ctx)
		assert.NoError(t, err) // Should be healthy in free tier mode
	})

	t.Run("generate code - Go", func(t *testing.T) {
		ctx := context.Background()
		req := &llm.GenerationRequest{
			UserID:   "test-user",
			Prompt:   "Create a hello world function",
			Language: "go",
			Metadata: map[string]string{"request_id": "test-123"},
		}

		resp, err := client.GenerateCode(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)

		assert.NotEmpty(t, resp.Content)
		assert.Contains(t, resp.Content, "package main")
		assert.Contains(t, resp.Content, "fmt.Println")
		assert.Equal(t, "openai", resp.Provider)
		assert.Equal(t, "mock-gpt-3.5-turbo", resp.Model)
		assert.Greater(t, resp.TokensUsed, 0)
		assert.Equal(t, "test-123", resp.RequestID)
	})

	t.Run("generate code - JavaScript", func(t *testing.T) {
		ctx := context.Background()
		req := &llm.GenerationRequest{
			UserID:   "test-user",
			Prompt:   "Create a hello world function",
			Language: "javascript",
			Metadata: map[string]string{"request_id": "test-456"},
		}

		resp, err := client.GenerateCode(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)

		assert.NotEmpty(t, resp.Content)
		assert.Contains(t, resp.Content, "function")
		assert.Contains(t, resp.Content, "console.log")
		assert.Equal(t, "openai", resp.Provider)
	})

	t.Run("close client", func(t *testing.T) {
		err := client.Close()
		assert.NoError(t, err)
	})
}

func TestVLLMClient(t *testing.T) {
	config := llm.ProviderConfig{
		FreeTierOnly: true,
		BaseURL:      "http://localhost:8000",
		Model:        "codellama-7b",
	}

	client, err := llm.NewVLLMClient(config)
	require.NoError(t, err)
	require.NotNil(t, client)

	t.Run("provider info", func(t *testing.T) {
		info := client.GetProviderInfo()
		assert.Equal(t, "vLLM", info.Name)
		assert.True(t, info.FreeTier)
		assert.Contains(t, info.Models, "codellama-7b")
		assert.Contains(t, info.Capabilities, "code_generation")
	})

	t.Run("provider limits", func(t *testing.T) {
		limits := client.GetLimits()
		assert.Equal(t, 10, limits.RequestsPerMinute)
		assert.Equal(t, 1000, limits.TokensPerMinute)
		assert.Equal(t, 1000, limits.DailyQuota)
		assert.Equal(t, 4096, limits.MaxTokensPerRequest)
	})

	t.Run("health check", func(t *testing.T) {
		ctx := context.Background()
		err := client.HealthCheck(ctx)
		assert.NoError(t, err) // Should be healthy in mock mode
	})

	t.Run("generate code - Python", func(t *testing.T) {
		ctx := context.Background()
		req := &llm.GenerationRequest{
			UserID:   "test-user",
			Prompt:   "Create a hello world function",
			Language: "python",
			Metadata: map[string]string{"request_id": "test-789"},
		}

		resp, err := client.GenerateCode(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)

		assert.NotEmpty(t, resp.Content)
		assert.Contains(t, resp.Content, "def generate_code")
		assert.Contains(t, resp.Content, "print(")
		assert.Equal(t, "vllm", resp.Provider)
		assert.Equal(t, "codellama-7b", resp.Model)
		assert.Greater(t, resp.TokensUsed, 0)
		assert.Equal(t, "test-789", resp.RequestID)
	})
}

func TestOpenAIClientWithAPIKey(t *testing.T) {
	config := llm.ProviderConfig{
		FreeTierOnly: true,
		APIKey:       "test-api-key", // Should still work in free tier mode
		Model:        "gpt-3.5-turbo",
	}

	client, err := llm.NewOpenAIClient(config)
	require.Error(t, err) // Should fail because free tier mode + API key
	require.Nil(t, client)
	assert.Contains(t, err.Error(), "Free tier mode enabled but API key provided")
}
