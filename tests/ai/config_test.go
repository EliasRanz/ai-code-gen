package ai_test

import (
	"testing"
	"time"

	"github.com/EliasRanz/ai-code-gen/internal/ai"
	"github.com/EliasRanz/ai-code-gen/internal/ai/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	config := ai.DefaultConfig()

	t.Run("rate limit config", func(t *testing.T) {
		assert.Equal(t, 60, config.RateLimit.RequestsPerMinute)
		assert.Equal(t, 10, config.RateLimit.BurstSize)
		assert.Equal(t, 10*time.Minute, config.RateLimit.CleanupInterval)
	})

	t.Run("quota config", func(t *testing.T) {
		assert.Equal(t, 100, config.Quota.DefaultDailyLimit)
		assert.Equal(t, 1000, config.Quota.PremiumDailyLimit)
		assert.Equal(t, "00:00", config.Quota.ResetTime)
		assert.True(t, config.Quota.TrackingEnabled)
		assert.Equal(t, time.Hour, config.Quota.CleanupInterval)
	})

	t.Run("LLM config", func(t *testing.T) {
		assert.Equal(t, "openai", config.LLM.DefaultProvider)
		assert.True(t, config.LLM.FreeTierOnly)
		assert.Equal(t, 8000, config.LLM.MaxPromptLength)
		assert.Equal(t, 2000, config.LLM.MaxTokensPerReq)
		assert.Equal(t, "gpt-3.5-turbo", config.LLM.DefaultModel)
		assert.Equal(t, 0.7, config.LLM.DefaultTemperature)
		assert.Equal(t, 1000, config.LLM.DefaultMaxTokens)
		assert.Equal(t, 30*time.Second, config.LLM.DefaultTimeout)
		assert.Equal(t, 5*time.Minute, config.LLM.HealthCheckInterval)
		assert.Equal(t, 10*time.Second, config.LLM.HealthCheckTimeout)

		// Check provider configs
		openaiConfig, exists := config.LLM.Providers["openai"]
		assert.True(t, exists)
		assert.True(t, openaiConfig.FreeTierOnly)
		assert.Equal(t, "gpt-3.5-turbo", openaiConfig.Model)
		assert.Equal(t, 30*time.Second, openaiConfig.Timeout)

		vllmConfig, exists := config.LLM.Providers["vllm"]
		assert.True(t, exists)
		assert.True(t, vllmConfig.FreeTierOnly)
		assert.Equal(t, "http://localhost:8000", vllmConfig.BaseURL)
		assert.Equal(t, "codellama-7b", vllmConfig.Model)
		assert.Equal(t, 60*time.Second, vllmConfig.Timeout)
	})

	t.Run("service config", func(t *testing.T) {
		assert.True(t, config.Service.EnableValidation)
		assert.True(t, config.Service.EnableCaching)
		assert.Equal(t, 30*time.Minute, config.Service.CacheTTL)
		assert.True(t, config.Service.EnableMetrics)
		assert.True(t, config.Service.EnableTracing)
		assert.Equal(t, "ai_service", config.Service.MetricsPrefix)
		assert.Equal(t, 3, config.Service.MaxRetries)
		assert.Equal(t, time.Second, config.Service.RetryDelay)
		assert.True(t, config.Service.CircuitBreakerEnabled)
	})
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  func() ai.Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid default config",
			config: func() ai.Config {
				return ai.DefaultConfig()
			},
			wantErr: false,
		},
		{
			name: "invalid rate limit requests per minute",
			config: func() ai.Config {
				c := ai.DefaultConfig()
				c.RateLimit.RequestsPerMinute = 0
				return c
			},
			wantErr: true,
			errMsg:  "rate limit requests per minute must be positive",
		},
		{
			name: "invalid rate limit burst size",
			config: func() ai.Config {
				c := ai.DefaultConfig()
				c.RateLimit.BurstSize = -1
				return c
			},
			wantErr: true,
			errMsg:  "rate limit burst size must be positive",
		},
		{
			name: "invalid default daily limit",
			config: func() ai.Config {
				c := ai.DefaultConfig()
				c.Quota.DefaultDailyLimit = 0
				return c
			},
			wantErr: true,
			errMsg:  "default daily limit must be positive",
		},
		{
			name: "empty default provider",
			config: func() ai.Config {
				c := ai.DefaultConfig()
				c.LLM.DefaultProvider = ""
				return c
			},
			wantErr: true,
			errMsg:  "default LLM provider must be specified",
		},
		{
			name: "invalid max prompt length",
			config: func() ai.Config {
				c := ai.DefaultConfig()
				c.LLM.MaxPromptLength = 0
				return c
			},
			wantErr: true,
			errMsg:  "max prompt length must be positive",
		},
		{
			name: "invalid temperature - too low",
			config: func() ai.Config {
				c := ai.DefaultConfig()
				c.LLM.DefaultTemperature = -0.1
				return c
			},
			wantErr: true,
			errMsg:  "default temperature must be between 0 and 2.0",
		},
		{
			name: "invalid temperature - too high",
			config: func() ai.Config {
				c := ai.DefaultConfig()
				c.LLM.DefaultTemperature = 2.1
				return c
			},
			wantErr: true,
			errMsg:  "default temperature must be between 0 and 2.0",
		},
		{
			name: "invalid max retries",
			config: func() ai.Config {
				c := ai.DefaultConfig()
				c.Service.MaxRetries = -1
				return c
			},
			wantErr: true,
			errMsg:  "max retries cannot be negative",
		},
		{
			name: "default provider not in providers map",
			config: func() ai.Config {
				c := ai.DefaultConfig()
				c.LLM.DefaultProvider = "nonexistent"
				return c
			},
			wantErr: true,
			errMsg:  "default provider nonexistent not found in providers configuration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tt.config()
			err := config.Validate()

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGetProviderConfig(t *testing.T) {
	config := ai.DefaultConfig()

	t.Run("get existing provider config", func(t *testing.T) {
		providerConfig, exists := config.GetProviderConfig("openai")
		assert.True(t, exists)
		assert.True(t, providerConfig.FreeTierOnly)
		assert.Equal(t, "gpt-3.5-turbo", providerConfig.Model)
	})

	t.Run("get nonexistent provider config", func(t *testing.T) {
		providerConfig, exists := config.GetProviderConfig("nonexistent")
		assert.False(t, exists)
		assert.Equal(t, llm.ProviderConfig{}, providerConfig)
	})
}

func TestConfigCacheDefaults(t *testing.T) {
	config := ai.DefaultConfig()

	assert.Equal(t, 30*time.Minute, config.Cache.TTL)
	assert.Equal(t, "ai:generation:", config.Cache.GenerationKeyPrefix)
	assert.Equal(t, "ai:rate_limit:", config.Cache.RateLimitKeyPrefix)
	assert.Equal(t, "ai:model_response:", config.Cache.ModelResponseKeyPrefix)
}
