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

func setupHistoryTest(ctrl *gomock.Controller) (*ai.Handler, *ai.Service, *gin.Engine) {
	gin.SetMode(gin.TestMode)

	mockLLM := mocks.NewMockLLMClient(ctrl)
	mockLLM.EXPECT().Generate(gomock.Any()).DoAndReturn(func(prompt string) (string, error) {
		return "<div>" + prompt + "</div>", nil
	}).AnyTimes()

	svc := ai.NewService(mockLLM)
	h := ai.NewHandler(svc)
	r := gin.New()
	group := r.Group("")
	h.RegisterRoutes(group)
	return h, svc, r
}

func TestHistoryEndpoint(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	_, svc, r := setupHistoryTest(ctrl)
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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	_, _, r := setupHistoryTest(ctrl)
	req, _ := http.NewRequest("GET", "/ai/history", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)
	assert.Contains(t, w.Body.String(), "user_id required")
}
