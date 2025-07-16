package llm

import (
	"fmt"
	"strings"
)

// LLMFactoryImpl implements LLMFactory interface
type LLMFactoryImpl struct {
	defaultConfig ProviderConfig
}

// NewLLMFactory creates a new LLM factory
func NewLLMFactory() LLMFactory {
	return &LLMFactoryImpl{
		defaultConfig: ProviderConfig{
			FreeTierOnly: true, // Enforce free tier by default
		},
	}
}

// CreateProvider creates an LLM provider based on type and configuration
func (f *LLMFactoryImpl) CreateProvider(providerType string, config ProviderConfig) (LLMProvider, error) {
	// Validate free tier enforcement BEFORE merging config
	if !config.FreeTierOnly {
		return nil, ErrFreeTierRequired
	}

	// Merge with defaults
	mergedConfig := f.mergeConfig(config)

	// Validate merged configuration
	if err := f.validateFreeTierConfig(mergedConfig); err != nil {
		return nil, err
	}

	switch strings.ToLower(providerType) {
	case "openai":
		return NewOpenAIClient(mergedConfig)
	case "vllm":
		return NewVLLMClient(mergedConfig)
	default:
		return nil, fmt.Errorf("%w: %s", ErrProviderNotFound, providerType)
	}
}

// ListAvailableProviders returns list of supported provider types
func (f *LLMFactoryImpl) ListAvailableProviders() []string {
	return []string{"openai", "vllm"}
}

// mergeConfig merges user config with factory defaults
func (f *LLMFactoryImpl) mergeConfig(config ProviderConfig) ProviderConfig {
	merged := config // Start with user config

	// Apply defaults only if not provided
	if merged.Timeout == 0 {
		merged.Timeout = f.defaultConfig.Timeout
	}
	if merged.Extra == nil {
		merged.Extra = f.defaultConfig.Extra
	}

	// Always enforce free tier - this is the key security requirement
	merged.FreeTierOnly = true

	return merged
}

// validateFreeTierConfig ensures free tier configuration is enforced
func (f *LLMFactoryImpl) validateFreeTierConfig(config ProviderConfig) error {
	// Always enforce free tier for safety
	if !config.FreeTierOnly {
		return ErrFreeTierRequired
	}

	return nil
}
