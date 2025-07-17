package llm

import (
	"context"
	"fmt"
	"time"
)

// LLMProvider defines the interface that all LLM providers must implement
type LLMProvider interface {
	LLMGenerationOperations
	LLMProviderOperations
	HealthOperations
}

// LLMGenerationOperations defines code generation operations
type LLMGenerationOperations interface {
	GenerateCode(ctx context.Context, req *GenerationRequest) (*GenerationResponse, error)
}

// LLMProviderOperations defines provider-specific operations
type LLMProviderOperations interface {
	GetProviderInfo() ProviderInfo
	GetLimits() ProviderLimits
}

// HealthOperations defines health and lifecycle operations
type HealthOperations interface {
	HealthCheck(ctx context.Context) error
	Close() error
}

// LLMFactory provides factory pattern for provider instantiation
type LLMFactory interface {
	CreateProvider(providerType string, config ProviderConfig) (LLMProvider, error)
	ListAvailableProviders() []string
}

// GenerationRequest represents standard request structure across all providers
type GenerationRequest struct {
	UserID      string            `json:"user_id" validate:"required"`
	Prompt      string            `json:"prompt" validate:"required,max=8000"`
	Language    string            `json:"language,omitempty"`
	MaxTokens   int               `json:"max_tokens,omitempty"`
	Temperature float64           `json:"temperature,omitempty"`
	Provider    string            `json:"provider,omitempty"`
	Model       string            `json:"model,omitempty"`
	Stream      bool              `json:"stream,omitempty"`
	Timeout     time.Duration     `json:"timeout,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// GenerationResponse represents standard response structure across all providers
type GenerationResponse struct {
	Content      string            `json:"content"`
	TokensUsed   int               `json:"tokens_used"`
	Provider     string            `json:"provider"`
	Model        string            `json:"model"`
	Latency      time.Duration     `json:"latency"`
	RequestID    string            `json:"request_id"`
	FinishReason string            `json:"finish_reason,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// ProviderInfo contains provider-specific metadata
type ProviderInfo struct {
	Name         string   `json:"name"`
	Version      string   `json:"version"`
	Models       []string `json:"models"`
	Capabilities []string `json:"capabilities"`
	FreeTier     bool     `json:"free_tier"`
}

// ProviderLimits contains provider-specific rate limits and quotas
type ProviderLimits struct {
	RequestsPerMinute   int `json:"requests_per_minute"`
	TokensPerMinute     int `json:"tokens_per_minute"`
	DailyQuota          int `json:"daily_quota"`
	MaxTokensPerRequest int `json:"max_tokens_per_request"`
}

// ProviderConfig holds configuration for provider instantiation
type ProviderConfig struct {
	APIKey       string            `json:"api_key,omitempty"`
	BaseURL      string            `json:"base_url,omitempty"`
	Model        string            `json:"model,omitempty"`
	FreeTierOnly bool              `json:"free_tier_only"`
	Timeout      time.Duration     `json:"timeout,omitempty"`
	Extra        map[string]string `json:"extra,omitempty"`
}

// LLMOrchestrator manages multiple providers consistently
type LLMOrchestrator struct {
	providers    map[string]LLMProvider
	factory      LLMFactory
	rateLimiter  RateLimiter
	quotaManager QuotaManager
}

// RateLimiter interface for rate limiting integration
type RateLimiter interface {
	Allow(userID string) bool
}

// QuotaManager interface for quota management
type QuotaManager interface {
	IsQuotaExceeded(userID string) bool
	IncrementUsage(userID string, tokens int) error
}

// Common errors
var (
	ErrRateLimitExceeded = fmt.Errorf("rate limit exceeded")
	ErrQuotaExceeded     = fmt.Errorf("daily quota exceeded")
	ErrProviderNotFound  = fmt.Errorf("provider not found")
	ErrInvalidConfig     = fmt.Errorf("invalid provider configuration")
	ErrFreeTierRequired  = fmt.Errorf("free tier configuration required")
)

// LLMError represents provider-specific errors
type LLMError struct {
	Provider string `json:"provider"`
	Code     string `json:"code"`
	Message  string `json:"message"`
	Details  string `json:"details,omitempty"`
}

func (e *LLMError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("[%s] %s: %s (%s)", e.Provider, e.Message, e.Details, e.Code)
	}
	return fmt.Sprintf("[%s] %s (%s)", e.Provider, e.Message, e.Code)
}

// NewLLMError creates a new LLM error
func NewLLMError(provider, code, message, details string) *LLMError {
	return &LLMError{
		Provider: provider,
		Code:     code,
		Message:  message,
		Details:  details,
	}
}
