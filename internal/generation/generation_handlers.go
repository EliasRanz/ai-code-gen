package generation

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

func RegisterRoutes(router *gin.Engine, service *Service) {
	// Generation group - auth is handled by API Gateway
	generationGroup := router.Group("/api/v1/generate")
	{
		generationGroup.POST("/stream", service.StreamGenerationHandler)
		generationGroup.POST("/request-response", service.RequestResponseHandler)
	}

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
}

// StreamGenerationHandler handles streaming AI generation requests
func (s *Service) StreamGenerationHandler(c *gin.Context) {
	// Extract user context set by API Gateway
	userID, _, _, authenticated := auth.GetUserContextFromMiddleware(c)
	if !authenticated {
		log.Error().Msg("No user context found - API Gateway authentication may have failed")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	var req struct {
		Model       string                 `json:"model" binding:"required"`
		Prompt      string                 `json:"prompt" binding:"required"`
		MaxTokens   int                    `json:"max_tokens"`
		Temperature float64                `json:"temperature"`
		UserID      string                 `json:"user_id"`
		ProjectID   string                 `json:"project_id"`
		Metadata    map[string]interface{} `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error().Err(err).Msg("Invalid request")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set user ID from context
	req.UserID = string(userID)

	// Create a stream for the user
	streamID := fmt.Sprintf("user-%s-stream-%d", string(userID), time.Now().UnixNano())

	log.Info().
		Str("model", req.Model).
		Str("user_id", req.UserID).
		Str("project_id", req.ProjectID).
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

	// Stream the response (simplified - single response)
	c.Writer.Write([]byte(fmt.Sprintf("data: %s\n\n", response.Content)))
	c.Writer.Flush()
}

// RequestResponseHandler handles non-streaming AI generation requests
func (s *Service) RequestResponseHandler(c *gin.Context) {
	// Extract user context set by API Gateway
	userID, _, _, authenticated := auth.GetUserContextFromMiddleware(c)
	if !authenticated {
		log.Error().Msg("No user context found - API Gateway authentication may have failed")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	var req struct {
		Model       string                 `json:"model" binding:"required"`
		Prompt      string                 `json:"prompt" binding:"required"`
		MaxTokens   int                    `json:"max_tokens"`
		Temperature float64                `json:"temperature"`
		UserID      string                 `json:"user_id"`
		ProjectID   string                 `json:"project_id"`
		Metadata    map[string]interface{} `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error().Err(err).Msg("Invalid request")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set user ID from context
	req.UserID = string(userID)

	// Generate response using AI service
	response, err := s.aiService.GenerateWithBuilder(c.Request.Context(), req.UserID, req.Prompt)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate response")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Generation failed"})
		return
	}

	// Convert to expected response format
	resp := map[string]interface{}{
		"content":     response.Content,
		"tokens_used": response.TokensUsed,
		"provider":    response.Provider,
		"model":       response.Model,
	}

	c.JSON(http.StatusOK, resp)
}

// GetModelsHandler returns available models
func (s *Service) GetModelsHandler(c *gin.Context) {
	providers := s.aiService.GetAvailableProviders()
	c.JSON(http.StatusOK, gin.H{"providers": providers})
}

// HealthHandler checks service health
func (s *Service) HealthHandler(c *gin.Context) {
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
func (s *Service) writeSSEEvent(c *gin.Context, event string, data interface{}, id string) {
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
func (s *Service) writeSSEError(c *gin.Context, errorCode, message string) {
	fmt.Fprintf(c.Writer, "event: error\n")
	fmt.Fprintf(c.Writer, "data: {\"error_code\":\"%s\",\"message\":\"%s\"}\n\n", errorCode, message)
}

// HealthCheckHandler handles health checks for the generation service
func (s *Service) HealthCheckHandler(c *gin.Context) {
	// Check AI service health
	aiHealthResults := s.aiService.HealthCheck(c.Request.Context())
	aiHealthy := len(aiHealthResults) == 0 || func() bool {
		for _, err := range aiHealthResults {
			if err != nil {
				return false
			}
		}
		return true
	}()

	// Check Redis client health
	redisHealthy := s.redisClient.Ping(c.Request.Context()) == nil

	if aiHealthy && redisHealthy {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"ai":     "healthy",
			"redis":  "healthy",
		})
	} else {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "error",
			"ai":     map[string]interface{}{"healthy": aiHealthy, "details": aiHealthResults},
			"redis":  map[string]interface{}{"healthy": redisHealthy},
		})
	}
}
