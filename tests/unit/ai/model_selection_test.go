package ai

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/EliasRanz/ai-code-gen/internal/ai"
)

func TestModelSelectionEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := ai.NewService(&mockLLM{})
	h := ai.NewHandler(svc)
	r := gin.New()
	group := r.Group("")
	h.RegisterRoutes(group)

	// Test with model parameters
	temp := 0.7
	maxTokens := 100
	req := ai.GenerateRequest{
		Prompt:      "create a button",
		UserID:      "user1",
		Model:       "gpt-4",
		Temperature: &temp,
		MaxTokens:   &maxTokens,
	}

	reqBody, _ := json.Marshal(req)
	httpReq, _ := http.NewRequest("POST", "/ai/generate", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httpReq)

	assert.Equal(t, 200, w.Code)
	
	var resp ai.GenerateResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Contains(t, resp.Code, "create a button")
}

func TestStreamWithModelParams(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := ai.NewService(&mockLLM{})
	h := ai.NewHandler(svc)
	r := gin.New()
	group := r.Group("")
	h.RegisterRoutes(group)

	req, _ := http.NewRequest("GET", "/ai/stream/abc?prompt=test&model=gpt-4&temperature=0.5&max_tokens=50", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "text/event-stream", w.Header().Get("Content-Type"))
}

func TestGenerationParams(t *testing.T) {
	svc := ai.NewService(&mockLLM{})
	
	temp := 0.8
	maxTokens := 150
	params := ai.GenerationParams{
		Model:       "gpt-3.5-turbo",
		Temperature: &temp,
		MaxTokens:   &maxTokens,
	}
	
	code, err := svc.GenerateCodeWithParams("test prompt", "user1", params)
	assert.NoError(t, err)
	assert.Contains(t, code, "test prompt")
	
	// Check history was updated
	history := svc.GetHistory("user1")
	assert.Len(t, history, 1)
	assert.Equal(t, "test prompt", history[0].Prompt)
}
