package config_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/EliasRanz/ai-code-gen/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigFactory(t *testing.T) {
	factory := config.NewConfigFactory()

	t.Run("ListAvailableProviders", func(t *testing.T) {
		providers := factory.ListAvailableProviders()
		assert.Contains(t, providers, "env")
		assert.Contains(t, providers, "yaml")
		assert.Contains(t, providers, "json")
	})

	t.Run("CreateEnvironmentProvider", func(t *testing.T) {
		provider, err := factory.CreateProvider("env", "TEST_")
		require.NoError(t, err)
		assert.NotNil(t, provider)

		// Cleanup
		err = provider.Close()
		assert.NoError(t, err)
	})

	t.Run("CreateProviderWithEmptySource", func(t *testing.T) {
		_, err := factory.CreateProvider("env", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "source cannot be empty")
	})

	t.Run("CreateProviderWithUnknownType", func(t *testing.T) {
		_, err := factory.CreateProvider("unknown", "test")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown provider type")
	})

	t.Run("RegisterCustomProvider", func(t *testing.T) {
		customFactory := func(source string) (config.ConfigProvider, error) {
			return config.NewEnvironmentProvider(source), nil
		}

		err := factory.RegisterProvider("custom", customFactory)
		assert.NoError(t, err)

		providers := factory.ListAvailableProviders()
		assert.Contains(t, providers, "custom")

		provider, err := factory.CreateProvider("custom", "TEST_")
		require.NoError(t, err)
		assert.NotNil(t, provider)

		err = provider.Close()
		assert.NoError(t, err)
	})

	t.Run("RegisterProviderWithEmptyType", func(t *testing.T) {
		customFactory := func(source string) (config.ConfigProvider, error) {
			return config.NewEnvironmentProvider(source), nil
		}

		err := factory.RegisterProvider("", customFactory)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "provider type cannot be empty")
	})

	t.Run("RegisterProviderWithNilFactory", func(t *testing.T) {
		err := factory.RegisterProvider("nil", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "provider factory cannot be nil")
	})
}

func TestEnvironmentProvider(t *testing.T) {
	// Set up test environment variables
	os.Setenv("TEST_SERVICE_NAME", "test-service")
	os.Setenv("TEST_SERVICE_PORT", "8080")
	os.Setenv("TEST_FEATURE_ENABLED", "true")
	os.Setenv("TEST_TIMEOUT", "30s")
	os.Setenv("TEST_TAGS", "tag1,tag2,tag3")
	defer func() {
		os.Unsetenv("TEST_SERVICE_NAME")
		os.Unsetenv("TEST_SERVICE_PORT")
		os.Unsetenv("TEST_FEATURE_ENABLED")
		os.Unsetenv("TEST_TIMEOUT")
		os.Unsetenv("TEST_TAGS")
	}()

	provider := config.NewEnvironmentProvider("TEST_")
	defer provider.Close()

	t.Run("LoadConfiguration", func(t *testing.T) {
		ctx := context.Background()
		data, err := provider.Load(ctx)
		require.NoError(t, err)

		assert.Equal(t, "test-service", data["service.name"])
		assert.Equal(t, 8080, data["service.port"])
		assert.Equal(t, true, data["feature.enabled"])
		assert.Equal(t, "30s", data["timeout"])
		assert.Equal(t, []string{"tag1", "tag2", "tag3"}, data["tags"])
	})

	t.Run("GetSpecificValue", func(t *testing.T) {
		ctx := context.Background()

		// Load first
		_, err := provider.Load(ctx)
		require.NoError(t, err)

		value, err := provider.Get(ctx, "service.name")
		require.NoError(t, err)
		assert.Equal(t, "test-service", value)

		value, err = provider.Get(ctx, "service.port")
		require.NoError(t, err)
		assert.Equal(t, 8080, value)
	})

	t.Run("GetNonExistentValue", func(t *testing.T) {
		ctx := context.Background()

		_, err := provider.Get(ctx, "nonexistent.key")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("HealthCheck", func(t *testing.T) {
		ctx := context.Background()
		err := provider.HealthCheck(ctx)
		assert.NoError(t, err)
	})

	t.Run("Validate", func(t *testing.T) {
		ctx := context.Background()
		data := config.ConfigData{
			"test.key": "test.value",
			"another":  42,
		}

		err := provider.Validate(ctx, data)
		assert.NoError(t, err)
	})

	t.Run("ValidateWithNilValue", func(t *testing.T) {
		ctx := context.Background()
		data := config.ConfigData{
			"test.key": nil,
		}

		err := provider.Validate(ctx, data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "nil value")
	})
}

func TestConfigManager(t *testing.T) {
	// Set up test environment variables
	os.Setenv("TEST_SERVICE_NAME", "test-service")
	os.Setenv("TEST_SERVICE_PORT", "8080")
	os.Setenv("TEST_FEATURE_ENABLED", "true")
	os.Setenv("TEST_TIMEOUT", "30s")
	defer func() {
		os.Unsetenv("TEST_SERVICE_NAME")
		os.Unsetenv("TEST_SERVICE_PORT")
		os.Unsetenv("TEST_FEATURE_ENABLED")
		os.Unsetenv("TEST_TIMEOUT")
	}()

	provider := config.NewEnvironmentProvider("TEST_")
	manager := config.NewConfigManager(provider)

	t.Run("LoadAndGetValues", func(t *testing.T) {
		ctx := context.Background()
		err := manager.LoadConfig(ctx)
		require.NoError(t, err)

		assert.Equal(t, "test-service", manager.GetString("service.name"))
		assert.Equal(t, 8080, manager.GetInt("service.port"))
		assert.Equal(t, true, manager.GetBool("feature.enabled"))
		assert.Equal(t, 30*time.Second, manager.GetDuration("timeout"))
	})

	t.Run("GetDefaultValues", func(t *testing.T) {
		ctx := context.Background()
		err := manager.LoadConfig(ctx)
		require.NoError(t, err)

		// These keys don't exist in environment
		assert.Equal(t, "", manager.GetString("nonexistent.string"))
		assert.Equal(t, 0, manager.GetInt("nonexistent.int"))
		assert.Equal(t, false, manager.GetBool("nonexistent.bool"))
		assert.Equal(t, time.Duration(0), manager.GetDuration("nonexistent.duration"))
		assert.Equal(t, []string{}, manager.GetStringSlice("nonexistent.slice"))
	})

	t.Run("HasKey", func(t *testing.T) {
		ctx := context.Background()
		err := manager.LoadConfig(ctx)
		require.NoError(t, err)

		assert.True(t, manager.HasKey("service.name"))
		assert.False(t, manager.HasKey("nonexistent.key"))
	})

	t.Run("GetRawData", func(t *testing.T) {
		ctx := context.Background()
		err := manager.LoadConfig(ctx)
		require.NoError(t, err)

		raw := manager.GetRaw()
		assert.NotEmpty(t, raw)
		assert.Contains(t, raw, "service.name")
		assert.Equal(t, "test-service", raw["service.name"])
	})

	t.Run("Validate", func(t *testing.T) {
		ctx := context.Background()
		err := manager.LoadConfig(ctx)
		require.NoError(t, err)

		err = manager.Validate()
		assert.NoError(t, err)
	})

	t.Run("Reload", func(t *testing.T) {
		ctx := context.Background()

		// Initial load
		err := manager.LoadConfig(ctx)
		require.NoError(t, err)

		// Reload
		err = manager.Reload(ctx)
		assert.NoError(t, err)
	})
}

func TestConfigValidator(t *testing.T) {
	validator := config.NewConfigValidator()

	t.Run("AddValidationRules", func(t *testing.T) {
		rule := config.ValidationRule{
			Key:      "test.key",
			Required: true,
			Type:     "string",
		}

		err := validator.AddRule(rule)
		assert.NoError(t, err)

		rules := validator.GetRules()
		assert.Len(t, rules, 1)
		assert.Equal(t, "test.key", rules[0].Key)
	})

	t.Run("AddRuleWithEmptyKey", func(t *testing.T) {
		rule := config.ValidationRule{
			Key:      "",
			Required: true,
		}

		err := validator.AddRule(rule)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "key cannot be empty")
	})

	t.Run("AddRuleWithInvalidType", func(t *testing.T) {
		rule := config.ValidationRule{
			Key:  "test.key",
			Type: "invalid",
		}

		err := validator.AddRule(rule)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid validation type")
	})

	t.Run("ValidateRequiredField", func(t *testing.T) {
		validator := config.NewConfigValidator()
		rule := config.ValidationRule{
			Key:      "required.field",
			Required: true,
		}
		validator.AddRule(rule)

		// Missing required field
		data := config.ConfigData{}
		err := validator.Validate(data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "required field")

		// Present required field
		data["required.field"] = "value"
		err = validator.Validate(data)
		assert.NoError(t, err)
	})

	t.Run("ValidateTypeConstraints", func(t *testing.T) {
		validator := config.NewConfigValidator()

		// String type validation
		validator.AddRule(config.ValidationRule{
			Key:  "string.field",
			Type: "string",
		})

		// Int type validation
		validator.AddRule(config.ValidationRule{
			Key:  "int.field",
			Type: "int",
		})

		// Bool type validation
		validator.AddRule(config.ValidationRule{
			Key:  "bool.field",
			Type: "bool",
		})

		// Valid data
		data := config.ConfigData{
			"string.field": "test",
			"int.field":    42,
			"bool.field":   true,
		}
		err := validator.Validate(data)
		assert.NoError(t, err)

		// Invalid data
		data["string.field"] = 123
		err = validator.Validate(data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be a string")
	})

	t.Run("ValidateRangeConstraints", func(t *testing.T) {
		validator := config.NewConfigValidator()
		validator.AddRule(config.ValidationRule{
			Key:      "port",
			Type:     "int",
			MinValue: 1,
			MaxValue: 65535,
		})

		// Valid port
		data := config.ConfigData{"port": 8080}
		err := validator.Validate(data)
		assert.NoError(t, err)

		// Port too low
		data["port"] = 0
		err = validator.Validate(data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "below minimum")

		// Port too high
		data["port"] = 70000
		err = validator.Validate(data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "above maximum")
	})

	t.Run("ValidatePatternConstraints", func(t *testing.T) {
		validator := config.NewConfigValidator()
		validator.AddRule(config.ValidationRule{
			Key:     "email",
			Type:    "string",
			Pattern: `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
		})

		// Valid email
		data := config.ConfigData{"email": "test@example.com"}
		err := validator.Validate(data)
		assert.NoError(t, err)

		// Invalid email
		data["email"] = "invalid-email"
		err = validator.Validate(data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "does not match pattern")
	})
}
