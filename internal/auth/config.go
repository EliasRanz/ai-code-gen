package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/EliasRanz/ai-code-gen/internal/config"
)

// AuthServiceConfig holds authentication service configuration
type AuthServiceConfig struct {
	Service       config.BaseServiceConfig `json:"service" yaml:"service"`
	JWT           JWTConfig                `json:"jwt" yaml:"jwt"`
	Session       SessionConfig            `json:"session" yaml:"session"`
	OAuth         OAuthConfig              `json:"oauth" yaml:"oauth"`
	Security      SecurityConfig           `json:"security" yaml:"security"`
	Database      config.DatabaseConfig    `json:"database" yaml:"database"`
	Redis         config.RedisConfig       `json:"redis" yaml:"redis"`
	Logging       LoggingConfig            `json:"logging" yaml:"logging"`
	Observability ObservabilityConfig      `json:"observability" yaml:"observability"`
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret               string        `json:"secret" yaml:"secret"`
	AccessTokenDuration  time.Duration `json:"access_token_duration" yaml:"access_token_duration"`
	RefreshTokenDuration time.Duration `json:"refresh_token_duration" yaml:"refresh_token_duration"`
	Issuer               string        `json:"issuer" yaml:"issuer"`
	Audience             string        `json:"audience" yaml:"audience"`
	Algorithm            string        `json:"algorithm" yaml:"algorithm"`
}

// SessionConfig holds session configuration
type SessionConfig struct {
	Duration     time.Duration `json:"duration" yaml:"duration"`
	CookieName   string        `json:"cookie_name" yaml:"cookie_name"`
	CookieDomain string        `json:"cookie_domain" yaml:"cookie_domain"`
	Secure       bool          `json:"secure" yaml:"secure"`
	HTTPOnly     bool          `json:"http_only" yaml:"http_only"`
	SameSite     string        `json:"same_site" yaml:"same_site"`
}

// OAuthConfig holds OAuth configuration
type OAuthConfig struct {
	Google GoogleOAuthConfig `json:"google" yaml:"google"`
}

// GoogleOAuthConfig holds Google OAuth configuration
type GoogleOAuthConfig struct {
	ClientID     string   `json:"client_id" yaml:"client_id"`
	ClientSecret string   `json:"client_secret" yaml:"client_secret"`
	RedirectURL  string   `json:"redirect_url" yaml:"redirect_url"`
	Scopes       []string `json:"scopes" yaml:"scopes"`
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	PasswordMinLength      int           `json:"password_min_length" yaml:"password_min_length"`
	PasswordRequireSpecial bool          `json:"password_require_special" yaml:"password_require_special"`
	MaxLoginAttempts       int           `json:"max_login_attempts" yaml:"max_login_attempts"`
	LockoutDuration        time.Duration `json:"lockout_duration" yaml:"lockout_duration"`
	BcryptCost             int           `json:"bcrypt_cost" yaml:"bcrypt_cost"`
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

// AuthConfigManager manages auth service configuration
type AuthConfigManager struct {
	manager config.ConfigManager
	config  *AuthServiceConfig
}

// NewAuthConfigManager creates a new auth configuration manager
func NewAuthConfigManager(provider config.ConfigProvider) *AuthConfigManager {
	manager := config.NewConfigManager(provider)

	// Add validation rules for auth configuration
	validator := config.NewConfigValidator()
	addAuthValidationRules(validator)

	return &AuthConfigManager{
		manager: manager,
		config:  &AuthServiceConfig{},
	}
}

// LoadConfig loads and validates auth service configuration
func (m *AuthConfigManager) LoadConfig(ctx context.Context) error {
	if err := m.manager.LoadConfig(ctx); err != nil {
		return fmt.Errorf("failed to load auth configuration: %w", err)
	}

	// Map configuration to struct
	if err := m.mapConfig(); err != nil {
		return fmt.Errorf("failed to map auth configuration: %w", err)
	}

	// Apply defaults
	m.applyDefaults()

	return nil
}

// GetConfig returns the current auth service configuration
func (m *AuthConfigManager) GetConfig() *AuthServiceConfig {
	return m.config
}

// Watch watches for configuration changes
func (m *AuthConfigManager) Watch(ctx context.Context, callback func()) error {
	return m.manager.Watch(ctx, func() {
		if err := m.mapConfig(); err == nil {
			m.applyDefaults()
			callback()
		}
	})
}

// Reload reloads the configuration
func (m *AuthConfigManager) Reload(ctx context.Context) error {
	return m.LoadConfig(ctx)
}

// mapConfig maps raw configuration data to auth config struct
func (m *AuthConfigManager) mapConfig() error {
	// Service configuration
	m.config.Service.Name = m.manager.GetString("service.name")
	m.config.Service.Host = m.manager.GetString("service.host")
	m.config.Service.Port = m.manager.GetInt("service.port")
	m.config.Service.Environment = m.manager.GetString("service.environment")
	m.config.Service.Version = m.manager.GetString("service.version")

	// JWT configuration
	m.config.JWT.Secret = m.manager.GetString("jwt.secret")
	m.config.JWT.AccessTokenDuration = m.manager.GetDuration("jwt.access.token.duration")
	m.config.JWT.RefreshTokenDuration = m.manager.GetDuration("jwt.refresh.token.duration")
	m.config.JWT.Issuer = m.manager.GetString("jwt.issuer")
	m.config.JWT.Audience = m.manager.GetString("jwt.audience")
	m.config.JWT.Algorithm = m.manager.GetString("jwt.algorithm")

	// Session configuration
	m.config.Session.Duration = m.manager.GetDuration("session.duration")
	m.config.Session.CookieName = m.manager.GetString("session.cookie.name")
	m.config.Session.CookieDomain = m.manager.GetString("session.cookie.domain")
	m.config.Session.Secure = m.manager.GetBool("session.secure")
	m.config.Session.HTTPOnly = m.manager.GetBool("session.http.only")
	m.config.Session.SameSite = m.manager.GetString("session.same.site")

	// OAuth configuration
	m.config.OAuth.Google.ClientID = m.manager.GetString("oauth.google.client.id")
	m.config.OAuth.Google.ClientSecret = m.manager.GetString("oauth.google.client.secret")
	m.config.OAuth.Google.RedirectURL = m.manager.GetString("oauth.google.redirect.url")
	m.config.OAuth.Google.Scopes = m.manager.GetStringSlice("oauth.google.scopes")

	// Security configuration
	m.config.Security.PasswordMinLength = m.manager.GetInt("security.password.min.length")
	m.config.Security.PasswordRequireSpecial = m.manager.GetBool("security.password.require.special")
	m.config.Security.MaxLoginAttempts = m.manager.GetInt("security.max.login.attempts")
	m.config.Security.LockoutDuration = m.manager.GetDuration("security.lockout.duration")
	m.config.Security.BcryptCost = m.manager.GetInt("security.bcrypt.cost")

	return nil
}

// applyDefaults applies default values to configuration
func (m *AuthConfigManager) applyDefaults() {
	if m.config.Service.Name == "" {
		m.config.Service.Name = "auth-service"
	}
	if m.config.Service.Host == "" {
		m.config.Service.Host = "0.0.0.0"
	}
	if m.config.Service.Port == 0 {
		m.config.Service.Port = 8081
	}
	if m.config.Service.Environment == "" {
		m.config.Service.Environment = "development"
	}

	if m.config.JWT.AccessTokenDuration == 0 {
		m.config.JWT.AccessTokenDuration = 15 * time.Minute
	}
	if m.config.JWT.RefreshTokenDuration == 0 {
		m.config.JWT.RefreshTokenDuration = 7 * 24 * time.Hour
	}
	if m.config.JWT.Algorithm == "" {
		m.config.JWT.Algorithm = "HS256"
	}

	if m.config.Session.Duration == 0 {
		m.config.Session.Duration = 24 * time.Hour
	}
	if m.config.Session.CookieName == "" {
		m.config.Session.CookieName = "auth_session"
	}
	if m.config.Session.SameSite == "" {
		m.config.Session.SameSite = "Lax"
	}

	if len(m.config.OAuth.Google.Scopes) == 0 {
		m.config.OAuth.Google.Scopes = []string{"openid", "email", "profile"}
	}

	if m.config.Security.PasswordMinLength == 0 {
		m.config.Security.PasswordMinLength = 8
	}
	if m.config.Security.MaxLoginAttempts == 0 {
		m.config.Security.MaxLoginAttempts = 5
	}
	if m.config.Security.LockoutDuration == 0 {
		m.config.Security.LockoutDuration = 30 * time.Minute
	}
	if m.config.Security.BcryptCost == 0 {
		m.config.Security.BcryptCost = 12
	}
}

// addAuthValidationRules adds validation rules for auth configuration
func addAuthValidationRules(validator config.ConfigValidator) {
	// Required fields
	validator.AddRule(config.ValidationRule{
		Key:      "jwt.secret",
		Required: true,
		Type:     "string",
	})

	// Port validation
	validator.AddRule(config.ValidationRule{
		Key:      "service.port",
		Type:     "int",
		MinValue: 1,
		MaxValue: 65535,
	})

	// Security validations
	validator.AddRule(config.ValidationRule{
		Key:      "security.password_min_length",
		Type:     "int",
		MinValue: 6,
		MaxValue: 50,
	})

	validator.AddRule(config.ValidationRule{
		Key:      "security.max_login_attempts",
		Type:     "int",
		MinValue: 1,
		MaxValue: 20,
	})

	validator.AddRule(config.ValidationRule{
		Key:      "security.bcrypt_cost",
		Type:     "int",
		MinValue: 4,
		MaxValue: 31,
	})
}
