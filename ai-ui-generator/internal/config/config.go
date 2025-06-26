package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds the application configuration
type Config struct {
	Server        ServerConfig        `json:"server"`
	Database      DatabaseConfig      `json:"database"`
	Redis         RedisConfig         `json:"redis"`
	Auth          AuthConfig          `json:"auth"`
	AI            AIConfig            `json:"ai"`
	Logging       LoggingConfig       `json:"logging"`
	Observability ObservabilityConfig `json:"observability"`
	// Service-specific configurations
	APIGateway   ServiceConfig `json:"api_gateway"`
	AuthService  ServiceConfig `json:"auth_service"`
	UserService  ServiceConfig `json:"user_service"`
	AIService    ServiceConfig `json:"ai_service"`
	AIGenService ServiceConfig `json:"ai_generation_service"`
	// Shared configurations
	LogLevel        string `json:"log_level"`
	TracingEndpoint string `json:"tracing_endpoint"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

// ServiceConfig holds service-specific configuration
type ServiceConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host            string `json:"host"`
	Port            int    `json:"port"`
	User            string `json:"user"`
	Password        string `json:"password"`
	DBName          string `json:"dbname"`
	SSLMode         string `json:"sslmode"`
	MaxOpenConns    int    `json:"max_open_conns"`
	MaxIdleConns    int    `json:"max_idle_conns"`
	ConnMaxLifetime string `json:"conn_max_lifetime"`
	ConnMaxIdleTime string `json:"conn_max_idle_time"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	JWTSecret string      `json:"jwt_secret"`
	JWTExpiry string      `json:"jwt_expiry"`
	OAuth     OAuthConfig `json:"oauth"`
}

// OAuthConfig holds OAuth configuration
type OAuthConfig struct {
	Google GoogleOAuthConfig `json:"google"`
}

// GoogleOAuthConfig holds Google OAuth configuration
type GoogleOAuthConfig struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectURL  string `json:"redirect_url"`
}

// AIConfig holds AI service configuration
type AIConfig struct {
	LLMEndpoint string    `json:"llm_endpoint"`
	ModelName   string    `json:"model_name"`
	MaxTokens   int       `json:"max_tokens"`
	Temperature float64   `json:"temperature"`
	LLM         LLMConfig `json:"llm"`
}

// LLMConfig holds LLM provider configuration
type LLMConfig struct {
	BaseURL    string        `json:"base_url"`
	APIKey     string        `json:"api_key"`
	Timeout    time.Duration `json:"timeout"`
	MaxRetries int           `json:"max_retries"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string `json:"level"`
	Format string `json:"format"`
}

// ObservabilityConfig holds observability configuration
type ObservabilityConfig struct {
	MetricsEnabled bool   `json:"metrics_enabled"`
	TracingEnabled bool   `json:"tracing_enabled"`
	JaegerEndpoint string `json:"jaeger_endpoint"`
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		// It's okay if .env doesn't exist in production
		fmt.Printf("Warning: Could not load .env file: %v\n", err)
	}

	config := &Config{
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
			Port: getEnvAsInt("SERVER_PORT", 8080),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DATABASE_HOST", "localhost"),
			Port:            getEnvAsInt("DATABASE_PORT", 5433),
			User:            getEnv("DATABASE_USER", "postgres"),
			Password:        getEnv("DATABASE_PASSWORD", "password"),
			DBName:          getEnv("DATABASE_NAME", "ai_ui_generator"),
			SSLMode:         getEnv("DATABASE_SSL_MODE", "disable"),
			MaxOpenConns:    getEnvAsInt("DATABASE_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("DATABASE_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnv("DATABASE_CONN_MAX_LIFETIME", "5m"),
			ConnMaxIdleTime: getEnv("DATABASE_CONN_MAX_IDLE_TIME", "1m"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvAsInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		Auth: AuthConfig{
			JWTSecret: getEnv("JWT_SECRET", "your-secret-key"),
			JWTExpiry: getEnv("JWT_EXPIRY", "24h"),
			OAuth: OAuthConfig{
				Google: GoogleOAuthConfig{
					ClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
					ClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
					RedirectURL:  getEnv("OAUTH_REDIRECT_URL", "http://localhost:3000/api/auth/callback/google"),
				},
			},
		},
		AI: AIConfig{
			LLMEndpoint: getEnv("LLM_ENDPOINT", "http://localhost:8000/v1"),
			ModelName:   getEnv("LLM_MODEL_NAME", "gpt-3.5-turbo"),
			MaxTokens:   getEnvAsInt("LLM_MAX_TOKENS", 4096),
			Temperature: getEnvAsFloat("LLM_TEMPERATURE", 0.7),
			LLM: LLMConfig{
				BaseURL:    getEnv("VLLM_BASE_URL", "http://localhost:8000"),
				APIKey:     getEnv("VLLM_API_KEY", ""),
				Timeout:    time.Duration(getEnvAsInt("VLLM_TIMEOUT_SECONDS", 30)) * time.Second,
				MaxRetries: getEnvAsInt("VLLM_MAX_RETRIES", 3),
			},
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
		Observability: ObservabilityConfig{
			MetricsEnabled: getEnvAsBool("METRICS_ENABLED", true),
			TracingEnabled: getEnvAsBool("TRACING_ENABLED", true),
			JaegerEndpoint: getEnv("JAEGER_ENDPOINT", "http://localhost:14268/api/traces"),
		},
		// Service-specific configurations
		APIGateway: ServiceConfig{
			Host: getEnv("API_GATEWAY_HOST", "0.0.0.0"),
			Port: getEnvAsInt("API_GATEWAY_PORT", 8080),
		},
		AuthService: ServiceConfig{
			Host: getEnv("AUTH_SERVICE_HOST", "0.0.0.0"),
			Port: getEnvAsInt("AUTH_SERVICE_PORT", 8081),
		},
		UserService: ServiceConfig{
			Host: getEnv("USER_SERVICE_HOST", "0.0.0.0"),
			Port: getEnvAsInt("USER_SERVICE_PORT", 8082),
		},
		AIService: ServiceConfig{
			Host: getEnv("AI_SERVICE_HOST", "0.0.0.0"),
			Port: getEnvAsInt("AI_SERVICE_PORT", 8083),
		},
		AIGenService: ServiceConfig{
			Host: getEnv("AI_GEN_SERVICE_HOST", "0.0.0.0"),
			Port: getEnvAsInt("AI_GEN_SERVICE_PORT", 8084),
		},
		// Shared configurations
		LogLevel:        getEnv("LOG_LEVEL", "info"),
		TracingEndpoint: getEnv("JAEGER_ENDPOINT", "http://localhost:14268/api/traces"),
	}

	return config, nil
}

// Helper functions for environment variable parsing
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsFloat(key string, defaultValue float64) float64 {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseFloat(valueStr, 64); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return defaultValue
}

// Validate validates the configuration
func (c *Config) Validate() error {
	var errors []string

	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		errors = append(errors, "invalid server port")
	}

	if c.Database.Host == "" {
		errors = append(errors, "database host is required")
	}

	if c.Auth.JWTSecret == "" || c.Auth.JWTSecret == "your-secret-key" {
		errors = append(errors, "JWT secret must be set and not use default value")
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed: %s", strings.Join(errors, ", "))
	}

	return nil
}

// DSN returns the PostgreSQL connection string
func (dc *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		dc.Host, dc.Port, dc.User, dc.Password, dc.DBName, dc.SSLMode,
	)
}
