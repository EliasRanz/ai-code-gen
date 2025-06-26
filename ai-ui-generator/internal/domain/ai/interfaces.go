// Package ai contains AI domain interfaces
package ai

import (
	"context"

	"github.com/EliasRanz/ai-code-gen/ai-ui-generator/internal/domain/common"
)

// Repository defines AI domain data access
type Repository interface {
	SaveGeneration(ctx context.Context, generation GenerationHistory) error
	GetHistory(ctx context.Context, userID common.UserID, limit int) ([]GenerationHistory, error)
	GetQuotaUsage(ctx context.Context, userID common.UserID) (QuotaStatus, error)
	UpdateQuotaUsage(ctx context.Context, userID common.UserID, tokens int) error
}

// LLMService defines the interface for LLM interactions
type LLMService interface {
	Generate(ctx context.Context, req GenerationRequest) (GenerationResult, error)
	GenerateStream(ctx context.Context, req GenerationRequest, ch chan<- StreamChunk) error
	Stream(ctx context.Context, req GenerationRequest, ch chan<- string) error // Legacy method
	Validate(ctx context.Context, code string) (ValidationResult, error)
}

// RateLimiter defines rate limiting interface
type RateLimiter interface {
	Allow(userID common.UserID) bool
	Reset(userID common.UserID)
}

// EventPublisher defines event publishing interface
type EventPublisher interface {
	PublishGenerationEvent(ctx context.Context, userID common.UserID, tokens int) error
}
