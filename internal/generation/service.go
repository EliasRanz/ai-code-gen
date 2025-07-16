package generation

import (
	"github.com/EliasRanz/ai-code-gen/internal/ai"
)

// Config holds configuration for the generation service
type Config struct {
	AIConfig    *ai.Config   `json:"ai"`
	RedisConfig *RedisConfig `json:"redis"`
}

// RedisConfig holds Redis configuration for pub/sub
type RedisConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

// Service provides AI generation functionality via AI service
type Service struct {
	aiService   *ai.AIService
	redisClient RedisClient
}

// NewService creates a new generation service using AI service
func NewService(aiService *ai.AIService, redisClient RedisClient) *Service {
	return &Service{
		aiService:   aiService,
		redisClient: redisClient,
	}
}

// Close shuts down the service gracefully
func (s *Service) Close() error {
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
