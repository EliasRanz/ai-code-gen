package ai_test

import (
	"testing"
	"time"

	"github.com/EliasRanz/ai-code-gen/internal/ai/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerationRequestBuilder(t *testing.T) {
	tests := []struct {
		name    string
		builder func() llm.GenerationRequestBuilder
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid basic request",
			builder: func() llm.GenerationRequestBuilder {
				return llm.NewGenerationRequestBuilder().
					SetUserID("user123").
					SetPrompt("Generate a hello world function")
			},
			wantErr: false,
		},
		{
			name: "empty user ID should fail",
			builder: func() llm.GenerationRequestBuilder {
				return llm.NewGenerationRequestBuilder().
					SetUserID("").
					SetPrompt("Generate a hello world function")
			},
			wantErr: true,
			errMsg:  "user ID cannot be empty",
		},
		{
			name: "empty prompt should fail",
			builder: func() llm.GenerationRequestBuilder {
				return llm.NewGenerationRequestBuilder().
					SetUserID("user123").
					SetPrompt("")
			},
			wantErr: true,
			errMsg:  "prompt cannot be empty",
		},
		{
			name: "prompt too long should fail",
			builder: func() llm.GenerationRequestBuilder {
				longPrompt := string(make([]byte, 9000)) // Exceed 8000 limit
				return llm.NewGenerationRequestBuilder().
					SetUserID("user123").
					SetPrompt(longPrompt)
			},
			wantErr: true,
			errMsg:  "prompt exceeds free tier limit",
		},
		{
			name: "negative max tokens should fail",
			builder: func() llm.GenerationRequestBuilder {
				return llm.NewGenerationRequestBuilder().
					SetUserID("user123").
					SetPrompt("test").
					SetMaxTokens(-1)
			},
			wantErr: true,
			errMsg:  "max tokens cannot be negative",
		},
		{
			name: "excessive max tokens should fail in free tier",
			builder: func() llm.GenerationRequestBuilder {
				return llm.NewGenerationRequestBuilder().
					SetUserID("user123").
					SetPrompt("test").
					SetMaxTokens(3000)
			},
			wantErr: true,
			errMsg:  "max tokens exceeds free tier limit",
		},
		{
			name: "invalid temperature should fail",
			builder: func() llm.GenerationRequestBuilder {
				return llm.NewGenerationRequestBuilder().
					SetUserID("user123").
					SetPrompt("test").
					SetTemperature(3.0)
			},
			wantErr: true,
			errMsg:  "temperature must be between 0 and 2.0",
		},
		{
			name: "invalid provider should fail",
			builder: func() llm.GenerationRequestBuilder {
				return llm.NewGenerationRequestBuilder().
					SetUserID("user123").
					SetPrompt("test").
					SetProvider("invalid-provider")
			},
			wantErr: true,
			errMsg:  "provider invalid-provider not in allowed list",
		},
		{
			name: "complete valid request with all options",
			builder: func() llm.GenerationRequestBuilder {
				return llm.NewGenerationRequestBuilder().
					SetUserID("user123").
					SetPrompt("Generate a hello world function").
					SetLanguage("go").
					SetMaxTokens(500).
					SetTemperature(0.7).
					AddMetadata("test", "value").
					SetProvider("openai").
					SetModel("gpt-3.5-turbo").
					EnableStreaming().
					SetTimeout(30 * time.Second)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := tt.builder()

			if tt.wantErr {
				err := builder.Validate()
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)

				_, buildErr := builder.Build()
				require.Error(t, buildErr)
			} else {
				err := builder.Validate()
				require.NoError(t, err)

				req, buildErr := builder.Build()
				require.NoError(t, buildErr)
				require.NotNil(t, req)

				// Verify defaults are applied
				assert.NotZero(t, req.MaxTokens)
				assert.NotZero(t, req.Temperature)
				assert.NotEmpty(t, req.Provider)
				assert.NotEmpty(t, req.Metadata["request_id"])
				assert.Equal(t, "true", req.Metadata["free_tier_only"])
			}
		})
	}
}

func TestBuilderDefaults(t *testing.T) {
	req, err := llm.NewGenerationRequestBuilder().
		SetUserID("user123").
		SetPrompt("test prompt").
		Build()

	require.NoError(t, err)
	require.NotNil(t, req)

	// Check defaults
	assert.Equal(t, 1000, req.MaxTokens)
	assert.Equal(t, 0.7, req.Temperature)
	assert.Equal(t, "openai", req.Provider)
	assert.NotEmpty(t, req.Metadata["request_id"])
	assert.Equal(t, "true", req.Metadata["free_tier_only"])
}

func TestBuilderMetadata(t *testing.T) {
	req, err := llm.NewGenerationRequestBuilder().
		SetUserID("user123").
		SetPrompt("test prompt").
		AddMetadata("key1", "value1").
		AddMetadata("key2", "value2").
		Build()

	require.NoError(t, err)
	require.NotNil(t, req)

	assert.Equal(t, "value1", req.Metadata["key1"])
	assert.Equal(t, "value2", req.Metadata["key2"])
	assert.NotEmpty(t, req.Metadata["request_id"])
}

func TestBuilderEmptyMetadataKey(t *testing.T) {
	err := llm.NewGenerationRequestBuilder().
		SetUserID("user123").
		SetPrompt("test").
		AddMetadata("", "value").
		Validate()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "metadata key cannot be empty")
}

func TestBuilderInvalidTimeout(t *testing.T) {
	err := llm.NewGenerationRequestBuilder().
		SetUserID("user123").
		SetPrompt("test").
		SetTimeout(-1 * time.Second).
		Validate()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "timeout must be positive")
}

func TestBuilderExcessiveTimeout(t *testing.T) {
	err := llm.NewGenerationRequestBuilder().
		SetUserID("user123").
		SetPrompt("test").
		SetTimeout(10 * time.Minute).
		Validate()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "timeout exceeds maximum of 5 minutes")
}

func TestBuilderFluentInterface(t *testing.T) {
	// Test that all methods return the builder for chaining
	builder := llm.NewGenerationRequestBuilder()

	result := builder.
		SetUserID("user123").
		SetPrompt("test").
		SetLanguage("go").
		SetMaxTokens(100).
		SetTemperature(0.5).
		AddMetadata("test", "value").
		SetProvider("openai").
		SetModel("gpt-3.5-turbo").
		EnableStreaming().
		SetTimeout(30 * time.Second)

	// Verify we can still call Build on the result
	req, err := result.Build()
	require.NoError(t, err)
	require.NotNil(t, req)

	assert.Equal(t, "user123", req.UserID)
	assert.Equal(t, "test", req.Prompt)
	assert.Equal(t, "go", req.Language)
	assert.Equal(t, 100, req.MaxTokens)
	assert.Equal(t, 0.5, req.Temperature)
	assert.Equal(t, "openai", req.Provider)
	assert.Equal(t, "gpt-3.5-turbo", req.Model)
	assert.True(t, req.Stream)
	assert.Equal(t, 30*time.Second, req.Timeout)
	assert.Equal(t, "value", req.Metadata["test"])
}
