// Package ai contains AI application use cases
package ai

import (
	"context"

	"github.com/EliasRanz/ai-code-gen/internal/utilities"
)

// GenerateCodeRequest represents a code generation request
type GenerateCodeRequest struct {
	Prompt     string               `json:"prompt" validate:"required,min=1,max=10000"`
	Language   string               `json:"language" validate:"required,oneof=javascript typescript python go java"`
	Framework  string               `json:"framework"`
	Style      string               `json:"style"`
	Complexity string               `json:"complexity" validate:"oneof=simple medium complex"`
	UserID     utilities.UserID     `json:"user_id" validate:"required"`
	ProjectID  *utilities.ProjectID `json:"project_id,omitempty"`
}

// GenerateCodeResponse represents a code generation response
type GenerateCodeResponse struct {
	ID            string  `json:"id"`
	Code          string  `json:"code"`
	Language      string  `json:"language"`
	Framework     string  `json:"framework"`
	Model         string  `json:"model"`
	UsedTokens    int     `json:"used_tokens"`
	EstimatedCost float64 `json:"estimated_cost"`
	CreatedAt     string  `json:"created_at"`
}

// GenerateCodeService handles code generation
type GenerateCodeService struct {
	repo        Repository
	llmService  LLMService
	rateLimiter RateLimiter
	publisher   EventPublisher
}

// NewGenerateCodeService creates a new GenerateCodeService
func NewGenerateCodeService(
	repo Repository,
	llmService LLMService,
	rateLimiter RateLimiter,
	publisher EventPublisher,
) *GenerateCodeService {
	return &GenerateCodeService{
		repo:        repo,
		llmService:  llmService,
		rateLimiter: rateLimiter,
		publisher:   publisher,
	}
}

// Execute executes the code generation
func (s *GenerateCodeService) Execute(ctx context.Context, req GenerateCodeRequest) (*GenerateCodeResponse, error) {
	// Convert to domain request
	domainReq := GenerationRequest{
		Prompt:     req.Prompt,
		Language:   req.Language,
		Framework:  req.Framework,
		Style:      req.Style,
		Complexity: req.Complexity,
		UserID:     req.UserID,
		ProjectID:  req.ProjectID,
	}

	// Validate request
	if err := domainReq.Validate(); err != nil {
		return nil, utilities.NewValidationError("invalid generation request", err)
	}

	// Check rate limit
	if !s.rateLimiter.Allow(req.UserID) {
		return nil, utilities.NewValidationError("rate limit exceeded", nil)
	}

	// Check quota
	quota, err := s.repo.GetQuotaUsage(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	if !quota.CanGenerate() {
		return nil, utilities.NewValidationError("quota exceeded", nil)
	}

	// Generate code
	result, err := s.llmService.Generate(ctx, domainReq)
	if err != nil {
		return nil, err
	}

	// Save to history
	history := AIGenerationHistory{
		UserID: req.UserID,
		Prompt: req.Prompt,
		Code:   result.Code,
		Model:  result.Model,
		Tokens: result.UsedTokens,
	}

	if err := s.repo.SaveGeneration(ctx, history); err != nil {
		// Log error but don't fail the request
	}

	// Update quota
	if err := s.repo.UpdateQuotaUsage(ctx, req.UserID, result.UsedTokens); err != nil {
		// Log error but don't fail the request
	}

	// Publish event
	if s.publisher != nil {
		_ = s.publisher.PublishGenerationEvent(ctx, req.UserID, result.UsedTokens)
	}

	// Convert to response
	response := &GenerateCodeResponse{
		ID:            result.ID,
		Code:          result.Code,
		Language:      req.Language,
		Framework:     req.Framework,
		Model:         result.Model,
		UsedTokens:    result.UsedTokens,
		EstimatedCost: result.EstimatedCost,
		CreatedAt:     result.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	return response, nil
}
