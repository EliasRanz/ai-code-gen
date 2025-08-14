package config_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/EliasRanz/ai-code-gen/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestNewYamlProvider(t *testing.T) {
	t.Run("successful creation with existing file", func(t *testing.T) {
		// Create temporary YAML file
		tempFile, err := createTempYamlFile(map[string]interface{}{
			"test_key": "test_value",
			"port":     8080,
		})
		require.NoError(t, err)
		defer os.Remove(tempFile)

		provider, err := config.NewYamlProvider(tempFile)
		assert.NoError(t, err)
		assert.NotNil(t, provider)
	})

	t.Run("empty file path should return error", func(t *testing.T) {
		provider, err := config.NewYamlProvider("")
		assert.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "file path cannot be empty")
	})

	t.Run("non-existent file should return error", func(t *testing.T) {
		provider, err := config.NewYamlProvider("/non/existent/file.yaml")
		assert.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "configuration file not found")
	})
}

func TestYamlProviderLoad(t *testing.T) {
	ctx := context.Background()

	t.Run("successful load with valid YAML", func(t *testing.T) {
		testData := map[string]interface{}{
			"app_name":     "test-app",
			"port":         8080,
			"debug":        true,
			"database_url": "postgres://localhost/test",
		}

		tempFile, err := createTempYamlFile(testData)
		require.NoError(t, err)
		defer os.Remove(tempFile)

		provider, err := config.NewYamlProvider(tempFile)
		require.NoError(t, err)

		data, err := provider.Load(ctx)
		assert.NoError(t, err)
		assert.Len(t, data, 4)
		assert.Equal(t, "test-app", data["app_name"])
		assert.Equal(t, 8080, data["port"]) // YAML preserves int type
		assert.Equal(t, true, data["debug"])
		assert.Equal(t, "postgres://localhost/test", data["database_url"])
	})

	t.Run("load with nested YAML structure", func(t *testing.T) {
		testData := map[string]interface{}{
			"database": map[string]interface{}{
				"host": "localhost",
				"port": 5432,
				"ssl":  false,
			},
			"redis": map[string]interface{}{
				"url":     "redis://localhost:6379",
				"timeout": 30,
			},
		}

		tempFile, err := createTempYamlFile(testData)
		require.NoError(t, err)
		defer os.Remove(tempFile)

		provider, err := config.NewYamlProvider(tempFile)
		require.NoError(t, err)

		data, err := provider.Load(ctx)
		assert.NoError(t, err)
		assert.Len(t, data, 2)

		// Check nested structures
		dbConfig, exists := data["database"]
		assert.True(t, exists)
		dbMap, ok := dbConfig.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "localhost", dbMap["host"])
		assert.Equal(t, 5432, dbMap["port"])
		assert.Equal(t, false, dbMap["ssl"])
	})

	t.Run("load with YAML arrays", func(t *testing.T) {
		testData := map[string]interface{}{
			"allowed_origins": []string{
				"http://localhost:3000",
				"https://api.example.com",
			},
			"ports": []int{8080, 8081, 8082},
		}

		tempFile, err := createTempYamlFile(testData)
		require.NoError(t, err)
		defer os.Remove(tempFile)

		provider, err := config.NewYamlProvider(tempFile)
		require.NoError(t, err)

		data, err := provider.Load(ctx)
		assert.NoError(t, err)
		assert.Len(t, data, 2)

		// Check array values
		origins, exists := data["allowed_origins"]
		assert.True(t, exists)
		originsList, ok := origins.([]interface{})
		assert.True(t, ok)
		assert.Len(t, originsList, 2)
		assert.Equal(t, "http://localhost:3000", originsList[0])
	})

	t.Run("load with invalid YAML should return error", func(t *testing.T) {
		invalidYaml := `
invalid_yaml:
  - item1
 - item2  # incorrect indentation
`
		tempFile, err := createTempFileWithYamlContent(invalidYaml)
		require.NoError(t, err)
		defer os.Remove(tempFile)

		provider, err := config.NewYamlProvider(tempFile)
		require.NoError(t, err)

		data, err := provider.Load(ctx)
		assert.Error(t, err)
		assert.Nil(t, data)
		assert.Contains(t, err.Error(), "failed to parse YAML")
	})

	t.Run("load with empty YAML file", func(t *testing.T) {
		tempFile, err := createTempYamlFile(map[string]interface{}{})
		require.NoError(t, err)
		defer os.Remove(tempFile)

		provider, err := config.NewYamlProvider(tempFile)
		require.NoError(t, err)

		data, err := provider.Load(ctx)
		assert.NoError(t, err)
		assert.Len(t, data, 0)
	})

	t.Run("load with YAML null values", func(t *testing.T) {
		yamlContent := `
app_name: test-app
debug: true
optional_setting: null
empty_setting: ""
`
		tempFile, err := createTempFileWithYamlContent(yamlContent)
		require.NoError(t, err)
		defer os.Remove(tempFile)

		provider, err := config.NewYamlProvider(tempFile)
		require.NoError(t, err)

		data, err := provider.Load(ctx)
		assert.NoError(t, err)
		assert.Equal(t, "test-app", data["app_name"])
		assert.Equal(t, true, data["debug"])
		assert.Nil(t, data["optional_setting"])
		assert.Equal(t, "", data["empty_setting"])
	})
}

func TestYamlProviderGet(t *testing.T) {
	ctx := context.Background()

	testData := map[string]interface{}{
		"existing_key": "existing_value",
		"port":         8080,
		"debug":        true,
		"float_value":  3.14,
	}

	tempFile, err := createTempYamlFile(testData)
	require.NoError(t, err)
	defer os.Remove(tempFile)

	provider, err := config.NewYamlProvider(tempFile)
	require.NoError(t, err)

	// Load data first
	_, err = provider.Load(ctx)
	require.NoError(t, err)

	t.Run("get existing string key", func(t *testing.T) {
		value, err := provider.Get(ctx, "existing_key")
		assert.NoError(t, err)
		assert.Equal(t, "existing_value", value)
	})

	t.Run("get non-existent key should return error", func(t *testing.T) {
		value, err := provider.Get(ctx, "non_existent_key")
		assert.Error(t, err)
		assert.Nil(t, value)
		assert.Contains(t, err.Error(), "key 'non_existent_key' not found")
	})

	t.Run("get integer value", func(t *testing.T) {
		value, err := provider.Get(ctx, "port")
		assert.NoError(t, err)
		assert.Equal(t, 8080, value)
	})

	t.Run("get boolean value", func(t *testing.T) {
		value, err := provider.Get(ctx, "debug")
		assert.NoError(t, err)
		assert.Equal(t, true, value)
	})

	t.Run("get float value", func(t *testing.T) {
		value, err := provider.Get(ctx, "float_value")
		assert.NoError(t, err)
		assert.Equal(t, 3.14, value)
	})
}

func TestYamlProviderValidate(t *testing.T) {
	ctx := context.Background()

	tempFile, err := createTempYamlFile(map[string]interface{}{"dummy": "value"})
	require.NoError(t, err)
	defer os.Remove(tempFile)

	provider, err := config.NewYamlProvider(tempFile)
	require.NoError(t, err)

	t.Run("validate valid config data", func(t *testing.T) {
		validData := config.ConfigData{
			"key1": "value1",
			"key2": 123,
			"key3": true,
			"key4": 3.14,
		}

		err := provider.Validate(ctx, validData)
		assert.NoError(t, err)
	})

	t.Run("validate config data with nil value should return error", func(t *testing.T) {
		invalidData := config.ConfigData{
			"key1":    "value1",
			"nil_key": nil,
			"key3":    true,
		}

		err := provider.Validate(ctx, invalidData)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "configuration key 'nil_key' has nil value")
	})

	t.Run("validate empty config data", func(t *testing.T) {
		emptyData := config.ConfigData{}

		err := provider.Validate(ctx, emptyData)
		assert.NoError(t, err)
	})
}

func TestYamlProviderHealthCheck(t *testing.T) {
	ctx := context.Background()

	t.Run("health check with valid YAML file", func(t *testing.T) {
		tempFile, err := createTempYamlFile(map[string]interface{}{
			"test":   "value",
			"number": 42,
		})
		require.NoError(t, err)
		defer os.Remove(tempFile)

		provider, err := config.NewYamlProvider(tempFile)
		require.NoError(t, err)

		err = provider.HealthCheck(ctx)
		assert.NoError(t, err)
	})

	t.Run("health check with invalid YAML file", func(t *testing.T) {
		invalidYaml := `
invalid_yaml:
  - item1
 - item2  # incorrect indentation causes error
`
		tempFile, err := createTempFileWithYamlContent(invalidYaml)
		require.NoError(t, err)
		defer os.Remove(tempFile)

		provider, err := config.NewYamlProvider(tempFile)
		require.NoError(t, err)

		err = provider.HealthCheck(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "configuration file is not valid")
	})
}

func TestYamlProviderWatch(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t.Run("watch functionality setup", func(t *testing.T) {
		tempFile, err := createTempYamlFile(map[string]interface{}{
			"initial": "value",
		})
		require.NoError(t, err)
		defer os.Remove(tempFile)

		provider, err := config.NewYamlProvider(tempFile)
		require.NoError(t, err)

		// Test watch setup
		callbackCalled := make(chan bool, 1)
		callback := func(data config.ConfigData) {
			callbackCalled <- true
		}

		err = provider.Watch(ctx, callback)
		assert.NoError(t, err)

		// Note: Full file watching test would require file modification and timing,
		// which is complex for unit tests. This tests the setup.
	})
}

func TestYamlProviderClose(t *testing.T) {
	tempFile, err := createTempYamlFile(map[string]interface{}{
		"test": "value",
	})
	require.NoError(t, err)
	defer os.Remove(tempFile)

	provider, err := config.NewYamlProvider(tempFile)
	require.NoError(t, err)

	t.Run("close should cleanup resources", func(t *testing.T) {
		err := provider.Close()
		assert.NoError(t, err)
	})
}

// Helper functions for YAML tests

func createTempYamlFile(data map[string]interface{}) (string, error) {
	content, err := yaml.Marshal(data)
	if err != nil {
		return "", err
	}

	return createTempFileWithYamlContent(string(content))
}

func createTempFileWithYamlContent(content string) (string, error) {
	tempDir, err := os.MkdirTemp("", "config_test")
	if err != nil {
		return "", err
	}

	tempFile := filepath.Join(tempDir, "test_config.yaml")
	err = os.WriteFile(tempFile, []byte(content), 0644)
	if err != nil {
		os.RemoveAll(tempDir)
		return "", err
	}

	return tempFile, nil
}
