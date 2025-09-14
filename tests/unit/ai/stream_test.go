package ai_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/EliasRanz/ai-code-gen/internal/ai"
	"github.com/EliasRanz/ai-code-gen/tests/mocks"
)

func newStreamTestHandler(ctrl *gomock.Controller) *ai.Handler {
	mockLLM := mocks.NewMockLLMClient(ctrl)
	mockLLM.EXPECT().Generate(gomock.Any()).Return("", nil).AnyTimes()
	mockLLM.EXPECT().StreamGenerate(gomock.Any(), gomock.Any()).DoAndReturn(func(prompt string, responseChannel chan string) error {
		for _, chunk := range []string{"chunk1", "chunk2", "chunk3"} {
			responseChannel <- chunk
		}
		return nil
	}).AnyTimes()

	svc := ai.NewService(mockLLM)
	return ai.NewHandler(svc)
}

func TestStreamHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	h := newStreamTestHandler(ctrl)
	r := gin.Default()
	r.GET("/ai/stream/:sessionId", h.Stream)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ai/stream/abc?prompt=test", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "chunk1")
	assert.Contains(t, w.Body.String(), "chunk2")
	assert.Contains(t, w.Body.String(), "chunk3")
}

func TestStreamHandler_MissingPrompt(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	h := newStreamTestHandler(ctrl)
	r := gin.Default()
	r.GET("/ai/stream/:sessionId", h.Stream)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ai/stream/abc", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)
	assert.Contains(t, w.Body.String(), "prompt required")
}
