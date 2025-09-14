package config_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/EliasRanz/ai-code-gen/internal/config"
)

func TestNewEnvironmentProvider(t *testing.T) {
	provider := config.NewEnvironmentProvider("TEST_")
	assert.NotNil(t, provider)
}

func TestEnvironmentProvider_Load(t *testing.T) {
	ctx := context.Background()

	t.Run("load with prefix", func(t *testing.T) {
		// Set test environment variables
		os.Setenv("TESTAPP_DATABASE_HOST", "localhost")
		os.Setenv("TESTAPP_DATABASE_PORT", "5432")
		os.Setenv("TESTAPP_DEBUG", "true")
		defer func() {
			os.Unsetenv("TESTAPP_DATABASE_HOST")
			os.Unsetenv("TESTAPP_DATABASE_PORT")
			os.Unsetenv("TESTAPP_DEBUG")
		}()

		provider := config.NewEnvironmentProvider("TESTAPP_")
		data, err := provider.Load(ctx)

		assert.NoError(t, err)
		assert.NotNil(t, data)

		// Check transformed keys (prefix removed, lowercase, underscores to dots)
		assert.Equal(t, "localhost", data["database.host"])
		assert.Equal(t, 5432, data["database.port"]) // Should be parsed as int
		assert.Equal(t, true, data["debug"])         // Should be parsed as bool
	})

	t.Run("load without prefix", func(t *testing.T) {
		// Set test environment variables
		os.Setenv("SERVER_PORT", "8080")
		os.Setenv("DEBUG_MODE", "false")
		defer func() {
			os.Unsetenv("SERVER_PORT")
			os.Unsetenv("DEBUG_MODE")
		}()

		provider := config.NewEnvironmentProvider("")
		data, err := provider.Load(ctx)

		assert.NoError(t, err)
		assert.NotNil(t, data)

		// Should include all environment variables (including system ones)
		assert.Contains(t, data, "server.port")
		assert.Contains(t, data, "debug.mode")
		assert.Equal(t, 8080, data["server.port"])
		assert.Equal(t, false, data["debug.mode"])
	})

	t.Run("empty environment", func(t *testing.T) {
		provider := config.NewEnvironmentProvider("NONEXISTENT_")
		data, err := provider.Load(ctx)

		assert.NoError(t, err)
		assert.NotNil(t, data)
		assert.Len(t, data, 0) // Should be empty since no vars match prefix
	})
}

func TestEnvironmentProvider_Get(t *testing.T) {
	ctx := context.Background()

	// Set test environment variables
	os.Setenv("MYAPP_CONFIG_VALUE", "test123")
	defer os.Unsetenv("MYAPP_CONFIG_VALUE")

	provider := config.NewEnvironmentProvider("MYAPP_")

	// Load data first
	_, err := provider.Load(ctx)
	assert.NoError(t, err)

	t.Run("get existing key", func(t *testing.T) {
		value, err := provider.Get(ctx, "config.value")
		assert.NoError(t, err)
		assert.Equal(t, "test123", value)
	})

	t.Run("get nonexistent key", func(t *testing.T) {
		value, err := provider.Get(ctx, "nonexistent")
		assert.Error(t, err)
		assert.Nil(t, value)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestEnvironmentProvider_Watch(t *testing.T) {
	ctx := context.Background()
	provider := config.NewEnvironmentProvider("WATCH_")

	callback := func(data config.ConfigData) {
		// Callback for testing - no assertion needed
	}

	err := provider.Watch(ctx, callback)
	assert.NoError(t, err)

	// Note: Environment variable watching is not implemented in practice
	// This tests the interface compliance
}

func TestEnvironmentProvider_Validate(t *testing.T) {
	ctx := context.Background()
	provider := config.NewEnvironmentProvider("TEST_")

	t.Run("valid config data", func(t *testing.T) {
		data := config.ConfigData{
			"database": map[string]interface{}{
				"host": "localhost",
				"port": 5432,
			},
		}

		err := provider.Validate(ctx, data)
		assert.NoError(t, err)
	})
}

func TestEnvironmentProvider_HealthCheck(t *testing.T) {
	ctx := context.Background()
	provider := config.NewEnvironmentProvider("HEALTH_")

	err := provider.HealthCheck(ctx)
	assert.NoError(t, err) // Environment provider should always be healthy
}

func TestEnvironmentProvider_Close(t *testing.T) {
	provider := config.NewEnvironmentProvider("CLOSE_")

	err := provider.Close()
	assert.NoError(t, err)
}

func TestEnvironmentProvider_ValueParsing(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		envVar   string
		envValue string
		expected interface{}
	}{
		{
			name:     "string value",
			envVar:   "STRING_VAL",
			envValue: "hello world",
			expected: "hello world",
		},
		{
			name:     "integer value",
			envVar:   "INT_VAL",
			envValue: "42",
			expected: 42,
		},
		{
			name:     "boolean true",
			envVar:   "BOOL_TRUE",
			envValue: "true",
			expected: true,
		},
		{
			name:     "boolean false",
			envVar:   "BOOL_FALSE",
			envValue: "false",
			expected: false,
		},
		{
			name:     "float value",
			envVar:   "FLOAT_VAL",
			envValue: "3.14",
			expected: 3.14,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fullEnvVar := "PARSE_" + tt.envVar
			os.Setenv(fullEnvVar, tt.envValue)
			defer os.Unsetenv(fullEnvVar)

			provider := config.NewEnvironmentProvider("PARSE_")
			data, err := provider.Load(ctx)

			assert.NoError(t, err)

			// Convert env var name to config key format
			key := strings.ToLower(strings.ReplaceAll(tt.envVar, "_", "."))
			assert.Equal(t, tt.expected, data[key])
		})
	}
}

func TestEnvironmentProvider_KeyTransformation(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		envVar   string
		expected string
	}{
		{
			name:     "simple key",
			envVar:   "SIMPLE",
			expected: "simple",
		},
		{
			name:     "nested key",
			envVar:   "DATABASE_HOST",
			expected: "database.host",
		},
		{
			name:     "deeply nested key",
			envVar:   "API_AUTH_JWT_SECRET",
			expected: "api.auth.jwt.secret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fullEnvVar := "TRANSFORM_" + tt.envVar
			os.Setenv(fullEnvVar, "value")
			defer os.Unsetenv(fullEnvVar)

			provider := config.NewEnvironmentProvider("TRANSFORM_")
			data, err := provider.Load(ctx)

			assert.NoError(t, err)
			assert.Contains(t, data, tt.expected)
			assert.Equal(t, "value", data[tt.expected])
		})
	}
}
