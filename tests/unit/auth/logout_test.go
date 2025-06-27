package authtest

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/EliasRanz/ai-code-gen/internal/auth"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestLogoutHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	service := CreateTestService()
	r.POST("/logout", func(c *gin.Context) {
		h := auth.NewHandler(service)
		h.Logout(c)
	})

	w := httptest.NewRecorder()
	body := `{"refresh_token": "sometoken"}`
	req, _ := http.NewRequest("POST", "/logout", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Logged out successfully")
}

func TestLogoutHandler_MissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	service := CreateTestService()
	r.POST("/logout", func(c *gin.Context) {
		h := auth.NewHandler(service)
		h.Logout(c)
	})

	w := httptest.NewRecorder()
	body := `{}`
	req, _ := http.NewRequest("POST", "/logout", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
