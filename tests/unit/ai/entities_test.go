// Tests for AI entity implementations
package ai_test

import (
	"testing"
	"time"

	"github.com/EliasRanz/ai-code-gen/internal/ai"
	"github.com/EliasRanz/ai-code-gen/internal/utilities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerationRequestValidation(t *testing.T) {
	userID := utilities.UserID("user-123")

	t.Run("Valid generation request", func(t *testing.T) {
		req := ai.GenerationRequest{
			Prompt:   "Generate a hello world function",
			Language: "go",
			UserID:   userID,
		}

		err := req.Validate()
		assert.NoError(t, err)
	})

	t.Run("Invalid generation request - missing prompt", func(t *testing.T) {
		req := ai.GenerationRequest{
			Language: "go",
			UserID:   userID,
		}

		err := req.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid input")
	})

	t.Run("Invalid generation request - empty user ID", func(t *testing.T) {
		req := ai.GenerationRequest{
			Prompt: "Generate code",
			UserID: utilities.UserID(""),
		}

		err := req.Validate()
		assert.Error(t, err)
	})

	t.Run("Invalid generation request - invalid temperature", func(t *testing.T) {
		temp := float64(3.0) // Invalid temperature > 2
		req := ai.GenerationRequest{
			Prompt:      "Generate code",
			UserID:      userID,
			Temperature: &temp,
		}

		err := req.Validate()
		assert.Error(t, err)
	})

	t.Run("Invalid generation request - invalid max tokens", func(t *testing.T) {
		tokens := 5000 // Invalid tokens > 4096
		req := ai.GenerationRequest{
			Prompt:    "Generate code",
			UserID:    userID,
			MaxTokens: &tokens,
		}

		err := req.Validate()
		assert.Error(t, err)
	})
}

func TestGenerationRequestDefaults(t *testing.T) {
	req := ai.GenerationRequest{
		Prompt: "Generate code",
		UserID: utilities.UserID("user-123"),
	}

	t.Run("Default model", func(t *testing.T) {
		assert.Equal(t, "gpt-3.5-turbo", req.GetModel())
	})

	t.Run("Default temperature", func(t *testing.T) {
		assert.Equal(t, 0.7, req.GetTemperature())
	})

	t.Run("Default max tokens", func(t *testing.T) {
		assert.Equal(t, 2048, req.GetMaxTokens())
	})

	t.Run("Custom values override defaults", func(t *testing.T) {
		customModel := "gpt-4"
		customTemp := 0.5
		customTokens := 1024

		customReq := ai.GenerationRequest{
			Prompt:      "Generate code",
			UserID:      utilities.UserID("user-123"),
			Model:       customModel,
			Temperature: &customTemp,
			MaxTokens:   &customTokens,
		}

		assert.Equal(t, customModel, customReq.GetModel())
		assert.Equal(t, customTemp, customReq.GetTemperature())
		assert.Equal(t, customTokens, customReq.GetMaxTokens())
	})
}

func TestGenerationResult(t *testing.T) {
	t.Run("Create new generation result", func(t *testing.T) {
		result := ai.NewGenerationResult("gen-123", "fmt.Println(\"hello\")", "gpt-4", 150, 0.002)

		assert.Equal(t, "gen-123", result.ID)
		assert.Equal(t, "fmt.Println(\"hello\")", result.Code)
		assert.Equal(t, "gpt-4", result.Model)
		assert.Equal(t, 150, result.UsedTokens)
		assert.Equal(t, 0.002, result.EstimatedCost)

		// Test DomainEntity interface implementation
		assert.Equal(t, "gen-123", result.GetID())
		assert.Equal(t, utilities.EntityTypeGeneration, result.GetType())
		assert.Equal(t, int64(1), result.GetVersion())
	})

	t.Run("Generation result validation", func(t *testing.T) {
		result := ai.NewGenerationResult("gen-123", "fmt.Println(\"hello\")", "gpt-4", 150, 0.002)

		err := result.Validate()
		assert.NoError(t, err)

		// Test validation rules
		rules := result.GetValidationRules()
		assert.Len(t, rules, 2)
		assert.Equal(t, "id", rules[0].Field)
		assert.Equal(t, "code", rules[1].Field)
	})

	t.Run("Generation result serialization", func(t *testing.T) {
		result := ai.NewGenerationResult("gen-123", "fmt.Println(\"hello\")", "gpt-4", 150, 0.002)

		// Test ToMap
		data := result.ToMap()
		assert.Equal(t, "gen-123", data["id"])
		assert.Equal(t, "fmt.Println(\"hello\")", data["code"])
		assert.Equal(t, "gpt-4", data["model"])
		assert.Equal(t, 150, data["used_tokens"])

		// Test ToJSON
		jsonData, err := result.ToJSON()
		require.NoError(t, err)
		assert.Contains(t, string(jsonData), "gen-123")
		assert.Contains(t, string(jsonData), "fmt.Println")

		// Test FromJSON
		newResult := &ai.GenerationResult{}
		err = newResult.FromJSON(jsonData)
		require.NoError(t, err)
		assert.Equal(t, "gen-123", newResult.ID)
		assert.Equal(t, "fmt.Println(\"hello\")", newResult.Code)
	})

	t.Run("Invalid generation result validation", func(t *testing.T) {
		// Empty ID
		result := &ai.GenerationResult{Code: "some code"}
		err := result.Validate()
		assert.Error(t, err)

		// Empty code
		result = &ai.GenerationResult{ID: "gen-123"}
		err = result.Validate()
		assert.Error(t, err)
	})
}

func TestValidationRequest(t *testing.T) {
	userID := utilities.UserID("user-123")

	t.Run("Valid validation request", func(t *testing.T) {
		req := ai.ValidationRequest{
			Code:   "fmt.Println(\"hello\")",
			UserID: userID,
		}

		err := req.Validate()
		assert.NoError(t, err)
	})

	t.Run("Invalid validation request - missing code", func(t *testing.T) {
		req := ai.ValidationRequest{
			UserID: userID,
		}

		err := req.Validate()
		assert.Error(t, err)
	})

	t.Run("Invalid validation request - empty user ID", func(t *testing.T) {
		req := ai.ValidationRequest{
			Code:   "fmt.Println(\"hello\")",
			UserID: utilities.UserID(""),
		}

		err := req.Validate()
		assert.Error(t, err)
	})
}

func TestAIGenerationHistory(t *testing.T) {
	t.Run("Create AI generation history", func(t *testing.T) {
		now := time.Now()
		history := ai.AIGenerationHistory{
			ID:     "hist-123",
			UserID: utilities.UserID("user-456"),
			Prompt: "Generate a function",
			Code:   "func hello() {}",
			Model:  "gpt-4",
			Tokens: 200,
			Timestamps: utilities.Timestamps{
				CreatedAt: now,
				UpdatedAt: now,
			},
		}

		assert.Equal(t, "hist-123", history.ID)
		assert.Equal(t, utilities.UserID("user-456"), history.UserID)
		assert.Equal(t, "Generate a function", history.Prompt)
		assert.Equal(t, "func hello() {}", history.Code)
		assert.Equal(t, "gpt-4", history.Model)
		assert.Equal(t, 200, history.Tokens)
	})
}

func TestQuotaStatus(t *testing.T) {
	t.Run("Can generate when remaining quota", func(t *testing.T) {
		quota := ai.QuotaStatus{
			UserID:     utilities.UserID("user-123"),
			DailyLimit: 1000,
			UsedToday:  500,
			Remaining:  500,
		}

		assert.True(t, quota.CanGenerate())
	})

	t.Run("Cannot generate when no remaining quota", func(t *testing.T) {
		quota := ai.QuotaStatus{
			UserID:     utilities.UserID("user-123"),
			DailyLimit: 1000,
			UsedToday:  1000,
			Remaining:  0,
		}

		assert.False(t, quota.CanGenerate())
	})
}

func TestStreamChunk(t *testing.T) {
	t.Run("Stream chunk properties", func(t *testing.T) {
		chunk := ai.StreamChunk{
			Content:    "partial content",
			TokenCount: 5,
			IsComplete: false,
			Model:      "gpt-4",
			Error:      nil,
		}

		assert.Equal(t, "partial content", chunk.Content)
		assert.Equal(t, 5, chunk.TokenCount)
		assert.False(t, chunk.IsComplete)
		assert.Equal(t, "gpt-4", chunk.Model)
		assert.NoError(t, chunk.Error)
	})

	t.Run("Stream chunk with error", func(t *testing.T) {
		testErr := assert.AnError
		chunk := ai.StreamChunk{
			Content:    "",
			TokenCount: 0,
			IsComplete: true,
			Model:      "gpt-4",
			Error:      testErr,
		}

		assert.Equal(t, testErr, chunk.Error)
		assert.True(t, chunk.IsComplete)
	})
}
