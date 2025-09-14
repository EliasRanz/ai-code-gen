package ai_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/EliasRanz/ai-code-gen/internal/ai"
	"github.com/EliasRanz/ai-code-gen/tests/mocks"
)

func setupModelTestHandler(ctrl *gomock.Controller) (*ai.Handler, *ai.Service, *gin.Engine) {
	gin.SetMode(gin.TestMode)

	mockLLM := mocks.NewMockLLMClient(ctrl)
	mockLLM.EXPECT().Generate(gomock.Any()).Return("<div>Generated UI</div>", nil).AnyTimes()
	mockLLM.EXPECT().StreamGenerate(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	svc := ai.NewService(mockLLM)
	h := ai.NewHandler(svc)
	r := gin.New()
	group := r.Group("")
	h.RegisterRoutes(group)
	return h, svc, r
}

func TestModelSelectionEndpoint(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	_, _, r := setupModelTestHandler(ctrl)

	// Test with model parameters
	temp := 0.7
	maxTokens := 100
	reqBody := ai.GenerateRequest{
		Prompt:      "create a button",
		Model:       "gpt-4",
		Temperature: &temp,
		MaxTokens:   &maxTokens,
	}
	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/ai/generate", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	var resp ai.GenerateResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Contains(t, resp.Code, "Generated UI")
}

func TestStreamWithModelParams(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	_, _, r := setupModelTestHandler(ctrl)

	req, _ := http.NewRequest("GET", "/ai/stream/abc?prompt=test&model=gpt-4&temperature=0.5&max_tokens=50", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "text/event-stream", w.Header().Get("Content-Type"))
}

func TestGenerationParams(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	_, svc, _ := setupModelTestHandler(ctrl)

	temp := 0.8
	maxTokens := 150
	params := ai.GenerationParams{
		Model:       "gpt-3.5-turbo",
		Temperature: &temp,
		MaxTokens:   &maxTokens,
	}

	// Test that parameters are accepted (service level test)
	code, err := svc.GenerateCodeWithParams("create a form", "user123", params)
	assert.NoError(t, err)
	assert.NotEmpty(t, code)
	assert.Contains(t, code, "Generated UI")
}
