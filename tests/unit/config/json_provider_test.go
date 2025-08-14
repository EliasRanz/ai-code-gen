package config_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/EliasRanz/ai-code-gen/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewJsonProvider(t *testing.T) {
	t.Run("successful creation with existing file", func(t *testing.T) {
		// Create temporary JSON file
		tempFile, err := createTempJsonFile(map[string]interface{}{
			"test_key": "test_value",
			"port":     8080,
		})
		require.NoError(t, err)
		defer os.Remove(tempFile)

		provider, err := config.NewJsonProvider(tempFile)
		assert.NoError(t, err)
		assert.NotNil(t, provider)
	})

	t.Run("empty file path should return error", func(t *testing.T) {
		provider, err := config.NewJsonProvider("")
		assert.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "file path cannot be empty")
	})

	t.Run("non-existent file should return error", func(t *testing.T) {
		provider, err := config.NewJsonProvider("/non/existent/file.json")
		assert.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "configuration file not found")
	})
}

func TestJsonProviderLoad(t *testing.T) {
	ctx := context.Background()

	t.Run("successful load with valid JSON", func(t *testing.T) {
		testData := map[string]interface{}{
			"app_name":     "test-app",
			"port":         8080,
			"debug":        true,
			"database_url": "postgres://localhost/test",
		}

		tempFile, err := createTempJsonFile(testData)
		require.NoError(t, err)
		defer os.Remove(tempFile)

		provider, err := config.NewJsonProvider(tempFile)
		require.NoError(t, err)

		data, err := provider.Load(ctx)
		assert.NoError(t, err)
		assert.Len(t, data, 4)
		assert.Equal(t, "test-app", data["app_name"])
		assert.Equal(t, float64(8080), data["port"]) // JSON numbers are float64
		assert.Equal(t, true, data["debug"])
		assert.Equal(t, "postgres://localhost/test", data["database_url"])
	})

	t.Run("load with nested JSON structure", func(t *testing.T) {
		testData := map[string]interface{}{
			"database": map[string]interface{}{
				"host": "localhost",
				"port": 5432,
			},
			"redis": map[string]interface{}{
				"url": "redis://localhost:6379",
			},
		}

		tempFile, err := createTempJsonFile(testData)
		require.NoError(t, err)
		defer os.Remove(tempFile)

		provider, err := config.NewJsonProvider(tempFile)
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
		assert.Equal(t, float64(5432), dbMap["port"])
	})

	t.Run("load with invalid JSON should return error", func(t *testing.T) {
		tempFile, err := createTempFileWithContent("invalid json content {")
		require.NoError(t, err)
		defer os.Remove(tempFile)

		provider, err := config.NewJsonProvider(tempFile)
		require.NoError(t, err)

		data, err := provider.Load(ctx)
		assert.Error(t, err)
		assert.Nil(t, data)
		assert.Contains(t, err.Error(), "failed to parse JSON")
	})

	t.Run("load with empty JSON file", func(t *testing.T) {
		tempFile, err := createTempJsonFile(map[string]interface{}{})
		require.NoError(t, err)
		defer os.Remove(tempFile)

		provider, err := config.NewJsonProvider(tempFile)
		require.NoError(t, err)

		data, err := provider.Load(ctx)
		assert.NoError(t, err)
		assert.Len(t, data, 0)
	})
}

func TestJsonProviderGet(t *testing.T) {
	ctx := context.Background()

	testData := map[string]interface{}{
		"existing_key": "existing_value",
		"port":         8080,
		"debug":        true,
	}

	tempFile, err := createTempJsonFile(testData)
	require.NoError(t, err)
	defer os.Remove(tempFile)

	provider, err := config.NewJsonProvider(tempFile)
	require.NoError(t, err)

	// Load data first
	_, err = provider.Load(ctx)
	require.NoError(t, err)

	t.Run("get existing key", func(t *testing.T) {
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

	t.Run("get numeric value", func(t *testing.T) {
		value, err := provider.Get(ctx, "port")
		assert.NoError(t, err)
		assert.Equal(t, float64(8080), value)
	})

	t.Run("get boolean value", func(t *testing.T) {
		value, err := provider.Get(ctx, "debug")
		assert.NoError(t, err)
		assert.Equal(t, true, value)
	})
}

func TestJsonProviderValidate(t *testing.T) {
	ctx := context.Background()

	tempFile, err := createTempJsonFile(map[string]interface{}{"dummy": "value"})
	require.NoError(t, err)
	defer os.Remove(tempFile)

	provider, err := config.NewJsonProvider(tempFile)
	require.NoError(t, err)

	t.Run("validate valid config data", func(t *testing.T) {
		validData := config.ConfigData{
			"key1": "value1",
			"key2": 123,
			"key3": true,
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

func TestJsonProviderHealthCheck(t *testing.T) {
	ctx := context.Background()

	t.Run("health check with valid file", func(t *testing.T) {
		tempFile, err := createTempJsonFile(map[string]interface{}{
			"test": "value",
		})
		require.NoError(t, err)
		defer os.Remove(tempFile)

		provider, err := config.NewJsonProvider(tempFile)
		require.NoError(t, err)

		err = provider.HealthCheck(ctx)
		assert.NoError(t, err)
	})

	t.Run("health check with invalid JSON file", func(t *testing.T) {
		tempFile, err := createTempFileWithContent("invalid json")
		require.NoError(t, err)
		defer os.Remove(tempFile)

		provider, err := config.NewJsonProvider(tempFile)
		require.NoError(t, err)

		err = provider.HealthCheck(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "configuration file is not valid")
	})
}

func TestJsonProviderWatch(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t.Run("watch functionality setup", func(t *testing.T) {
		tempFile, err := createTempJsonFile(map[string]interface{}{
			"initial": "value",
		})
		require.NoError(t, err)
		defer os.Remove(tempFile)

		provider, err := config.NewJsonProvider(tempFile)
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

func TestJsonProviderClose(t *testing.T) {
	tempFile, err := createTempJsonFile(map[string]interface{}{
		"test": "value",
	})
	require.NoError(t, err)
	defer os.Remove(tempFile)

	provider, err := config.NewJsonProvider(tempFile)
	require.NoError(t, err)

	t.Run("close should cleanup resources", func(t *testing.T) {
		err := provider.Close()
		assert.NoError(t, err)
	})
}

// Helper functions

func createTempJsonFile(data map[string]interface{}) (string, error) {
	content, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}

	return createTempFileWithContent(string(content))
}

func createTempFileWithContent(content string) (string, error) {
	tempDir, err := os.MkdirTemp("", "config_test")
	if err != nil {
		return "", err
	}

	tempFile := filepath.Join(tempDir, "test_config.json")
	err = os.WriteFile(tempFile, []byte(content), 0644)
	if err != nil {
		os.RemoveAll(tempDir)
		return "", err
	}

	return tempFile, nil
}
