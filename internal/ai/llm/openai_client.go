package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// OpenAIClient implements LLMProvider using OpenAI API
type OpenAIClient struct {
	config     ProviderConfig
	httpClient *http.Client
	info       ProviderInfo
	limits     ProviderLimits
}

// NewOpenAIClient creates a new OpenAI client with free tier configuration
func NewOpenAIClient(config ProviderConfig) (*OpenAIClient, error) {
	if config.FreeTierOnly && config.APIKey != "" {
		return nil, NewLLMError("openai", "PAID_API_KEY", "Free tier mode enabled but API key provided", "Remove API key for free tier testing")
	}

	if config.BaseURL == "" {
		config.BaseURL = "https://api.openai.com/v1"
	}
	if config.Model == "" {
		config.Model = "gpt-3.5-turbo"
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	client := &OpenAIClient{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		info: ProviderInfo{
			Name:         "OpenAI",
			Version:      "v1",
			Models:       []string{"gpt-3.5-turbo", "gpt-4"},
			Capabilities: []string{"code_generation", "chat", "completion"},
			FreeTier:     config.FreeTierOnly,
		},
		limits: ProviderLimits{
			RequestsPerMinute:   3,   // Free tier limit
			TokensPerMinute:     200, // Free tier limit
			DailyQuota:          100, // Free tier limit
			MaxTokensPerRequest: 2000,
		},
	}

	return client, nil
}

// GenerateCode implements LLMProvider interface
func (c *OpenAIClient) GenerateCode(ctx context.Context, req *GenerationRequest) (*GenerationResponse, error) {
	if c.config.FreeTierOnly {
		return c.generateMockResponse(req)
	}

	startTime := time.Now()

	openAIReq := openAIRequest{
		Model: c.config.Model,
		Messages: []message{
			{Role: "user", Content: req.Prompt},
		},
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		Stream:      false,
	}

	resp, err := c.makeRequest(ctx, openAIReq)
	if err != nil {
		return nil, NewLLMError("openai", "REQUEST_FAILED", "Failed to make API request", err.Error())
	}
	defer resp.Body.Close()

	var openAIResp openAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return nil, NewLLMError("openai", "DECODE_FAILED", "Failed to decode response", err.Error())
	}

	if len(openAIResp.Choices) == 0 {
		return nil, NewLLMError("openai", "NO_CHOICES", "No choices in response", "")
	}

	return &GenerationResponse{
		Content:      openAIResp.Choices[0].Message.Content,
		TokensUsed:   openAIResp.Usage.TotalTokens,
		Provider:     "openai",
		Model:        openAIResp.Model,
		Latency:      time.Since(startTime),
		RequestID:    req.Metadata["request_id"],
		FinishReason: "stop",
		Metadata:     req.Metadata,
	}, nil
}

// HealthCheck implements LLMProvider interface
func (c *OpenAIClient) HealthCheck(ctx context.Context) error {
	if c.config.FreeTierOnly {
		return nil // Always healthy in free tier mode
	}

	req := openAIRequest{
		Model: c.config.Model,
		Messages: []message{
			{Role: "user", Content: "ping"},
		},
		MaxTokens: 1,
	}

	resp, err := c.makeRequest(ctx, req)
	if err != nil {
		return NewLLMError("openai", "HEALTH_CHECK_FAILED", "Health check failed", err.Error())
	}
	resp.Body.Close()

	return nil
}

// GetProviderInfo implements LLMProvider interface
func (c *OpenAIClient) GetProviderInfo() ProviderInfo {
	return c.info
}

// GetLimits implements LLMProvider interface
func (c *OpenAIClient) GetLimits() ProviderLimits {
	return c.limits
}

// Close implements LLMProvider interface
func (c *OpenAIClient) Close() error {
	return nil // HTTP client doesn't require explicit closing
}

// generateMockResponse creates a mock response for free tier testing
func (c *OpenAIClient) generateMockResponse(req *GenerationRequest) (*GenerationResponse, error) {
	// Generate simple mock based on language
	var content string
	switch req.Language {
	case "go":
		content = `package main

import "fmt"

func main() {
    fmt.Println("Hello from AI-generated Go code!")
}`
	case "javascript":
		content = `function helloWorld() {
    console.log("Hello from AI-generated JavaScript code!");
}

helloWorld();`
	case "python":
		content = `def hello_world():
    print("Hello from AI-generated Python code!")

hello_world()`
	default:
		content = fmt.Sprintf("// AI-generated code for: %s\n// This is a mock response for free tier testing", req.Prompt)
	}

	return &GenerationResponse{
		Content:      content,
		TokensUsed:   len(content) / 4, // Rough token estimation
		Provider:     "openai",
		Model:        "mock-gpt-3.5-turbo",
		Latency:      100 * time.Millisecond,
		RequestID:    req.Metadata["request_id"],
		FinishReason: "stop",
		Metadata:     req.Metadata,
	}, nil
}

// makeRequest creates and sends HTTP request to OpenAI API
func (c *OpenAIClient) makeRequest(ctx context.Context, reqBody openAIRequest) (*http.Response, error) {
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.config.BaseURL+"/chat/completions", strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return resp, nil
}

// OpenAI API types
type openAIRequest struct {
	Model       string    `json:"model"`
	Messages    []message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []choice `json:"choices"`
	Usage   usage    `json:"usage"`
}

type choice struct {
	Index        int     `json:"index"`
	Message      message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

type usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}
