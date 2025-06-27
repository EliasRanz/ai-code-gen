package authtest

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/EliasRanz/ai-code-gen/ai-ui-generator/internal/auth"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRefreshTokenHandlerSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	// Setup TokenManager and Service with a known userID
	secret := "testsecret"
	issuer := "testissuer"
	tm := auth.NewTokenManager(secret, issuer)
	userID := "user123"
	refreshToken, err := tm.GenerateRefreshToken(userID)
	assert.NoError(t, err)

	service := auth.NewService(&MockUserRepository{}, tm)
	r.POST("/refresh", func(c *gin.Context) {
		h := auth.NewHandler(service)
		h.RefreshToken(c)
	})

	w := httptest.NewRecorder()
	body := `{"refresh_token": "` + refreshToken + `"}`
	req, _ := http.NewRequest("POST", "/refresh", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "access_token")
	assert.Contains(t, w.Body.String(), "refresh_token")
}

func TestRefreshTokenHandlerInvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	secret := "testsecret"
	issuer := "testissuer"
	tm := auth.NewTokenManager(secret, issuer)
	service := auth.NewService(&MockUserRepository{}, tm)
	r.POST("/refresh", func(c *gin.Context) {
		h := auth.NewHandler(service)
		h.RefreshToken(c)
	})

	w := httptest.NewRecorder()
	body := `{"refresh_token": "invalidtoken"}`
	req, _ := http.NewRequest("POST", "/refresh", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
