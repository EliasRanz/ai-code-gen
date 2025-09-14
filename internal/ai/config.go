package ai

import (
	"context"
	"fmt"
	"time"

	"github.com/EliasRanz/ai-code-gen/internal/ai/llm"
	"github.com/EliasRanz/ai-code-gen/internal/config"
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

// AIServiceConfig holds enhanced AI service configuration following new pattern
type AIServiceConfig struct {
	Service       config.BaseServiceConfig `json:"service" yaml:"service"`
	Database      config.DatabaseConfig    `json:"database" yaml:"database"`
	Redis         config.RedisConfig       `json:"redis" yaml:"redis"`
	LLM           LLMConfig                `json:"llm" yaml:"llm"`
	RateLimit     RateLimitConfig          `json:"rate_limit" yaml:"rate_limit"`
	Quota         QuotaConfig              `json:"quota" yaml:"quota"`
	Cache         CacheConfig              `json:"cache" yaml:"cache"`
	Behavior      ServiceConfig            `json:"behavior" yaml:"behavior"`
	Logging       LoggingConfig            `json:"logging" yaml:"logging"`
	Observability ObservabilityConfig      `json:"observability" yaml:"observability"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string `json:"level" yaml:"level"`
	Format string `json:"format" yaml:"format"`
}

// ObservabilityConfig holds observability configuration
type ObservabilityConfig struct {
	MetricsEnabled bool   `json:"metrics_enabled" yaml:"metrics_enabled"`
	TracingEnabled bool   `json:"tracing_enabled" yaml:"tracing_enabled"`
	JaegerEndpoint string `json:"jaeger_endpoint" yaml:"jaeger_endpoint"`
}

// AIConfigManager manages AI service configuration
type AIConfigManager struct {
	manager config.ConfigManager
	config  *AIServiceConfig
}

// NewAIConfigManager creates a new AI configuration manager
func NewAIConfigManager(provider config.ConfigProvider) *AIConfigManager {
	manager := config.NewConfigManager(provider)

	// Add validation rules for AI configuration
	validator := config.NewConfigValidator()
	if err := addAIValidationRules(validator); err != nil {
		// Log error and continue with degraded validation
		// In production, proper logging framework would be used
		fmt.Printf("Warning: Failed to add AI validation rules: %v\n", err)
	}

	return &AIConfigManager{
		manager: manager,
		config:  &AIServiceConfig{},
	}
}

// LoadConfig loads and validates AI service configuration
func (m *AIConfigManager) LoadConfig(ctx context.Context) error {
	if err := m.manager.LoadConfig(ctx); err != nil {
		return fmt.Errorf("failed to load AI configuration: %w", err)
	}

	// Map configuration to struct
	if err := m.mapConfig(); err != nil {
		return fmt.Errorf("failed to map AI configuration: %w", err)
	}

	// Apply defaults
	m.applyDefaults()

	return nil
}

// GetConfig returns the current AI service configuration
func (m *AIConfigManager) GetConfig() *AIServiceConfig {
	return m.config
}

// Watch watches for configuration changes
func (m *AIConfigManager) Watch(ctx context.Context, callback func()) error {
	return m.manager.Watch(ctx, func() {
		if err := m.mapConfig(); err == nil {
			m.applyDefaults()
			callback()
		}
	})
}

// Reload reloads the configuration
func (m *AIConfigManager) Reload(ctx context.Context) error {
	return m.LoadConfig(ctx)
}

// mapConfig maps raw configuration data to AI config struct
func (m *AIConfigManager) mapConfig() error {
	// Service configuration
	m.config.Service.Name = m.manager.GetString("service.name")
	m.config.Service.Host = m.manager.GetString("service.host")
	m.config.Service.Port = m.manager.GetInt("service.port")
	m.config.Service.Environment = m.manager.GetString("service.environment")
	m.config.Service.Version = m.manager.GetString("service.version")

	// LLM configuration
	m.config.LLM.DefaultProvider = m.manager.GetString("llm.default.provider")
	m.config.LLM.FreeTierOnly = m.manager.GetBool("llm.free.tier.only")
	m.config.LLM.MaxPromptLength = m.manager.GetInt("llm.max.prompt.length")
	m.config.LLM.MaxTokensPerReq = m.manager.GetInt("llm.max.tokens.per.request")
	m.config.LLM.DefaultModel = m.manager.GetString("llm.default.model")
	m.config.LLM.DefaultTemperature = m.manager.GetFloat64("llm.default.temperature")
	m.config.LLM.DefaultMaxTokens = m.manager.GetInt("llm.default.max.tokens")
	m.config.LLM.DefaultTimeout = m.manager.GetDuration("llm.default.timeout")

	// Rate limiting configuration
	m.config.RateLimit.RequestsPerMinute = m.manager.GetInt("rate.limit.requests.per.minute")
	m.config.RateLimit.BurstSize = m.manager.GetInt("rate.limit.burst.size")
	m.config.RateLimit.CleanupInterval = m.manager.GetDuration("rate.limit.cleanup.interval")

	// Quota configuration
	m.config.Quota.DefaultDailyLimit = m.manager.GetInt("quota.default.daily.limit")
	m.config.Quota.PremiumDailyLimit = m.manager.GetInt("quota.premium.daily.limit")
	m.config.Quota.ResetTime = m.manager.GetString("quota.reset.time")
	m.config.Quota.TrackingEnabled = m.manager.GetBool("quota.tracking.enabled")
	m.config.Quota.CleanupInterval = m.manager.GetDuration("quota.cleanup.interval")

	return nil
}

// applyDefaults applies default values to configuration
func (m *AIConfigManager) applyDefaults() {
	if m.config.Service.Name == "" {
		m.config.Service.Name = "ai-service"
	}
	if m.config.Service.Host == "" {
		m.config.Service.Host = "0.0.0.0"
	}
	if m.config.Service.Port == 0 {
		m.config.Service.Port = 8083
	}
	if m.config.Service.Environment == "" {
		m.config.Service.Environment = "development"
	}

	if m.config.LLM.DefaultProvider == "" {
		m.config.LLM.DefaultProvider = "openai"
	}
	if m.config.LLM.MaxPromptLength == 0 {
		m.config.LLM.MaxPromptLength = 10000
	}
	if m.config.LLM.MaxTokensPerReq == 0 {
		m.config.LLM.MaxTokensPerReq = 4096
	}
	if m.config.LLM.DefaultModel == "" {
		m.config.LLM.DefaultModel = "gpt-3.5-turbo"
	}
	if m.config.LLM.DefaultTemperature == 0 {
		m.config.LLM.DefaultTemperature = 0.7
	}
	if m.config.LLM.DefaultMaxTokens == 0 {
		m.config.LLM.DefaultMaxTokens = 4096
	}
	if m.config.LLM.DefaultTimeout == 0 {
		m.config.LLM.DefaultTimeout = 30 * time.Second
	}

	if m.config.RateLimit.RequestsPerMinute == 0 {
		m.config.RateLimit.RequestsPerMinute = 60
	}
	if m.config.RateLimit.BurstSize == 0 {
		m.config.RateLimit.BurstSize = 10
	}
	if m.config.RateLimit.CleanupInterval == 0 {
		m.config.RateLimit.CleanupInterval = 10 * time.Minute
	}

	if m.config.Quota.DefaultDailyLimit == 0 {
		m.config.Quota.DefaultDailyLimit = 100
	}
	if m.config.Quota.PremiumDailyLimit == 0 {
		m.config.Quota.PremiumDailyLimit = 1000
	}
	if m.config.Quota.ResetTime == "" {
		m.config.Quota.ResetTime = "00:00"
	}
	if m.config.Quota.CleanupInterval == 0 {
		m.config.Quota.CleanupInterval = time.Hour
	}
}

// addAIValidationRules adds validation rules for AI configuration
func addAIValidationRules(validator config.ConfigValidator) error {
	// Service port validation
	if err := validator.AddRule(config.ValidationRule{
		Key:      "service.port",
		Type:     "int",
		MinValue: 1,
		MaxValue: 65535,
	}); err != nil {
		return fmt.Errorf("failed to add service.port validation rule: %w", err)
	}

	// LLM validation
	if err := validator.AddRule(config.ValidationRule{
		Key:      "llm.default_provider",
		Required: true,
		Type:     "string",
	}); err != nil {
		return fmt.Errorf("failed to add llm.default_provider validation rule: %w", err)
	}

	if err := validator.AddRule(config.ValidationRule{
		Key:      "llm.max_prompt_length",
		Type:     "int",
		MinValue: 100,
		MaxValue: 100000,
	}); err != nil {
		return fmt.Errorf("failed to add llm.max_prompt_length validation rule: %w", err)
	}

	if err := validator.AddRule(config.ValidationRule{
		Key:      "llm.default_temperature",
		Type:     "float",
		MinValue: 0.0,
		MaxValue: 2.0,
	}); err != nil {
		return fmt.Errorf("failed to add llm.default_temperature validation rule: %w", err)
	}

	// Rate limiting validation
	if err := validator.AddRule(config.ValidationRule{
		Key:      "rate_limit.requests_per_minute",
		Type:     "int",
		MinValue: 1,
		MaxValue: 10000,
	}); err != nil {
		return fmt.Errorf("failed to add rate_limit.requests_per_minute validation rule: %w", err)
	}

	if err := validator.AddRule(config.ValidationRule{
		Key:      "rate_limit.burst_size",
		Type:     "int",
		MinValue: 1,
		MaxValue: 1000,
	}); err != nil {
		return fmt.Errorf("failed to add rate_limit.burst_size validation rule: %w", err)
	}

	// Quota validation
	if err := validator.AddRule(config.ValidationRule{
		Key:      "quota.default_daily_limit",
		Type:     "int",
		MinValue: 1,
		MaxValue: 100000,
	}); err != nil {
		return fmt.Errorf("failed to add quota.default_daily_limit validation rule: %w", err)
	}

	return nil
}
