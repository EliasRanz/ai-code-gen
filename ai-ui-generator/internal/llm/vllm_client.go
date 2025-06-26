package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
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

// GetModels returns available models
func (c *VLLMClient) GetModels(ctx context.Context) ([]Model, error) {
	log.Info().Msg("VLLM GetModels request")

	if c.baseURL == "" {
		// Fallback to stubbed models if no baseURL configured
		return c.getStubModels(), nil
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/v1/models", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	if c.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	// Execute request
	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to get models from VLLM, falling back to stubs")
		return c.getStubModels(), nil
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		log.Warn().Int("status", httpResp.StatusCode).Msg("VLLM models API returned error, falling back to stubs")
		return c.getStubModels(), nil
	}

	// Parse response
	var response struct {
		Data []struct {
			ID      string `json:"id"`
			Object  string `json:"object"`
			Created int64  `json:"created"`
		} `json:"data"`
	}

	if err := json.NewDecoder(httpResp.Body).Decode(&response); err != nil {
		log.Warn().Err(err).Msg("Failed to parse models response, falling back to stubs")
		return c.getStubModels(), nil
	}

	// Convert to standard format
	models := make([]Model, len(response.Data))
	for i, model := range response.Data {
		models[i] = Model{
			ID:          model.ID,
			Name:        model.ID, // VLLM doesn't provide separate name
			Description: fmt.Sprintf("Model %s served by VLLM", model.ID),
			Provider:    "vllm",
			MaxTokens:   4096, // Default assumption
			CreatedAt:   time.Unix(model.Created, 0),
		}
	}

	return models, nil
}

// Health checks if the VLLM service is healthy
func (c *VLLMClient) Health(ctx context.Context) error {
	log.Info().Msg("VLLM Health check")

	if c.baseURL == "" {
		// Consider healthy if no baseURL configured (stub mode)
		return nil
	}

	// Create health check request
	httpReq, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/health", nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	// Set headers if needed
	if c.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	// Execute request with shorter timeout for health check
	client := &http.Client{Timeout: 5 * time.Second}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("health check request failed: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return fmt.Errorf("VLLM health check failed with status %d", httpResp.StatusCode)
	}

	return nil
}

// Close closes the client connection
func (c *VLLMClient) Close() error {
	log.Info().Msg("VLLM client closed")

	// Close the HTTP client's idle connections
	if c.httpClient != nil {
		c.httpClient.CloseIdleConnections()
	}

	return nil
}

// Helper functions

func generateResponseID() string {
	return fmt.Sprintf("vllm-resp-%d", time.Now().UnixNano())
}

func generateStubbedResponse(prompt string) string {
	// Generate a simple stubbed response based on the prompt
	responses := []string{
		"This is a stubbed response from the VLLM client. The actual implementation will call the real VLLM API.",
		"Here's a generated response that demonstrates the streaming functionality of the AI Generation Service.",
		"The VLLM client is currently stubbed and will be implemented to call the actual VLLM inference server.",
		"This response shows how the AI generation system will work once fully integrated with real LLM services.",
	}

	// Simple hash to make response somewhat deterministic
	hash := 0
	for _, char := range prompt {
		hash += int(char)
	}

	return responses[hash%len(responses)]
}

func estimateTokens(text string) int {
	// Rough estimation: ~4 characters per token
	return len(text) / 4
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func stringPtr(s string) *string {
	return &s
}

// generateStubResponse creates a fallback response when VLLM is not configured
func (c *VLLMClient) generateStubResponse(req *GenerationRequest) *GenerationResponse {
	// Simulate some processing time
	time.Sleep(100 * time.Millisecond)

	return &GenerationResponse{
		ID:     generateResponseID(),
		Object: "text_completion",
		Model:  req.Model,
		Choices: []Choice{
			{
				Index:        0,
				Text:         generateStubbedResponse(req.Prompt),
				FinishReason: stringPtr("stop"),
			},
		},
		Usage: &Usage{
			PromptTokens:     estimateTokens(req.Prompt),
			CompletionTokens: 50, // Stubbed
			TotalTokens:      estimateTokens(req.Prompt) + 50,
		},
	}
}

// generateStubStreamResponse creates a fallback streaming response when VLLM is not configured
func (c *VLLMClient) generateStubStreamResponse(ctx context.Context, req *GenerationRequest) <-chan *GenerationResponse {
	responseChan := make(chan *GenerationResponse, 10)

	go func() {
		defer close(responseChan)

		stubbedText := generateStubbedResponse(req.Prompt)
		words := strings.Fields(stubbedText)

		for i, word := range words {
			select {
			case <-ctx.Done():
				return
			case <-time.After(50 * time.Millisecond): // Simulate streaming delay
			}

			response := &GenerationResponse{
				ID:     generateResponseID(),
				Object: "text_completion.chunk",
				Model:  req.Model,
				Choices: []Choice{
					{
						Index: 0,
						Delta: &Delta{
							Content: word + " ",
						},
					},
				},
			}

			// Send the last chunk with finish reason
			if i == len(words)-1 {
				response.Choices[0].FinishReason = stringPtr("stop")
				response.Usage = &Usage{
					PromptTokens:     estimateTokens(req.Prompt),
					CompletionTokens: len(words),
					TotalTokens:      estimateTokens(req.Prompt) + len(words),
				}
			}

			select {
			case responseChan <- response:
			case <-ctx.Done():
				return
			}
		}
	}()

	return responseChan
}

// processStreamResponse processes the streaming HTTP response from VLLM
func (c *VLLMClient) processStreamResponse(ctx context.Context, body io.ReadCloser, responseChan chan<- *GenerationResponse) {
	defer body.Close()
	defer close(responseChan)

	scanner := bufio.NewScanner(body)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
		}

		line := scanner.Text()
		if line == "" || !strings.HasPrefix(line, "data: ") {
			continue
		}

		// Extract JSON data from SSE format
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			return
		}

		// Parse the streaming response
		var vllmResp vllmResponse
		if err := json.Unmarshal([]byte(data), &vllmResp); err != nil {
			log.Error().Err(err).Str("data", data).Msg("Failed to parse streaming response")
			continue
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

		select {
		case responseChan <- response:
		case <-ctx.Done():
			return
		}
	}

	if err := scanner.Err(); err != nil {
		log.Error().Err(err).Msg("Error reading streaming response")
	}
}

// getStubModels returns stub models when VLLM is not available
func (c *VLLMClient) getStubModels() []Model {
	return []Model{
		{
			ID:          "llama-2-7b-chat",
			Name:        "Llama 2 7B Chat",
			Description: "Meta's Llama 2 7B parameter chat model",
			Provider:    "vllm",
			MaxTokens:   4096,
			CreatedAt:   time.Now().Add(-24 * time.Hour),
		},
		{
			ID:          "llama-2-13b-chat",
			Name:        "Llama 2 13B Chat",
			Description: "Meta's Llama 2 13B parameter chat model",
			Provider:    "vllm",
			MaxTokens:   4096,
			CreatedAt:   time.Now().Add(-24 * time.Hour),
		},
		{
			ID:          "code-llama-7b-instruct",
			Name:        "Code Llama 7B Instruct",
			Description: "Meta's Code Llama 7B parameter instruction model",
			Provider:    "vllm",
			MaxTokens:   4096,
			CreatedAt:   time.Now().Add(-24 * time.Hour),
		},
	}
}

/*
TODO: LLM Implementation Roadmap - Replace Stubbed Implementations

PRIORITY 1: Core LLM Provider Integration
1. OpenAI Integration
   - Add OpenAI Go SDK: go get github.com/sashabaranov/go-openai
   - Implement OpenAI client in internal/llm/openai_client.go
   - Support GPT-4, GPT-3.5-turbo, and code models
   - Handle OpenAI API keys and rate limiting

2. Anthropic Claude Integration  
   - Add Anthropic SDK or HTTP client
   - Implement Claude client in internal/llm/claude_client.go
   - Support Claude-3, Claude-2 models
   - Handle Anthropic API authentication

3. Local LLM Integration
   - Ollama integration for local models
   - Add Ollama Go client: go get github.com/ollama/ollama/api
   - Support local Llama, CodeLlama, Mistral models
   - Implement in internal/llm/ollama_client.go

PRIORITY 2: Production VLLM Integration
4. Real VLLM Server Setup
   - Document VLLM server deployment (Docker/Kubernetes)
   - Add authentication and security for VLLM endpoints
   - Implement proper model loading and management
   - Add monitoring and health checks for VLLM instances

5. Multi-Provider Support
   - Implement LLM provider factory pattern
   - Add provider selection logic based on model requirements
   - Implement load balancing across providers
   - Add fallback mechanisms between providers

PRIORITY 3: Advanced Features
6. Model Management
   - Implement dynamic model discovery
   - Add model capabilities metadata (context length, pricing, etc.)
   - Implement model selection algorithms based on task requirements
   - Add model performance monitoring

7. Cost and Usage Tracking
   - Implement token counting and cost calculation
   - Add usage analytics and reporting
   - Implement budget controls and alerts
   - Add user quota management per provider

8. Streaming and Performance
   - Optimize streaming performance across all providers
   - Implement connection pooling for HTTP clients
   - Add caching for model metadata and frequent responses
   - Implement request batching where supported

FILES TO CREATE/MODIFY:
- internal/llm/openai_client.go (NEW)
- internal/llm/claude_client.go (NEW) 
- internal/llm/ollama_client.go (NEW)
- internal/llm/factory.go (NEW)
- internal/llm/provider_config.go (NEW)
- internal/config/llm_config.go (UPDATE)
- internal/generation/service.go (UPDATE - provider selection)
- cmd/server/main.go (UPDATE - configuration)

CONFIGURATION UPDATES NEEDED:
- Add provider-specific configuration sections
- Add API key management (preferably via secrets/env vars)
- Add model selection and routing configuration
- Add cost and usage tracking configuration
*/
