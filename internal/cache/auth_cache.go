package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// UserContext represents cached user authentication context
type UserContext struct {
	UserID   string    `json:"user_id"`
	Email    string    `json:"email"`
	Role     string    `json:"role"`
	CachedAt time.Time `json:"cached_at"`
}

// AuthCache provides Redis-based caching for authentication results
type AuthCache struct {
	client *redis.Client
	ttl    time.Duration
}

// NewAuthCache creates a new auth cache instance
func NewAuthCache(redisURL string, ttl time.Duration) (*AuthCache, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	client := redis.NewClient(opt)
	
	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &AuthCache{
		client: client,
		ttl:    ttl,
	}, nil
}

// GetUserContext retrieves cached user context by token hash
func (ac *AuthCache) GetUserContext(ctx context.Context, tokenHash string) (*UserContext, error) {
	key := ac.generateKey(tokenHash)
	
	data, err := ac.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss
		}
		return nil, fmt.Errorf("failed to get from cache: %w", err)
	}

	var userContext UserContext
	if err := json.Unmarshal([]byte(data), &userContext); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached data: %w", err)
	}

	return &userContext, nil
}

// SetUserContext caches user context with TTL
func (ac *AuthCache) SetUserContext(ctx context.Context, tokenHash string, userContext *UserContext) error {
	userContext.CachedAt = time.Now()
	
	data, err := json.Marshal(userContext)
	if err != nil {
		return fmt.Errorf("failed to marshal user context: %w", err)
	}

	key := ac.generateKey(tokenHash)
	if err := ac.client.Set(ctx, key, data, ac.ttl).Err(); err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	return nil
}

// InvalidateUserContext removes cached auth result
func (ac *AuthCache) InvalidateUserContext(ctx context.Context, tokenHash string) error {
	key := ac.generateKey(tokenHash)
	return ac.client.Del(ctx, key).Err()
}

// InvalidateUserSessions removes all cached sessions for a user
func (ac *AuthCache) InvalidateUserSessions(ctx context.Context, userID string) error {
	pattern := fmt.Sprintf("auth:user:%s:*", userID)
	keys, err := ac.client.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to find user sessions: %w", err)
	}

	if len(keys) > 0 {
		return ac.client.Del(ctx, keys...).Err()
	}

	return nil
}

// Close closes the Redis connection
func (ac *AuthCache) Close() error {
	return ac.client.Close()
}

// generateKey creates a Redis key for token-based caching
func (ac *AuthCache) generateKey(tokenHash string) string {
	return fmt.Sprintf("auth:token:%s", tokenHash)
}

// HashToken creates a SHA256 hash of the token for safe caching
func HashToken(token string) string {
	hasher := sha256.New()
	hasher.Write([]byte(token))
	return hex.EncodeToString(hasher.Sum(nil))
}

// HealthCheck verifies Redis connectivity
func (ac *AuthCache) HealthCheck(ctx context.Context) error {
	return ac.client.Ping(ctx).Err()
}
