package ai

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GenerateRequest represents a request to generate AI code
type GenerateRequest struct {
	Prompt      string   `json:"prompt" binding:"required"`
	UserID      string   `json:"user_id,omitempty"`
	Model       string   `json:"model,omitempty"`
	Temperature *float64 `json:"temperature,omitempty"`
	MaxTokens   *int     `json:"max_tokens,omitempty"`
}

// GenerateResponse represents a response from AI code generation
type GenerateResponse struct {
	Code    string `json:"code"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// ValidateRequest represents a request to validate code
type ValidateRequest struct {
	Code string `json:"code" binding:"required"`
}

// ValidateResponse represents a response from code validation
type ValidateResponse struct {
	Valid   bool     `json:"valid"`
	Errors  []string `json:"errors,omitempty"`
	Message string   `json:"message,omitempty"`
	Error   string   `json:"error,omitempty"`
}

// Handler handles AI-related HTTP requests
type Handler struct {
	service *Service
}

// NewHandler creates a new AI handler
func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// Generate handles AI code generation requests
func (h *Handler) Generate(c *gin.Context) {
	var req GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, GenerateResponse{
			Error: "invalid request",
		})
		return
	}

	// Use userID from context if available, otherwise from request
	userID := req.UserID
	if contextUserID, exists := c.Get("user_id"); exists {
		if uid, ok := contextUserID.(string); ok {
			userID = uid
		}
	}

	var code string
	var err error

	// Check if model parameters are provided
	if req.Model != "" || req.Temperature != nil || req.MaxTokens != nil {
		params := GenerationParams{
			Model:       req.Model,
			Temperature: req.Temperature,
			MaxTokens:   req.MaxTokens,
		}
		code, err = h.service.GenerateCodeWithParams(req.Prompt, userID, params)
	} else {
		code, err = h.service.GenerateCode(req.Prompt, userID)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, GenerateResponse{
			Error: "Failed to generate code: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, GenerateResponse{
		Code:    code,
		Message: "Code generated successfully",
	})
}

// Stream handles streaming AI code generation
func (h *Handler) Stream(c *gin.Context) {
	_ = c.Param("sessionId") // sessionId parameter available but not used in mock
	prompt := c.Query("prompt")

	if prompt == "" {
		c.JSON(http.StatusBadRequest, GenerateResponse{
			Error: "prompt required",
		})
		return
	}

	userID := c.Query("user_id")
	if contextUserID, exists := c.Get("user_id"); exists {
		if uid, ok := contextUserID.(string); ok {
			userID = uid
		}
	}

	// Parse optional model parameters
	model := c.Query("model")
	var temperature *float64
	var maxTokens *int

	if tempStr := c.Query("temperature"); tempStr != "" {
		if temp, err := strconv.ParseFloat(tempStr, 64); err == nil {
			temperature = &temp
		}
	}

	if tokenStr := c.Query("max_tokens"); tokenStr != "" {
		if tokens, err := strconv.Atoi(tokenStr); err == nil {
			maxTokens = &tokens
		}
	}

	// Set proper headers for streaming
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	responseChannel := make(chan string, 10)

	params := GenerationParams{
		Model:       model,
		Temperature: temperature,
		MaxTokens:   maxTokens,
	}

	// Start streaming in a goroutine
	go func() {
		defer close(responseChannel)
		err := h.service.StreamGenerationWithParams(prompt, userID, params, responseChannel)
		if err != nil {
			responseChannel <- "error: " + err.Error()
		}
	}()

	// Stream the response
	for chunk := range responseChannel {
		if chunk != "" {
			c.Writer.WriteString("data: " + chunk + "\n\n")
			c.Writer.Flush()
		}
	}
}

// ValidateCode handles code validation requests
func (h *Handler) ValidateCode(c *gin.Context) {
	var req ValidateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ValidateResponse{
			Error: "invalid request",
		})
		return
	}

	valid, errors, err := h.service.ValidateGeneratedCode(req.Code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ValidateResponse{
			Error: "Failed to validate code: " + err.Error(),
		})
		return
	}

	response := ValidateResponse{
		Valid:   valid,
		Errors:  errors,
		Message: "Code validation completed",
	}

	c.JSON(http.StatusOK, response)
}

// GetQuota handles quota checking requests
func (h *Handler) GetQuota(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user_id required",
		})
		return
	}

	// Return mock quota data for testing
	quota := gin.H{
		"user_id":     userID,
		"daily_limit": 1000,
		"used":        450,
		"remaining":   550,
	}

	c.JSON(http.StatusOK, quota)
}

// History handles request history retrieval
func (h *Handler) History(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user_id required",
		})
		return
	}

	history := h.service.GetHistory(userID)
	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"history": history,
	})
}

// RegisterRoutes registers all AI-related routes
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	ai := r.Group("/ai")
	ai.POST("/generate", h.Generate)
	ai.GET("/stream/:sessionId", h.Stream)
	ai.POST("/validate", h.ValidateCode)
	ai.GET("/quota", h.GetQuota)
	ai.GET("/history", h.History)
}
