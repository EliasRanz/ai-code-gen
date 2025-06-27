package ai

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type mockService struct{}

func (m *mockService) GenerateCode(prompt string, userID string) (string, error) {
	if prompt == "fail" {
		return "", assert.AnError
	}
	return "<div>Generated UI</div>", nil
}

func (m *mockService) StreamGeneration(prompt string, userID string, responseChannel chan string) error { return nil }
func (m *mockService) ValidateGeneratedCode(code string) (bool, []string, error) { return true, nil, nil }

type mockLLMClient struct{}

func (m *mockLLMClient) Generate(prompt string) (string, error) {
	if prompt == "fail" {
		return "", assert.AnError
	}
	return "<div>Generated UI</div>", nil
}
func (m *mockLLMClient) StreamGenerate(prompt string, responseChannel chan string) error { return nil }

func newTestHandler() *Handler {
	svc := NewService(&mockLLMClient{})
	return &Handler{service: svc}
}

func TestGenerateHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newTestHandler()
	r := gin.Default()
	r.POST("/ai/generate", h.Generate)
	body, _ := json.Marshal(GenerateRequest{Prompt: "hello"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/ai/generate", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	var resp GenerateResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "<div>Generated UI</div>", resp.Code)
}

func TestGenerateHandler_InvalidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newTestHandler()
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
	h := newTestHandler()
	r := gin.Default()
	r.POST("/ai/generate", h.Generate)
	body, _ := json.Marshal(GenerateRequest{Prompt: "fail"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/ai/generate", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, 500, w.Code)
	assert.Contains(t, w.Body.String(), "assert.AnError")
}
