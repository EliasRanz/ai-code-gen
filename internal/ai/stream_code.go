// Package ai contains AI streaming code generation
package ai

import (
	"context"
	"fmt"

	"github.com/EliasRanz/ai-code-gen/internal/utilities"
)

// StreamCodeRequest represents a streaming code generation request
type StreamCodeRequest struct {
	Prompt     string               `json:"prompt" validate:"required,min=1,max=10000"`
	Language   string               `json:"language" validate:"required,oneof=javascript typescript python go java"`
	Framework  string               `json:"framework"`
	Style      string               `json:"style"`
	Complexity string               `json:"complexity" validate:"oneof=simple medium complex"`
	UserID     utilities.UserID     `json:"user_id" validate:"required"`
	ProjectID  *utilities.ProjectID `json:"project_id,omitempty"`
}

// StreamCodeResponse represents a streaming code generation response chunk
type StreamCodeResponse struct {
	Type       string `json:"type"` // "chunk", "complete", "error"
	Content    string `json:"content"`
	TokenCount int    `json:"token_count,omitempty"`
	IsComplete bool   `json:"is_complete"`
	Error      string `json:"error,omitempty"`
}

// StreamCodeService handles streaming code generation
type StreamCodeService struct {
	repo        Repository
	llmService  LLMService
	rateLimiter RateLimiter
	publisher   EventPublisher
}

// NewStreamCodeService creates a new StreamCodeService
func NewStreamCodeService(
	repo Repository,
	llmService LLMService,
	rateLimiter RateLimiter,
	publisher EventPublisher,
) *StreamCodeService {
	return &StreamCodeService{
		repo:        repo,
		llmService:  llmService,
		rateLimiter: rateLimiter,
		publisher:   publisher,
	}
}

// Execute executes streaming code generation
func (s *StreamCodeService) Execute(ctx context.Context, req StreamCodeRequest) (<-chan StreamCodeResponse, error) {
	responseChan := make(chan StreamCodeResponse, 10)

	go func() {
		defer close(responseChan)

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
			responseChan <- StreamCodeResponse{
				Type:  "error",
				Error: "Invalid generation request: " + err.Error(),
			}
			return
		}

		// Check rate limit
		if !s.rateLimiter.Allow(req.UserID) {
			responseChan <- StreamCodeResponse{
				Type:  "error",
				Error: "Rate limit exceeded",
			}
			return
		}

		// Check quota
		quota, err := s.repo.GetQuotaUsage(ctx, req.UserID)
		if err != nil {
			responseChan <- StreamCodeResponse{
				Type:  "error",
				Error: "Failed to check quota: " + err.Error(),
			}
			return
		}

		if !quota.CanGenerate() {
			responseChan <- StreamCodeResponse{
				Type:  "error",
				Error: "Quota exceeded",
			}
			return
		}

		// Start streaming generation
		streamChan := make(chan StreamChunk)
		go func() {
			defer close(streamChan)
			err := s.llmService.GenerateStream(ctx, domainReq, streamChan)
			if err != nil {
				streamChan <- StreamChunk{
					Error: fmt.Errorf("failed to start generation: %w", err),
				}
				return
			}
		}()

		totalTokens := 0
		var fullCode string

		for chunk := range streamChan {
			if chunk.Error != nil {
				responseChan <- StreamCodeResponse{
					Type:  "error",
					Error: chunk.Error.Error(),
				}
				return
			}

			totalTokens += chunk.TokenCount
			fullCode += chunk.Content

			responseChan <- StreamCodeResponse{
				Type:       "chunk",
				Content:    chunk.Content,
				TokenCount: chunk.TokenCount,
				IsComplete: chunk.IsComplete,
			}

			if chunk.IsComplete {
				// Save to history
				history := AIGenerationHistory{
					UserID: req.UserID,
					Prompt: req.Prompt,
					Code:   fullCode,
					Model:  chunk.Model,
					Tokens: totalTokens,
				}

				if err := s.repo.SaveGeneration(ctx, history); err != nil {
					// Log error but don't fail the stream
				}

				// Update quota
				if err := s.repo.UpdateQuotaUsage(ctx, req.UserID, totalTokens); err != nil {
					// Log error but don't fail the stream
				}

				// Publish event
				if s.publisher != nil {
					_ = s.publisher.PublishGenerationEvent(ctx, req.UserID, totalTokens)
				}

				responseChan <- StreamCodeResponse{
					Type:       "complete",
					IsComplete: true,
				}
				break
			}
		}
	}()

	return responseChan, nil
}
