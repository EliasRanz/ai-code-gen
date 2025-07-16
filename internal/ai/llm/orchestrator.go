package llm

import (
	"context"
	"fmt"
	"sync"
)

// LLMOrchestratorImpl manages multiple LLM providers consistently
type LLMOrchestratorImpl struct {
	providers    map[string]LLMProvider
	factory      LLMFactory
	rateLimiter  RateLimiter
	quotaManager QuotaManager
	mu           sync.RWMutex
}

// NewLLMOrchestrator creates a new LLM orchestrator
func NewLLMOrchestrator(rateLimiter RateLimiter, quotaManager QuotaManager) *LLMOrchestratorImpl {
	return &LLMOrchestratorImpl{
		providers:    make(map[string]LLMProvider),
		factory:      NewLLMFactory(),
		rateLimiter:  rateLimiter,
		quotaManager: quotaManager,
	}
}

// GenerateCode orchestrates code generation with rate limiting and provider management
func (o *LLMOrchestratorImpl) GenerateCode(ctx context.Context, req *GenerationRequest) (*GenerationResponse, error) {
	// Apply rate limiting
	if !o.rateLimiter.Allow(req.UserID) {
		return nil, ErrRateLimitExceeded
	}

	// Check daily quota
	if o.quotaManager.IsQuotaExceeded(req.UserID) {
		return nil, ErrQuotaExceeded
	}

	// Get or create provider
	provider, err := o.getProvider(req.Provider)
	if err != nil {
		return nil, err
	}

	// Execute generation
	response, err := provider.GenerateCode(ctx, req)
	if err != nil {
		return nil, err
	}

	// Update usage tracking
	if err := o.quotaManager.IncrementUsage(req.UserID, response.TokensUsed); err != nil {
		// Log error but don't fail the request
		// In production, this might be logged to monitoring
	}

	return response, nil
}

// RegisterProvider registers a pre-configured provider
func (o *LLMOrchestratorImpl) RegisterProvider(name string, provider LLMProvider) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.providers[name] = provider
}

// GetAvailableProviders returns list of available providers
func (o *LLMOrchestratorImpl) GetAvailableProviders() []string {
	o.mu.RLock()
	defer o.mu.RUnlock()

	providers := make([]string, 0, len(o.providers))
	for name := range o.providers {
		providers = append(providers, name)
	}

	// Add factory-supported providers not yet instantiated
	factoryProviders := o.factory.ListAvailableProviders()
	for _, fp := range factoryProviders {
		found := false
		for _, p := range providers {
			if p == fp {
				found = true
				break
			}
		}
		if !found {
			providers = append(providers, fp)
		}
	}

	return providers
}

// HealthCheck checks health of all providers
func (o *LLMOrchestratorImpl) HealthCheck(ctx context.Context) map[string]error {
	o.mu.RLock()
	defer o.mu.RUnlock()

	results := make(map[string]error)
	for name, provider := range o.providers {
		results[name] = provider.HealthCheck(ctx)
	}

	return results
}

// GetProviderInfo returns information about a specific provider
func (o *LLMOrchestratorImpl) GetProviderInfo(providerName string) (ProviderInfo, error) {
	provider, err := o.getProvider(providerName)
	if err != nil {
		return ProviderInfo{}, err
	}

	return provider.GetProviderInfo(), nil
}

// Close closes all providers
func (o *LLMOrchestratorImpl) Close() error {
	o.mu.Lock()
	defer o.mu.Unlock()

	var lastErr error
	for name, provider := range o.providers {
		if err := provider.Close(); err != nil {
			lastErr = fmt.Errorf("failed to close provider %s: %w", name, err)
		}
	}

	return lastErr
}

// getProvider gets or creates a provider
func (o *LLMOrchestratorImpl) getProvider(providerName string) (LLMProvider, error) {
	if providerName == "" {
		providerName = "openai" // Default provider
	}

	// Check if already exists
	o.mu.RLock()
	if provider, exists := o.providers[providerName]; exists {
		o.mu.RUnlock()
		return provider, nil
	}
	o.mu.RUnlock()

	// Create new provider
	o.mu.Lock()
	defer o.mu.Unlock()

	// Double-check after acquiring write lock
	if provider, exists := o.providers[providerName]; exists {
		return provider, nil
	}

	// Create with free tier configuration
	config := ProviderConfig{
		FreeTierOnly: true,
	}

	provider, err := o.factory.CreateProvider(providerName, config)
	if err != nil {
		return nil, err
	}

	o.providers[providerName] = provider
	return provider, nil
}
