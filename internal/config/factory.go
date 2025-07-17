package config

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// StandardConfigFactory implements ConfigFactory interface
type StandardConfigFactory struct {
	providers map[string]ProviderFactory
	mu        sync.RWMutex
}

// NewConfigFactory creates a new configuration factory
func NewConfigFactory() ConfigFactory {
	factory := &StandardConfigFactory{
		providers: make(map[string]ProviderFactory),
	}

	// Register default providers
	factory.registerDefaultProviders()

	return factory
}

// CreateProvider creates a configuration provider instance
func (f *StandardConfigFactory) CreateProvider(providerType string, source string) (ConfigProvider, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	factory, exists := f.providers[providerType]
	if !exists {
		return nil, fmt.Errorf("unknown provider type: %s", providerType)
	}

	if source == "" {
		return nil, fmt.Errorf("source cannot be empty for provider type: %s", providerType)
	}

	provider, err := factory(source)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider %s: %w", providerType, err)
	}

	return provider, nil
}

// ListAvailableProviders returns list of available provider types
func (f *StandardConfigFactory) ListAvailableProviders() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	var providers []string
	for providerType := range f.providers {
		providers = append(providers, providerType)
	}

	return providers
}

// RegisterProvider registers a new provider factory
func (f *StandardConfigFactory) RegisterProvider(providerType string, factory ProviderFactory) error {
	if providerType == "" {
		return fmt.Errorf("provider type cannot be empty")
	}

	if factory == nil {
		return fmt.Errorf("provider factory cannot be nil")
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	f.providers[providerType] = factory
	return nil
}

// registerDefaultProviders registers built-in provider factories
func (f *StandardConfigFactory) registerDefaultProviders() {
	// Environment variable provider
	f.providers["env"] = func(source string) (ConfigProvider, error) {
		return NewEnvironmentProvider(source), nil
	}

	// YAML file provider
	f.providers["yaml"] = func(source string) (ConfigProvider, error) {
		return NewYamlProvider(source)
	}

	// JSON file provider
	f.providers["json"] = func(source string) (ConfigProvider, error) {
		return NewJsonProvider(source)
	}
}

// StandardConfigManager implements ConfigManager interface
type StandardConfigManager struct {
	provider  ConfigProvider
	data      ConfigData
	validator ConfigValidator
	watcher   ConfigWatcher
	mu        sync.RWMutex
	callbacks []func()
}

// NewConfigManager creates a new configuration manager
func NewConfigManager(provider ConfigProvider) ConfigManager {
	return &StandardConfigManager{
		provider:  provider,
		data:      make(ConfigData),
		validator: NewConfigValidator(),
		callbacks: make([]func(), 0),
	}
}

// LoadConfig loads configuration from provider
func (m *StandardConfigManager) LoadConfig(ctx context.Context) error {
	data, err := m.provider.Load(ctx)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	if err := m.validator.Validate(data); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	m.mu.Lock()
	m.data = data
	m.mu.Unlock()

	// Notify callbacks of configuration change
	for _, callback := range m.callbacks {
		go callback()
	}

	return nil
}

// GetString gets string configuration value
func (m *StandardConfigManager) GetString(key string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if val, exists := m.data[key]; exists {
		if str, ok := val.(string); ok {
			return str
		}
	}

	return ""
}

// GetInt gets integer configuration value
func (m *StandardConfigManager) GetInt(key string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if val, exists := m.data[key]; exists {
		if i, ok := val.(int); ok {
			return i
		}
		if f, ok := val.(float64); ok {
			return int(f)
		}
	}

	return 0
}

// GetFloat64 gets float64 configuration value
func (m *StandardConfigManager) GetFloat64(key string) float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if val, exists := m.data[key]; exists {
		if f, ok := val.(float64); ok {
			return f
		}
		if i, ok := val.(int); ok {
			return float64(i)
		}
	}

	return 0.0
}

// GetBool gets boolean configuration value
func (m *StandardConfigManager) GetBool(key string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if val, exists := m.data[key]; exists {
		if b, ok := val.(bool); ok {
			return b
		}
	}

	return false
}

// GetDuration gets duration configuration value
func (m *StandardConfigManager) GetDuration(key string) time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if val, exists := m.data[key]; exists {
		if str, ok := val.(string); ok {
			if d, err := time.ParseDuration(str); err == nil {
				return d
			}
		}
	}

	return 0
}

// GetStringSlice gets string slice configuration value
func (m *StandardConfigManager) GetStringSlice(key string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if val, exists := m.data[key]; exists {
		// Handle []string type (from environment provider)
		if slice, ok := val.([]string); ok {
			return slice
		}
		// Handle []interface{} type (from YAML/JSON providers)
		if slice, ok := val.([]interface{}); ok {
			result := make([]string, len(slice))
			for i, item := range slice {
				if str, ok := item.(string); ok {
					result[i] = str
				}
			}
			return result
		}
	}

	return []string{}
}

// HasKey checks if configuration key exists
func (m *StandardConfigManager) HasKey(key string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, exists := m.data[key]
	return exists
}

// Validate validates current configuration
func (m *StandardConfigManager) Validate() error {
	m.mu.RLock()
	data := m.data
	m.mu.RUnlock()

	return m.validator.Validate(data)
}

// Watch watches for configuration changes
func (m *StandardConfigManager) Watch(ctx context.Context, callback func()) error {
	m.callbacks = append(m.callbacks, callback)

	return m.provider.Watch(ctx, func(data ConfigData) {
		m.mu.Lock()
		m.data = data
		m.mu.Unlock()

		callback()
	})
}

// GetRaw returns raw configuration data
func (m *StandardConfigManager) GetRaw() ConfigData {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(ConfigData)
	for k, v := range m.data {
		result[k] = v
	}

	return result
}

// Reload reloads configuration from provider
func (m *StandardConfigManager) Reload(ctx context.Context) error {
	return m.LoadConfig(ctx)
}
