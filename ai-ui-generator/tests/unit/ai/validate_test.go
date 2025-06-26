package ai

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"strings"
)

type mockValidateService struct{}

func (m *mockValidateService) GenerateCode(prompt string, userID string) (string, error) { return "", nil }
func (m *mockValidateService) StreamGeneration(prompt string, userID string, responseChannel chan string) error { return nil }
func (m *mockValidateService) ValidateGeneratedCode(code string) (bool, []string, error) {
	if code == "bad" {
		return false, []string{"syntax error", "security issue"}, nil
	}
	if code == "error" {
		return false, nil, assert.AnError
	}
	return true, nil, nil
}

type mockValidateLLMClient struct{}
func (m *mockValidateLLMClient) Generate(prompt string) (string, error) { return "", nil }
func (m *mockValidateLLMClient) StreamGenerate(prompt string, responseChannel chan string) error { return nil }

func newValidateTestHandler() *Handler {
	// Use the real Service with real validation logic
	svc := NewService(&mockValidateLLMClient{})
	return &Handler{service: svc}
}

func TestValidateCodeHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newValidateTestHandler()
	r := gin.Default()
	r.POST("/ai/validate", h.ValidateCode)
	body, _ := json.Marshal(ValidateRequest{Code: "good"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/ai/validate", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	var resp ValidateResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.True(t, resp.Valid)
	assert.Empty(t, resp.Errors)
}

func TestValidateCodeHandler_InvalidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newValidateTestHandler()
	r := gin.Default()
	r.POST("/ai/validate", h.ValidateCode)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/ai/validate", bytes.NewBuffer([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)
	assert.Contains(t, w.Body.String(), "invalid request")
}

func TestValidateCodeHandler_Failure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newValidateTestHandler()
	r := gin.Default()
	r.POST("/ai/validate", h.ValidateCode)
	// This code contains a <script> tag (security error)
	body, _ := json.Marshal(ValidateRequest{Code: "<div><script>alert('x')</script>"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/ai/validate", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	var resp ValidateResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.False(t, resp.Valid)
	assert.True(t, len(resp.Errors) > 0)
	foundScript := false
	for _, e := range resp.Errors {
		if strings.Contains(e, "<script>") || strings.Contains(e, "Security issue") {
			foundScript = true
		}
	}
	assert.True(t, foundScript)
}
