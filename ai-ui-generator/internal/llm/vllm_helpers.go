package llm

import (
	"fmt"
	"time"
)

// Helper functions for VLLM client operations

// generateResponseID creates a unique response ID
func generateResponseID() string {
	return fmt.Sprintf("vllm-resp-%d", time.Now().UnixNano())
}

// generateStubbedResponse creates a stubbed response based on prompt
func generateStubbedResponse(prompt string) string {
	// Generate a simple stubbed response based on the prompt
	responses := []string{
		"This is a stubbed response from the VLLM client. The actual implementation will call the real VLLM API.",
		"Here's a generated response that demonstrates the streaming functionality of the AI Generation Service.",
		"The VLLM client is currently stubbed and will be implemented to call the actual VLLM inference server.",
		"This response shows how the AI generation system will work once fully integrated with real LLM services.",
	}

	// Simple hash to make response somewhat deterministic
	hash := 0
	for _, char := range prompt {
		hash += int(char)
	}

	return responses[hash%len(responses)]
}

// estimateTokens provides a rough estimation of token count
func estimateTokens(text string) int {
	// Rough estimation: ~4 characters per token
	return len(text) / 4
}
