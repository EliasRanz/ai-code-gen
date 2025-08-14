// Package http provides HTTP interface adapters
package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/EliasRanz/ai-code-gen/internal/ai"
	"github.com/EliasRanz/ai-code-gen/internal/observability"
	"github.com/EliasRanz/ai-code-gen/internal/utilities"
)

// AIHandler handles HTTP requests for AI operations
type AIHandler struct {
	generateCodeService *ai.GenerateCodeService
	streamCodeService   *ai.StreamCodeService
	logger              observability.Logger
}

// NewAIHandler creates a new AI handler
func NewAIHandler(
	generateCodeService *ai.GenerateCodeService,
	streamCodeService *ai.StreamCodeService,
	logger observability.Logger,
) *AIHandler {
	return &AIHandler{
		generateCodeService: generateCodeService,
		streamCodeService:   streamCodeService,
		logger:              logger,
	}
}

// GenerateCode handles POST /ai/generate
func (h *AIHandler) GenerateCode(c *gin.Context) {
	var req ai.GenerateCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid generate code request", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	resp, err := h.generateCodeService.Execute(c.Request.Context(), req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	h.logger.Info("Code generated successfully", map[string]interface{}{
		"prompt_length": len(req.Prompt),
		"response_id":   resp.ID,
	})

	c.JSON(http.StatusOK, resp)
}

// StreamCode handles POST /ai/stream
func (h *AIHandler) StreamCode(c *gin.Context) {
	var req ai.StreamCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid stream code request", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Set headers for Server-Sent Events
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// Start streaming
	responseChan, err := h.streamCodeService.Execute(c.Request.Context(), req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Send streaming responses
	for {
		select {
		case resp, ok := <-responseChan:
			if !ok {
				// Channel closed, streaming complete
				h.logger.Info("Code streaming completed", map[string]interface{}{
					"prompt_length": len(req.Prompt),
				})
				return
			}

			// Send SSE event
			c.SSEvent("data", resp)
			c.Writer.Flush()

			// If there's an error in the response, stop streaming
			if resp.Type == "error" {
				h.logger.Error("Code streaming failed", nil, map[string]interface{}{
					"prompt_length": len(req.Prompt),
					"error":         resp.Error,
				})
				return
			}

		case <-c.Request.Context().Done():
			// Client disconnected
			h.logger.Info("Client disconnected during streaming")
			return
		}
	}
}

// handleError handles different types of domain errors
func (h *AIHandler) handleError(c *gin.Context, err error) {
	h.logger.Error("AI request failed", err, map[string]interface{}{
		"path":   c.Request.URL.Path,
		"method": c.Request.Method,
	})

	if utilities.IsValidationError(err) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if utilities.IsNotFoundError(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Rate limiting or quota exceeded
	if err.Error() == "rate_limit_exceeded" || err.Error() == "quota_exceeded" {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
		return
	}

	// Default to internal server error
	c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
}
