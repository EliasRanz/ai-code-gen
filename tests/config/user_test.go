package config_test

import (
	"context"
	"os"
	"testing"

	"github.com/EliasRanz/ai-code-gen/internal/config"
	"github.com/EliasRanz/ai-code-gen/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserConfigManager(t *testing.T) {
	// Set up test environment variables
	os.Setenv("USER_SERVICE_NAME", "user-service")
	os.Setenv("USER_SERVICE_PORT", "8082")
	os.Setenv("USER_DATABASE_HOST", "localhost")
	os.Setenv("USER_DATABASE_PORT", "5432")
	os.Setenv("USER_DATABASE_USER", "postgres")
	os.Setenv("USER_DATABASE_DBNAME", "test_db")
	os.Setenv("USER_PAGINATION_DEFAULT_LIMIT", "20")
	os.Setenv("USER_PAGINATION_MAX_LIMIT", "100")
	os.Setenv("USER_VALIDATION_USERNAME_MIN_LENGTH", "3")
	defer func() {
		os.Unsetenv("USER_SERVICE_NAME")
		os.Unsetenv("USER_SERVICE_PORT")
		os.Unsetenv("USER_DATABASE_HOST")
		os.Unsetenv("USER_DATABASE_PORT")
		os.Unsetenv("USER_DATABASE_USER")
		os.Unsetenv("USER_DATABASE_DBNAME")
		os.Unsetenv("USER_PAGINATION_DEFAULT_LIMIT")
		os.Unsetenv("USER_PAGINATION_MAX_LIMIT")
		os.Unsetenv("USER_VALIDATION_USERNAME_MIN_LENGTH")
	}()

	provider := config.NewEnvironmentProvider("USER_")
	manager := user.NewUserConfigManager(provider)

	t.Run("LoadConfiguration", func(t *testing.T) {
		ctx := context.Background()
		err := manager.LoadConfig(ctx)
		require.NoError(t, err)

		config := manager.GetConfig()
		assert.NotNil(t, config)

		// Check service configuration
		assert.Equal(t, "user-service", config.Service.Name)
		assert.Equal(t, 8082, config.Service.Port)

		// Check database configuration
		assert.Equal(t, "localhost", config.Database.Host)
		assert.Equal(t, 5432, config.Database.Port)
		assert.Equal(t, "postgres", config.Database.User)
		assert.Equal(t, "test_db", config.Database.DBName)

		// Check pagination configuration
		assert.Equal(t, 20, config.Pagination.DefaultLimit)
		assert.Equal(t, 100, config.Pagination.MaxLimit)

		// Check validation configuration
		assert.Equal(t, 3, config.Validation.UsernameMinLength)
	})

	t.Run("ApplyDefaults", func(t *testing.T) {
		// Test with minimal environment variables
		os.Unsetenv("USER_SERVICE_PORT")
		os.Unsetenv("USER_DATABASE_PORT")
		os.Unsetenv("USER_PAGINATION_DEFAULT_LIMIT")

		ctx := context.Background()
		err := manager.LoadConfig(ctx)
		require.NoError(t, err)

		config := manager.GetConfig()

		// Check defaults are applied
		assert.Equal(t, "0.0.0.0", config.Service.Host)
		assert.Equal(t, 8082, config.Service.Port)          // Default port
		assert.Equal(t, 5432, config.Database.Port)         // Default DB port
		assert.Equal(t, 20, config.Pagination.DefaultLimit) // Default pagination
		assert.Equal(t, "disable", config.Database.SSLMode) // Default SSL mode
		assert.Equal(t, 25, config.Database.MaxOpenConns)   // Default max connections

		// Restore environment variables for other tests
		os.Setenv("USER_SERVICE_PORT", "8082")
		os.Setenv("USER_DATABASE_PORT", "5432")
		os.Setenv("USER_PAGINATION_DEFAULT_LIMIT", "20")
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
