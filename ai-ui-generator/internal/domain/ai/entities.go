// Package ai contains AI domain entities and business rules
package ai

import (
	"github.com/EliasRanz/ai-code-gen/ai-ui-generator/internal/domain/common"
)

// GenerationRequest represents a request to generate code
type GenerationRequest struct {
	Prompt      string
	Language    string
	Framework   string
	Style       string
	Complexity  string
	UserID      common.UserID
	ProjectID   *common.ProjectID
	Model       string
	Temperature *float64
	MaxTokens   *int
}

// Validate validates the generation request
func (r GenerationRequest) Validate() error {
	if r.Prompt == "" {
		return common.ErrInvalidInput
	}
	if r.UserID.IsEmpty() {
		return common.ErrInvalidInput
	}
	if r.Temperature != nil && (*r.Temperature < 0 || *r.Temperature > 2) {
		return common.ErrInvalidInput
	}
	if r.MaxTokens != nil && (*r.MaxTokens < 1 || *r.MaxTokens > 4096) {
		return common.ErrInvalidInput
	}
	return nil
}

// GetModel returns the model to use for generation, with a default value
func (r GenerationRequest) GetModel() string {
	if r.Model == "" {
		return "gpt-3.5-turbo" // Default model
	}
	return r.Model
}

// GetTemperature returns the temperature to use for generation, with a default value
func (r GenerationRequest) GetTemperature() float64 {
	if r.Temperature == nil {
		return 0.7 // Default temperature
	}
	return *r.Temperature
}

// GetMaxTokens returns the max tokens to use for generation, with a default value
func (r GenerationRequest) GetMaxTokens() int {
	if r.MaxTokens == nil {
		return 2048 // Default max tokens
	}
	return *r.MaxTokens
}

// GenerationResult represents the result of code generation
type GenerationResult struct {
	ID            string
	Code          string
	Model         string
	UsedTokens    int
	EstimatedCost float64
	common.Timestamps
}

// ValidationRequest represents a request to validate code
type ValidationRequest struct {
	Code   string
	UserID common.UserID
}

// Validate validates the validation request
func (r ValidationRequest) Validate() error {
	if r.Code == "" {
		return common.ErrInvalidInput
	}
	if r.UserID.IsEmpty() {
		return common.ErrInvalidInput
	}
	return nil
}

// ValidationResult represents the result of code validation
type ValidationResult struct {
	Valid  bool
	Errors []string
}

// GenerationHistory represents a user's generation history entry
type GenerationHistory struct {
	ID     string
	UserID common.UserID
	Prompt string
	Code   string
	Model  string
	Tokens int
	common.Timestamps
}

// QuotaStatus represents a user's quota status
type QuotaStatus struct {
	UserID     common.UserID
	DailyLimit int
	UsedToday  int
	Remaining  int
	ResetTime  string
}

// CanGenerate returns true if the user can generate more content
func (q QuotaStatus) CanGenerate() bool {
	return q.Remaining > 0
}

// StreamChunk represents a chunk of streaming content
type StreamChunk struct {
	Content    string
	TokenCount int
	IsComplete bool
	Model      string // Model name used for generation
	Error      error
}
