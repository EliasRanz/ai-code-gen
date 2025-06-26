// Package ai contains AI application use cases
package ai

import (
	"context"

	"github.com/EliasRanz/ai-code-gen/ai-ui-generator/internal/domain/ai"
	"github.com/EliasRanz/ai-code-gen/ai-ui-generator/internal/domain/common"
)

// StreamCodeRequest represents a streaming code generation request
type StreamCodeRequest struct {
	Prompt     string            `json:"prompt" validate:"required,min=1,max=10000"`
	Language   string            `json:"language" validate:"required,oneof=javascript typescript python go java"`
	Framework  string            `json:"framework"`
	Style      string            `json:"style"`
	Complexity string            `json:"complexity" validate:"oneof=simple medium complex"`
	UserID     common.UserID     `json:"user_id" validate:"required"`
	ProjectID  *common.ProjectID `json:"project_id,omitempty"`
}

// StreamCodeResponse represents a streaming code generation response chunk
type StreamCodeResponse struct {
	Type       string `json:"type"` // "chunk", "complete", "error"
	Content    string `json:"content"`
	TokenCount int    `json:"token_count,omitempty"`
	IsComplete bool   `json:"is_complete"`
	Error      string `json:"error,omitempty"`
}

// StreamCodeUseCase handles streaming code generation
type StreamCodeUseCase struct {
	repo        ai.Repository
	llmService  ai.LLMService
	rateLimiter ai.RateLimiter
	publisher   ai.EventPublisher
}

// NewStreamCodeUseCase creates a new StreamCodeUseCase
func NewStreamCodeUseCase(
	repo ai.Repository,
	llmService ai.LLMService,
	rateLimiter ai.RateLimiter,
	publisher ai.EventPublisher,
) *StreamCodeUseCase {
	return &StreamCodeUseCase{
		repo:        repo,
		llmService:  llmService,
		rateLimiter: rateLimiter,
		publisher:   publisher,
	}
}

// Execute executes the streaming code generation use case
func (uc *StreamCodeUseCase) Execute(ctx context.Context, req StreamCodeRequest, responseChan chan<- StreamCodeResponse) error {
	// Convert to domain request
	domainReq := ai.GenerationRequest{
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
		responseChan <- StreamCodeResponse{
			Type:  "error",
			Error: "Invalid request: " + err.Error(),
		}
		return common.NewValidationError("invalid generation request", err)
	}

	// Check rate limit
	if !uc.rateLimiter.Allow(req.UserID) {
		responseChan <- StreamCodeResponse{
			Type:  "error",
			Error: "Rate limit exceeded",
		}
		return common.NewValidationError("rate limit exceeded", nil)
	}

	// Check quota
	quota, err := uc.repo.GetQuotaUsage(ctx, req.UserID)
	if err != nil {
		responseChan <- StreamCodeResponse{
			Type:  "error",
			Error: "Failed to check quota",
		}
		return err
	}

	if !quota.CanGenerate() {
		responseChan <- StreamCodeResponse{
			Type:  "error",
			Error: "Quota exceeded",
		}
		return common.NewValidationError("quota exceeded", nil)
	}

	// Create streaming channel for domain chunks
	streamChan := make(chan ai.StreamChunk, 10)

	// Start streaming from LLM service
	go func() {
		defer close(streamChan)
		err := uc.llmService.GenerateStream(ctx, domainReq, streamChan)
		if err != nil {
			streamChan <- ai.StreamChunk{
				Error: err,
			}
		}
	}()

	// Forward stream chunks to response channel
	totalTokens := 0
	fullContent := ""
	var modelName string

	for chunk := range streamChan {
		if chunk.Error != nil {
			responseChan <- StreamCodeResponse{
				Type:  "error",
				Error: chunk.Error.Error(),
			}
			return chunk.Error
		}

		fullContent += chunk.Content
		totalTokens += chunk.TokenCount

		// Capture model name from the first chunk that has it
		if chunk.Model != "" && modelName == "" {
			modelName = chunk.Model
		}

		responseChan <- StreamCodeResponse{
			Type:       "chunk",
			Content:    chunk.Content,
			TokenCount: chunk.TokenCount,
			IsComplete: chunk.IsComplete,
		}

		if chunk.IsComplete {
			break
		}
	}

	// Use captured model name or fallback to default
	if modelName == "" {
		modelName = "unknown-model"
	}

	// Save to history
	history := ai.GenerationHistory{
		UserID: req.UserID,
		Prompt: req.Prompt,
		Code:   fullContent,
		Model:  modelName, // Now using actual model from stream
		Tokens: totalTokens,
	}

	if err := uc.repo.SaveGeneration(ctx, history); err != nil {
		// Log error but don't fail the request
	}

	// Update quota
	if err := uc.repo.UpdateQuotaUsage(ctx, req.UserID, totalTokens); err != nil {
		// Log error but don't fail the request
	}

	// Publish event
	if uc.publisher != nil {
		_ = uc.publisher.PublishGenerationEvent(ctx, req.UserID, totalTokens)
	}

	// Send completion response
	responseChan <- StreamCodeResponse{
		Type:       "complete",
		Content:    "",
		TokenCount: totalTokens,
		IsComplete: true,
	}

	return nil
}
