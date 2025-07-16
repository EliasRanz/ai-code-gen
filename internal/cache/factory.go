package cache

import (
	"fmt"
	"strings"
)

// providerType constants for supported cache providers
const (
	ProviderTypeRedis  = "redis"
	ProviderTypeMemory = "memory"
	ProviderTypeMulti  = "multi"
)

// cacheFactory implements the Factory Pattern for cache provider creation
type cacheFactory struct {
	providers map[string]func(CacheConfig) (CacheProvider, error)
}

// NewCacheFactory creates a new cache factory with all supported providers
func NewCacheFactory() CacheFactory {
	factory := &cacheFactory{
		providers: make(map[string]func(CacheConfig) (CacheProvider, error)),
	}

	// Register built-in providers
	factory.RegisterProvider(ProviderTypeRedis, createRedisProvider)
	factory.RegisterProvider(ProviderTypeMemory, createMemoryProvider)
	factory.RegisterProvider(ProviderTypeMulti, createMultiProvider)

	return factory
}

// CreateProvider creates a cache provider instance based on type and configuration
func (f *cacheFactory) CreateProvider(providerType string, config CacheConfig) (CacheProvider, error) {
	providerType = strings.ToLower(strings.TrimSpace(providerType))

	if providerType == "" {
		return nil, fmt.Errorf("provider type cannot be empty")
	}

	createFunc, exists := f.providers[providerType]
	if !exists {
		return nil, fmt.Errorf("unsupported cache provider type: %s", providerType)
	}

	provider, err := createFunc(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create %s provider: %w", providerType, err)
	}

	return provider, nil
}

// ListAvailableProviders returns a list of all registered provider types
func (f *cacheFactory) ListAvailableProviders() []string {
	providers := make([]string, 0, len(f.providers))
	for providerType := range f.providers {
		providers = append(providers, providerType)
	}
	return providers
}

// RegisterProvider allows registration of custom cache providers
func (f *cacheFactory) RegisterProvider(providerType string, createFunc func(CacheConfig) (CacheProvider, error)) {
	if createFunc == nil {
		return
	}

	providerType = strings.ToLower(strings.TrimSpace(providerType))
	if providerType != "" {
		f.providers[providerType] = createFunc
	}
}

// Provider creation functions

// createRedisProvider creates a Redis cache provider with circuit breaker
func createRedisProvider(config CacheConfig) (CacheProvider, error) {
	if config.Host == "" {
		return nil, fmt.Errorf("Redis host is required")
	}

	if config.Port <= 0 {
		config.Port = 6379 // Default Redis port
	}

	// Apply defaults for circuit breaker if not set
	if config.FailureThreshold <= 0 {
		config.FailureThreshold = 5
	}
	if config.RecoveryTimeout <= 0 {
		config.RecoveryTimeout = 30000000000 // 30 seconds in nanoseconds
	}

	return NewRedisProvider(config)
}

// createMemoryProvider creates an in-memory cache provider (fallback)
func createMemoryProvider(config CacheConfig) (CacheProvider, error) {
	return NewMemoryProvider(config)
}

// createMultiProvider creates a multi-tier cache provider (Redis + Memory)
func createMultiProvider(config CacheConfig) (CacheProvider, error) {
	return NewMultiProvider(config)
}

// DefaultCacheConfig returns a sensible default cache configuration
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		Host:                   "localhost",
		Port:                   6379,
		Password:               "",
		DB:                     0,
		MaxConnections:         100,
		MaxIdleConnections:     10,
		ConnectionTimeout:      5000000000,   // 5 seconds in nanoseconds
		IdleTimeout:            300000000000, // 5 minutes in nanoseconds
		FailureThreshold:       5,
		RequestVolumeThreshold: 10,
		RecoveryTimeout:        30000000000,  // 30 seconds in nanoseconds
		DefaultTTL:             300000000000, // 5 minutes in nanoseconds
	}
}

// ValidateCacheConfig validates a cache configuration
func ValidateCacheConfig(config CacheConfig) error {
	if config.Host == "" {
		return fmt.Errorf("host is required")
	}

	if config.Port <= 0 || config.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}

	if config.MaxConnections <= 0 {
		return fmt.Errorf("max connections must be positive")
	}

	if config.MaxIdleConnections < 0 {
		return fmt.Errorf("max idle connections cannot be negative")
	}

	if config.MaxIdleConnections > config.MaxConnections {
		return fmt.Errorf("max idle connections cannot exceed max connections")
	}

	if config.FailureThreshold <= 0 {
		return fmt.Errorf("failure threshold must be positive")
	}

	if config.RequestVolumeThreshold <= 0 {
		return fmt.Errorf("request volume threshold must be positive")
	}

	if config.RecoveryTimeout <= 0 {
		return fmt.Errorf("recovery timeout must be positive")
	}

	return nil
}
