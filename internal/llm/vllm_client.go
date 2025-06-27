package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

// VLLMClient implements LLMClient for VLLM (Volunteer-run LLM) service
type VLLMClient struct {
	baseURL    string
	apiKey     string
	timeout    time.Duration
	maxRetries int
	httpClient *http.Client
}

// vllmRequest represents a request to the VLLM API
type vllmRequest struct {
	Model       string  `json:"model"`
	Prompt      string  `json:"prompt"`
	MaxTokens   int     `json:"max_tokens,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
	Stream      bool    `json:"stream,omitempty"`
}

// vllmResponse represents a response from the VLLM API
type vllmResponse struct {
	ID      string       `json:"id"`
	Object  string       `json:"object"`
	Model   string       `json:"model"`
	Choices []vllmChoice `json:"choices"`
	Usage   *vllmUsage   `json:"usage,omitempty"`
}

// vllmChoice represents a choice in the VLLM response
type vllmChoice struct {
	Index        int     `json:"index"`
	Text         string  `json:"text,omitempty"`
	Delta        *Delta  `json:"delta,omitempty"`
	FinishReason *string `json:"finish_reason,omitempty"`
}

// vllmUsage represents usage statistics from VLLM
type vllmUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// VLLMConfig holds configuration for VLLM client
type VLLMConfig struct {
	BaseURL    string        `json:"base_url"`
	APIKey     string        `json:"api_key"`
	Timeout    time.Duration `json:"timeout"`
	MaxRetries int           `json:"max_retries"`
}

// NewVLLMClient creates a new VLLM client
func NewVLLMClient(config *VLLMConfig) *VLLMClient {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}

	return &VLLMClient{
		baseURL:    config.BaseURL,
		apiKey:     config.APIKey,
		timeout:    config.Timeout,
		maxRetries: config.MaxRetries,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// Generate performs a single generation request
func (c *VLLMClient) Generate(ctx context.Context, req *GenerationRequest) (*GenerationResponse, error) {
	log.Info().
		Str("model", req.Model).
		Str("prompt_preview", truncateString(req.Prompt, 100)).
		Int("max_tokens", req.MaxTokens).
		Msg("VLLM Generate request")

	if c.baseURL == "" {
		// Fallback to stubbed response if no baseURL configured
		return c.generateStubResponse(req), nil
	}

	// Create VLLM API request
	vllmReq := &vllmRequest{
		Model:       req.Model,
		Prompt:      req.Prompt,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		Stream:      false,
	}

	// Marshal request
	reqBody, err := json.Marshal(vllmReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/v1/completions", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	// Execute request with retries
	var httpResp *http.Response
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		httpResp, err = c.httpClient.Do(httpReq)
		if err == nil && httpResp.StatusCode < 500 {
			break
		}

		if attempt < c.maxRetries {
			waitTime := time.Duration(attempt+1) * time.Second
			log.Warn().
				Err(err).
				Int("attempt", attempt+1).
				Dur("wait", waitTime).
				Msg("Request failed, retrying")

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(waitTime):
			}
		}
	}

	if err != nil {
		return nil, fmt.Errorf("request failed after %d attempts: %w", c.maxRetries+1, err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(httpResp.Body)
		return nil, fmt.Errorf("VLLM API error %d: %s", httpResp.StatusCode, string(body))
	}

	// Parse response
	var vllmResp vllmResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&vllmResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert to standard response format
	response := &GenerationResponse{
		ID:     vllmResp.ID,
		Object: vllmResp.Object,
		Model:  vllmResp.Model,
		Choices: make([]Choice, len(vllmResp.Choices)),
	}

	for i, choice := range vllmResp.Choices {
		response.Choices[i] = Choice(choice)
	}

	if vllmResp.Usage != nil {
		response.Usage = &Usage{
			PromptTokens:     vllmResp.Usage.PromptTokens,
			CompletionTokens: vllmResp.Usage.CompletionTokens,
			TotalTokens:      vllmResp.Usage.TotalTokens,
		}
	}

	return response, nil
}

// GenerateStream performs a streaming generation request
func (c *VLLMClient) GenerateStream(ctx context.Context, req *GenerationRequest) (<-chan *GenerationResponse, error) {
	log.Info().
		Str("model", req.Model).
		Str("prompt_preview", truncateString(req.Prompt, 100)).
		Msg("VLLM GenerateStream request")

	if c.baseURL == "" {
		// Fallback to stubbed streaming response if no baseURL configured
		return c.generateStubStreamResponse(ctx, req), nil
	}

	// Create VLLM API request for streaming
	vllmReq := &vllmRequest{
		Model:       req.Model,
		Prompt:      req.Prompt,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		Stream:      true,
	}

	// Marshal request
	reqBody, err := json.Marshal(vllmReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/v1/completions", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers for streaming
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")
	httpReq.Header.Set("Cache-Control", "no-cache")
	if c.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	// Execute request
	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		httpResp.Body.Close()
		body, _ := io.ReadAll(httpResp.Body)
		return nil, fmt.Errorf("VLLM API error %d: %s", httpResp.StatusCode, string(body))
	}

	// Create response channel
	responseChan := make(chan *GenerationResponse, 10)

	// Start streaming response processor
	go c.processStreamResponse(ctx, httpResp.Body, responseChan)

	return responseChan, nil
}

// Helper functions for VLLM client operations

// truncateString truncates a string to maxLen with ellipsis
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// generateStubResponse creates a fallback response when VLLM is not configured
func (c *VLLMClient) generateStubResponse(req *GenerationRequest) *GenerationResponse {
	// TODO: Move to vllm_helpers.go
	return &GenerationResponse{
		ID:     "stub-response",
		Object: "text_completion",
		Model:  req.Model,
		Choices: []Choice{
			{
				Index:        0,
				Text:         "This is a stubbed response.",
				FinishReason: stringPtr("stop"),
			},
		},
	}
}

// generateStubStreamResponse creates a fallback streaming response when VLLM is not configured
func (c *VLLMClient) generateStubStreamResponse(ctx context.Context, req *GenerationRequest) <-chan *GenerationResponse {
	// TODO: Move to vllm_helpers.go
	responseChan := make(chan *GenerationResponse, 1)
	go func() {
		defer close(responseChan)
		responseChan <- c.generateStubResponse(req)
	}()
	return responseChan
}

// processStreamResponse processes the streaming HTTP response from VLLM
func (c *VLLMClient) processStreamResponse(ctx context.Context, body io.ReadCloser, responseChan chan<- *GenerationResponse) {
	// TODO: Move to vllm_helpers.go
	defer body.Close()
	defer close(responseChan)
	// Stub implementation for now
}

// stringPtr returns a pointer to a string
func stringPtr(s string) *string {
	return &s
}

// Close closes the VLLMClient connection
func (c *VLLMClient) Close() error {
	// VLLM client uses HTTP client which doesn't need explicit closing
	// The HTTP client will be garbage collected
	return nil
}

// GetModels returns available models from the VLLM service
func (c *VLLMClient) GetModels(ctx context.Context) ([]Model, error) {
	// TODO: Implement actual models endpoint call
	// For now, return a default model
	models := []Model{
		{
			ID:          "default",
			Name:        "Default VLLM Model",
			Description: "Default model served by VLLM",
			Provider:    "VLLM",
			MaxTokens:   4096,
			CreatedAt:   time.Now(),
		},
	}
	return models, nil
}

// Health checks if the VLLM service is healthy
func (c *VLLMClient) Health(ctx context.Context) error {
	// TODO: Implement actual health check endpoint
	// For now, assume the service is healthy
	return nil
}
