package ai_test

import (
	"context"
	"testing"

	"github.com/EliasRanz/ai-code-gen/internal/ai/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockRateLimiter implements llm.RateLimiter for testing
type mockRateLimiter struct {
	allowed map[string]bool
}

func (m *mockRateLimiter) Allow(userID string) bool {
	if m.allowed == nil {
		return true
	}
	return m.allowed[userID]
}

// mockQuotaManager implements llm.QuotaManager for testing
type mockQuotaManager struct {
	exceeded map[string]bool
}

func (m *mockQuotaManager) IsQuotaExceeded(userID string) bool {
	if m.exceeded == nil {
		return false
	}
	return m.exceeded[userID]
}

func (m *mockQuotaManager) IncrementUsage(userID string, tokens int) error {
	return nil
}

func TestLLMOrchestrator(t *testing.T) {
	rateLimiter := &mockRateLimiter{}
	quotaManager := &mockQuotaManager{}

	orchestrator := llm.NewLLMOrchestrator(rateLimiter, quotaManager)
	require.NotNil(t, orchestrator)

	t.Run("get available providers", func(t *testing.T) {
		providers := orchestrator.GetAvailableProviders()
		assert.Contains(t, providers, "openai")
		assert.Contains(t, providers, "vllm")
	})

	t.Run("generate code with default provider", func(t *testing.T) {
		ctx := context.Background()
		req := &llm.GenerationRequest{
			UserID:   "test-user",
			Prompt:   "Create a hello world function",
			Language: "go",
			Metadata: map[string]string{},
		}

		resp, err := orchestrator.GenerateCode(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)

		assert.NotEmpty(t, resp.Content)
		assert.Equal(t, "openai", resp.Provider)
		assert.Greater(t, resp.TokensUsed, 0)
	})

	t.Run("generate code with specific provider", func(t *testing.T) {
		ctx := context.Background()
		req := &llm.GenerationRequest{
			UserID:   "test-user",
			Prompt:   "Create a hello world function",
			Language: "python",
			Provider: "vllm",
			Metadata: map[string]string{},
		}

		resp, err := orchestrator.GenerateCode(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)

		assert.NotEmpty(t, resp.Content)
		assert.Equal(t, "vllm", resp.Provider)
		assert.Greater(t, resp.TokensUsed, 0)
	})

	t.Run("rate limit exceeded", func(t *testing.T) {
		rateLimiter.allowed = map[string]bool{"blocked-user": false}

		ctx := context.Background()
		req := &llm.GenerationRequest{
			UserID:   "blocked-user",
			Prompt:   "test",
			Metadata: map[string]string{},
		}

		resp, err := orchestrator.GenerateCode(ctx, req)
		require.Error(t, err)
		require.Nil(t, resp)
		assert.Equal(t, llm.ErrRateLimitExceeded, err)
	})

	t.Run("quota exceeded", func(t *testing.T) {
		rateLimiter.allowed = nil // Allow rate limiting
		quotaManager.exceeded = map[string]bool{"quota-exceeded-user": true}

		ctx := context.Background()
		req := &llm.GenerationRequest{
			UserID:   "quota-exceeded-user",
			Prompt:   "test",
			Metadata: map[string]string{},
		}

		resp, err := orchestrator.GenerateCode(ctx, req)
		require.Error(t, err)
		require.Nil(t, resp)
		assert.Equal(t, llm.ErrQuotaExceeded, err)
	})

	t.Run("invalid provider", func(t *testing.T) {
		ctx := context.Background()
		req := &llm.GenerationRequest{
			UserID:   "test-user",
			Prompt:   "test",
			Provider: "invalid-provider",
			Metadata: map[string]string{},
		}

		resp, err := orchestrator.GenerateCode(ctx, req)
		require.Error(t, err)
		require.Nil(t, resp)
		assert.Contains(t, err.Error(), "provider not found")
	})

	t.Run("health check all providers", func(t *testing.T) {
		ctx := context.Background()

		// First, trigger provider creation by making requests
		req1 := &llm.GenerationRequest{
			UserID:   "test-user",
			Prompt:   "test",
			Provider: "openai",
			Metadata: map[string]string{},
		}
		_, err := orchestrator.GenerateCode(ctx, req1)
		require.NoError(t, err)

		req2 := &llm.GenerationRequest{
			UserID:   "test-user",
			Prompt:   "test",
			Provider: "vllm",
			Metadata: map[string]string{},
		}
		_, err = orchestrator.GenerateCode(ctx, req2)
		require.NoError(t, err)

		// Now check health of all providers
		results := orchestrator.HealthCheck(ctx)
		assert.Contains(t, results, "openai")
		assert.Contains(t, results, "vllm")
		assert.NoError(t, results["openai"])
		assert.NoError(t, results["vllm"])
	})

	t.Run("get provider info", func(t *testing.T) {
		info, err := orchestrator.GetProviderInfo("openai")
		require.NoError(t, err)
		assert.Equal(t, "OpenAI", info.Name)
		assert.True(t, info.FreeTier)

		info, err = orchestrator.GetProviderInfo("vllm")
		require.NoError(t, err)
		assert.Equal(t, "vLLM", info.Name)
		assert.True(t, info.FreeTier)

		_, err = orchestrator.GetProviderInfo("invalid")
		require.Error(t, err)
	})

	t.Run("register custom provider", func(t *testing.T) {
		mockProvider := &mockLLMProvider{}
		orchestrator.RegisterProvider("mock", mockProvider)

		providers := orchestrator.GetAvailableProviders()
		assert.Contains(t, providers, "mock")
	})

	t.Run("close orchestrator", func(t *testing.T) {
		err := orchestrator.Close()
		assert.NoError(t, err)
	})
}

// mockLLMProvider implements llm.LLMProvider for testing
type mockLLMProvider struct{}

func (m *mockLLMProvider) GenerateCode(ctx context.Context, req *llm.GenerationRequest) (*llm.GenerationResponse, error) {
	return &llm.GenerationResponse{
		Content:    "mock response",
		TokensUsed: 10,
		Provider:   "mock",
		Model:      "mock-model",
		RequestID:  req.Metadata["request_id"],
	}, nil
}

func (m *mockLLMProvider) HealthCheck(ctx context.Context) error {
	return nil
}

func (m *mockLLMProvider) GetProviderInfo() llm.ProviderInfo {
	return llm.ProviderInfo{
		Name:         "Mock",
		Version:      "1.0",
		Models:       []string{"mock-model"},
		Capabilities: []string{"testing"},
		FreeTier:     true,
	}
}

func (m *mockLLMProvider) GetLimits() llm.ProviderLimits {
	return llm.ProviderLimits{
		RequestsPerMinute:   100,
		TokensPerMinute:     1000,
		DailyQuota:          10000,
		MaxTokensPerRequest: 4096,
	}
}

func (m *mockLLMProvider) Close() error {
	return nil
}
