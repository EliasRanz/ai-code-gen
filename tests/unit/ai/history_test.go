package ai

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type mockLLM struct{}
func (m *mockLLM) Generate(prompt string) (string, error) { return "<div>" + prompt + "</div>", nil }
func (m *mockLLM) StreamGenerate(prompt string, ch chan string) error { return nil }

func setupHistoryTest() (*Handler, *Service, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	svc := NewService(&mockLLM{})
	h := NewHandler(svc)
	r := gin.New()
	group := r.Group("")
	h.RegisterRoutes(group)
	return h, svc, r
}

func TestHistoryEndpoint(t *testing.T) {
	_, svc, r := setupHistoryTest()
	userID := "user1"
	// Simulate generations
	svc.GenerateCode("prompt1", userID)
	svc.GenerateCode("prompt2", userID)
	req, _ := http.NewRequest("GET", "/ai/history?user_id="+userID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "prompt1")
	assert.Contains(t, w.Body.String(), "prompt2")
}

func TestHistoryEndpointNoUser(t *testing.T) {
	_, _, r := setupHistoryTest()
	req, _ := http.NewRequest("GET", "/ai/history", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)
	assert.Contains(t, w.Body.String(), "user_id required")
}
