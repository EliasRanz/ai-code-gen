package ai_test

import (
	"context"
	"testing"

	"github.com/EliasRanz/ai-code-gen/internal/ai"
	"github.com/EliasRanz/ai-code-gen/internal/utilities"
	"github.com/stretchr/testify/assert"
)

// TestGenerationRequest_Validate tests request validation
func TestGenerationRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request ai.GenerationRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid request",
			request: ai.GenerationRequest{
				Prompt:   "Generate a hello world function",
				Language: "go",
				UserID:   utilities.UserID("user-123"),
				Model:    "gpt-3.5-turbo",
			},
			wantErr: false,
		},
		{
			name: "Empty prompt",
			request: ai.GenerationRequest{
				Language: "go",
				UserID:   utilities.UserID("user-123"),
				Model:    "gpt-3.5-turbo",
			},
			wantErr: true,
		},
		{
			name: "Empty user ID",
			request: ai.GenerationRequest{
				Prompt:   "Generate a hello world function",
				Language: "go",
				Model:    "gpt-3.5-turbo",
			},
			wantErr: true,
		},
		{
			name: "Invalid temperature - too low",
			request: ai.GenerationRequest{
				Prompt:      "Generate a hello world function",
				Language:    "go",
				UserID:      utilities.UserID("user-123"),
				Model:       "gpt-3.5-turbo",
				Temperature: func() *float64 { v := -0.1; return &v }(),
			},
			wantErr: true,
		},
		{
			name: "Invalid temperature - too high",
			request: ai.GenerationRequest{
				Prompt:      "Generate a hello world function",
				Language:    "go",
				UserID:      utilities.UserID("user-123"),
				Model:       "gpt-3.5-turbo",
				Temperature: func() *float64 { v := 2.1; return &v }(),
			},
			wantErr: true,
		},
		{
			name: "Invalid max tokens - too low",
			request: ai.GenerationRequest{
				Prompt:    "Generate a hello world function",
				Language:  "go",
				UserID:    utilities.UserID("user-123"),
				Model:     "gpt-3.5-turbo",
				MaxTokens: func() *int { v := 0; return &v }(),
			},
			wantErr: true,
		},
		{
			name: "Invalid max tokens - too high",
			request: ai.GenerationRequest{
				Prompt:    "Generate a hello world function",
				Language:  "go",
				UserID:    utilities.UserID("user-123"),
				Model:     "gpt-3.5-turbo",
				MaxTokens: func() *int { v := 5000; return &v }(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestRedisClientFactory tests Redis client factory
func TestRedisClientFactory(t *testing.T) {
	factory := ai.NewRedisClientFactory()

	tests := []struct {
		name    string
		config  *ai.RedisConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid Redis config",
			config: &ai.RedisConfig{
				Host: "localhost",
				Port: 6379,
				DB:   0,
			},
			wantErr: false,
		},
		{
			name:    "Nil config returns stub client",
			config:  nil,
			wantErr: false,
		},
		{
			name: "Invalid config - empty host",
			config: &ai.RedisConfig{
				Port: 6379,
				DB:   0,
			},
			wantErr: true,
			errMsg:  "Redis host is required",
		},
		{
			name: "Invalid config - invalid port",
			config: &ai.RedisConfig{
				Host: "localhost",
				Port: -1,
				DB:   0,
			},
			wantErr: true,
			errMsg:  "Redis port must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := factory.CreateClient(tt.config)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, client)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
			}
		})
	}
}

// TestValidateGenerationConfig tests generation configuration validation
func TestValidateGenerationConfig(t *testing.T) {
	tests := []struct {
		name        string
		aiConfig    *ai.Config
		redisConfig *ai.RedisConfig
		wantErr     bool
		errMsg      string
	}{
		{
			name:     "Valid configs",
			aiConfig: &ai.Config{}, // Assuming empty config is valid
			redisConfig: &ai.RedisConfig{
				Host: "localhost",
				Port: 6379,
			},
			wantErr: false,
		},
		{
			name:        "Nil AI config",
			aiConfig:    nil,
			redisConfig: &ai.RedisConfig{Host: "localhost", Port: 6379},
			wantErr:     true,
			errMsg:      "AI configuration is required",
		},
		{
			name:        "Nil Redis config is valid",
			aiConfig:    &ai.Config{},
			redisConfig: nil,
			wantErr:     false,
		},
		{
			name:     "Invalid Redis config",
			aiConfig: &ai.Config{},
			redisConfig: &ai.RedisConfig{
				Port: 6379,
			},
			wantErr: true,
			errMsg:  "Redis host is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ai.ValidateGenerationConfig(tt.aiConfig, tt.redisConfig)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestStubRedisClient tests the stub Redis client implementation
func TestStubRedisClient(t *testing.T) {
	client := ai.NewRedisClient(nil) // Should return stub client

	// Test ping
	err := client.Ping(context.Background())
	assert.NoError(t, err)

	// Test close
	err = client.Close()
	assert.NoError(t, err)
}

// BenchmarkGenerationRequest benchmarks request validation
func BenchmarkGenerationRequest_Validate(b *testing.B) {
	request := ai.GenerationRequest{
		Prompt:   "Generate a hello world function",
		Language: "go",
		UserID:   utilities.UserID("user-123"),
		Model:    "gpt-3.5-turbo",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		request.Validate()
	}
}
