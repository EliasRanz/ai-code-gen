package ai

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/EliasRanz/ai-code-gen/internal/ai"
)

type mockStreamLLMClient struct{}

func (m *mockStreamLLMClient) Generate(prompt string) (string, error) { return "", nil }
func (m *mockStreamLLMClient) StreamGenerate(prompt string, responseChannel chan string) error {
	for _, chunk := range []string{"chunk1", "chunk2", "chunk3"} {
		responseChannel <- chunk
	}
	return nil
}

func newStreamTestHandler() *ai.Handler {
	svc := ai.NewService(&mockStreamLLMClient{})
	return ai.NewHandler(svc)
}

func TestStreamHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newStreamTestHandler()
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
	h := newStreamTestHandler()
	r := gin.Default()
	r.GET("/ai/stream/:sessionId", h.Stream)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ai/stream/abc", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)
	assert.Contains(t, w.Body.String(), "prompt required")
}
