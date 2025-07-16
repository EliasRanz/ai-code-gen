package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/EliasRanz/ai-code-gen/internal/cache"
)

// CacheManager handles AI-specific caching operations
type CacheManager struct {
	provider cache.CacheProvider
	config   CacheConfig
}

// CacheConfig holds AI cache configuration
type CacheConfig struct {
	TTL                    time.Duration `json:"ttl"`
	GenerationKeyPrefix    string        `json:"generation_key_prefix"`
	RateLimitKeyPrefix     string        `json:"rate_limit_key_prefix"`
	ModelResponseKeyPrefix string        `json:"model_response_key_prefix"`
}

// DefaultCacheConfig returns default AI cache configuration
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		TTL:                    30 * time.Minute,
		GenerationKeyPrefix:    "ai:generation:",
		RateLimitKeyPrefix:     "ai:rate_limit:",
		ModelResponseKeyPrefix: "ai:model_response:",
	}
}

// NewCacheManager creates a new AI cache manager
func NewCacheManager(provider cache.CacheProvider, config CacheConfig) *CacheManager {
	return &CacheManager{
		provider: provider,
		config:   config,
	}
}

// CacheGeneration stores AI generation result in cache
func (cm *CacheManager) CacheGeneration(ctx context.Context, generationID string, generation *CachedGenerationResult) error {
	if generationID == "" {
		return fmt.Errorf("generation ID cannot be empty")
	}
	if generation == nil {
		return fmt.Errorf("generation result cannot be nil")
	}

	generation.CachedAt = time.Now()
	key := cm.GenerateKey("generation", generationID)

	return cm.SetJSON(ctx, key, generation, cm.config.TTL)
}

// GetGeneration retrieves AI generation result from cache
func (cm *CacheManager) GetGeneration(ctx context.Context, generationID string) (*CachedGenerationResult, error) {
	if generationID == "" {
		return nil, fmt.Errorf("generation ID cannot be empty")
	}

	key := cm.GenerateKey("generation", generationID)

	var generation CachedGenerationResult
	if err := cm.GetJSON(ctx, key, &generation); err != nil {
		return nil, err
	}

	// Check if we got data
	if generation.GenerationID == "" {
		return nil, nil // Cache miss
	}

	return &generation, nil
}

// InvalidateGeneration removes generation from cache
func (cm *CacheManager) InvalidateGeneration(ctx context.Context, generationID string) error {
	if generationID == "" {
		return fmt.Errorf("generation ID cannot be empty")
	}

	key := cm.GenerateKey("generation", generationID)
	return cm.provider.Delete(ctx, key)
}

// CacheUserGenerations stores user generations list in cache
func (cm *CacheManager) CacheUserGenerations(ctx context.Context, userID string, generations []GenerationSummary) error {
	if userID == "" {
		return fmt.Errorf("user ID cannot be empty")
	}

	key := cm.GenerateKey("user_generations", userID)

	data := UserGenerationsCache{
		UserID:      userID,
		Generations: generations,
		CachedAt:    time.Now(),
	}

	return cm.SetJSON(ctx, key, data, cm.config.TTL)
}

// GetUserGenerations retrieves user generations from cache
func (cm *CacheManager) GetUserGenerations(ctx context.Context, userID string) ([]GenerationSummary, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}

	key := cm.GenerateKey("user_generations", userID)

	var data UserGenerationsCache
	if err := cm.GetJSON(ctx, key, &data); err != nil {
		return nil, err
	}

	// Check if we got data
	if data.UserID == "" {
		return nil, nil // Cache miss
	}

	return data.Generations, nil
}

// InvalidateUserGenerations removes user generations from cache
func (cm *CacheManager) InvalidateUserGenerations(ctx context.Context, userID string) error {
	if userID == "" {
		return fmt.Errorf("user ID cannot be empty")
	}

	key := cm.GenerateKey("user_generations", userID)
	return cm.provider.Delete(ctx, key)
}

// CacheModelResponse stores LLM model response for potential reuse
func (cm *CacheManager) CacheModelResponse(ctx context.Context, requestHash string, response *ModelResponse) error {
	if requestHash == "" {
		return fmt.Errorf("request hash cannot be empty")
	}
	if response == nil {
		return fmt.Errorf("model response cannot be nil")
	}

	response.CachedAt = time.Now()
	key := cm.GenerateKey("model_response", requestHash)

	// Cache model responses for shorter time (to reduce costs)
	cacheTTL := 15 * time.Minute
	return cm.SetJSON(ctx, key, response, cacheTTL)
}

// GetModelResponse retrieves cached LLM model response
func (cm *CacheManager) GetModelResponse(ctx context.Context, requestHash string) (*ModelResponse, error) {
	if requestHash == "" {
		return nil, fmt.Errorf("request hash cannot be empty")
	}

	key := cm.GenerateKey("model_response", requestHash)

	var response ModelResponse
	if err := cm.GetJSON(ctx, key, &response); err != nil {
		return nil, err
	}

	// Check if we got data
	if response.RequestHash == "" {
		return nil, nil // Cache miss
	}

	return &response, nil
}

// CacheRateLimit stores rate limiting information
func (cm *CacheManager) CacheRateLimit(ctx context.Context, userID string, rateLimitData *RateLimitData) error {
	if userID == "" {
		return fmt.Errorf("user ID cannot be empty")
	}
	if rateLimitData == nil {
		return fmt.Errorf("rate limit data cannot be nil")
	}

	rateLimitData.CachedAt = time.Now()
	key := cm.GenerateKey("rate_limit", userID)

	// Rate limit data should have shorter TTL (1 hour max)
	rateLimitTTL := time.Hour
	return cm.SetJSON(ctx, key, rateLimitData, rateLimitTTL)
}

// GetRateLimit retrieves rate limiting information
func (cm *CacheManager) GetRateLimit(ctx context.Context, userID string) (*RateLimitData, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}

	key := cm.GenerateKey("rate_limit", userID)

	var rateLimitData RateLimitData
	if err := cm.GetJSON(ctx, key, &rateLimitData); err != nil {
		return nil, err
	}

	// Check if we got data
	if rateLimitData.UserID == "" {
		return nil, nil // Cache miss
	}

	return &rateLimitData, nil
}

// InvalidateRateLimit removes rate limit data from cache
func (cm *CacheManager) InvalidateRateLimit(ctx context.Context, userID string) error {
	if userID == "" {
		return fmt.Errorf("user ID cannot be empty")
	}

	key := cm.GenerateKey("rate_limit", userID)
	return cm.provider.Delete(ctx, key)
}

// GenerateKey creates cache keys with proper prefixes
func (cm *CacheManager) GenerateKey(keyType string, identifiers ...string) string {
	switch keyType {
	case "generation":
		if len(identifiers) >= 1 {
			return fmt.Sprintf("%s%s", cm.config.GenerationKeyPrefix, identifiers[0])
		}
	case "user_generations":
		if len(identifiers) >= 1 {
			return fmt.Sprintf("ai:user_generations:%s", identifiers[0])
		}
	case "model_response":
		if len(identifiers) >= 1 {
			return fmt.Sprintf("%s%s", cm.config.ModelResponseKeyPrefix, identifiers[0])
		}
	case "rate_limit":
		if len(identifiers) >= 1 {
			return fmt.Sprintf("%s%s", cm.config.RateLimitKeyPrefix, identifiers[0])
		}
	}
	return fmt.Sprintf("ai:unknown:%s", identifiers[0])
}

// InvalidateByPattern removes all keys matching a pattern
func (cm *CacheManager) InvalidateByPattern(ctx context.Context, pattern string) error {
	return cm.provider.DeleteByPattern(ctx, pattern)
}

// InvalidateByUser removes all cached AI data for a specific user
func (cm *CacheManager) InvalidateByUser(ctx context.Context, userID string) error {
	patterns := []string{
		fmt.Sprintf("ai:*:%s:*", userID),
		fmt.Sprintf("ai:*:%s", userID),
	}

	for _, pattern := range patterns {
		if err := cm.provider.DeleteByPattern(ctx, pattern); err != nil {
			return err
		}
	}

	return nil
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

// Cache data structures

// CachedGenerationResult represents cached AI generation result with caching metadata
type CachedGenerationResult struct {
	GenerationID string    `json:"generation_id"`
	UserID       string    `json:"user_id"`
	ProjectID    string    `json:"project_id"`
	Prompt       string    `json:"prompt"`
	Response     string    `json:"response"`
	Status       string    `json:"status"`
	Model        string    `json:"model"`
	TokensUsed   int       `json:"tokens_used"`
	CreatedAt    time.Time `json:"created_at"`
	CachedAt     time.Time `json:"cached_at"`
}

// GenerationSummary represents a summarized generation for lists
type GenerationSummary struct {
	GenerationID string    `json:"generation_id"`
	ProjectID    string    `json:"project_id"`
	Prompt       string    `json:"prompt"`
	Status       string    `json:"status"`
	TokensUsed   int       `json:"tokens_used"`
	CreatedAt    time.Time `json:"created_at"`
}

// UserGenerationsCache represents cached generations list
type UserGenerationsCache struct {
	UserID      string              `json:"user_id"`
	Generations []GenerationSummary `json:"generations"`
	CachedAt    time.Time           `json:"cached_at"`
}

// ModelResponse represents cached LLM model response
type ModelResponse struct {
	RequestHash string    `json:"request_hash"`
	Model       string    `json:"model"`
	Response    string    `json:"response"`
	TokensUsed  int       `json:"tokens_used"`
	CachedAt    time.Time `json:"cached_at"`
}

// RateLimitData represents cached rate limiting information
type RateLimitData struct {
	UserID       string    `json:"user_id"`
	RequestCount int       `json:"request_count"`
	TokensUsed   int       `json:"tokens_used"`
	WindowStart  time.Time `json:"window_start"`
	WindowEnd    time.Time `json:"window_end"`
	Tier         string    `json:"tier"`
	IsBlocked    bool      `json:"is_blocked"`
	BlockedUntil time.Time `json:"blocked_until"`
	CachedAt     time.Time `json:"cached_at"`
}
