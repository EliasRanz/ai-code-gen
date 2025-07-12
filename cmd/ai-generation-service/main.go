package main

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/EliasRanz/ai-code-gen/internal/config"
	"github.com/EliasRanz/ai-code-gen/internal/generation"
	"github.com/EliasRanz/ai-code-gen/internal/llm"
	"github.com/EliasRanz/ai-code-gen/internal/service"
)

func main() {
	// Initialize configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Override port for this service
	cfg.Server.Port = cfg.AIGenService.Port

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatal().Err(err).Msg("Configuration validation failed")
	}

	// Create service
	svc := service.New("ai-generation-service", "1.0.0", cfg)

	// Initialize observability
	if err := svc.Initialize(); err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize service")
	}

	// Initialize generation service (auth-agnostic - trusts API Gateway)
	genConfig := &generation.Config{
		LLMConfig: &llm.VLLMConfig{
			BaseURL:    cfg.AI.LLM.BaseURL,
			APIKey:     cfg.AI.LLM.APIKey,
			Timeout:    cfg.AI.LLM.Timeout,
			MaxRetries: cfg.AI.LLM.MaxRetries,
		},
		RedisConfig: &generation.RedisConfig{
			Host:     cfg.Redis.Host,
			Port:     cfg.Redis.Port,
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		},
	}

	// Initialize LLM client and Redis client
	llmClient := llm.NewVLLMClient(genConfig.LLMConfig)
	redisClient := generation.NewRedisClient(genConfig.RedisConfig)
	genService := generation.NewService(llmClient, redisClient)

	// Setup HTTP router
	router := setupGenerationRouter(cfg, genService)

	// Setup HTTP server
	svc.SetupHTTPServer(router)

	// Start service
	if err := svc.Start(); err != nil {
		log.Fatal().Err(err).Msg("Failed to start AI generation service")
	}
}

// setupGenerationRouter configures all generation routes
func setupGenerationRouter(cfg *config.Config, genService *generation.Service) *gin.Engine {
	router := gin.Default()

	// Setup middleware
	router.Use(gin.Recovery())

	// Setup routes (auth-agnostic - trusts API Gateway)
	generation.RegisterRoutes(router, genService)

	return router
}
