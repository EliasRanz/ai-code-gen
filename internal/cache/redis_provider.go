package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	redis "github.com/redis/go-redis/v9"
)

// redisProvider implements CacheProvider with Redis backend and circuit breaker
type redisProvider struct {
	client         *redis.Client
	circuitBreaker CircuitBreaker
	config         CacheConfig
}

// NewRedisProvider creates a new Redis cache provider with circuit breaker protection
func NewRedisProvider(config CacheConfig) (CacheProvider, error) {
	if err := ValidateCacheConfig(config); err != nil {
		return nil, fmt.Errorf("invalid Redis config: %w", err)
	}

	// Create Redis client options
	opt := &redis.Options{
		Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password:     config.Password,
		DB:           config.DB,
		PoolSize:     config.MaxConnections,
		MinIdleConns: config.MaxIdleConnections,
		DialTimeout:  config.ConnectionTimeout,
	}

	client := redis.NewClient(opt)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), config.ConnectionTimeout)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	// Create circuit breaker
	cbConfig := CircuitBreakerConfig{
		FailureThreshold:       config.FailureThreshold,
		RequestVolumeThreshold: config.RequestVolumeThreshold,
		RecoveryTimeout:        config.RecoveryTimeout,
		MaxConcurrentRequests:  config.MaxConnections,
	}
	circuitBreaker := NewCircuitBreaker(cbConfig)

	return &redisProvider{
		client:         client,
		circuitBreaker: circuitBreaker,
		config:         config,
	}, nil
}

// Get retrieves a value from Redis through the circuit breaker
func (r *redisProvider) Get(ctx context.Context, key string) (string, error) {
	if key == "" {
		return "", fmt.Errorf("key cannot be empty")
	}

	result, err := r.circuitBreaker.Execute(ctx, func() (interface{}, error) {
		return r.client.Get(ctx, key).Result()
	})

	if err != nil {
		if result == nil && redis.Nil.Error() == err.Error() {
			return "", nil // Cache miss, not an error
		}
		return "", &CacheError{
			Op:       "get",
			Key:      key,
			Provider: "redis",
			Err:      err,
		}
	}

	return result.(string), nil
}

// Set stores a value in Redis through the circuit breaker
func (r *redisProvider) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	if ttl <= 0 {
		ttl = r.config.DefaultTTL
	}

	_, err := r.circuitBreaker.Execute(ctx, func() (interface{}, error) {
		return nil, r.client.Set(ctx, key, value, ttl).Err()
	})

	if err != nil {
		return &CacheError{
			Op:       "set",
			Key:      key,
			Provider: "redis",
			Err:      err,
		}
	}

	return nil
}

// Delete removes a key from Redis through the circuit breaker
func (r *redisProvider) Delete(ctx context.Context, key string) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	_, err := r.circuitBreaker.Execute(ctx, func() (interface{}, error) {
		return nil, r.client.Del(ctx, key).Err()
	})

	if err != nil {
		return &CacheError{
			Op:       "delete",
			Key:      key,
			Provider: "redis",
			Err:      err,
		}
	}

	return nil
}

// Exists checks if a key exists in Redis through the circuit breaker
func (r *redisProvider) Exists(ctx context.Context, key string) (bool, error) {
	if key == "" {
		return false, fmt.Errorf("key cannot be empty")
	}

	result, err := r.circuitBreaker.Execute(ctx, func() (interface{}, error) {
		count, err := r.client.Exists(ctx, key).Result()
		return count > 0, err
	})

	if err != nil {
		return false, &CacheError{
			Op:       "exists",
			Key:      key,
			Provider: "redis",
			Err:      err,
		}
	}

	return result.(bool), nil
}

// MGet retrieves multiple values from Redis through the circuit breaker
func (r *redisProvider) MGet(ctx context.Context, keys []string) ([]string, error) {
	if len(keys) == 0 {
		return []string{}, nil
	}

	for _, key := range keys {
		if key == "" {
			return nil, fmt.Errorf("key cannot be empty")
		}
	}

	result, err := r.circuitBreaker.Execute(ctx, func() (interface{}, error) {
		return r.client.MGet(ctx, keys...).Result()
	})

	if err != nil {
		return nil, &CacheError{
			Op:       "mget",
			Key:      fmt.Sprintf("[%d keys]", len(keys)),
			Provider: "redis",
			Err:      err,
		}
	}

	// Convert interface{} slice to string slice
	values := result.([]interface{})
	strings := make([]string, len(values))
	for i, v := range values {
		if v != nil {
			strings[i] = v.(string)
		}
	}

	return strings, nil
}

// MSet stores multiple key-value pairs in Redis through the circuit breaker
func (r *redisProvider) MSet(ctx context.Context, pairs map[string]string, ttl time.Duration) error {
	if len(pairs) == 0 {
		return nil
	}

	if ttl <= 0 {
		ttl = r.config.DefaultTTL
	}

	_, err := r.circuitBreaker.Execute(ctx, func() (interface{}, error) {
		pipe := r.client.Pipeline()

		for key, value := range pairs {
			if key == "" {
				return nil, fmt.Errorf("key cannot be empty")
			}
			pipe.Set(ctx, key, value, ttl)
		}

		_, err := pipe.Exec(ctx)
		return nil, err
	})

	if err != nil {
		return &CacheError{
			Op:       "mset",
			Key:      fmt.Sprintf("[%d pairs]", len(pairs)),
			Provider: "redis",
			Err:      err,
		}
	}

	return nil
}

// MDelete removes multiple keys from Redis through the circuit breaker
func (r *redisProvider) MDelete(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	for _, key := range keys {
		if key == "" {
			return fmt.Errorf("key cannot be empty")
		}
	}

	_, err := r.circuitBreaker.Execute(ctx, func() (interface{}, error) {
		return nil, r.client.Del(ctx, keys...).Err()
	})

	if err != nil {
		return &CacheError{
			Op:       "mdelete",
			Key:      fmt.Sprintf("[%d keys]", len(keys)),
			Provider: "redis",
			Err:      err,
		}
	}

	return nil
}

// Keys retrieves keys matching a pattern from Redis through the circuit breaker
func (r *redisProvider) Keys(ctx context.Context, pattern string) ([]string, error) {
	if pattern == "" {
		return []string{}, nil
	}

	result, err := r.circuitBreaker.Execute(ctx, func() (interface{}, error) {
		return r.client.Keys(ctx, pattern).Result()
	})

	if err != nil {
		return nil, &CacheError{
			Op:       "keys",
			Key:      pattern,
			Provider: "redis",
			Err:      err,
		}
	}

	return result.([]string), nil
}

// DeleteByPattern removes all keys matching a pattern from Redis
func (r *redisProvider) DeleteByPattern(ctx context.Context, pattern string) error {
	if pattern == "" {
		return nil
	}

	keys, err := r.Keys(ctx, pattern)
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		return r.MDelete(ctx, keys)
	}

	return nil
}

// HealthCheck verifies Redis connectivity through the circuit breaker
func (r *redisProvider) HealthCheck(ctx context.Context) error {
	_, err := r.circuitBreaker.Execute(ctx, func() (interface{}, error) {
		return nil, r.client.Ping(ctx).Err()
	})

	if err != nil {
		return &CacheError{
			Op:       "healthcheck",
			Key:      "ping",
			Provider: "redis",
			Err:      err,
		}
	}

	return nil
}

// Close closes the Redis connection
func (r *redisProvider) Close() error {
	if r.client != nil {
		return r.client.Close()
	}
	return nil
}

// Helper methods for JSON operations (commonly used by service cache managers)

// GetJSON retrieves and unmarshals JSON data from Redis
func (r *redisProvider) GetJSON(ctx context.Context, key string, target interface{}) error {
	data, err := r.Get(ctx, key)
	if err != nil {
		return err
	}

	if data == "" {
		return nil // Cache miss
	}

	if err := json.Unmarshal([]byte(data), target); err != nil {
		return &CacheError{
			Op:       "get_json",
			Key:      key,
			Provider: "redis",
			Err:      fmt.Errorf("failed to unmarshal JSON: %w", err),
		}
	}

	return nil
}

// SetJSON marshals and stores JSON data in Redis
func (r *redisProvider) SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return &CacheError{
			Op:       "set_json",
			Key:      key,
			Provider: "redis",
			Err:      fmt.Errorf("failed to marshal JSON: %w", err),
		}
	}

	return r.Set(ctx, key, string(data), ttl)
}
