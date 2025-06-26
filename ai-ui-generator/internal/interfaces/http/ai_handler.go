// Package http provides HTTP interface adapters
package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/EliasRanz/ai-code-gen/ai-ui-generator/internal/application/ai"
	"github.com/EliasRanz/ai-code-gen/ai-ui-generator/internal/domain/common"
	"github.com/EliasRanz/ai-code-gen/ai-ui-generator/internal/infrastructure/observability"
)

// AIHandler handles HTTP requests for AI operations
type AIHandler struct {
	generateCodeUC *ai.GenerateCodeUseCase
	streamCodeUC   *ai.StreamCodeUseCase
	logger         observability.Logger
}

// NewAIHandler creates a new AI handler
func NewAIHandler(
	generateCodeUC *ai.GenerateCodeUseCase,
	streamCodeUC *ai.StreamCodeUseCase,
	logger observability.Logger,
) *AIHandler {
	return &AIHandler{
		generateCodeUC: generateCodeUC,
		streamCodeUC:   streamCodeUC,
		logger:         logger,
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

	resp, err := h.generateCodeUC.Execute(c.Request.Context(), req)
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

	// Create a channel to receive streaming responses
	responseChan := make(chan ai.StreamCodeResponse, 10)
	errorChan := make(chan error, 1)

	// Start streaming in a goroutine
	go func() {
		defer close(responseChan)
		defer close(errorChan)

		err := h.streamCodeUC.Execute(c.Request.Context(), req, responseChan)
		if err != nil {
			errorChan <- err
		}
	}()

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

		case err := <-errorChan:
			if err != nil {
				h.logger.Error("Code streaming failed", err, map[string]interface{}{
					"prompt_length": len(req.Prompt),
				})
				c.SSEvent("error", gin.H{"error": err.Error()})
				c.Writer.Flush()
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

	if common.IsValidationError(err) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if common.IsNotFoundError(err) {
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
