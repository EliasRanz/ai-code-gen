package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/EliasRanz/ai-code-gen/internal/cache"
)

// UserContext represents cached user authentication context
type UserContext struct {
	UserID   string    `json:"user_id"`
	Email    string    `json:"email"`
	Role     string    `json:"role"`
	CachedAt time.Time `json:"cached_at"`
}

// CacheManager handles auth-specific caching operations
type CacheManager struct {
	provider cache.CacheProvider
	config   CacheConfig
}

// CacheConfig holds auth cache configuration
type CacheConfig struct {
	TTL              time.Duration `json:"ttl"`
	TokenKeyPrefix   string        `json:"token_key_prefix"`
	SessionKeyPrefix string        `json:"session_key_prefix"`
}

// DefaultCacheConfig returns default auth cache configuration
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		TTL:              5 * time.Minute,
		TokenKeyPrefix:   "auth:token:",
		SessionKeyPrefix: "auth:session:",
	}
}

// NewCacheManager creates a new auth cache manager
func NewCacheManager(provider cache.CacheProvider, config CacheConfig) *CacheManager {
	return &CacheManager{
		provider: provider,
		config:   config,
	}
}

// GetUserContext retrieves cached user context by token hash
func (cm *CacheManager) GetUserContext(ctx context.Context, tokenHash string) (*UserContext, error) {
	if tokenHash == "" {
		return nil, fmt.Errorf("token hash cannot be empty")
	}

	key := cm.GenerateKey("token", tokenHash)

	data, err := cm.provider.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get user context from cache: %w", err)
	}

	if data == "" {
		return nil, nil // Cache miss
	}

	var userContext UserContext
	if err := json.Unmarshal([]byte(data), &userContext); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached user context: %w", err)
	}

	return &userContext, nil
}

// SetUserContext caches user context with TTL
func (cm *CacheManager) SetUserContext(ctx context.Context, tokenHash string, userContext *UserContext) error {
	if tokenHash == "" {
		return fmt.Errorf("token hash cannot be empty")
	}
	if userContext == nil {
		return fmt.Errorf("user context cannot be nil")
	}

	userContext.CachedAt = time.Now()

	data, err := json.Marshal(userContext)
	if err != nil {
		return fmt.Errorf("failed to marshal user context: %w", err)
	}

	key := cm.GenerateKey("token", tokenHash)
	if err := cm.provider.Set(ctx, key, string(data), cm.config.TTL); err != nil {
		return fmt.Errorf("failed to set user context in cache: %w", err)
	}

	return nil
}

// InvalidateUserContext removes cached auth result
func (cm *CacheManager) InvalidateUserContext(ctx context.Context, tokenHash string) error {
	if tokenHash == "" {
		return fmt.Errorf("token hash cannot be empty")
	}

	key := cm.GenerateKey("token", tokenHash)
	return cm.provider.Delete(ctx, key)
}

// InvalidateUserSessions removes all cached sessions for a user
func (cm *CacheManager) InvalidateUserSessions(ctx context.Context, userID string) error {
	if userID == "" {
		return fmt.Errorf("user ID cannot be empty")
	}

	pattern := cm.GenerateKey("user", userID, "*")
	return cm.provider.DeleteByPattern(ctx, pattern)
}

// CacheSession stores a session in cache
func (cm *CacheManager) CacheSession(ctx context.Context, sessionID, userID string, sessionData *SessionData) error {
	if sessionID == "" {
		return fmt.Errorf("session ID cannot be empty")
	}
	if userID == "" {
		return fmt.Errorf("user ID cannot be empty")
	}
	if sessionData == nil {
		return fmt.Errorf("session data cannot be nil")
	}

	sessionData.CachedAt = time.Now()

	data, err := json.Marshal(sessionData)
	if err != nil {
		return fmt.Errorf("failed to marshal session data: %w", err)
	}

	key := cm.GenerateKey("session", sessionID)
	return cm.provider.Set(ctx, key, string(data), cm.config.TTL)
}

// GetSession retrieves a session from cache
func (cm *CacheManager) GetSession(ctx context.Context, sessionID string) (*SessionData, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("session ID cannot be empty")
	}

	key := cm.GenerateKey("session", sessionID)

	data, err := cm.provider.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get session from cache: %w", err)
	}

	if data == "" {
		return nil, nil // Cache miss
	}

	var sessionData SessionData
	if err := json.Unmarshal([]byte(data), &sessionData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached session data: %w", err)
	}

	return &sessionData, nil
}

// InvalidateSession removes a session from cache
func (cm *CacheManager) InvalidateSession(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return fmt.Errorf("session ID cannot be empty")
	}

	key := cm.GenerateKey("session", sessionID)
	return cm.provider.Delete(ctx, key)
}

// GenerateKey creates cache keys with proper prefixes
func (cm *CacheManager) GenerateKey(keyType string, identifiers ...string) string {
	switch keyType {
	case "token":
		if len(identifiers) >= 1 {
			return fmt.Sprintf("%s%s", cm.config.TokenKeyPrefix, identifiers[0])
		}
	case "session":
		if len(identifiers) >= 1 {
			return fmt.Sprintf("%s%s", cm.config.SessionKeyPrefix, identifiers[0])
		}
	case "user":
		if len(identifiers) >= 2 {
			return fmt.Sprintf("auth:user:%s:%s", identifiers[0], identifiers[1])
		} else if len(identifiers) >= 1 {
			return fmt.Sprintf("auth:user:%s", identifiers[0])
		}
	}
	return fmt.Sprintf("auth:unknown:%s", identifiers[0])
}

// InvalidateByPattern removes all keys matching a pattern
func (cm *CacheManager) InvalidateByPattern(ctx context.Context, pattern string) error {
	return cm.provider.DeleteByPattern(ctx, pattern)
}

// InvalidateByUser removes all cached data for a specific user
func (cm *CacheManager) InvalidateByUser(ctx context.Context, userID string) error {
	return cm.InvalidateUserSessions(ctx, userID)
}

// HealthCheck verifies cache connectivity
func (cm *CacheManager) HealthCheck(ctx context.Context) error {
	return cm.provider.HealthCheck(ctx)
}

// GetJSON retrieves and unmarshals JSON data from cache
func (cm *CacheManager) GetJSON(ctx context.Context, key string, target interface{}) error {
	data, err := cm.provider.Get(ctx, key)
	if err != nil {
		return err
	}

	if data == "" {
		return nil // Cache miss
	}

	return json.Unmarshal([]byte(data), target)
}

// SetJSON marshals and stores JSON data in cache
func (cm *CacheManager) SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return cm.provider.Set(ctx, key, string(data), ttl)
}

// SessionData represents cached session information
type SessionData struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	ExpiresAt time.Time `json:"expires_at"`
	CachedAt  time.Time `json:"cached_at"`
}

// HashToken creates a SHA256 hash of the token for safe caching
func HashToken(token string) string {
	hasher := sha256.New()
	hasher.Write([]byte(token))
	return hex.EncodeToString(hasher.Sum(nil))
}
