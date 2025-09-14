package config_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/EliasRanz/ai-code-gen/internal/config"
)

// MockProvider for testing
type MockProvider struct {
	mock.Mock
}

func (m *MockProvider) Load(ctx context.Context) (config.ConfigData, error) {
	args := m.Called(ctx)
	return args.Get(0).(config.ConfigData), args.Error(1)
}

func (m *MockProvider) Watch(ctx context.Context, callback func(config.ConfigData)) error {
	args := m.Called(ctx, callback)
	return args.Error(0)
}

func (m *MockProvider) Get(ctx context.Context, key string) (interface{}, error) {
	args := m.Called(ctx, key)
	return args.Get(0), args.Error(1)
}

func (m *MockProvider) Validate(ctx context.Context, data config.ConfigData) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

func (m *MockProvider) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockProvider) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestNewConfigFactory(t *testing.T) {
	factory := config.NewConfigFactory()
	assert.NotNil(t, factory)

	// Check that default providers are registered
	providers := factory.ListAvailableProviders()
	assert.Contains(t, providers, "env")
	assert.Contains(t, providers, "json")
	assert.Contains(t, providers, "yaml")
}

func TestStandardConfigFactory_CreateProvider(t *testing.T) {
	factory := config.NewConfigFactory()

	t.Run("successful environment provider creation", func(t *testing.T) {
		provider, err := factory.CreateProvider("env", "TEST_")
		assert.NoError(t, err)
		assert.NotNil(t, provider)
	})

	t.Run("json provider with nonexistent file", func(t *testing.T) {
		// JSON provider should return error for nonexistent file
		provider, err := factory.CreateProvider("json", "nonexistent.json")
		assert.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "configuration file not found")
	})

	t.Run("yaml provider with nonexistent file", func(t *testing.T) {
		// YAML provider should return error for nonexistent file
		provider, err := factory.CreateProvider("yaml", "nonexistent.yaml")
		assert.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "configuration file not found")
	})

	t.Run("json provider with existing file", func(t *testing.T) {
		// Create temporary JSON file
		tmpFile, err := os.CreateTemp("", "test-config-*.json")
		assert.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		_, err = tmpFile.WriteString(`{"test": "value"}`)
		assert.NoError(t, err)
		tmpFile.Close()

		provider, err := factory.CreateProvider("json", tmpFile.Name())
		assert.NoError(t, err)
		assert.NotNil(t, provider)
	})

	t.Run("yaml provider with existing file", func(t *testing.T) {
		// Create temporary YAML file
		tmpFile, err := os.CreateTemp("", "test-config-*.yaml")
		assert.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		_, err = tmpFile.WriteString("test: value\n")
		assert.NoError(t, err)
		tmpFile.Close()

		provider, err := factory.CreateProvider("yaml", tmpFile.Name())
		assert.NoError(t, err)
		assert.NotNil(t, provider)
	})

	t.Run("unknown provider type", func(t *testing.T) {
		provider, err := factory.CreateProvider("unknown", "source")
		assert.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "unknown provider type")
	})

	t.Run("empty source", func(t *testing.T) {
		provider, err := factory.CreateProvider("env", "")
		assert.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "source cannot be empty")
	})
}

func TestStandardConfigFactory_ListAvailableProviders(t *testing.T) {
	factory := config.NewConfigFactory()
	providers := factory.ListAvailableProviders()

	// Should have at least the default providers
	assert.GreaterOrEqual(t, len(providers), 3)
	assert.Contains(t, providers, "env")
	assert.Contains(t, providers, "json")
	assert.Contains(t, providers, "yaml")
}

func TestStandardConfigFactory_RegisterProvider(t *testing.T) {
	factory := config.NewConfigFactory()

	t.Run("successful registration", func(t *testing.T) {
		mockFactory := func(source string) (config.ConfigProvider, error) {
			return &MockProvider{}, nil
		}

		err := factory.RegisterProvider("mock", mockFactory)
		assert.NoError(t, err)

		// Verify it was registered
		providers := factory.ListAvailableProviders()
		assert.Contains(t, providers, "mock")

		// Verify it can create the provider
		provider, err := factory.CreateProvider("mock", "test-source")
		assert.NoError(t, err)
		assert.NotNil(t, provider)
	})

	t.Run("empty provider type", func(t *testing.T) {
		mockFactory := func(source string) (config.ConfigProvider, error) {
			return &MockProvider{}, nil
		}

		err := factory.RegisterProvider("", mockFactory)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "provider type cannot be empty")
	})

	t.Run("nil factory", func(t *testing.T) {
		err := factory.RegisterProvider("nil-factory", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "provider factory cannot be nil")
	})

	t.Run("override existing provider", func(t *testing.T) {
		// Should be able to override existing providers
		mockFactory := func(source string) (config.ConfigProvider, error) {
			return &MockProvider{}, nil
		}

		err := factory.RegisterProvider("env", mockFactory)
		assert.NoError(t, err)

		// Should still be able to create provider (now using mock implementation)
		provider, err := factory.CreateProvider("env", "test")
		assert.NoError(t, err)
		assert.NotNil(t, provider)
	})
}

func TestStandardConfigFactory_ConcurrentAccess(t *testing.T) {
	factory := config.NewConfigFactory()

	// Test concurrent access to factory methods
	done := make(chan bool)

	// Concurrent provider creation
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()
			_, _ = factory.CreateProvider("env", "TEST_")
		}()
	}

	// Concurrent provider registration
	for i := 0; i < 5; i++ {
		go func(i int) {
			defer func() { done <- true }()
			mockFactory := func(source string) (config.ConfigProvider, error) {
				return &MockProvider{}, nil
			}
			_ = factory.RegisterProvider(fmt.Sprintf("mock%d", i), mockFactory)
		}(i)
	}

	// Concurrent provider listing
	for i := 0; i < 5; i++ {
		go func() {
			defer func() { done <- true }()
			_ = factory.ListAvailableProviders()
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 20; i++ {
		<-done
	}

	// Verify factory is still functional
	providers := factory.ListAvailableProviders()
	assert.GreaterOrEqual(t, len(providers), 3)
}

func TestStandardConfigFactory_ErrorHandling(t *testing.T) {
	factory := config.NewConfigFactory()

	t.Run("provider factory returns error", func(t *testing.T) {
		failingFactory := func(source string) (config.ConfigProvider, error) {
			return nil, fmt.Errorf("factory creation failed")
		}

		err := factory.RegisterProvider("failing", failingFactory)
		assert.NoError(t, err)

		provider, err := factory.CreateProvider("failing", "source")
		assert.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "failed to create provider")
	})
}
