package generation

import (
	"github.com/EliasRanz/ai-code-gen/internal/auth"
	"github.com/EliasRanz/ai-code-gen/internal/llm"
)

// Config holds configuration for the generation service
type Config struct {
	LLMConfig   *llm.VLLMConfig `json:"llm"`
	RedisConfig *RedisConfig    `json:"redis"`
}

// RedisConfig holds Redis configuration for pub/sub
type RedisConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

// Service provides AI generation functionality
type Service struct {
	llmClient    llm.LLMClient
	redisClient  RedisClient
	authService  *auth.Service
}

// NewService creates a new generation service  
func NewService(llmClient llm.LLMClient, redisClient RedisClient, authService *auth.Service) *Service {
	return &Service{
		llmClient:   llmClient,
		redisClient: redisClient,
		authService: authService,
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

	// Close LLM client connection  
	if s.llmClient != nil {
		if closeErr := s.llmClient.Close(); closeErr != nil {
			err = closeErr
		}
	}

	return err
}
