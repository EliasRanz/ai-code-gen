package user

import (
	"context"
	"fmt"
	"time"

	"github.com/EliasRanz/ai-code-gen/internal/config"
)

// UserServiceConfig holds user service configuration
type UserServiceConfig struct {
	Service       config.BaseServiceConfig `json:"service" yaml:"service"`
	Database      config.DatabaseConfig    `json:"database" yaml:"database"`
	Redis         config.RedisConfig       `json:"redis" yaml:"redis"`
	Pagination    PaginationConfig         `json:"pagination" yaml:"pagination"`
	Validation    ValidationConfig         `json:"validation" yaml:"validation"`
	Logging       LoggingConfig            `json:"logging" yaml:"logging"`
	Observability ObservabilityConfig      `json:"observability" yaml:"observability"`
}

// PaginationConfig holds pagination configuration
type PaginationConfig struct {
	DefaultLimit int `json:"default_limit" yaml:"default_limit"`
	MaxLimit     int `json:"max_limit" yaml:"max_limit"`
}

// ValidationConfig holds validation configuration
type ValidationConfig struct {
	UsernameMinLength int      `json:"username_min_length" yaml:"username_min_length"`
	UsernameMaxLength int      `json:"username_max_length" yaml:"username_max_length"`
	EmailDomains      []string `json:"allowed_email_domains" yaml:"allowed_email_domains"`
	ProjectNameMin    int      `json:"project_name_min_length" yaml:"project_name_min_length"`
	ProjectNameMax    int      `json:"project_name_max_length" yaml:"project_name_max_length"`
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

// UserConfigManager manages user service configuration
type UserConfigManager struct {
	manager config.ConfigManager
	config  *UserServiceConfig
}

// NewUserConfigManager creates a new user configuration manager
func NewUserConfigManager(provider config.ConfigProvider) *UserConfigManager {
	manager := config.NewConfigManager(provider)

	// Add validation rules for user configuration
	validator := config.NewConfigValidator()
	addUserValidationRules(validator)

	return &UserConfigManager{
		manager: manager,
		config:  &UserServiceConfig{},
	}
}

// LoadConfig loads and validates user service configuration
func (m *UserConfigManager) LoadConfig(ctx context.Context) error {
	if err := m.manager.LoadConfig(ctx); err != nil {
		return fmt.Errorf("failed to load user configuration: %w", err)
	}

	// Map configuration to struct
	if err := m.mapConfig(); err != nil {
		return fmt.Errorf("failed to map user configuration: %w", err)
	}

	// Apply defaults
	m.applyDefaults()

	return nil
}

// GetConfig returns the current user service configuration
func (m *UserConfigManager) GetConfig() *UserServiceConfig {
	return m.config
}

// Watch watches for configuration changes
func (m *UserConfigManager) Watch(ctx context.Context, callback func()) error {
	return m.manager.Watch(ctx, func() {
		if err := m.mapConfig(); err == nil {
			m.applyDefaults()
			callback()
		}
	})
}

// Reload reloads the configuration
func (m *UserConfigManager) Reload(ctx context.Context) error {
	return m.LoadConfig(ctx)
}

// mapConfig maps raw configuration data to user config struct
func (m *UserConfigManager) mapConfig() error {
	// Service configuration
	m.config.Service.Name = m.manager.GetString("service.name")
	m.config.Service.Host = m.manager.GetString("service.host")
	m.config.Service.Port = m.manager.GetInt("service.port")
	m.config.Service.Environment = m.manager.GetString("service.environment")
	m.config.Service.Version = m.manager.GetString("service.version")

	// Database configuration
	m.config.Database.Host = m.manager.GetString("database.host")
	m.config.Database.Port = m.manager.GetInt("database.port")
	m.config.Database.User = m.manager.GetString("database.user")
	m.config.Database.Password = m.manager.GetString("database.password")
	m.config.Database.DBName = m.manager.GetString("database.dbname")
	m.config.Database.SSLMode = m.manager.GetString("database.ssl_mode")
	m.config.Database.MaxOpenConns = m.manager.GetInt("database.max_open_conns")
	m.config.Database.MaxIdleConns = m.manager.GetInt("database.max_idle_conns")

	// These are stored as strings in the existing config
	if lifetime := m.manager.GetString("database.conn_max_lifetime"); lifetime != "" {
		m.config.Database.ConnMaxLifetime = lifetime
	}
	if idleTime := m.manager.GetString("database.conn_max_idle_time"); idleTime != "" {
		m.config.Database.ConnMaxIdleTime = idleTime
	}

	// Redis configuration
	m.config.Redis.Host = m.manager.GetString("redis.host")
	m.config.Redis.Port = m.manager.GetInt("redis.port")
	m.config.Redis.Password = m.manager.GetString("redis.password")
	m.config.Redis.DB = m.manager.GetInt("redis.db")
	m.config.Redis.MaxConnections = m.manager.GetInt("redis.max_connections")
	m.config.Redis.MaxIdleConnections = m.manager.GetInt("redis.max_idle_connections")
	m.config.Redis.ConnectionTimeout = m.manager.GetDuration("redis.connection_timeout")
	m.config.Redis.IdleTimeout = m.manager.GetDuration("redis.idle_timeout")

	// Pagination configuration
	m.config.Pagination.DefaultLimit = m.manager.GetInt("pagination.default_limit")
	m.config.Pagination.MaxLimit = m.manager.GetInt("pagination.max_limit")

	// Validation configuration
	m.config.Validation.UsernameMinLength = m.manager.GetInt("validation.username_min_length")
	m.config.Validation.UsernameMaxLength = m.manager.GetInt("validation.username_max_length")
	m.config.Validation.EmailDomains = m.manager.GetStringSlice("validation.allowed_email_domains")
	m.config.Validation.ProjectNameMin = m.manager.GetInt("validation.project_name_min_length")
	m.config.Validation.ProjectNameMax = m.manager.GetInt("validation.project_name_max_length")

	return nil
}

// applyDefaults applies default values to configuration
func (m *UserConfigManager) applyDefaults() {
	if m.config.Service.Name == "" {
		m.config.Service.Name = "user-service"
	}
	if m.config.Service.Host == "" {
		m.config.Service.Host = "0.0.0.0"
	}
	if m.config.Service.Port == 0 {
		m.config.Service.Port = 8082
	}
	if m.config.Service.Environment == "" {
		m.config.Service.Environment = "development"
	}

	if m.config.Database.Host == "" {
		m.config.Database.Host = "localhost"
	}
	if m.config.Database.Port == 0 {
		m.config.Database.Port = 5432
	}
	if m.config.Database.User == "" {
		m.config.Database.User = "postgres"
	}
	if m.config.Database.DBName == "" {
		m.config.Database.DBName = "ai_ui_generator"
	}
	if m.config.Database.SSLMode == "" {
		m.config.Database.SSLMode = "disable"
	}
	if m.config.Database.MaxOpenConns == 0 {
		m.config.Database.MaxOpenConns = 25
	}
	if m.config.Database.MaxIdleConns == 0 {
		m.config.Database.MaxIdleConns = 5
	}
	if m.config.Database.ConnMaxLifetime == "" {
		m.config.Database.ConnMaxLifetime = "5m"
	}
	if m.config.Database.ConnMaxIdleTime == "" {
		m.config.Database.ConnMaxIdleTime = "1m"
	}

	if m.config.Redis.Host == "" {
		m.config.Redis.Host = "localhost"
	}
	if m.config.Redis.Port == 0 {
		m.config.Redis.Port = 6379
	}
	if m.config.Redis.MaxConnections == 0 {
		m.config.Redis.MaxConnections = 10
	}
	if m.config.Redis.MaxIdleConnections == 0 {
		m.config.Redis.MaxIdleConnections = 3
	}
	if m.config.Redis.ConnectionTimeout == 0 {
		m.config.Redis.ConnectionTimeout = 5 * time.Second
	}
	if m.config.Redis.IdleTimeout == 0 {
		m.config.Redis.IdleTimeout = 300 * time.Second
	}

	if m.config.Pagination.DefaultLimit == 0 {
		m.config.Pagination.DefaultLimit = 20
	}
	if m.config.Pagination.MaxLimit == 0 {
		m.config.Pagination.MaxLimit = 100
	}

	if m.config.Validation.UsernameMinLength == 0 {
		m.config.Validation.UsernameMinLength = 3
	}
	if m.config.Validation.UsernameMaxLength == 0 {
		m.config.Validation.UsernameMaxLength = 50
	}
	if m.config.Validation.ProjectNameMin == 0 {
		m.config.Validation.ProjectNameMin = 3
	}
	if m.config.Validation.ProjectNameMax == 0 {
		m.config.Validation.ProjectNameMax = 100
	}
}

// addUserValidationRules adds validation rules for user configuration
func addUserValidationRules(validator config.ConfigValidator) {
	// Port validation
	validator.AddRule(config.ValidationRule{
		Key:      "service.port",
		Type:     "int",
		MinValue: 1,
		MaxValue: 65535,
	})

	// Database validation
	validator.AddRule(config.ValidationRule{
		Key:      "database.host",
		Required: true,
		Type:     "string",
	})

	validator.AddRule(config.ValidationRule{
		Key:      "database.port",
		Type:     "int",
		MinValue: 1,
		MaxValue: 65535,
	})

	// Pagination validation
	validator.AddRule(config.ValidationRule{
		Key:      "pagination.default_limit",
		Type:     "int",
		MinValue: 1,
		MaxValue: 1000,
	})

	validator.AddRule(config.ValidationRule{
		Key:      "pagination.max_limit",
		Type:     "int",
		MinValue: 1,
		MaxValue: 1000,
	})

	// Username validation
	validator.AddRule(config.ValidationRule{
		Key:      "validation.username_min_length",
		Type:     "int",
		MinValue: 1,
		MaxValue: 50,
	})

	validator.AddRule(config.ValidationRule{
		Key:      "validation.username_max_length",
		Type:     "int",
		MinValue: 1,
		MaxValue: 100,
	})
}
