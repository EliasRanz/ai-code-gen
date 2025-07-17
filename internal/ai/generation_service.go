package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/EliasRanz/ai-code-gen/internal/ai/llm"
	"github.com/EliasRanz/ai-code-gen/internal/utilities"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

// GenerationResponse represents the response from AI generation
type GenerationResponse struct {
	Content    string               `json:"content"`
	TokensUsed int                  `json:"tokens_used"`
	Provider   string               `json:"provider"`
	Model      string               `json:"model"`
	UserID     utilities.UserID     `json:"user_id"`
	ProjectID  *utilities.ProjectID `json:"project_id,omitempty"`
	Timestamp  time.Time            `json:"timestamp"`
}

// RedisConfig holds Redis configuration for pub/sub
type RedisConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

// GenerationService provides AI generation functionality
type GenerationService struct {
	aiService   *AIService
	redisClient RedisClient
}

// NewGenerationService creates a new generation service using AI service
func NewGenerationService(aiService *AIService, redisClient RedisClient) *GenerationService {
	return &GenerationService{
		aiService:   aiService,
		redisClient: redisClient,
	}
}

// publishToRedis publishes generation response to Redis channels
func (s *GenerationService) publishToRedis(resp *llm.GenerationResponse, userID, projectID string) {
	ctx := context.Background()

	message := gin.H{
		"response":   resp,
		"user_id":    userID,
		"project_id": projectID,
		"timestamp":  time.Now().UTC(),
	}

	jsonMessage, err := json.Marshal(message)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal message for Redis")
		return
	}

	// Publish to user-specific channel
	if userID != "" {
		channel := fmt.Sprintf("user:%s:generations", userID)
		if err := s.redisClient.Publish(ctx, channel, jsonMessage).Err(); err != nil {
			log.Error().Err(err).Str("channel", channel).Msg("Failed to publish to user channel")
		}
	}

	// Publish to project-specific channel
	if projectID != "" {
		channel := fmt.Sprintf("project:%s:generations", projectID)
		if err := s.redisClient.Publish(ctx, channel, jsonMessage).Err(); err != nil {
			log.Error().Err(err).Str("channel", channel).Msg("Failed to publish to project channel")
		}
	}

	// Publish to global channel
	if err := s.redisClient.Publish(ctx, "global:generations", jsonMessage).Err(); err != nil {
		log.Error().Err(err).Msg("Failed to publish to global channel")
	}
}

// SubscribeToUserChannel subscribes to user-specific generation events
func (s *GenerationService) SubscribeToUserChannel(ctx context.Context, userID string) (*redis.PubSub, error) {
	if s.redisClient == nil {
		return nil, fmt.Errorf("redis client is not initialized")
	}
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	channel := fmt.Sprintf("user:%s:generations", userID)
	pubsub := s.redisClient.Subscribe(ctx, channel)

	return pubsub, nil
}

// SubscribeToProjectChannel subscribes to project-specific generation events
func (s *GenerationService) SubscribeToProjectChannel(ctx context.Context, projectID string) (*redis.PubSub, error) {
	if s.redisClient == nil {
		return nil, fmt.Errorf("redis client is not initialized")
	}
	if projectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}

	channel := fmt.Sprintf("project:%s:generations", projectID)
	pubsub := s.redisClient.Subscribe(ctx, channel)

	return pubsub, nil
}

// SubscribeToGlobalChannel subscribes to global generation events
func (s *GenerationService) SubscribeToGlobalChannel(ctx context.Context) (*redis.PubSub, error) {
	if s.redisClient == nil {
		return nil, fmt.Errorf("redis client is not initialized")
	}
	pubsub := s.redisClient.Subscribe(ctx, "global:generations")
	return pubsub, nil
}

// Close shuts down the service gracefully
func (s *GenerationService) Close() error {
	var err error

	// Close Redis connection
	if s.redisClient != nil {
		if closeErr := s.redisClient.Close(); closeErr != nil {
			err = closeErr
		}
	}

	// Close AI service
	if s.aiService != nil {
		if closeErr := s.aiService.Close(); closeErr != nil {
			err = closeErr
		}
	}

	return err
}

// GetMetrics returns generation service metrics
func (s *GenerationService) GetMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})

	// Add AI service metrics if available
	if s.aiService != nil {
		// This would depend on AIService having a GetMetrics method
		metrics["ai_service"] = map[string]interface{}{
			"status": "active",
		}
	}

	// Add Redis metrics if available
	if s.redisClient != nil {
		metrics["redis"] = map[string]interface{}{
			"status": "connected",
		}
	}

	return metrics
}

// ValidateConfig validates the generation service configuration
func ValidateGenerationConfig(aiConfig *Config, redisConfig *RedisConfig) error {
	// Validate AI config
	if aiConfig == nil {
		return fmt.Errorf("AI configuration is required")
	}

	// Validate Redis config
	if redisConfig != nil {
		if redisConfig.Host == "" {
			return fmt.Errorf("Redis host is required")
		}
		if redisConfig.Port <= 0 {
			return fmt.Errorf("Redis port must be positive")
		}
	}

	return nil
}
