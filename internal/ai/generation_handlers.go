package ai

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/EliasRanz/ai-code-gen/internal/auth"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// RegisterGenerationRoutes registers AI generation routes
func RegisterGenerationRoutes(router *gin.Engine, service *GenerationService) {
	// Generation group - auth is handled by API Gateway
	generationGroup := router.Group("/api/v1/generate")
	{
		generationGroup.POST("/stream", service.StreamGenerationHandler)
		generationGroup.POST("/request-response", service.RequestResponseHandler)
		generationGroup.GET("/models", service.GetModelsHandler)
	}

	// Health check endpoint
	router.GET("/health", service.HealthHandler)
}

// StreamGenerationHandler handles streaming AI generation requests
func (s *GenerationService) StreamGenerationHandler(c *gin.Context) {
	// Extract user context set by API Gateway
	userID, _, _, authenticated := auth.GetUserContextFromMiddleware(c)
	if !authenticated {
		log.Error().Msg("No user context found - API Gateway authentication may have failed")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	var req GenerationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error().Err(err).Msg("Invalid request")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set user ID from context
	req.UserID = userID

	if err := req.Validate(); err != nil {
		log.Error().Err(err).Msg("Request validation failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create a stream for the user
	streamID := fmt.Sprintf("user-%s-stream-%d", string(userID), time.Now().UnixNano())

	log.Info().
		Str("model", req.Model).
		Str("user_id", string(req.UserID)).
		Str("project_id", func() string {
			if req.ProjectID != nil {
				return string(*req.ProjectID)
			}
			return ""
		}()).
		Int("prompt_length", len(req.Prompt)).
		Str("stream_id", streamID).
		Msg("Streaming generation request")

	// Create AI generation request using the builder pattern
	response, err := s.aiService.GenerateWithBuilder(c.Request.Context(), string(userID), req.Prompt)
	if err != nil {
		log.Error().Err(err).Msg("Failed to start stream generation")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start generation"})
		return
	}

	// Set headers for SSE
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// Stream the response
	s.writeSSEEvent(c, "generation", response, streamID)
	c.Writer.Flush()

	// Publish to Redis for real-time updates
	s.publishToRedis(response, string(req.UserID), func() string {
		if req.ProjectID != nil {
			return string(*req.ProjectID)
		}
		return ""
	}())
}

// RequestResponseHandler handles non-streaming AI generation requests
func (s *GenerationService) RequestResponseHandler(c *gin.Context) {
	// Extract user context set by API Gateway
	userID, _, _, authenticated := auth.GetUserContextFromMiddleware(c)
	if !authenticated {
		log.Error().Msg("No user context found - API Gateway authentication may have failed")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	var req GenerationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error().Err(err).Msg("Invalid request")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set user ID from context
	req.UserID = userID

	if err := req.Validate(); err != nil {
		log.Error().Err(err).Msg("Request validation failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate response using AI service
	response, err := s.aiService.GenerateWithBuilder(c.Request.Context(), string(req.UserID), req.Prompt)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate response")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Generation failed"})
		return
	}

	// Convert to expected response format
	resp := GenerationResponse{
		Content:    response.Content,
		TokensUsed: response.TokensUsed,
		Provider:   response.Provider,
		Model:      response.Model,
		UserID:     req.UserID,
		ProjectID:  req.ProjectID,
		Timestamp:  time.Now().UTC(),
	}

	// Publish to Redis for real-time updates
	s.publishToRedis(response, string(req.UserID), func() string {
		if req.ProjectID != nil {
			return string(*req.ProjectID)
		}
		return ""
	}())

	c.JSON(http.StatusOK, resp)
}

// GetModelsHandler returns available models
func (s *GenerationService) GetModelsHandler(c *gin.Context) {
	providers := s.aiService.GetAvailableProviders()
	c.JSON(http.StatusOK, gin.H{"providers": providers})
}

// HealthHandler checks service health
func (s *GenerationService) HealthHandler(c *gin.Context) {
	health := gin.H{
		"status":    "ok",
		"timestamp": time.Now().UTC(),
		"services":  gin.H{},
	}

	// Check AI service health
	healthResults := s.aiService.HealthCheck(c.Request.Context())
	if len(healthResults) > 0 {
		// Check if any providers are unhealthy
		hasUnhealthy := false
		for provider, err := range healthResults {
			if err != nil {
				log.Warn().Err(err).Str("provider", provider).Msg("AI provider health check failed")
				hasUnhealthy = true
			}
		}

		if hasUnhealthy {
			health["services"].(gin.H)["ai"] = gin.H{
				"status":    "degraded",
				"providers": healthResults,
			}
			health["status"] = "degraded"
		} else {
			health["services"].(gin.H)["ai"] = gin.H{"status": "healthy"}
		}
	} else {
		health["services"].(gin.H)["ai"] = gin.H{"status": "healthy"}
	}

	// Check Redis health
	if err := s.redisClient.Ping(c.Request.Context()); err != nil {
		log.Warn().Err(err).Msg("Redis health check failed")
		health["services"].(gin.H)["redis"] = gin.H{
			"status": "unhealthy",
			"error":  err.Error(),
		}
		if health["status"] != "degraded" {
			health["status"] = "degraded"
		}
	} else {
		health["services"].(gin.H)["redis"] = gin.H{"status": "healthy"}
	}

	statusCode := http.StatusOK
	if health["status"] != "ok" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, health)
}

// writeSSEEvent writes a Server-Sent Event
func (s *GenerationService) writeSSEEvent(c *gin.Context, event string, data interface{}, id string) {
	if id != "" {
		fmt.Fprintf(c.Writer, "id: %s\n", id)
	}
	fmt.Fprintf(c.Writer, "event: %s\n", event)

	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal SSE data")
		s.writeSSEError(c, "marshal_error", "Failed to encode response")
		return
	}

	// Split data into lines for proper SSE format
	lines := strings.Split(string(jsonData), "\n")
	for _, line := range lines {
		fmt.Fprintf(c.Writer, "data: %s\n", line)
	}
	fmt.Fprintf(c.Writer, "\n")
}

// writeSSEError writes an error event in SSE format
func (s *GenerationService) writeSSEError(c *gin.Context, errorCode, message string) {
	fmt.Fprintf(c.Writer, "event: error\n")
	fmt.Fprintf(c.Writer, "data: {\"error_code\":\"%s\",\"message\":\"%s\"}\n\n", errorCode, message)
}
