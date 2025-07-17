package config_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/EliasRanz/ai-code-gen/internal/config"
	"github.com/EliasRanz/ai-code-gen/internal/gateway"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGatewayConfigManager(t *testing.T) {
	// Set up test environment variables
	os.Setenv("GATEWAY_SERVICE_NAME", "api-gateway")
	os.Setenv("GATEWAY_SERVICE_PORT", "8080")
	os.Setenv("GATEWAY_RATE_LIMIT_ENABLED", "true")
	os.Setenv("GATEWAY_RATE_LIMIT_REQUESTS_PER_SECOND", "100")
	os.Setenv("GATEWAY_CORS_ENABLED", "true")
	os.Setenv("GATEWAY_CORS_ALLOWED_ORIGINS", "http://localhost:3000,https://example.com")
	os.Setenv("GATEWAY_AUTH_ENABLED", "true")
	os.Setenv("GATEWAY_AUTH_AUTH_SERVICE_URL", "http://localhost:8081")
	defer func() {
		os.Unsetenv("GATEWAY_SERVICE_NAME")
		os.Unsetenv("GATEWAY_SERVICE_PORT")
		os.Unsetenv("GATEWAY_RATE_LIMIT_ENABLED")
		os.Unsetenv("GATEWAY_RATE_LIMIT_REQUESTS_PER_SECOND")
		os.Unsetenv("GATEWAY_CORS_ENABLED")
		os.Unsetenv("GATEWAY_CORS_ALLOWED_ORIGINS")
		os.Unsetenv("GATEWAY_AUTH_ENABLED")
		os.Unsetenv("GATEWAY_AUTH_AUTH_SERVICE_URL")
	}()

	provider := config.NewEnvironmentProvider("GATEWAY_")
	manager := gateway.NewGatewayConfigManager(provider)

	t.Run("LoadConfiguration", func(t *testing.T) {
		ctx := context.Background()
		err := manager.LoadConfig(ctx)
		require.NoError(t, err)

		config := manager.GetConfig()
		assert.NotNil(t, config)

		// Check service configuration
		assert.Equal(t, "api-gateway", config.Service.Name)
		assert.Equal(t, 8080, config.Service.Port)

		// Check rate limiting configuration
		assert.True(t, config.RateLimit.Enabled)
		assert.Equal(t, 100, config.RateLimit.RequestsPerSec)

		// Check CORS configuration
		assert.True(t, config.CORS.Enabled)
		assert.Equal(t, []string{"http://localhost:3000", "https://example.com"}, config.CORS.AllowedOrigins)

		// Check auth proxy configuration
		assert.True(t, config.Auth.Enabled)
		assert.Equal(t, "http://localhost:8081", config.Auth.AuthServiceURL)
	})

	t.Run("ApplyDefaults", func(t *testing.T) {
		// Test with minimal environment variables
		os.Unsetenv("GATEWAY_SERVICE_PORT")
		os.Unsetenv("GATEWAY_RATE_LIMIT_REQUESTS_PER_SECOND")
		os.Unsetenv("GATEWAY_CORS_ALLOWED_ORIGINS")

		ctx := context.Background()
		err := manager.LoadConfig(ctx)
		require.NoError(t, err)

		config := manager.GetConfig()

		// Check defaults are applied
		assert.Equal(t, "0.0.0.0", config.Service.Host)
		assert.Equal(t, 8080, config.Service.Port)                 // Default port
		assert.Equal(t, 100, config.RateLimit.RequestsPerSec)      // Default rate limit
		assert.Equal(t, []string{"*"}, config.CORS.AllowedOrigins) // Default CORS
		assert.Equal(t, "Authorization", config.Auth.TokenHeader)  // Default header
		assert.Equal(t, 5*time.Second, config.Auth.Timeout)        // Default timeout

		// Restore environment variables for other tests
		os.Setenv("GATEWAY_SERVICE_PORT", "8080")
		os.Setenv("GATEWAY_RATE_LIMIT_REQUESTS_PER_SECOND", "100")
		os.Setenv("GATEWAY_CORS_ALLOWED_ORIGINS", "http://localhost:3000,https://example.com")
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
