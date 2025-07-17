package gateway

import (
	"context"
	"fmt"
	"time"

	"github.com/EliasRanz/ai-code-gen/internal/config"
)

// GatewayServiceConfig holds gateway service configuration
type GatewayServiceConfig struct {
	Service       config.BaseServiceConfig `json:"service" yaml:"service"`
	Routing       RoutingConfig            `json:"routing" yaml:"routing"`
	RateLimit     RateLimitConfig          `json:"rate_limit" yaml:"rate_limit"`
	CORS          CORSConfig               `json:"cors" yaml:"cors"`
	Auth          AuthProxyConfig          `json:"auth" yaml:"auth"`
	LoadBalancer  LoadBalancerConfig       `json:"load_balancer" yaml:"load_balancer"`
	Logging       LoggingConfig            `json:"logging" yaml:"logging"`
	Observability ObservabilityConfig      `json:"observability" yaml:"observability"`
}

// RoutingConfig holds routing configuration
type RoutingConfig struct {
	Routes        []RouteConfig `json:"routes" yaml:"routes"`
	DefaultRoute  string        `json:"default_route" yaml:"default_route"`
	StripPrefixes []string      `json:"strip_prefixes" yaml:"strip_prefixes"`
}

// RouteConfig defines a routing rule
type RouteConfig struct {
	Path        string   `json:"path" yaml:"path"`
	Target      string   `json:"target" yaml:"target"`
	Methods     []string `json:"methods" yaml:"methods"`
	StripPrefix bool     `json:"strip_prefix" yaml:"strip_prefix"`
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled         bool          `json:"enabled" yaml:"enabled"`
	RequestsPerSec  int           `json:"requests_per_second" yaml:"requests_per_second"`
	BurstSize       int           `json:"burst_size" yaml:"burst_size"`
	CleanupInterval time.Duration `json:"cleanup_interval" yaml:"cleanup_interval"`
	IPWhitelist     []string      `json:"ip_whitelist" yaml:"ip_whitelist"`
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	Enabled          bool     `json:"enabled" yaml:"enabled"`
	AllowedOrigins   []string `json:"allowed_origins" yaml:"allowed_origins"`
	AllowedMethods   []string `json:"allowed_methods" yaml:"allowed_methods"`
	AllowedHeaders   []string `json:"allowed_headers" yaml:"allowed_headers"`
	ExposedHeaders   []string `json:"exposed_headers" yaml:"exposed_headers"`
	AllowCredentials bool     `json:"allow_credentials" yaml:"allow_credentials"`
	MaxAge           int      `json:"max_age" yaml:"max_age"`
}

// AuthProxyConfig holds authentication proxy configuration
type AuthProxyConfig struct {
	Enabled        bool          `json:"enabled" yaml:"enabled"`
	AuthServiceURL string        `json:"auth_service_url" yaml:"auth_service_url"`
	ExcludePaths   []string      `json:"exclude_paths" yaml:"exclude_paths"`
	TokenHeader    string        `json:"token_header" yaml:"token_header"`
	Timeout        time.Duration `json:"timeout" yaml:"timeout"`
}

// LoadBalancerConfig holds load balancing configuration
type LoadBalancerConfig struct {
	Strategy       string         `json:"strategy" yaml:"strategy"`
	HealthCheck    HealthCheck    `json:"health_check" yaml:"health_check"`
	RetryPolicy    RetryPolicy    `json:"retry_policy" yaml:"retry_policy"`
	CircuitBreaker CircuitBreaker `json:"circuit_breaker" yaml:"circuit_breaker"`
}

// HealthCheck configuration
type HealthCheck struct {
	Enabled  bool          `json:"enabled" yaml:"enabled"`
	Interval time.Duration `json:"interval" yaml:"interval"`
	Timeout  time.Duration `json:"timeout" yaml:"timeout"`
	Path     string        `json:"path" yaml:"path"`
}

// RetryPolicy configuration
type RetryPolicy struct {
	MaxRetries int           `json:"max_retries" yaml:"max_retries"`
	Delay      time.Duration `json:"delay" yaml:"delay"`
	MaxDelay   time.Duration `json:"max_delay" yaml:"max_delay"`
}

// CircuitBreaker configuration
type CircuitBreaker struct {
	Enabled                bool          `json:"enabled" yaml:"enabled"`
	FailureThreshold       int           `json:"failure_threshold" yaml:"failure_threshold"`
	RecoveryTimeout        time.Duration `json:"recovery_timeout" yaml:"recovery_timeout"`
	RequestVolumeThreshold int           `json:"request_volume_threshold" yaml:"request_volume_threshold"`
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

// GatewayConfigManager manages gateway service configuration
type GatewayConfigManager struct {
	manager config.ConfigManager
	config  *GatewayServiceConfig
}

// NewGatewayConfigManager creates a new gateway configuration manager
func NewGatewayConfigManager(provider config.ConfigProvider) *GatewayConfigManager {
	manager := config.NewConfigManager(provider)

	// Add validation rules for gateway configuration
	validator := config.NewConfigValidator()
	addGatewayValidationRules(validator)

	return &GatewayConfigManager{
		manager: manager,
		config:  &GatewayServiceConfig{},
	}
}

// LoadConfig loads and validates gateway service configuration
func (m *GatewayConfigManager) LoadConfig(ctx context.Context) error {
	if err := m.manager.LoadConfig(ctx); err != nil {
		return fmt.Errorf("failed to load gateway configuration: %w", err)
	}

	// Map configuration to struct
	if err := m.mapConfig(); err != nil {
		return fmt.Errorf("failed to map gateway configuration: %w", err)
	}

	// Apply defaults
	m.applyDefaults()

	return nil
}

// GetConfig returns the current gateway service configuration
func (m *GatewayConfigManager) GetConfig() *GatewayServiceConfig {
	return m.config
}

// Watch watches for configuration changes
func (m *GatewayConfigManager) Watch(ctx context.Context, callback func()) error {
	return m.manager.Watch(ctx, func() {
		if err := m.mapConfig(); err == nil {
			m.applyDefaults()
			callback()
		}
	})
}

// Reload reloads the configuration
func (m *GatewayConfigManager) Reload(ctx context.Context) error {
	return m.LoadConfig(ctx)
}

// mapConfig maps raw configuration data to gateway config struct
func (m *GatewayConfigManager) mapConfig() error {
	// Service configuration
	m.config.Service.Name = m.manager.GetString("service.name")
	m.config.Service.Host = m.manager.GetString("service.host")
	m.config.Service.Port = m.manager.GetInt("service.port")
	m.config.Service.Environment = m.manager.GetString("service.environment")
	m.config.Service.Version = m.manager.GetString("service.version")

	// Rate limiting configuration
	m.config.RateLimit.Enabled = m.manager.GetBool("rate.limit.enabled")
	m.config.RateLimit.RequestsPerSec = m.manager.GetInt("rate.limit.requests.per.second")
	m.config.RateLimit.BurstSize = m.manager.GetInt("rate.limit.burst.size")
	m.config.RateLimit.CleanupInterval = m.manager.GetDuration("rate.limit.cleanup.interval")
	m.config.RateLimit.IPWhitelist = m.manager.GetStringSlice("rate.limit.ip.whitelist")

	// CORS configuration
	m.config.CORS.Enabled = m.manager.GetBool("cors.enabled")
	m.config.CORS.AllowedOrigins = m.manager.GetStringSlice("cors.allowed.origins")
	m.config.CORS.AllowedMethods = m.manager.GetStringSlice("cors.allowed.methods")
	m.config.CORS.AllowedHeaders = m.manager.GetStringSlice("cors.allowed.headers")
	m.config.CORS.AllowCredentials = m.manager.GetBool("cors.allow.credentials")
	m.config.CORS.MaxAge = m.manager.GetInt("cors.max.age")

	// Auth proxy configuration
	m.config.Auth.Enabled = m.manager.GetBool("auth.enabled")
	m.config.Auth.AuthServiceURL = m.manager.GetString("auth.auth.service.url")
	m.config.Auth.ExcludePaths = m.manager.GetStringSlice("auth.exclude.paths")
	m.config.Auth.TokenHeader = m.manager.GetString("auth.token.header")
	m.config.Auth.Timeout = m.manager.GetDuration("auth.timeout")

	return nil
}

// applyDefaults applies default values to configuration
func (m *GatewayConfigManager) applyDefaults() {
	if m.config.Service.Name == "" {
		m.config.Service.Name = "api-gateway"
	}
	if m.config.Service.Host == "" {
		m.config.Service.Host = "0.0.0.0"
	}
	if m.config.Service.Port == 0 {
		m.config.Service.Port = 8080
	}
	if m.config.Service.Environment == "" {
		m.config.Service.Environment = "development"
	}

	if m.config.RateLimit.RequestsPerSec == 0 {
		m.config.RateLimit.RequestsPerSec = 100
	}
	if m.config.RateLimit.BurstSize == 0 {
		m.config.RateLimit.BurstSize = 20
	}
	if m.config.RateLimit.CleanupInterval == 0 {
		m.config.RateLimit.CleanupInterval = 5 * time.Minute
	}

	if len(m.config.CORS.AllowedOrigins) == 0 {
		m.config.CORS.AllowedOrigins = []string{"*"}
	}
	if len(m.config.CORS.AllowedMethods) == 0 {
		m.config.CORS.AllowedMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	}
	if len(m.config.CORS.AllowedHeaders) == 0 {
		m.config.CORS.AllowedHeaders = []string{"*"}
	}
	if m.config.CORS.MaxAge == 0 {
		m.config.CORS.MaxAge = 86400 // 24 hours
	}

	if m.config.Auth.TokenHeader == "" {
		m.config.Auth.TokenHeader = "Authorization"
	}
	if m.config.Auth.Timeout == 0 {
		m.config.Auth.Timeout = 5 * time.Second
	}
}

// addGatewayValidationRules adds validation rules for gateway configuration
func addGatewayValidationRules(validator config.ConfigValidator) {
	// Service port validation
	validator.AddRule(config.ValidationRule{
		Key:      "service.port",
		Type:     "int",
		MinValue: 1,
		MaxValue: 65535,
	})

	// Rate limiting validation
	validator.AddRule(config.ValidationRule{
		Key:      "rate_limit.requests_per_second",
		Type:     "int",
		MinValue: 1,
		MaxValue: 10000,
	})

	validator.AddRule(config.ValidationRule{
		Key:      "rate_limit.burst_size",
		Type:     "int",
		MinValue: 1,
		MaxValue: 1000,
	})

	// CORS validation
	validator.AddRule(config.ValidationRule{
		Key:      "cors.max_age",
		Type:     "int",
		MinValue: 0,
		MaxValue: 86400 * 7, // 7 days max
	})
}
