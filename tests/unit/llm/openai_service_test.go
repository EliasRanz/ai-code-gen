package llm

import (
	"context"
	"testing"

	"github.com/EliasRanz/ai-code-gen/internal/domain/ai"
	"github.com/EliasRanz/ai-code-gen/internal/domain/common"
)

func TestOpenAIService_Validate(t *testing.T) {
	service := NewOpenAIService("test-api-key")

	tests := []struct {
		name     string
		code     string
		expected bool
	}{
		{
			name:     "Valid Go function",
			code:     "func main() {\n    fmt.Println(\"Hello, World!\")\n}",
			expected: true,
		},
		{
			name:     "Valid Python function",
			code:     "def hello():\n    print(\"Hello, World!\")",
			expected: true,
		},
		{
			name:     "Valid JavaScript class",
			code:     "class Calculator {\n    add(a, b) { return a + b; }\n}",
			expected: true,
		},
		{
			name:     "Empty code",
			code:     "",
			expected: false,
		},
		{
			name:     "Whitespace only",
			code:     "   \n\t  \n",
			expected: false,
		},
		{
			name:     "Random text without code structure",
			code:     "This is just random text without any code structure",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.Validate(context.Background(), tt.code)
			if err != nil {
				t.Fatalf("Validate() error = %v", err)
			}

			if result.Valid != tt.expected {
				t.Errorf("Validate() = %v, want %v", result.Valid, tt.expected)
			}

			// Check that we get appropriate error messages for invalid code
			if !tt.expected && len(result.Errors) == 0 {
				t.Error("Expected validation errors for invalid code, got none")
			}
		})
	}
}

func TestOpenAIService_StreamingInterface(t *testing.T) {
	service := NewOpenAIService("test-api-key")

	// Test that the streaming interface can be called without API key errors
	// (This test focuses on interface compatibility, not actual API calls)
	req := ai.GenerationRequest{
		Prompt:   "Generate a hello world function in Go",
		Language: "go",
		UserID:   common.UserID("test-user"),
	}

	t.Run("Legacy Stream method signature", func(t *testing.T) {
		// Test that the legacy Stream method can be called
		ch := make(chan string, 1)
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately to avoid actual API calls

		err := service.Stream(ctx, req, ch)
		// We expect a context cancellation error, not a method signature error
		if err == nil {
			t.Error("Expected context cancellation error")
		}
	})

	t.Run("New GenerateStream method signature", func(t *testing.T) {
		// Test that the new GenerateStream method can be called
		ch := make(chan ai.StreamChunk, 1)
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately to avoid actual API calls

		err := service.GenerateStream(ctx, req, ch)
		// We expect a context cancellation error, not a method signature error
		if err == nil {
			t.Error("Expected context cancellation error")
		}
	})
}

func TestOpenAIService_GenerationRequest(t *testing.T) {
	service := NewOpenAIService("test-api-key")

	req := ai.GenerationRequest{
		Prompt:   "Generate a hello world function",
		Language: "go",
		UserID:   common.UserID("test-user"),
	}

	// Test non-streaming generation (will fail without valid API key, but tests interface)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately to avoid actual API calls

	_, err := service.Generate(ctx, req)
	// We expect some kind of error (API key, network, etc.) but not a compilation error
	if err == nil {
		t.Error("Expected error due to cancelled context or API issues")
	}
}
