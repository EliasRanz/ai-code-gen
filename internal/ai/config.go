package ai

import (
	"fmt"
	"time"

	"github.com/EliasRanz/ai-code-gen/internal/ai/llm"
)

// Config holds AI service configuration including LLM settings
type Config struct {
	// Rate limiting configuration
	RateLimit RateLimitConfig `json:"rate_limit" yaml:"rate_limit"`

	// Quota configuration
	Quota QuotaConfig `json:"quota" yaml:"quota"`

	// LLM provider configuration
	LLM LLMConfig `json:"llm" yaml:"llm"`

	// Cache configuration
	Cache CacheConfig `json:"cache" yaml:"cache"`

	// Service configuration
	Service ServiceConfig `json:"service" yaml:"service"`
}

// RateLimitConfig configures rate limiting
type RateLimitConfig struct {
	RequestsPerMinute int           `json:"requests_per_minute" yaml:"requests_per_minute"`
	BurstSize         int           `json:"burst_size" yaml:"burst_size"`
	CleanupInterval   time.Duration `json:"cleanup_interval" yaml:"cleanup_interval"`
}

// QuotaConfig configures daily quotas
type QuotaConfig struct {
	DefaultDailyLimit int           `json:"default_daily_limit" yaml:"default_daily_limit"`
	PremiumDailyLimit int           `json:"premium_daily_limit" yaml:"premium_daily_limit"`
	ResetTime         string        `json:"reset_time" yaml:"reset_time"`
	TrackingEnabled   bool          `json:"tracking_enabled" yaml:"tracking_enabled"`
	CleanupInterval   time.Duration `json:"cleanup_interval" yaml:"cleanup_interval"`
}

// LLMConfig configures LLM providers and settings
type LLMConfig struct {
	// Provider configuration
	DefaultProvider string                        `json:"default_provider" yaml:"default_provider"`
	Providers       map[string]llm.ProviderConfig `json:"providers" yaml:"providers"`

	// Free tier enforcement
	FreeTierOnly    bool `json:"free_tier_only" yaml:"free_tier_only"`
	MaxPromptLength int  `json:"max_prompt_length" yaml:"max_prompt_length"`
	MaxTokensPerReq int  `json:"max_tokens_per_request" yaml:"max_tokens_per_request"`

	// Request defaults
	DefaultModel       string        `json:"default_model" yaml:"default_model"`
	DefaultTemperature float64       `json:"default_temperature" yaml:"default_temperature"`
	DefaultMaxTokens   int           `json:"default_max_tokens" yaml:"default_max_tokens"`
	DefaultTimeout     time.Duration `json:"default_timeout" yaml:"default_timeout"`

	// Health check configuration
	HealthCheckInterval time.Duration `json:"health_check_interval" yaml:"health_check_interval"`
	HealthCheckTimeout  time.Duration `json:"health_check_timeout" yaml:"health_check_timeout"`
}

// ServiceConfig configures AI service behavior
type ServiceConfig struct {
	// Generation settings
	EnableValidation bool          `json:"enable_validation" yaml:"enable_validation"`
	EnableCaching    bool          `json:"enable_caching" yaml:"enable_caching"`
	CacheTTL         time.Duration `json:"cache_ttl" yaml:"cache_ttl"`

	// Monitoring and observability
	EnableMetrics bool   `json:"enable_metrics" yaml:"enable_metrics"`
	EnableTracing bool   `json:"enable_tracing" yaml:"enable_tracing"`
	MetricsPrefix string `json:"metrics_prefix" yaml:"metrics_prefix"`

	// Error handling
	MaxRetries            int           `json:"max_retries" yaml:"max_retries"`
	RetryDelay            time.Duration `json:"retry_delay" yaml:"retry_delay"`
	CircuitBreakerEnabled bool          `json:"circuit_breaker_enabled" yaml:"circuit_breaker_enabled"`
}

// DefaultConfig returns the default AI service configuration
func DefaultConfig() Config {
	return Config{
		RateLimit: RateLimitConfig{
			RequestsPerMinute: 60,
			BurstSize:         10,
			CleanupInterval:   10 * time.Minute,
		},
		Quota: QuotaConfig{
			DefaultDailyLimit: 100,
			PremiumDailyLimit: 1000,
			ResetTime:         "00:00",
			TrackingEnabled:   true,
			CleanupInterval:   time.Hour,
		},
		LLM: LLMConfig{
			DefaultProvider:     "openai",
			FreeTierOnly:        true,
			MaxPromptLength:     8000,
			MaxTokensPerReq:     2000,
			DefaultModel:        "gpt-3.5-turbo",
			DefaultTemperature:  0.7,
			DefaultMaxTokens:    1000,
			DefaultTimeout:      30 * time.Second,
			HealthCheckInterval: 5 * time.Minute,
			HealthCheckTimeout:  10 * time.Second,
			Providers: map[string]llm.ProviderConfig{
				"openai": {
					FreeTierOnly: true,
					Model:        "gpt-3.5-turbo",
					Timeout:      30 * time.Second,
				},
				"vllm": {
					BaseURL:      "http://localhost:8000",
					FreeTierOnly: true,
					Model:        "codellama-7b",
					Timeout:      60 * time.Second,
				},
			},
		},
		Cache: DefaultCacheConfig(),
		Service: ServiceConfig{
			EnableValidation:      true,
			EnableCaching:         true,
			CacheTTL:              30 * time.Minute,
			EnableMetrics:         true,
			EnableTracing:         true,
			MetricsPrefix:         "ai_service",
			MaxRetries:            3,
			RetryDelay:            time.Second,
			CircuitBreakerEnabled: true,
		},
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.RateLimit.RequestsPerMinute <= 0 {
		return fmt.Errorf("rate limit requests per minute must be positive")
	}
	if c.RateLimit.BurstSize <= 0 {
		return fmt.Errorf("rate limit burst size must be positive")
	}
	if c.Quota.DefaultDailyLimit <= 0 {
		return fmt.Errorf("default daily limit must be positive")
	}
	if c.LLM.DefaultProvider == "" {
		return fmt.Errorf("default LLM provider must be specified")
	}
	if c.LLM.MaxPromptLength <= 0 {
		return fmt.Errorf("max prompt length must be positive")
	}
	if c.LLM.DefaultTemperature < 0 || c.LLM.DefaultTemperature > 2.0 {
		return fmt.Errorf("default temperature must be between 0 and 2.0")
	}
	if c.Service.MaxRetries < 0 {
		return fmt.Errorf("max retries cannot be negative")
	}

	// Validate provider configurations
	if _, exists := c.LLM.Providers[c.LLM.DefaultProvider]; !exists {
		return fmt.Errorf("default provider %s not found in providers configuration", c.LLM.DefaultProvider)
	}

	return nil
}

// GetProviderConfig returns configuration for a specific provider
func (c *Config) GetProviderConfig(providerName string) (llm.ProviderConfig, bool) {
	config, exists := c.LLM.Providers[providerName]
	return config, exists
}
