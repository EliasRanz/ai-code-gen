package auth

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/EliasRanz/ai-code-gen/ai-ui-generator/internal/user"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRegisterHandler_Validation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	// Provide a dummy logger in context
	r.Use(func(c *gin.Context) {
		c.Set("logger", &testLogger{})
		c.Next()
	})
	r.POST("/register", func(c *gin.Context) {
		handler := NewHandler(&Service{})
		handler.Register(c)
	})

	// Test missing fields
	w := httptest.NewRecorder()
	body := `{"email": "", "password": ""}`
	req, _ := http.NewRequest("POST", "/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRegisterHandler_DuplicateUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.Use(func(c *gin.Context) {
		c.Set("logger", &testLogger{})
		c.Next()
	})
	service := &Service{userRepo: &mockUserRepo{exists: true}}
	r.POST("/register", func(c *gin.Context) {
		handler := NewHandler(service)
		handler.Register(c)
	})

	w := httptest.NewRecorder()
	body := `{"email": "test@example.com", "password": "password123"}`
	req, _ := http.NewRequest("POST", "/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestRegisterHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.Use(func(c *gin.Context) {
		c.Set("logger", &testLogger{})
		c.Next()
	})

	// Create token manager and service with proper initialization
	tokenManager := NewTokenManager("test-secret", "test-issuer")
	service := &Service{
		userRepo:     &mockUserRepo{exists: false, createOK: true},
		TokenManager: tokenManager,
	}

	r.POST("/register", func(c *gin.Context) {
		handler := NewHandler(service)
		handler.Register(c)
	})

	w := httptest.NewRecorder()
	body := `{"email": "newuser@example.com", "password": "password123"}`
	req, _ := http.NewRequest("POST", "/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "access_token")
	assert.Contains(t, w.Body.String(), "refresh_token")
}

type testLogger struct{}

func (l *testLogger) Info() interface{ Msg(string) } { return l }
func (l *testLogger) Msg(msg string)                 {}

type mockUserRepo struct {
	exists   bool
	createOK bool
}

func (m *mockUserRepo) GetByEmail(email string) (*user.User, error) {
	if m.exists {
		return &user.User{Email: email}, nil
	}
	return nil, nil
}
func (m *mockUserRepo) GetByID(id string) (*user.User, error) { return nil, nil }
func (m *mockUserRepo) Create(u *user.User) error {
	if m.createOK {
		u.ID = "mock-id"
		return nil
	}
	return fmt.Errorf("create failed")
}
func (m *mockUserRepo) Update(id string, updates map[string]interface{}) (*user.User, error) {
	return nil, nil
}
func (m *mockUserRepo) Delete(id string) error                       { return nil }
func (m *mockUserRepo) List(limit, offset int) ([]*user.User, error) { return nil, nil }
