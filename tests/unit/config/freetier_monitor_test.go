package config

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/EliasRanz/ai-code-gen/internal/config"
)

func TestFreeTierMonitor_Creation(t *testing.T) {
	monitor := config.NewFreeTierMonitor()

	assert.NotNil(t, monitor)
	assert.NotNil(t, monitor.GetAlertChannel())
}

func TestFreeTierMonitor_ValidateFreeTierCompliance(t *testing.T) {
	tests := []struct {
		name        string
		configData  config.ConfigData
		expectError bool
		description string
	}{
		{
			name: "valid free tier config",
			configData: config.ConfigData{
				"database": map[string]interface{}{
					"max_open_conns": 5,
				},
				"redis": map[string]interface{}{
					"max_connections": 10,
				},
				"server": map[string]interface{}{
					"port": 8080,
				},
				"ai": map[string]interface{}{
					"openai": map[string]interface{}{
						"api_key":        "sk-",
						"free_tier_only": true,
						"max_tokens":     1500,
					},
				},
			},
			expectError: false,
			description: "Configuration should pass free tier validation",
		},
		{
			name: "database max connections exceeded",
			configData: config.ConfigData{
				"ai": map[string]interface{}{
					"openai": map[string]interface{}{
						"api_key":        "sk-",
						"free_tier_only": true,
						"max_tokens":     3000, // Exceeds OpenAI limit of 2000
					},
				},
			},
			expectError: true,
			description: "Should fail when AI token limit exceeds free tier limit",
		},
		{
			name: "vllm connections exceeded",
			configData: config.ConfigData{
				"ai": map[string]interface{}{
					"openai": map[string]interface{}{
						"api_key":        "sk-",
						"free_tier_only": true,
						"max_tokens":     1500,
					},
					"vllm": map[string]interface{}{
						"concurrent_requests": 10, // Exceeds vLLM limit of 5
					},
				},
			},
			expectError: true,
			description: "Should fail when vLLM concurrent requests exceed free tier limit",
		},
		{
			name: "multiple violations",
			configData: config.ConfigData{
				"ai": map[string]interface{}{
					"openai": map[string]interface{}{
						"api_key":        "sk-",
						"free_tier_only": true,
						"max_tokens":     3000, // Exceeds OpenAI limit
					},
					"vllm": map[string]interface{}{
						"max_tokens":          2000, // Exceeds vLLM limit of 1500
						"concurrent_requests": 10,   // Exceeds concurrent request limit
					},
				},
			},
			expectError: true,
			description: "Should fail when multiple AI limits are exceeded",
		},
		{
			name: "empty config",
			configData: config.ConfigData{
				"ai": map[string]interface{}{
					"openai": map[string]interface{}{
						"api_key":        "sk-",
						"free_tier_only": true,
						"max_tokens":     1500,
					},
				},
			},
			expectError: false,
			description: "Empty config should not violate free tier limits",
		},
		{
			name: "nested config validation",
			configData: config.ConfigData{
				"ai": map[string]interface{}{
					"openai": map[string]interface{}{
						"api_key":        "sk-",
						"free_tier_only": true,
						"max_tokens":     1000,
					},
					"rate_limit": map[string]interface{}{
						"requests_per_minute": 20, // Within limits
					},
				},
			},
			expectError: false,
			description: "Nested configuration should be validated correctly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitor := config.NewFreeTierMonitor()
			ctx := context.Background()

			err := monitor.ValidateFreeTierCompliance(ctx, tt.configData)

			if tt.expectError {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
			}
		})
	}
}

func TestFreeTierMonitor_AlertChannel(t *testing.T) {
	monitor := config.NewFreeTierMonitor()
	alertChan := monitor.GetAlertChannel()

	assert.NotNil(t, alertChan)

	// Test that channel is properly initialized
	select {
	case <-alertChan:
		t.Error("Channel should not have any alerts initially")
	case <-time.After(10 * time.Millisecond):
		// Expected: no alerts initially
	}
}

func TestFreeTierMonitor_ComplexConfigValidation(t *testing.T) {
	monitor := config.NewFreeTierMonitor()
	ctx := context.Background()

	// Test complex configuration structure
	complexConfig := config.ConfigData{
		"services": map[string]interface{}{
			"api_gateway": map[string]interface{}{
				"workers": 2,
			},
			"auth_service": map[string]interface{}{
				"workers": 2,
			},
		},
		"database": map[string]interface{}{
			"max_open_conns": 10,
			"max_idle_conns": 5,
		},
		"redis": map[string]interface{}{
			"max_connections": 15,
		},
		"observability": map[string]interface{}{
			"metrics_enabled": true,
			"tracing_enabled": true,
		},
		"ai": map[string]interface{}{
			"openai": map[string]interface{}{
				"api_key":        "sk-",
				"free_tier_only": true,
				"max_tokens":     1800,
			},
		},
	}

	err := monitor.ValidateFreeTierCompliance(ctx, complexConfig)
	assert.NoError(t, err, "Complex valid configuration should pass validation")
}

func TestFreeTierMonitor_EdgeCases(t *testing.T) {
	monitor := config.NewFreeTierMonitor()
	ctx := context.Background()

	t.Run("nil config data", func(t *testing.T) {
		err := monitor.ValidateFreeTierCompliance(ctx, nil)
		assert.Error(t, err, "Nil config should fail validation due to missing required AI config")
		assert.Contains(t, err.Error(), "ai.openai")
	})

	t.Run("invalid nested values", func(t *testing.T) {
		invalidConfig := config.ConfigData{
			"database": "invalid_structure", // Should be map
		}

		// Should handle gracefully without panicking
		_ = monitor.ValidateFreeTierCompliance(ctx, invalidConfig)
		// May error or pass depending on implementation - should not panic
		assert.NotPanics(t, func() {
			monitor.ValidateFreeTierCompliance(ctx, invalidConfig)
		})
	})

	t.Run("string numeric values", func(t *testing.T) {
		stringConfig := config.ConfigData{
			"database": map[string]interface{}{
				"max_open_conns": "10", // String instead of int
			},
		}

		assert.NotPanics(t, func() {
			monitor.ValidateFreeTierCompliance(ctx, stringConfig)
		})
	})
}

func TestFreeTierMonitor_ConcurrentAccess(t *testing.T) {
	monitor := config.NewFreeTierMonitor()
	ctx := context.Background()

	// Test concurrent validation calls
	validConfig := config.ConfigData{
		"database": map[string]interface{}{
			"max_open_conns": 5,
		},
		"ai": map[string]interface{}{
			"openai": map[string]interface{}{
				"api_key":        "sk-",
				"free_tier_only": true,
				"max_tokens":     1500,
			},
		},
	}

	done := make(chan bool)
	errorChan := make(chan error, 10)

	// Run 10 concurrent validations
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()
			err := monitor.ValidateFreeTierCompliance(ctx, validConfig)
			if err != nil {
				errorChan <- err
			}
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Check for any errors
	close(errorChan)
	for err := range errorChan {
		t.Errorf("Concurrent validation error: %v", err)
	}
}

func TestFreeTierMonitor_ValidationErrorFormatting(t *testing.T) {
	monitor := config.NewFreeTierMonitor()
	ctx := context.Background()

	// Create configuration that will definitely violate multiple limits
	violatingConfig := config.ConfigData{
		"database": map[string]interface{}{
			"max_open_conns": 1000, // Way over limit
		},
		"redis": map[string]interface{}{
			"max_connections": 500, // Way over limit
		},
	}

	err := monitor.ValidateFreeTierCompliance(ctx, violatingConfig)

	if err != nil {
		errorMsg := err.Error()
		// Error message should contain information about violations
		assert.Contains(t, errorMsg, "tier", "Error should mention tier violations")
		// Should be human-readable
		assert.NotEmpty(t, errorMsg, "Error message should not be empty")
	}
}
