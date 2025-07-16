package ai

import (
	"context"
	"fmt"

	"github.com/EliasRanz/ai-code-gen/internal/ai/llm"
	"github.com/EliasRanz/ai-code-gen/internal/utilities"
)

// AIService provides comprehensive AI generation with rate limiting and LLM orchestration
type AIService struct {
	orchestrator *llm.LLMOrchestratorImpl
	rateLimiter  *RateLimiter
	quotaManager *QuotaManager
	cache        *CacheManager
}

// NewAIService creates a new AI service with integrated LLM orchestration
func NewAIService(rateLimiter *RateLimiter, quotaManager *QuotaManager, cache *CacheManager) *AIService {
	// Create rate limiter adapter for LLM orchestrator
	rlAdapter := &rateLimiterAdapter{rateLimiter: rateLimiter}
	qmAdapter := &quotaManagerAdapter{quotaManager: quotaManager}

	orchestrator := llm.NewLLMOrchestrator(rlAdapter, qmAdapter)

	return &AIService{
		orchestrator: orchestrator,
		rateLimiter:  rateLimiter,
		quotaManager: quotaManager,
		cache:        cache,
	}
}

// GenerateWithBuilder generates code using the builder pattern for complex requests
func (s *AIService) GenerateWithBuilder(ctx context.Context, userID, prompt string) (*llm.GenerationResponse, error) {
	request, err := llm.NewGenerationRequestBuilder().
		SetUserID(userID).
		SetPrompt(prompt).
		SetLanguage("go").
		SetMaxTokens(500).
		SetTemperature(0.7).
		AddMetadata("service", "ai").
		SetProvider("openai"). // FREE TIER only
		Build()

	if err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	return s.orchestrator.GenerateCode(ctx, request)
}

// GenerateCode generates code with simple interface (legacy compatibility)
func (s *AIService) GenerateCode(ctx context.Context, req GenerationRequest) (GenerationResult, error) {
	// Convert to new LLM request format
	llmReq := &llm.GenerationRequest{
		UserID:      string(req.UserID),
		Prompt:      req.Prompt,
		Language:    req.Language,
		MaxTokens:   req.GetMaxTokens(),
		Temperature: req.GetTemperature(),
		Provider:    "openai", // Default to OpenAI free tier
		Metadata:    make(map[string]string),
	}

	// Add metadata
	llmReq.Metadata["legacy_request"] = "true"
	if req.ProjectID != nil {
		llmReq.Metadata["project_id"] = string(*req.ProjectID)
	}

	response, err := s.orchestrator.GenerateCode(ctx, llmReq)
	if err != nil {
		return GenerationResult{}, err
	}

	// Convert back to legacy format
	return GenerationResult{
		Code:       response.Content,
		Model:      response.Model,
		UsedTokens: response.TokensUsed,
	}, nil
}

// GetAvailableProviders returns list of available LLM providers
func (s *AIService) GetAvailableProviders() []string {
	return s.orchestrator.GetAvailableProviders()
}

// GetProviderInfo returns information about a specific provider
func (s *AIService) GetProviderInfo(providerName string) (llm.ProviderInfo, error) {
	return s.orchestrator.GetProviderInfo(providerName)
}

// HealthCheck checks health of all LLM providers
func (s *AIService) HealthCheck(ctx context.Context) map[string]error {
	return s.orchestrator.HealthCheck(ctx)
}

// Close closes the AI service and all providers
func (s *AIService) Close() error {
	return s.orchestrator.Close()
}

// rateLimiterAdapter adapts RateLimiter to LLM interface
type rateLimiterAdapter struct {
	rateLimiter *RateLimiter
}

func (r *rateLimiterAdapter) Allow(userID string) bool {
	return r.rateLimiter.Allow(utilities.UserID(userID))
}

// quotaManagerAdapter adapts QuotaManager to LLM interface
type quotaManagerAdapter struct {
	quotaManager *QuotaManager
}

func (q *quotaManagerAdapter) IsQuotaExceeded(userID string) bool {
	// Check with a default daily limit of 100
	return !q.quotaManager.CheckQuota(userID, 100)
}

func (q *quotaManagerAdapter) IncrementUsage(userID string, tokens int) error {
	q.quotaManager.UseQuota(userID)
	return nil
}
