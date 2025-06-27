package generation

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/EliasRanz/ai-code-gen/internal/llm"
)

// StreamGenerationHandler handles streaming AI generation requests
func (s *Service) StreamGenerationHandler(c *gin.Context) {
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

	log.Info().
		Str("model", req.Model).
		Str("user_id", req.UserID).
		Str("project_id", req.ProjectID).
		Int("prompt_length", len(req.Prompt)).
		Msg("Streaming generation request")

	// Create generation request
	genReq := &llm.GenerationRequest{
		Model:       req.Model,
		Prompt:      req.Prompt,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		Metadata:    req.Metadata,
	}

	// Start streaming
	respChan, err := s.llmClient.GenerateStream(c.Request.Context(), genReq)
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

	// Stream responses
	s.streamResponse(c, respChan, req.UserID, req.ProjectID)
}

// NonStreamGenerationHandler handles non-streaming AI generation requests
func (s *Service) NonStreamGenerationHandler(c *gin.Context) {
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

	log.Info().
		Str("model", req.Model).
		Str("user_id", req.UserID).
		Str("project_id", req.ProjectID).
		Int("prompt_length", len(req.Prompt)).
		Msg("Non-streaming generation request")

	// Create generation request
	genReq := &llm.GenerationRequest{
		Model:       req.Model,
		Prompt:      req.Prompt,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		Metadata:    req.Metadata,
	}

	// Generate response
	resp, err := s.llmClient.Generate(c.Request.Context(), genReq)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate response")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate response"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetModelsHandler returns available models
func (s *Service) GetModelsHandler(c *gin.Context) {
	models, err := s.llmClient.GetModels(c.Request.Context())
	if err != nil {
		log.Error().Err(err).Msg("Failed to get models")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get models"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"models": models})
}

// HealthHandler checks service health
func (s *Service) HealthHandler(c *gin.Context) {
	health := gin.H{
		"status":    "ok",
		"timestamp": time.Now().UTC(),
		"services":  gin.H{},
	}

	// Check LLM client health
	if err := s.llmClient.Health(c.Request.Context()); err != nil {
		log.Warn().Err(err).Msg("LLM client health check failed")
		health["services"].(gin.H)["llm"] = gin.H{
			"status": "unhealthy",
			"error":  err.Error(),
		}
		health["status"] = "degraded"
	} else {
		health["services"].(gin.H)["llm"] = gin.H{"status": "healthy"}
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

// streamResponse handles streaming responses to client
func (s *Service) streamResponse(c *gin.Context, respChan <-chan *llm.GenerationResponse, userID, projectID string) {
	ctx := c.Request.Context()
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		log.Error().Msg("Streaming not supported")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Streaming not supported"})
		return
	}

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Client disconnected")
			return
		case resp, ok := <-respChan:
			if !ok {
				// Channel closed, send final event
				s.writeSSEEvent(c, "done", gin.H{"message": "Generation complete"}, "")
				flusher.Flush()
				return
			}

			// Send response
			s.writeSSEEvent(c, "data", resp, resp.ID)
			flusher.Flush()

			// Publish to Redis if configured
			if userID != "" || projectID != "" {
				s.publishToRedis(resp, userID, projectID)
			}
		}
	}
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
