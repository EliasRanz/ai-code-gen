package config_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/EliasRanz/ai-code-gen/internal/ai"
	"github.com/EliasRanz/ai-code-gen/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAIConfigManager(t *testing.T) {
	// Set up test environment variables
	os.Setenv("AI_SERVICE_NAME", "ai-service")
	os.Setenv("AI_SERVICE_PORT", "8083")
	os.Setenv("AI_LLM_DEFAULT_PROVIDER", "openai")
	os.Setenv("AI_LLM_DEFAULT_MODEL", "gpt-4")
	os.Setenv("AI_LLM_DEFAULT_TEMPERATURE", "0.7")
	os.Setenv("AI_LLM_MAX_PROMPT_LENGTH", "10000")
	os.Setenv("AI_RATE_LIMIT_REQUESTS_PER_MINUTE", "60")
	os.Setenv("AI_QUOTA_DEFAULT_DAILY_LIMIT", "100")
	defer func() {
		os.Unsetenv("AI_SERVICE_NAME")
		os.Unsetenv("AI_SERVICE_PORT")
		os.Unsetenv("AI_LLM_DEFAULT_PROVIDER")
		os.Unsetenv("AI_LLM_DEFAULT_MODEL")
		os.Unsetenv("AI_LLM_DEFAULT_TEMPERATURE")
		os.Unsetenv("AI_LLM_MAX_PROMPT_LENGTH")
		os.Unsetenv("AI_RATE_LIMIT_REQUESTS_PER_MINUTE")
		os.Unsetenv("AI_QUOTA_DEFAULT_DAILY_LIMIT")
	}()

	provider := config.NewEnvironmentProvider("AI_")
	manager := ai.NewAIConfigManager(provider)

	t.Run("LoadConfiguration", func(t *testing.T) {
		ctx := context.Background()
		err := manager.LoadConfig(ctx)
		require.NoError(t, err)

		config := manager.GetConfig()
		assert.NotNil(t, config)

		// Check service configuration
		assert.Equal(t, "ai-service", config.Service.Name)
		assert.Equal(t, 8083, config.Service.Port)

		// Check LLM configuration
		assert.Equal(t, "openai", config.LLM.DefaultProvider)
		assert.Equal(t, "gpt-4", config.LLM.DefaultModel)
		assert.Equal(t, 0.7, config.LLM.DefaultTemperature)
		assert.Equal(t, 10000, config.LLM.MaxPromptLength)

		// Check rate limiting configuration
		assert.Equal(t, 60, config.RateLimit.RequestsPerMinute)

		// Check quota configuration
		assert.Equal(t, 100, config.Quota.DefaultDailyLimit)
	})

	t.Run("ApplyDefaults", func(t *testing.T) {
		// Test with minimal environment variables
		os.Unsetenv("AI_SERVICE_PORT")
		os.Unsetenv("AI_LLM_DEFAULT_MODEL")
		os.Unsetenv("AI_RATE_LIMIT_REQUESTS_PER_MINUTE")

		ctx := context.Background()
		err := manager.LoadConfig(ctx)
		require.NoError(t, err)

		config := manager.GetConfig()

		// Check defaults are applied
		assert.Equal(t, "0.0.0.0", config.Service.Host)
		assert.Equal(t, 8083, config.Service.Port)                 // Default port
		assert.Equal(t, "gpt-3.5-turbo", config.LLM.DefaultModel)  // Default model
		assert.Equal(t, 60, config.RateLimit.RequestsPerMinute)    // Default rate limit
		assert.Equal(t, 30*time.Second, config.LLM.DefaultTimeout) // Default timeout

		// Restore environment variables for other tests
		os.Setenv("AI_SERVICE_PORT", "8083")
		os.Setenv("AI_LLM_DEFAULT_MODEL", "gpt-4")
		os.Setenv("AI_RATE_LIMIT_REQUESTS_PER_MINUTE", "60")
	})

	t.Run("Reload", func(t *testing.T) {
		ctx := context.Background()

		// Initial load
		err := manager.LoadConfig(ctx)
		require.NoError(t, err)

		// Reload
		err = manager.Reload(ctx)
		assert.NoError(t, err)

		config := manager.GetConfig()
		assert.NotNil(t, config)
	})
}
