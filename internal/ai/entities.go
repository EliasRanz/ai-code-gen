// Package ai contains AI service entities and business logic (consolidated from domain layer)
package ai

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/EliasRanz/ai-code-gen/internal/utilities"
)

// Common domain errors
var (
	ErrNotFound     = errors.New("resource not found")
	ErrUnauthorized = errors.New("unauthorized access")
	ErrInvalidInput = errors.New("invalid input")
	ErrConflict     = errors.New("resource conflict")
	ErrInternal     = errors.New("internal error")
)

// GenerationRequest represents a request to generate code
type GenerationRequest struct {
	Prompt      string
	Language    string
	Framework   string
	Style       string
	Complexity  string
	UserID      utilities.UserID
	ProjectID   *utilities.ProjectID
	Model       string
	Temperature *float64
	MaxTokens   *int
}

// Validate validates the generation request
func (r GenerationRequest) Validate() error {
	if r.Prompt == "" {
		return ErrInvalidInput
	}
	if r.UserID.IsEmpty() {
		return ErrInvalidInput
	}
	if r.Temperature != nil && (*r.Temperature < 0 || *r.Temperature > 2) {
		return ErrInvalidInput
	}
	if r.MaxTokens != nil && (*r.MaxTokens < 1 || *r.MaxTokens > 4096) {
		return ErrInvalidInput
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
	utilities.BaseEntity
	ID            string
	Code          string
	Model         string
	UsedTokens    int
	EstimatedCost float64
	utilities.Timestamps
}

// Validate validates the generation result
func (g GenerationResult) Validate() error {
	if g.ID == "" {
		return ErrInvalidInput
	}
	if g.Code == "" {
		return ErrInvalidInput
	}
	return nil
}

// GetValidationRules returns validation rules for generation result
func (g GenerationResult) GetValidationRules() []utilities.ValidationRule {
	return []utilities.ValidationRule{
		{Field: "id", Rule: "required", Message: "ID is required"},
		{Field: "code", Rule: "required", Message: "Code is required"},
	}
}

// ToMap converts generation result to map
func (g GenerationResult) ToMap() map[string]interface{} {
	// Start with BaseEntity map
	baseMap := g.BaseEntity.ToMap()
	// Add GenerationResult specific fields
	baseMap["id"] = g.ID
	baseMap["code"] = g.Code
	baseMap["model"] = g.Model
	baseMap["used_tokens"] = g.UsedTokens
	baseMap["estimated_cost"] = g.EstimatedCost
	baseMap["created_at"] = g.CreatedAt
	baseMap["updated_at"] = g.UpdatedAt
	return baseMap
}

// ToJSON serializes generation result to JSON
func (g GenerationResult) ToJSON() ([]byte, error) {
	data := g.ToMap()
	return json.Marshal(data)
}

// FromJSON deserializes generation result from JSON
func (g *GenerationResult) FromJSON(data []byte) error {
	var mapData map[string]interface{}
	if err := json.Unmarshal(data, &mapData); err != nil {
		return err
	}
	if id, ok := mapData["id"].(string); ok {
		g.ID = id
	}
	if code, ok := mapData["code"].(string); ok {
		g.Code = code
	}
	if model, ok := mapData["model"].(string); ok {
		g.Model = model
	}
	return nil
}

// NewGenerationResult creates a new generation result with entity properties
func NewGenerationResult(id, code, model string, tokens int, cost float64) *GenerationResult {
	now := time.Now()
	return &GenerationResult{
		BaseEntity:    utilities.NewBaseEntity(id, utilities.EntityTypeGeneration),
		ID:            id,
		Code:          code,
		Model:         model,
		UsedTokens:    tokens,
		EstimatedCost: cost,
		Timestamps: utilities.Timestamps{
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
}

// ValidationRequest represents a request to validate code
type ValidationRequest struct {
	Code   string
	UserID utilities.UserID
}

// Validate validates the validation request
func (r ValidationRequest) Validate() error {
	if r.Code == "" {
		return ErrInvalidInput
	}
	if r.UserID.IsEmpty() {
		return ErrInvalidInput
	}
	return nil
}

// ValidationResult represents the result of code validation
type ValidationResult struct {
	Valid  bool
	Errors []string
}

// AIGenerationHistory represents a user's generation history entry
type AIGenerationHistory struct {
	ID     string
	UserID utilities.UserID
	Prompt string
	Code   string
	Model  string
	Tokens int
	utilities.Timestamps
}

// QuotaStatus represents a user's quota status
type QuotaStatus struct {
	UserID     utilities.UserID
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
