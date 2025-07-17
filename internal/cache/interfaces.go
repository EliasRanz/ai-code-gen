package cache

import (
	"context"
	"fmt"
	"time"
)

// CacheProvider defines the core interface that all cache providers must implement
// This implements the Cache Interface Pattern for provider-agnostic caching
type CacheProvider interface {
	BasicCacheOperations
	BatchCacheOperations
	PatternCacheOperations
	HealthOperations
}

// BasicCacheOperations defines single-key cache operations
type BasicCacheOperations interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
}

// BatchCacheOperations defines multi-key cache operations
type BatchCacheOperations interface {
	MGet(ctx context.Context, keys []string) ([]string, error)
	MSet(ctx context.Context, pairs map[string]string, ttl time.Duration) error
	MDelete(ctx context.Context, keys []string) error
}

// PatternCacheOperations defines pattern-based cache operations
type PatternCacheOperations interface {
	Keys(ctx context.Context, pattern string) ([]string, error)
	DeleteByPattern(ctx context.Context, pattern string) error
}

// HealthOperations defines health and lifecycle operations
type HealthOperations interface {
	HealthCheck(ctx context.Context) error
	Close() error
}

// CacheFactory implements the Factory Pattern for cache provider instantiation
type CacheFactory interface {
	CreateProvider(providerType string, config CacheConfig) (CacheProvider, error)
	ListAvailableProviders() []string
}

// CacheConfig holds configuration for cache providers
type CacheConfig struct {
	// Connection settings
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	DB       int    `json:"db"`

	// Connection pooling
	MaxConnections     int           `json:"max_connections"`
	MaxIdleConnections int           `json:"max_idle_connections"`
	ConnectionTimeout  time.Duration `json:"connection_timeout"`
	IdleTimeout        time.Duration `json:"idle_timeout"`

	// Circuit breaker settings
	FailureThreshold       int           `json:"failure_threshold"`
	RequestVolumeThreshold int           `json:"request_volume_threshold"`
	RecoveryTimeout        time.Duration `json:"recovery_timeout"`

	// Default TTL settings
	DefaultTTL time.Duration `json:"default_ttl"`
}

// CacheError represents cache-specific errors with proper classification
type CacheError struct {
	Op       string // Operation being performed
	Key      string // Cache key involved
	Provider string // Cache provider type
	Err      error  // Underlying error
}

func (e *CacheError) Error() string {
	return fmt.Sprintf("cache %s operation failed for key '%s' on provider '%s': %v",
		e.Op, e.Key, e.Provider, e.Err)
}

func (e *CacheError) Unwrap() error {
	return e.Err
}

// Cache operation result types
type CacheResult struct {
	Value string
	Found bool
	TTL   time.Duration
}

// CacheMetrics interface for monitoring cache operations
type CacheMetrics interface {
	IncrementHit(provider, operation string)
	IncrementMiss(provider, operation string)
	IncrementError(provider, operation string)
	RecordLatency(provider, operation string, duration time.Duration)
}

// ServiceCacheManager defines the interface for service-specific cache managers
// Each service implements this interface to manage its own cache operations
type ServiceCacheManager interface {
	// Service-specific cache operations with type safety
	GetJSON(ctx context.Context, key string, target interface{}) error
	SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error

	// Service-specific key generation
	GenerateKey(keyType string, identifiers ...string) string

	// Service-specific invalidation patterns
	InvalidateByPattern(ctx context.Context, pattern string) error
	InvalidateByUser(ctx context.Context, userID string) error

	// Health check for service cache
	HealthCheck(ctx context.Context) error
}
