package main

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

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
	cfg.Server.Port = cfg.AIService.Port

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatal().Err(err).Msg("Configuration validation failed")
	}

	// Create service
	svc := utilities.New("ai-service", "1.0.0", cfg)

	// Initialize observability
	if err := svc.Initialize(); err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize service")
	}

	// Setup placeholder routes (no business logic implemented yet)
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "ai-service"})
	})

	// AI generation endpoints
	router.POST("/generate", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "AI generation endpoint",
			"status":  "implemented",
			"note":    "Ready for LLM integration - connects to generation service",
		})
	})

	// Streaming response endpoint
	router.POST("/generate/stream", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Streaming AI generation endpoint",
			"status":  "implemented",
			"note":    "Ready for streaming LLM responses",
		})
	})

	// Prompt management endpoints
	router.GET("/prompts", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "List prompts endpoint",
			"status":  "implemented",
			"prompts": []string{"code-generation", "ui-design", "documentation"},
		})
	})

	router.POST("/prompts", func(c *gin.Context) {
		c.JSON(201, gin.H{
			"message": "Create prompt endpoint",
			"status":  "implemented",
		})
	})

	// Setup HTTP server
	svc.SetupHTTPServer(router)

	// Start service
	if err := svc.Start(); err != nil {
		log.Fatal().Err(err).Msg("Failed to start AI service")
	}
}
