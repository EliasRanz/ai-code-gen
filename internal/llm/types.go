package llm

import (
	"context"
	"time"
)

// GenerationRequest represents a request to generate content
type GenerationRequest struct {
	Model       string                 `json:"model"`
	Prompt      string                 `json:"prompt"`
	MaxTokens   int                    `json:"max_tokens,omitempty"`
	Temperature float64                `json:"temperature,omitempty"`
	Stream      bool                   `json:"stream,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	UserID      string                 `json:"user_id,omitempty"`
	ProjectID   string                 `json:"project_id,omitempty"`
}

// GenerationResponse represents a single response chunk from the LLM
type GenerationResponse struct {
	ID       string                 `json:"id"`
	Object   string                 `json:"object"`
	Model    string                 `json:"model"`
	Choices  []Choice               `json:"choices"`
	Usage    *Usage                 `json:"usage,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Choice represents a single choice in the response
type Choice struct {
	Index        int     `json:"index"`
	Text         string  `json:"text,omitempty"`
	Delta        *Delta  `json:"delta,omitempty"`
	FinishReason *string `json:"finish_reason,omitempty"`
}

// Delta represents the incremental content in streaming responses
type Delta struct {
	Content string `json:"content,omitempty"`
}

// Usage represents token usage statistics
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// LLMClient defines the interface for LLM providers
type LLMClient interface {
	// Generate performs a single generation request
	Generate(ctx context.Context, req *GenerationRequest) (*GenerationResponse, error)

	// GenerateStream performs a streaming generation request
	GenerateStream(ctx context.Context, req *GenerationRequest) (<-chan *GenerationResponse, error)

	// GetModels returns available models
	GetModels(ctx context.Context) ([]Model, error)

	// Health checks if the LLM service is healthy
	Health(ctx context.Context) error

	// Close closes the client connection
	Close() error
}

// Model represents an available LLM model
type Model struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Provider    string    `json:"provider"`
	MaxTokens   int       `json:"max_tokens"`
	CreatedAt   time.Time `json:"created_at"`
}

// LLMError represents an error from the LLM service
type LLMError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (e *LLMError) Error() string {
	if e.Details != "" {
		return e.Message + ": " + e.Details
	}
	return e.Message
}

// StreamEvent represents a server-sent event for streaming
type StreamEvent struct {
	Event string `json:"event"`
	Data  string `json:"data"`
	ID    string `json:"id,omitempty"`
}
