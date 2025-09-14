package ai_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"strings"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/EliasRanz/ai-code-gen/internal/ai"
	"github.com/EliasRanz/ai-code-gen/tests/mocks"
)

func newValidateTestHandler(ctrl *gomock.Controller) *ai.Handler {
	mockLLM := mocks.NewMockLLMClient(ctrl)
	mockLLM.EXPECT().Generate(gomock.Any()).Return("", nil).AnyTimes()
	mockLLM.EXPECT().StreamGenerate(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	// Create service with validation function
	validateFunc := func(code string) (bool, []string, error) {
		if code == "bad" {
			return false, []string{"syntax error", "security issue"}, nil
		}
		if code == "error" {
			return false, nil, assert.AnError
		}
		return true, nil, nil
	}

	svc := ai.NewServiceWithValidation(mockLLM, validateFunc)
	return ai.NewHandler(svc)
}

func TestValidateHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	h := newValidateTestHandler(ctrl)
	r := gin.Default()
	r.POST("/ai/validate", h.ValidateCode)
	body, _ := json.Marshal(ai.ValidateRequest{Code: "console.log('hello');"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/ai/validate", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	var resp ai.ValidateResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.True(t, resp.Valid)
	assert.Empty(t, resp.Errors)
}

func TestValidateHandler_InvalidCode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	h := newValidateTestHandler(ctrl)
	r := gin.Default()
	r.POST("/ai/validate", h.ValidateCode)
	body, _ := json.Marshal(ai.ValidateRequest{Code: "bad"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/ai/validate", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	var resp ai.ValidateResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.False(t, resp.Valid)
	assert.Contains(t, resp.Errors, "syntax error")
	assert.Contains(t, resp.Errors, "security issue")
}

func TestValidateHandler_ValidationError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	h := newValidateTestHandler(ctrl)
	r := gin.Default()
	r.POST("/ai/validate", h.ValidateCode)
	body, _ := json.Marshal(ai.ValidateRequest{Code: "error"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/ai/validate", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, 500, w.Code)
	assert.Contains(t, strings.ToLower(w.Body.String()), "failed to validate code")
}

func TestValidateHandler_InvalidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	h := newValidateTestHandler(ctrl)
	r := gin.Default()
	r.POST("/ai/validate", h.ValidateCode)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/ai/validate", bytes.NewBuffer([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)
	assert.Contains(t, strings.ToLower(w.Body.String()), "invalid request")
}
