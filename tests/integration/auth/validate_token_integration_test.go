package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	appAuth "github.com/EliasRanz/ai-code-gen/internal/auth"
	httpInterface "github.com/EliasRanz/ai-code-gen/internal/interfaces/http"
	"github.com/EliasRanz/ai-code-gen/internal/observability"
)

// Mock implementations for integration testing
type MockTokenProvider struct {
	mock.Mock
}

func (m *MockTokenProvider) GenerateAccessToken(userID appAuth.UserID) (string, error) {
	args := m.Called(userID)
	return args.String(0), args.Error(1)
}

func (m *MockTokenProvider) GenerateRefreshToken(userID appAuth.UserID) (string, error) {
	args := m.Called(userID)
	return args.String(0), args.Error(1)
}

func (m *MockTokenProvider) ValidateAccessToken(token string) (appAuth.UserID, error) {
	args := m.Called(token)
	return args.Get(0).(appAuth.UserID), args.Error(1)
}

func (m *MockTokenProvider) ValidateRefreshToken(token string) (appAuth.UserID, error) {
	args := m.Called(token)
	return args.Get(0).(appAuth.UserID), args.Error(1)
}

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetByID(ctx context.Context, id appAuth.UserID) (appAuth.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(appAuth.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (appAuth.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(appAuth.User), args.Error(1)
}

func (m *MockUserRepository) Create(ctx context.Context, u appAuth.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

func (m *MockUserRepository) Update(ctx context.Context, u appAuth.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id appAuth.UserID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, params appAuth.PaginationParams, search string) ([]appAuth.User, error) {
	args := m.Called(ctx, params, search)
	return args.Get(0).([]appAuth.User), args.Error(1)
}

func (m *MockUserRepository) Count(ctx context.Context, search string) (int, error) {
	args := m.Called(ctx, search)
	return args.Get(0).(int), args.Error(1)
}

// TestValidateTokenIntegration tests the full integration of the validate token endpoint
func TestValidateTokenIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("end-to-end token validation success", func(t *testing.T) {
		// Setup
		logger := observability.NewLogger("debug", "console")
		tokenProvider := new(MockTokenProvider)
		userRepo := new(MockUserRepository)

		// Create real use case with mocked dependencies
		validateTokenUC := appAuth.NewValidateToken(tokenProvider, userRepo)

		// Create handler with real use case
		handler := httpInterface.NewAuthHandler(nil, nil, nil, validateTokenUC, nil, nil, nil, logger)

		router := gin.New()
		authGroup := router.Group("/auth")
		authGroup.POST("/validate", handler.ValidateToken)

		// Setup test data
		userID := appAuth.UserID("test-user-id")
		testUser := appAuth.User{
			ID:       userID,
			Email:    "test@example.com",
			Username: "testuser",
			Name:     "Test User",
			Active:   true,
			Roles:    []string{"user", "admin"},
		}

		// Setup mock expectations
		tokenProvider.On("ValidateAccessToken", "valid-token").Return(userID, nil)
		userRepo.On("GetByID", mock.Anything, userID).Return(testUser, nil)

		// Prepare request
		reqBody := appAuth.ValidateTokenRequest{
			AccessToken: "valid-token",
		}
		body, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/auth/validate", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		// Assert response
		assert.Equal(t, http.StatusOK, resp.Code)

		var responseBody appAuth.ValidateTokenResponse
		err := json.Unmarshal(resp.Body.Bytes(), &responseBody)
		assert.NoError(t, err)

		// Verify response structure
		assert.True(t, responseBody.Valid)
		assert.Empty(t, responseBody.Error)
		assert.NotNil(t, responseBody.UserContext)

		// Verify user context details
		assert.Equal(t, userID, responseBody.UserContext.UserID)
		assert.Equal(t, "test@example.com", responseBody.UserContext.Email)
		assert.Equal(t, "testuser", responseBody.UserContext.Username)
		assert.Equal(t, "Test User", responseBody.UserContext.Name)
		assert.True(t, responseBody.UserContext.Active)
		assert.Equal(t, "user", responseBody.UserContext.Role) // First role
		assert.Equal(t, []string{"user", "admin"}, responseBody.UserContext.Roles)

		// Verify mocks were called correctly
		tokenProvider.AssertExpectations(t)
		userRepo.AssertExpectations(t)
	})

	t.Run("end-to-end token validation failure", func(t *testing.T) {
		// Setup
		logger := observability.NewLogger("debug", "console")
		tokenProvider := new(MockTokenProvider)
		userRepo := new(MockUserRepository)

		validateTokenUC := appAuth.NewValidateToken(tokenProvider, userRepo)
		handler := httpInterface.NewAuthHandler(nil, nil, nil, validateTokenUC, nil, nil, nil, logger)

		router := gin.New()
		authGroup := router.Group("/auth")
		authGroup.POST("/validate", handler.ValidateToken)

		// Setup mock for invalid token
		tokenProvider.On("ValidateAccessToken", "invalid-token").Return(appAuth.UserID(""), assert.AnError)

		// Prepare request
		reqBody := appAuth.ValidateTokenRequest{
			AccessToken: "invalid-token",
		}
		body, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/auth/validate", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		// Assert response
		assert.Equal(t, http.StatusOK, resp.Code)

		var responseBody appAuth.ValidateTokenResponse
		err := json.Unmarshal(resp.Body.Bytes(), &responseBody)
		assert.NoError(t, err)

		assert.False(t, responseBody.Valid)
		assert.Equal(t, "invalid or expired token", responseBody.Error)
		assert.Nil(t, responseBody.UserContext)

		tokenProvider.AssertExpectations(t)
	})
}
