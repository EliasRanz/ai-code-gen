package config

import (
	"context"
	"time"
)

// ConfigProvider defines the core interface that all configuration providers must implement
// This implements the Configuration Interface Pattern for source-agnostic configuration loading
type ConfigProvider interface {
	// Load configuration from source
	Load(ctx context.Context) (ConfigData, error)

	// Watch for configuration changes
	Watch(ctx context.Context, callback func(ConfigData)) error

	// Get specific configuration value
	Get(ctx context.Context, key string) (interface{}, error)

	// Validate configuration structure
	Validate(ctx context.Context, data ConfigData) error

	// Health check for configuration source
	HealthCheck(ctx context.Context) error

	// Close provider and cleanup resources
	Close() error
}

// ConfigFactory implements the Factory Pattern for configuration provider instantiation
type ConfigFactory interface {
	CreateProvider(providerType string, source string) (ConfigProvider, error)
	ListAvailableProviders() []string
	RegisterProvider(providerType string, factory ProviderFactory) error
}

// ProviderFactory creates specific provider instances
type ProviderFactory func(source string) (ConfigProvider, error)

// ConfigData represents configuration data in a provider-agnostic format
type ConfigData map[string]interface{}

// ConfigManager manages configuration for a specific service
type ConfigManager interface {
	// Load service-specific configuration
	LoadConfig(ctx context.Context) error

	// Get configuration value with type conversion
	GetString(key string) string
	GetInt(key string) int
	GetFloat64(key string) float64
	GetBool(key string) bool
	GetDuration(key string) time.Duration
	GetStringSlice(key string) []string

	// Check if configuration key exists
	HasKey(key string) bool

	// Validate current configuration
	Validate() error

	// Watch for configuration changes
	Watch(ctx context.Context, callback func()) error

	// Get raw configuration data
	GetRaw() ConfigData

	// Reload configuration
	Reload(ctx context.Context) error
}

// BaseServiceConfig represents base configuration for any service
type BaseServiceConfig struct {
	Name        string            `json:"name" yaml:"name"`
	Host        string            `json:"host" yaml:"host"`
	Port        int               `json:"port" yaml:"port"`
	Environment string            `json:"environment" yaml:"environment"`
	Version     string            `json:"version" yaml:"version"`
	Metadata    map[string]string `json:"metadata" yaml:"metadata"`
}

// ValidationRule defines configuration validation rules
type ValidationRule struct {
	Key       string
	Required  bool
	Type      string
	MinValue  interface{}
	MaxValue  interface{}
	Pattern   string
	Validator func(interface{}) error
}

// ConfigValidator validates configuration against rules
type ConfigValidator interface {
	AddRule(rule ValidationRule) error
	Validate(data ConfigData) error
	GetRules() []ValidationRule
}

// ConfigWatcher handles configuration change notifications
type ConfigWatcher interface {
	Start(ctx context.Context) error
	Stop() error
	AddCallback(callback func(ConfigData)) error
	RemoveCallback(callbackID string) error
}

// HotReloadConfig enables hot reloading capabilities
type HotReloadConfig struct {
	Enabled       bool          `json:"enabled" yaml:"enabled"`
	CheckInterval time.Duration `json:"check_interval" yaml:"check_interval"`
	WatchFiles    []string      `json:"watch_files" yaml:"watch_files"`
}
