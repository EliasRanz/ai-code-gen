package llm

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// GenerationRequestBuilder provides fluent API for complex LLM request configuration
type GenerationRequestBuilder interface {
	SetUserID(userID string) GenerationRequestBuilder
	SetPrompt(prompt string) GenerationRequestBuilder
	SetLanguage(language string) GenerationRequestBuilder
	SetMaxTokens(tokens int) GenerationRequestBuilder
	SetTemperature(temp float64) GenerationRequestBuilder
	AddMetadata(key, value string) GenerationRequestBuilder
	SetProvider(provider string) GenerationRequestBuilder
	SetModel(model string) GenerationRequestBuilder
	EnableStreaming() GenerationRequestBuilder
	SetTimeout(timeout time.Duration) GenerationRequestBuilder
	Build() (*GenerationRequest, error)
	Validate() error
}

// LLMRequestBuilder implements GenerationRequestBuilder
type LLMRequestBuilder struct {
	request *GenerationRequest
	config  *LLMConfig
	errors  []error
}

// LLMConfig holds configuration for request building
type LLMConfig struct {
	FreeTierOnly       bool
	MaxPromptLength    int
	DefaultMaxTokens   int
	DefaultTemperature float64
	DefaultTimeout     time.Duration
	AllowedProviders   []string
}

// NewGenerationRequestBuilder creates a new request builder
func NewGenerationRequestBuilder() GenerationRequestBuilder {
	return &LLMRequestBuilder{
		request: &GenerationRequest{
			Metadata: make(map[string]string),
		},
		config: &LLMConfig{
			FreeTierOnly:       true,
			MaxPromptLength:    8000,
			DefaultMaxTokens:   1000,
			DefaultTemperature: 0.7,
			DefaultTimeout:     30 * time.Second,
			AllowedProviders:   []string{"openai", "vllm"},
		},
	}
}

// SetUserID sets the user ID with validation
func (b *LLMRequestBuilder) SetUserID(userID string) GenerationRequestBuilder {
	if userID == "" {
		b.errors = append(b.errors, fmt.Errorf("user ID cannot be empty"))
		return b
	}
	b.request.UserID = userID
	return b
}

// SetPrompt sets the prompt with validation
func (b *LLMRequestBuilder) SetPrompt(prompt string) GenerationRequestBuilder {
	if len(prompt) == 0 {
		b.errors = append(b.errors, fmt.Errorf("prompt cannot be empty"))
		return b
	}
	if len(prompt) > b.config.MaxPromptLength {
		b.errors = append(b.errors, fmt.Errorf("prompt exceeds free tier limit of %d characters", b.config.MaxPromptLength))
		return b
	}
	b.request.Prompt = prompt
	return b
}

// SetLanguage sets the programming language
func (b *LLMRequestBuilder) SetLanguage(language string) GenerationRequestBuilder {
	if language != "" {
		b.request.Language = language
	}
	return b
}

// SetMaxTokens sets max tokens with free tier validation
func (b *LLMRequestBuilder) SetMaxTokens(tokens int) GenerationRequestBuilder {
	if tokens < 0 {
		b.errors = append(b.errors, fmt.Errorf("max tokens cannot be negative"))
		return b
	}
	if b.config.FreeTierOnly && tokens > 2000 {
		b.errors = append(b.errors, fmt.Errorf("max tokens exceeds free tier limit of 2000"))
		return b
	}
	b.request.MaxTokens = tokens
	return b
}

// SetTemperature sets temperature with validation
func (b *LLMRequestBuilder) SetTemperature(temp float64) GenerationRequestBuilder {
	if temp < 0 || temp > 2.0 {
		b.errors = append(b.errors, fmt.Errorf("temperature must be between 0 and 2.0"))
		return b
	}
	b.request.Temperature = temp
	return b
}

// AddMetadata adds metadata key-value pair
func (b *LLMRequestBuilder) AddMetadata(key, value string) GenerationRequestBuilder {
	if key == "" {
		b.errors = append(b.errors, fmt.Errorf("metadata key cannot be empty"))
		return b
	}
	b.request.Metadata[key] = value
	return b
}

// SetProvider sets the LLM provider with validation
func (b *LLMRequestBuilder) SetProvider(provider string) GenerationRequestBuilder {
	if provider == "" {
		return b
	}

	// Validate against allowed providers
	allowed := false
	for _, p := range b.config.AllowedProviders {
		if p == provider {
			allowed = true
			break
		}
	}
	if !allowed {
		b.errors = append(b.errors, fmt.Errorf("provider %s not in allowed list: %v", provider, b.config.AllowedProviders))
		return b
	}

	b.request.Provider = provider
	return b
}

// SetModel sets the model name
func (b *LLMRequestBuilder) SetModel(model string) GenerationRequestBuilder {
	if model != "" {
		b.request.Model = model
	}
	return b
}

// EnableStreaming enables streaming response
func (b *LLMRequestBuilder) EnableStreaming() GenerationRequestBuilder {
	b.request.Stream = true
	return b
}

// SetTimeout sets request timeout
func (b *LLMRequestBuilder) SetTimeout(timeout time.Duration) GenerationRequestBuilder {
	if timeout <= 0 {
		b.errors = append(b.errors, fmt.Errorf("timeout must be positive"))
		return b
	}
	if timeout > 5*time.Minute {
		b.errors = append(b.errors, fmt.Errorf("timeout exceeds maximum of 5 minutes"))
		return b
	}
	b.request.Timeout = timeout
	return b
}

// Validate performs validation without building
func (b *LLMRequestBuilder) Validate() error {
	if len(b.errors) > 0 {
		return fmt.Errorf("validation errors: %v", b.errors)
	}

	// Required field validation
	if b.request.UserID == "" {
		return fmt.Errorf("user ID is required")
	}
	if b.request.Prompt == "" {
		return fmt.Errorf("prompt is required")
	}

	return nil
}

// Build creates the final GenerationRequest with defaults applied
func (b *LLMRequestBuilder) Build() (*GenerationRequest, error) {
	// Check for accumulated errors
	if err := b.Validate(); err != nil {
		return nil, err
	}

	// Apply free tier defaults
	if b.request.MaxTokens == 0 {
		b.request.MaxTokens = b.config.DefaultMaxTokens
	}
	if b.request.Temperature == 0 {
		b.request.Temperature = b.config.DefaultTemperature
	}
	if b.request.Timeout == 0 {
		b.request.Timeout = b.config.DefaultTimeout
	}
	if b.request.Provider == "" {
		b.request.Provider = "openai" // FREE TIER default
	}

	// Add request ID if not present
	if _, exists := b.request.Metadata["request_id"]; !exists {
		b.request.Metadata["request_id"] = uuid.New().String()
	}

	// Add free tier enforcement flag
	if b.config.FreeTierOnly {
		b.request.Metadata["free_tier_only"] = "true"
	}

	return b.request, nil
}
