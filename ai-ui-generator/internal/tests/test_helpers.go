package tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/EliasRanz/ai-code-gen/ai-ui-generator/internal/infrastructure/config"
)

// TestConfig provides a test configuration for unit tests
func TestConfig() *config.Config {
	return &config.Config{
		Server: config.ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			Name:     "test_db",
			Username: "test_user",
			Password: "test_pass",
			SSLMode:  "disable",
		},
		Logging: config.LoggingConfig{
			Level:  "debug",
			Format: "json",
		},
		Auth: config.AuthConfig{
			JWTSecret:            "test-secret-key-for-testing-only",
			AccessTokenDuration:  15 * time.Minute,
			RefreshTokenDuration: 7 * 24 * time.Hour,
			SessionDuration:      24 * time.Hour,
		},
		LLM: config.LLMConfig{
			Provider:    "openai",
			APIKey:      "test-api-key",
			Model:       "gpt-3.5-turbo",
			MaxTokens:   4096,
			Temperature: 0.7,
		},
	}
}

// TestContext returns a context for testing with timeout
func TestContext() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), time.Minute)
	return ctx
}

// RequireNoError is a helper for testing that fails the test if error is not nil
func RequireNoError(t *testing.T, err error, msgAndArgs ...interface{}) {
	require.NoError(t, err, msgAndArgs...)
}

// RequireError is a helper for testing that fails the test if error is nil
func RequireError(t *testing.T, err error, msgAndArgs ...interface{}) {
	require.Error(t, err, msgAndArgs...)
}
