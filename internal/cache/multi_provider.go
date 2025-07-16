package cache

import (
	"context"
	"time"
)

// multiProvider implements CacheProvider with Redis primary and memory fallback
// Provides automatic fallback when Redis is unavailable
type multiProvider struct {
	primary  CacheProvider // Redis provider
	fallback CacheProvider // Memory provider
	config   CacheConfig
}

// NewMultiProvider creates a multi-tier cache provider (Redis + Memory)
func NewMultiProvider(config CacheConfig) (CacheProvider, error) {
	// Create Redis primary
	primary, err := NewRedisProvider(config)
	if err != nil {
		return nil, err
	}

	// Create memory fallback
	fallback, err := NewMemoryProvider(config)
	if err != nil {
		primary.Close() // Clean up primary
		return nil, err
	}

	return &multiProvider{
		primary:  primary,
		fallback: fallback,
		config:   config,
	}, nil
}

// Get attempts to retrieve from primary, falls back to secondary
func (mp *multiProvider) Get(ctx context.Context, key string) (string, error) {
	// Try primary first
	value, err := mp.primary.Get(ctx, key)
	if err == nil {
		return value, nil
	}

	// Fall back to memory cache
	return mp.fallback.Get(ctx, key)
}

// Set stores in both primary and fallback
func (mp *multiProvider) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	// Always store in fallback (memory)
	if err := mp.fallback.Set(ctx, key, value, ttl); err != nil {
		return err
	}

	// Try to store in primary (Redis)
	// Don't fail if Redis is down, fallback is sufficient
	mp.primary.Set(ctx, key, value, ttl)

	return nil
}

// Delete removes from both primary and fallback
func (mp *multiProvider) Delete(ctx context.Context, key string) error {
	// Remove from both, don't fail if one fails
	mp.primary.Delete(ctx, key)
	return mp.fallback.Delete(ctx, key)
}

// Exists checks both primary and fallback
func (mp *multiProvider) Exists(ctx context.Context, key string) (bool, error) {
	// Check primary first
	exists, err := mp.primary.Exists(ctx, key)
	if err == nil && exists {
		return true, nil
	}

	// Check fallback
	return mp.fallback.Exists(ctx, key)
}

// MGet attempts to retrieve from primary, falls back to secondary
func (mp *multiProvider) MGet(ctx context.Context, keys []string) ([]string, error) {
	// Try primary first
	values, err := mp.primary.MGet(ctx, keys)
	if err == nil {
		return values, nil
	}

	// Fall back to memory cache
	return mp.fallback.MGet(ctx, keys)
}

// MSet stores in both primary and fallback
func (mp *multiProvider) MSet(ctx context.Context, pairs map[string]string, ttl time.Duration) error {
	// Always store in fallback (memory)
	if err := mp.fallback.MSet(ctx, pairs, ttl); err != nil {
		return err
	}

	// Try to store in primary (Redis)
	// Don't fail if Redis is down, fallback is sufficient
	mp.primary.MSet(ctx, pairs, ttl)

	return nil
}

// MDelete removes from both primary and fallback
func (mp *multiProvider) MDelete(ctx context.Context, keys []string) error {
	// Remove from both, don't fail if one fails
	mp.primary.MDelete(ctx, keys)
	return mp.fallback.MDelete(ctx, keys)
}

// Keys retrieves from primary, falls back to secondary
func (mp *multiProvider) Keys(ctx context.Context, pattern string) ([]string, error) {
	// Try primary first
	keys, err := mp.primary.Keys(ctx, pattern)
	if err == nil {
		return keys, nil
	}

	// Fall back to memory cache
	return mp.fallback.Keys(ctx, pattern)
}

// DeleteByPattern removes from both primary and fallback
func (mp *multiProvider) DeleteByPattern(ctx context.Context, pattern string) error {
	// Remove from both, don't fail if one fails
	mp.primary.DeleteByPattern(ctx, pattern)
	return mp.fallback.DeleteByPattern(ctx, pattern)
}

// HealthCheck checks both providers
func (mp *multiProvider) HealthCheck(ctx context.Context) error {
	// Check primary health
	if err := mp.primary.HealthCheck(ctx); err == nil {
		return nil // Primary is healthy
	}

	// Primary is down, check fallback
	return mp.fallback.HealthCheck(ctx)
}

// Close closes both providers
func (mp *multiProvider) Close() error {
	// Close both providers
	var primaryErr, fallbackErr error

	if mp.primary != nil {
		primaryErr = mp.primary.Close()
	}

	if mp.fallback != nil {
		fallbackErr = mp.fallback.Close()
	}

	// Return the first error encountered
	if primaryErr != nil {
		return primaryErr
	}
	return fallbackErr
}
