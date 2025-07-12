// Package ai contains AI service interfaces (consolidated from domain layer)
package ai

import (
	"context"
)

// Repository defines AI domain data access
type Repository interface {
	SaveGeneration(ctx context.Context, generation AIGenerationHistory) error
	GetHistory(ctx context.Context, userID UserID, limit int) ([]AIGenerationHistory, error)
	GetQuotaUsage(ctx context.Context, userID UserID) (QuotaStatus, error)
	UpdateQuotaUsage(ctx context.Context, userID UserID, tokens int) error
}

// LLMService defines the interface for LLM interactions
type LLMService interface {
	Generate(ctx context.Context, req GenerationRequest) (GenerationResult, error)
	GenerateStream(ctx context.Context, req GenerationRequest, ch chan<- StreamChunk) error
	Stream(ctx context.Context, req GenerationRequest, ch chan<- string) error // Legacy method
	Validate(ctx context.Context, code string) (ValidationResult, error)
}

// EventPublisher defines event publishing interface
type EventPublisher interface {
	PublishGenerationEvent(ctx context.Context, userID UserID, tokens int) error
}
