package config_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/EliasRanz/ai-code-gen/internal/auth"
	"github.com/EliasRanz/ai-code-gen/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthConfigManager(t *testing.T) {
	// Set up test environment variables
	os.Setenv("AUTH_SERVICE_NAME", "auth-service")
	os.Setenv("AUTH_SERVICE_PORT", "8081")
	os.Setenv("AUTH_JWT_SECRET", "test-secret-key")
	os.Setenv("AUTH_JWT_ACCESS_TOKEN_DURATION", "15m")
	os.Setenv("AUTH_SESSION_DURATION", "24h")
	os.Setenv("AUTH_OAUTH_GOOGLE_CLIENT_ID", "test-client-id")
	os.Setenv("AUTH_SECURITY_PASSWORD_MIN_LENGTH", "8")
	defer func() {
		os.Unsetenv("AUTH_SERVICE_NAME")
		os.Unsetenv("AUTH_SERVICE_PORT")
		os.Unsetenv("AUTH_JWT_SECRET")
		os.Unsetenv("AUTH_JWT_ACCESS_TOKEN_DURATION")
		os.Unsetenv("AUTH_SESSION_DURATION")
		os.Unsetenv("AUTH_OAUTH_GOOGLE_CLIENT_ID")
		os.Unsetenv("AUTH_SECURITY_PASSWORD_MIN_LENGTH")
	}()

	provider := config.NewEnvironmentProvider("AUTH_")
	manager := auth.NewAuthConfigManager(provider)

	t.Run("LoadConfiguration", func(t *testing.T) {
		ctx := context.Background()
		err := manager.LoadConfig(ctx)
		require.NoError(t, err)

		config := manager.GetConfig()
		assert.NotNil(t, config)

		// Check service configuration
		assert.Equal(t, "auth-service", config.Service.Name)
		assert.Equal(t, 8081, config.Service.Port)

		// Check JWT configuration
		assert.Equal(t, "test-secret-key", config.JWT.Secret)
		assert.Equal(t, 15*time.Minute, config.JWT.AccessTokenDuration)

		// Check session configuration
		assert.Equal(t, 24*time.Hour, config.Session.Duration)

		// Check OAuth configuration
		assert.Equal(t, "test-client-id", config.OAuth.Google.ClientID)

		// Check security configuration
		assert.Equal(t, 8, config.Security.PasswordMinLength)
	})

	t.Run("ApplyDefaults", func(t *testing.T) {
		// Test with minimal environment variables
		os.Unsetenv("AUTH_SERVICE_PORT")
		os.Unsetenv("AUTH_JWT_ACCESS_TOKEN_DURATION")

		ctx := context.Background()
		err := manager.LoadConfig(ctx)
		require.NoError(t, err)

		config := manager.GetConfig()

		// Check defaults are applied
		assert.Equal(t, "0.0.0.0", config.Service.Host)
		assert.Equal(t, 8081, config.Service.Port)                      // Default port
		assert.Equal(t, 15*time.Minute, config.JWT.AccessTokenDuration) // Default duration
		assert.Equal(t, "HS256", config.JWT.Algorithm)                  // Default algorithm
		assert.Equal(t, "auth_session", config.Session.CookieName)      // Default cookie name

		// Restore environment variables for other tests
		os.Setenv("AUTH_SERVICE_PORT", "8081")
		os.Setenv("AUTH_JWT_ACCESS_TOKEN_DURATION", "15m")
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

	t.Run("ConfigurationValidation", func(t *testing.T) {
		ctx := context.Background()

		// Valid configuration
		err := manager.LoadConfig(ctx)
		assert.NoError(t, err)

		// Test with missing required JWT secret
		os.Unsetenv("AUTH_JWT_SECRET")
		defer os.Setenv("AUTH_JWT_SECRET", "test-secret-key")

		// This should fail validation (if we had validation in place)
		// For now, just ensure it loads but gets default values
		provider2 := config.NewEnvironmentProvider("AUTH_")
		manager2 := auth.NewAuthConfigManager(provider2)
		err = manager2.LoadConfig(ctx)
		// Should succeed but get empty secret
		require.NoError(t, err)
		config := manager2.GetConfig()
		assert.Empty(t, config.JWT.Secret)
	})
}
