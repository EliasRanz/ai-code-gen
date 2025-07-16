package cache

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// memoryItem represents a cached item with expiration
type memoryItem struct {
	value     string
	expiresAt time.Time
}

// isExpired checks if the item has expired
func (item *memoryItem) isExpired() bool {
	return !item.expiresAt.IsZero() && time.Now().After(item.expiresAt)
}

// memoryProvider implements CacheProvider with in-memory storage
// Used as fallback when Redis is unavailable
type memoryProvider struct {
	mutex        sync.RWMutex
	data         map[string]*memoryItem
	config       CacheConfig
	cleaner      *time.Ticker
	cleanerMutex sync.RWMutex
	done         chan bool
}

// NewMemoryProvider creates a new in-memory cache provider
func NewMemoryProvider(config CacheConfig) (CacheProvider, error) {
	provider := &memoryProvider{
		data:   make(map[string]*memoryItem),
		config: config,
		done:   make(chan bool),
	}

	// Start cleanup goroutine to remove expired items
	provider.cleaner = time.NewTicker(time.Minute)
	go provider.cleanupExpired()

	return provider, nil
}

// Get retrieves a value from memory cache
func (m *memoryProvider) Get(ctx context.Context, key string) (string, error) {
	if key == "" {
		return "", fmt.Errorf("key cannot be empty")
	}

	m.mutex.RLock()
	item, exists := m.data[key]
	m.mutex.RUnlock()

	if !exists {
		return "", nil // Cache miss
	}

	if item.isExpired() {
		// Remove expired item
		m.mutex.Lock()
		delete(m.data, key)
		m.mutex.Unlock()
		return "", nil // Cache miss
	}

	return item.value, nil
}

// Set stores a value in memory cache with TTL
func (m *memoryProvider) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	if ttl <= 0 {
		ttl = m.config.DefaultTTL
	}

	item := &memoryItem{
		value:     value,
		expiresAt: time.Now().Add(ttl),
	}

	m.mutex.Lock()
	m.data[key] = item
	m.mutex.Unlock()

	return nil
}

// Delete removes a key from memory cache
func (m *memoryProvider) Delete(ctx context.Context, key string) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	m.mutex.Lock()
	delete(m.data, key)
	m.mutex.Unlock()

	return nil
}

// Exists checks if a key exists in memory cache
func (m *memoryProvider) Exists(ctx context.Context, key string) (bool, error) {
	if key == "" {
		return false, fmt.Errorf("key cannot be empty")
	}

	m.mutex.RLock()
	item, exists := m.data[key]
	m.mutex.RUnlock()

	if !exists {
		return false, nil
	}

	if item.isExpired() {
		// Remove expired item
		m.mutex.Lock()
		delete(m.data, key)
		m.mutex.Unlock()
		return false, nil
	}

	return true, nil
}

// MGet retrieves multiple values from memory cache
func (m *memoryProvider) MGet(ctx context.Context, keys []string) ([]string, error) {
	if len(keys) == 0 {
		return []string{}, nil
	}

	for _, key := range keys {
		if key == "" {
			return nil, fmt.Errorf("key cannot be empty")
		}
	}

	values := make([]string, len(keys))

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for i, key := range keys {
		if item, exists := m.data[key]; exists && !item.isExpired() {
			values[i] = item.value
		}
		// Leave empty string for missing/expired keys
	}

	return values, nil
}

// MSet stores multiple key-value pairs in memory cache
func (m *memoryProvider) MSet(ctx context.Context, pairs map[string]string, ttl time.Duration) error {
	if len(pairs) == 0 {
		return nil
	}

	if ttl <= 0 {
		ttl = m.config.DefaultTTL
	}

	expiresAt := time.Now().Add(ttl)

	m.mutex.Lock()
	defer m.mutex.Unlock()

	for key, value := range pairs {
		if key == "" {
			return fmt.Errorf("key cannot be empty")
		}
		m.data[key] = &memoryItem{
			value:     value,
			expiresAt: expiresAt,
		}
	}

	return nil
}

// MDelete removes multiple keys from memory cache
func (m *memoryProvider) MDelete(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	for _, key := range keys {
		if key == "" {
			return fmt.Errorf("key cannot be empty")
		}
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, key := range keys {
		delete(m.data, key)
	}

	return nil
}

// Keys retrieves keys matching a pattern from memory cache
func (m *memoryProvider) Keys(ctx context.Context, pattern string) ([]string, error) {
	if pattern == "" {
		return []string{}, nil
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var matchingKeys []string
	for key, item := range m.data {
		if item.isExpired() {
			continue // Skip expired items
		}

		if m.matchesPattern(key, pattern) {
			matchingKeys = append(matchingKeys, key)
		}
	}

	return matchingKeys, nil
}

// DeleteByPattern removes all keys matching a pattern from memory cache
func (m *memoryProvider) DeleteByPattern(ctx context.Context, pattern string) error {
	if pattern == "" {
		return nil
	}

	keys, err := m.Keys(ctx, pattern)
	if err != nil {
		return err
	}

	return m.MDelete(ctx, keys)
}

// HealthCheck always returns nil for memory cache (always healthy)
func (m *memoryProvider) HealthCheck(ctx context.Context) error {
	return nil
}

// Close stops the cleanup goroutine and clears the cache
func (m *memoryProvider) Close() error {
	// Close the done channel first to signal the cleanup goroutine to stop
	select {
	case <-m.done:
		// Channel is already closed
	default:
		close(m.done)
	}

	// Stop the cleanup goroutine if it's running (with proper synchronization)
	m.cleanerMutex.Lock()
	if m.cleaner != nil {
		m.cleaner.Stop()
		m.cleaner = nil
	}
	m.cleanerMutex.Unlock()

	m.mutex.Lock()
	m.data = make(map[string]*memoryItem)
	m.mutex.Unlock()

	return nil
}

// cleanupExpired removes expired items from the cache
func (m *memoryProvider) cleanupExpired() {
	for {
		// Use a read lock to safely access the cleaner
		m.cleanerMutex.RLock()
		cleaner := m.cleaner
		m.cleanerMutex.RUnlock()

		if cleaner == nil {
			return
		}

		select {
		case <-cleaner.C:
			m.removeExpiredItems()
		case <-m.done:
			return
		}
	}
}

// removeExpiredItems removes all expired items from the cache
func (m *memoryProvider) removeExpiredItems() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for key, item := range m.data {
		if item.isExpired() {
			delete(m.data, key)
		}
	}
}

// matchesPattern implements basic pattern matching for Redis-style patterns
// Supports * for wildcard matching
func (m *memoryProvider) matchesPattern(key, pattern string) bool {
	// Simple implementation: convert Redis pattern to basic matching
	if pattern == "*" {
		return true
	}

	// Handle patterns with * wildcards
	if strings.Contains(pattern, "*") {
		parts := strings.Split(pattern, "*")
		if len(parts) == 2 {
			prefix := parts[0]
			suffix := parts[1]
			return strings.HasPrefix(key, prefix) && strings.HasSuffix(key, suffix)
		}
	}

	// Exact match
	return key == pattern
}
