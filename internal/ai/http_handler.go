// Package ai provides HTTP handler for AI operations
package ai

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/EliasRanz/ai-code-gen/internal/observability"
	"github.com/EliasRanz/ai-code-gen/internal/utilities"
)

// HTTPHandler handles HTTP requests for AI operations
type HTTPHandler struct {
	generateCodeService *GenerateCodeService
	streamCodeService   *StreamCodeService
	logger              observability.Logger
}

// NewHTTPHandler creates a new AI HTTP handler
func NewHTTPHandler(
	generateCodeService *GenerateCodeService,
	streamCodeService *StreamCodeService,
	logger observability.Logger,
) *HTTPHandler {
	return &HTTPHandler{
		generateCodeService: generateCodeService,
		streamCodeService:   streamCodeService,
		logger:              logger,
	}
}

// RegisterRoutes implements utilities.HTTPHandler interface
func (h *HTTPHandler) RegisterRoutes(router utilities.Router) error {
	if router == nil {
		return utilities.NewValidationError("router cannot be nil", nil)
	}

	// Register AI routes
	apiGroup := router.Group("/api")
	h.registerAIRoutes(apiGroup)

	return nil
}

// HealthCheck implements utilities.HTTPHandler interface
func (h *HTTPHandler) HealthCheck() error {
	// Check if all required dependencies are available
	if h.generateCodeService == nil || h.streamCodeService == nil {
		return utilities.NewValidationError("AI handler dependencies not properly initialized", nil)
	}
	return nil
}

// ValidateRoutes implements utilities.HTTPHandler interface
func (h *HTTPHandler) ValidateRoutes() error {
	// Validate that all required services are available
	if h.generateCodeService == nil {
		return utilities.NewValidationError("generate code service is required", nil)
	}
	if h.streamCodeService == nil {
		return utilities.NewValidationError("stream code service is required", nil)
	}
	return nil
}

// registerAIRoutes registers all AI-related routes
func (h *HTTPHandler) registerAIRoutes(rg utilities.RouterGroup) {
	aiGroup := rg.Group("/ai")
	aiGroup.POST("/generate", h.adaptHandlerFunc(h.GenerateCode))
	aiGroup.POST("/stream", h.adaptHandlerFunc(h.StreamCode))
}

// adaptHandlerFunc adapts gin handler to utilities.HandlerFunc
func (h *HTTPHandler) adaptHandlerFunc(ginHandler gin.HandlerFunc) utilities.HandlerFunc {
	return func(ctx utilities.Context) {
		// For now, we'll handle the adaptation in the router layer
		// This is a placeholder implementation
	}
}

// GenerateCode handles POST /ai/generate
func (h *HTTPHandler) GenerateCode(c *gin.Context) {
	var req GenerateCodeRequest
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
func (h *HTTPHandler) StreamCode(c *gin.Context) {
	var req StreamCodeRequest
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
func (h *HTTPHandler) handleError(c *gin.Context, err error) {
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
