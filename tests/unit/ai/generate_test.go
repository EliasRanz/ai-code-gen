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

func newTestHandler(ctrl *gomock.Controller) *ai.Handler {
	mockLLM := mocks.NewMockLLMClient(ctrl)
	mockLLM.EXPECT().Generate(gomock.Any()).DoAndReturn(func(prompt string) (string, error) {
		if prompt == "fail" {
			return "", assert.AnError
		}
		return "<div>Generated UI</div>", nil
	}).AnyTimes()

	svc := ai.NewService(mockLLM)
	return ai.NewHandler(svc)
}

func TestGenerateHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	h := newTestHandler(ctrl)
	r := gin.Default()
	r.POST("/ai/generate", h.Generate)
	body, _ := json.Marshal(ai.GenerateRequest{Prompt: "hello"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/ai/generate", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	var resp ai.GenerateResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "<div>Generated UI</div>", resp.Code)
}

func TestGenerateHandler_InvalidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	h := newTestHandler(ctrl)
	r := gin.Default()
	r.POST("/ai/generate", h.Generate)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/ai/generate", bytes.NewBuffer([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)
	assert.Contains(t, w.Body.String(), "invalid request")
}

func TestGenerateHandler_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a specific mock for error case
	mockLLM := mocks.NewMockLLMClient(ctrl)
	mockLLM.EXPECT().Generate("fail").Return("", assert.AnError)

	svc := ai.NewService(mockLLM)
	h := ai.NewHandler(svc)
	r := gin.Default()
	r.POST("/ai/generate", h.Generate)
	body, _ := json.Marshal(ai.GenerateRequest{Prompt: "fail"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/ai/generate", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, 500, w.Code)
}
