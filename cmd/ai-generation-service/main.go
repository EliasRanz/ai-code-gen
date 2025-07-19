package main

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/EliasRanz/ai-code-gen/internal/ai"
	"github.com/EliasRanz/ai-code-gen/internal/cache"
	"github.com/EliasRanz/ai-code-gen/internal/config"
	"github.com/EliasRanz/ai-code-gen/internal/utilities"
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
	svc := utilities.New("ai-generation-service", "1.0.0", cfg)

	// Initialize observability
	if err := svc.Initialize(); err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize service")
	}

	// Initialize generation service (auth-agnostic - trusts API Gateway)
	redisConfig := &ai.RedisConfig{
		Host:     cfg.Redis.Host,
		Port:     cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}
	// Initialize AI service components with simple configuration
	rateLimiter := ai.NewRateLimiter(10, 5) // 10 requests per second, burst of 5
	quotaManager := ai.NewQuotaManager()

	// Create a basic cache provider for testing
	cacheConfig := cache.CacheConfig{
		Host:                   "localhost",
		Port:                   6379,
		DefaultTTL:             300 * time.Second, // 5 minutes
		MaxConnections:         10,
		MaxIdleConnections:     5,
		ConnectionTimeout:      30 * time.Second,
		FailureThreshold:       5,
		RequestVolumeThreshold: 10,
		RecoveryTimeout:        60 * time.Second,
	}
	cacheProvider, err := cache.NewMemoryProvider(cacheConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create cache provider")
	}

	aiCacheConfig := ai.DefaultCacheConfig()
	cacheManager := ai.NewCacheManager(cacheProvider, aiCacheConfig)

	// Create AI service
	aiService := ai.NewAIService(rateLimiter, quotaManager, cacheManager)

	// Initialize Redis client and generation service
	redisClient := ai.NewRedisClient(redisConfig)
	genService := ai.NewGenerationService(aiService, redisClient)

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
func setupGenerationRouter(cfg *config.Config, genService *ai.GenerationService) *gin.Engine {
	router := gin.Default()

	// Setup middleware
	router.Use(gin.Recovery())

	// Setup routes (auth-agnostic - trusts API Gateway)
	ai.RegisterGenerationRoutes(router, genService)

	return router
}
