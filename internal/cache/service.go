package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/EliasRanz/ai-code-gen/internal/config"
)

// Service provides a unified cache service for the entire application
type Service struct {
	factory  CacheFactory
	provider CacheProvider
	config   ServiceConfig
}

// ServiceConfig holds cache service configuration
type ServiceConfig struct {
	ProviderType           string        `json:"provider_type"`
	Host                   string        `json:"host"`
	Port                   int           `json:"port"`
	Password               string        `json:"password"`
	DB                     int           `json:"db"`
	MaxConnections         int           `json:"max_connections"`
	MaxIdleConnections     int           `json:"max_idle_connections"`
	ConnectionTimeout      time.Duration `json:"connection_timeout"`
	FailureThreshold       int           `json:"failure_threshold"`
	RequestVolumeThreshold int           `json:"request_volume_threshold"`
	RecoveryTimeout        time.Duration `json:"recovery_timeout"`
	DefaultTTL             time.Duration `json:"default_ttl"`
}

// NewService creates a new cache service with the given configuration
func NewService(cfg ServiceConfig) (*Service, error) {
	factory := NewCacheFactory()

	// Convert service config to cache config
	cacheConfig := CacheConfig{
		Host:                   cfg.Host,
		Port:                   cfg.Port,
		Password:               cfg.Password,
		DB:                     cfg.DB,
		MaxConnections:         cfg.MaxConnections,
		MaxIdleConnections:     cfg.MaxIdleConnections,
		ConnectionTimeout:      cfg.ConnectionTimeout,
		FailureThreshold:       cfg.FailureThreshold,
		RequestVolumeThreshold: cfg.RequestVolumeThreshold,
		RecoveryTimeout:        cfg.RecoveryTimeout,
		DefaultTTL:             cfg.DefaultTTL,
	}

	// Create cache provider
	provider, err := factory.CreateProvider(cfg.ProviderType, cacheConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create cache provider: %w", err)
	}

	return &Service{
		factory:  factory,
		provider: provider,
		config:   cfg,
	}, nil
}

// NewServiceFromConfig creates a new cache service from the application config
func NewServiceFromConfig(cfg *config.Config) (*Service, error) {
	serviceConfig := ServiceConfig{
		ProviderType:           "memory", // Use memory for testing/development
		Host:                   cfg.Redis.Host,
		Port:                   cfg.Redis.Port,
		Password:               cfg.Redis.Password,
		DB:                     cfg.Redis.DB,
		MaxConnections:         cfg.Redis.MaxConnections,
		MaxIdleConnections:     cfg.Redis.MaxIdleConnections,
		ConnectionTimeout:      cfg.Redis.ConnectionTimeout,
		FailureThreshold:       cfg.Redis.FailureThreshold,
		RequestVolumeThreshold: cfg.Redis.RequestVolumeThreshold,
		RecoveryTimeout:        cfg.Redis.RecoveryTimeout,
		DefaultTTL:             5 * time.Minute, // Default TTL
	}

	// Apply defaults if not set
	if serviceConfig.MaxConnections <= 0 {
		serviceConfig.MaxConnections = 100
	}
	if serviceConfig.MaxIdleConnections <= 0 {
		serviceConfig.MaxIdleConnections = 10
	}
	if serviceConfig.ConnectionTimeout <= 0 {
		serviceConfig.ConnectionTimeout = 5 * time.Second
	}
	if serviceConfig.FailureThreshold <= 0 {
		serviceConfig.FailureThreshold = 5
	}
	if serviceConfig.RequestVolumeThreshold <= 0 {
		serviceConfig.RequestVolumeThreshold = 10
	}
	if serviceConfig.RecoveryTimeout <= 0 {
		serviceConfig.RecoveryTimeout = 30 * time.Second
	}

	return NewService(serviceConfig)
}

// GetProvider returns the underlying cache provider for service-specific managers
func (s *Service) GetProvider() CacheProvider {
	return s.provider
}

// HealthCheck verifies the cache service is healthy
func (s *Service) HealthCheck(ctx context.Context) error {
	return s.provider.HealthCheck(ctx)
}

// Close shuts down the cache service
func (s *Service) Close() error {
	if s.provider != nil {
		return s.provider.Close()
	}
	return nil
}

// GetConfig returns the cache service configuration
func (s *Service) GetConfig() ServiceConfig {
	return s.config
}

// Migrate migrates from old auth cache to new cache service
func (s *Service) Migrate(ctx context.Context, oldAuthCache interface{}) error {
	// This method can be used to migrate data from the old auth cache
	// to the new unified cache service if needed

	// For now, this is a placeholder that ensures the new cache is working
	return s.HealthCheck(ctx)
}

// DefaultServiceConfig returns a sensible default cache service configuration
func DefaultServiceConfig() ServiceConfig {
	return ServiceConfig{
		ProviderType:           "memory", // Safe default for development
		Host:                   "localhost",
		Port:                   6379,
		Password:               "",
		DB:                     0,
		MaxConnections:         100,
		MaxIdleConnections:     10,
		ConnectionTimeout:      5 * time.Second,
		FailureThreshold:       5,
		RequestVolumeThreshold: 10,
		RecoveryTimeout:        30 * time.Second,
		DefaultTTL:             5 * time.Minute,
	}
}
