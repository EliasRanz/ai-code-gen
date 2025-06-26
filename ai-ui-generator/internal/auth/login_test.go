package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/EliasRanz/ai-code-gen/ai-ui-generator/internal/user"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestLoginHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	// Create token manager and service with proper initialization
	tokenManager := NewTokenManager("test-secret", "test-issuer")
	service := &Service{
		userRepo:     &mockUserRepoLogin{valid: true},
		TokenManager: tokenManager,
	}

	handler := NewHandler(service)
	r.POST("/login", handler.Login)

	w := httptest.NewRecorder()
	body := `{"email": "user@example.com", "password": "password123"}`
	req, _ := http.NewRequest("POST", "/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "access_token")
}

func TestLoginHandler_InvalidCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	// Create token manager and service with proper initialization
	tokenManager := NewTokenManager("test-secret", "test-issuer")
	service := &Service{
		userRepo:     &mockUserRepoLogin{valid: false},
		TokenManager: tokenManager,
	}

	handler := NewHandler(service)
	r.POST("/login", handler.Login)

	w := httptest.NewRecorder()
	body := `{"email": "user@example.com", "password": "wrongpass"}`
	req, _ := http.NewRequest("POST", "/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

type mockUserRepoLogin struct{ valid bool }

func (m *mockUserRepoLogin) GetByEmail(email string) (*user.User, error) {
	if m.valid {
		hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
		return &user.User{ID: "mock-id", Email: email, PasswordHash: string(hash)}, nil
	}
	return &user.User{ID: "mock-id", Email: email, PasswordHash: "$2a$10$invalidhash"}, nil
}
func (m *mockUserRepoLogin) GetByID(id string) (*user.User, error) { return nil, nil }
func (m *mockUserRepoLogin) Create(u *user.User) error             { return nil }
func (m *mockUserRepoLogin) Update(id string, updates map[string]interface{}) (*user.User, error) {
	return nil, nil
}
func (m *mockUserRepoLogin) Delete(id string) error                       { return nil }
func (m *mockUserRepoLogin) List(limit, offset int) ([]*user.User, error) { return nil, nil }
