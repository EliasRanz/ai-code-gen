// Package config provides configuration management for the application
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	LLM      LLMConfig
	Auth     AuthConfig
	Logging  LoggingConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Host         string
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	GRPCPort     int
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	Name     string
	SSLMode  string
}

// LLMConfig holds LLM provider configuration
type LLMConfig struct {
	Provider    string
	APIKey      string
	Model       string
	MaxTokens   int
	Temperature float64
	BaseURL     string
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	JWTSecret            string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
	SessionDuration      time.Duration
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string
	Format string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Host:         getEnvOrDefault("SERVER_HOST", "localhost"),
			Port:         getEnvAsIntOrDefault("SERVER_PORT", 8080),
			ReadTimeout:  getEnvAsDurationOrDefault("SERVER_READ_TIMEOUT", 10*time.Second),
			WriteTimeout: getEnvAsDurationOrDefault("SERVER_WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:  getEnvAsDurationOrDefault("SERVER_IDLE_TIMEOUT", 60*time.Second),
			GRPCPort:     getEnvAsIntOrDefault("GRPC_PORT", 9090),
		},
		Database: DatabaseConfig{
			Host:     getEnvOrDefault("DB_HOST", "localhost"),
			Port:     getEnvAsIntOrDefault("DB_PORT", 5432),
			Username: getEnvOrDefault("DB_USERNAME", "postgres"),
			Password: getEnvOrDefault("DB_PASSWORD", ""),
			Name:     getEnvOrDefault("DB_NAME", "ai_ui_generator"),
			SSLMode:  getEnvOrDefault("DB_SSL_MODE", "disable"),
		},
		LLM: LLMConfig{
			Provider:    getEnvOrDefault("LLM_PROVIDER", "openai"),
			APIKey:      getEnvOrDefault("LLM_API_KEY", ""),
			Model:       getEnvOrDefault("LLM_MODEL", "gpt-4"),
			MaxTokens:   getEnvAsIntOrDefault("LLM_MAX_TOKENS", 4096),
			Temperature: getEnvAsFloatOrDefault("LLM_TEMPERATURE", 0.7),
			BaseURL:     getEnvOrDefault("LLM_BASE_URL", ""),
		},
		Auth: AuthConfig{
			JWTSecret:            getEnvOrDefault("JWT_SECRET", "your-secret-key"),
			AccessTokenDuration:  getEnvAsDurationOrDefault("ACCESS_TOKEN_DURATION", 15*time.Minute),
			RefreshTokenDuration: getEnvAsDurationOrDefault("REFRESH_TOKEN_DURATION", 7*24*time.Hour),
			SessionDuration:      getEnvAsDurationOrDefault("SESSION_DURATION", 24*time.Hour),
		},
		Logging: LoggingConfig{
			Level:  getEnvOrDefault("LOG_LEVEL", "info"),
			Format: getEnvOrDefault("LOG_FORMAT", "json"),
		},
	}

	// Validate required configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Auth.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	if c.LLM.APIKey == "" {
		return fmt.Errorf("LLM_API_KEY is required")
	}
	if c.Database.Password == "" {
		return fmt.Errorf("DB_PASSWORD is required")
	}
	return nil
}

// Helper functions for environment variable parsing
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsFloatOrDefault(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

func getEnvAsDurationOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
