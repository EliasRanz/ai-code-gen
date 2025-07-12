package authtest

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/EliasRanz/ai-code-gen/internal/auth"
	"github.com/EliasRanz/ai-code-gen/internal/domain/common"
	"github.com/EliasRanz/ai-code-gen/internal/domain/user"
)

// MockUserRepositoryForMiddleware mocks the user.Repository for middleware tests
type MockUserRepositoryForMiddleware struct {
	mock.Mock
}

func (m *MockUserRepositoryForMiddleware) GetByID(ctx context.Context, id common.UserID) (user.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return user.User{}, args.Error(1)
	}
	return args.Get(0).(user.User), args.Error(1)
}

func (m *MockUserRepositoryForMiddleware) GetByEmail(ctx context.Context, email string) (user.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return user.User{}, args.Error(1)
	}
	return args.Get(0).(user.User), args.Error(1)
}

func (m *MockUserRepositoryForMiddleware) Create(ctx context.Context, u user.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

func (m *MockUserRepositoryForMiddleware) Update(ctx context.Context, u user.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

func (m *MockUserRepositoryForMiddleware) Delete(ctx context.Context, id common.UserID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepositoryForMiddleware) List(ctx context.Context, params common.PaginationParams, search string) ([]user.User, error) {
	args := m.Called(ctx, params, search)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]user.User), args.Error(1)
}

func (m *MockUserRepositoryForMiddleware) Count(ctx context.Context, search string) (int, error) {
	args := m.Called(ctx, search)
	return args.Int(0), args.Error(1)
}

func TestAuthMiddleware_NoHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tokenValidator := auth.NewJWTTokenProvider("test-secret", "test-issuer")
	mockUserRepo := &MockUserRepositoryForMiddleware{}

	router := gin.New()
	router.Use(auth.AuthMiddleware(tokenValidator, mockUserRepo))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusUnauthorized, resp.Code)
	assert.Contains(t, resp.Body.String(), "Authorization header required")
}

func TestAuthMiddleware_InvalidHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tokenValidator := auth.NewJWTTokenProvider("test-secret", "test-issuer")
	mockUserRepo := &MockUserRepositoryForMiddleware{}

	router := gin.New()
	router.Use(auth.AuthMiddleware(tokenValidator, mockUserRepo))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "InvalidToken")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusUnauthorized, resp.Code)
	assert.Contains(t, resp.Body.String(), "Invalid authorization header format")
}

func TestAuthMiddleware_EmptyToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tokenValidator := auth.NewJWTTokenProvider("test-secret", "test-issuer")
	mockUserRepo := &MockUserRepositoryForMiddleware{}

	router := gin.New()
	router.Use(auth.AuthMiddleware(tokenValidator, mockUserRepo))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer ")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusUnauthorized, resp.Code)
	assert.Contains(t, resp.Body.String(), "Token not provided or invalid")
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tokenValidator := auth.NewJWTTokenProvider("test-secret", "test-issuer")
	mockUserRepo := &MockUserRepositoryForMiddleware{}

	router := gin.New()
	router.Use(auth.AuthMiddleware(tokenValidator, mockUserRepo))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusUnauthorized, resp.Code)
	assert.Contains(t, resp.Body.String(), "Invalid or expired token")
}

func TestAuthMiddleware_UserNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tokenValidator := auth.NewJWTTokenProvider("test-secret", "test-issuer")
	token, _ := tokenValidator.GenerateAccessToken(auth.UserID("1"))

	mockUserRepo := &MockUserRepositoryForMiddleware{}
	mockUserRepo.On("GetByID", mock.Anything, common.UserID("1")).Return(user.User{}, errors.New("user not found"))

	router := gin.New()
	router.Use(auth.AuthMiddleware(tokenValidator, mockUserRepo))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusUnauthorized, resp.Code)
	assert.Contains(t, resp.Body.String(), "User not found")
}

func TestAuthMiddleware_InactiveUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tokenValidator := auth.NewJWTTokenProvider("test-secret", "test-issuer")
	token, _ := tokenValidator.GenerateAccessToken(auth.UserID("1"))

	mockUserRepo := &MockUserRepositoryForMiddleware{}
	mockUserRepo.On("GetByID", mock.Anything, common.UserID("1")).Return(user.User{
		ID:       common.UserID("1"),
		Username: "testuser",
		Email:    "test@example.com",
		Active:   false,
	}, nil)

	router := gin.New()
	router.Use(auth.AuthMiddleware(tokenValidator, mockUserRepo))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusForbidden, resp.Code)
	assert.Contains(t, resp.Body.String(), "User account is inactive")
}

func TestAuthMiddleware_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tokenValidator := auth.NewJWTTokenProvider("test-secret", "test-issuer")
	token, _ := tokenValidator.GenerateAccessToken(auth.UserID("1"))

	mockUserRepo := &MockUserRepositoryForMiddleware{}
	mockUserRepo.On("GetByID", mock.Anything, common.UserID("1")).Return(user.User{
		ID:       common.UserID("1"),
		Username: "testuser",
		Email:    "test@example.com",
		Active:   true,
		Roles:    []string{"user"},
	}, nil)

	router := gin.New()
	router.Use(auth.AuthMiddleware(tokenValidator, mockUserRepo))
	router.GET("/test", func(c *gin.Context) {
		authenticated, exists := c.Get("authenticated")
		assert.True(t, exists)
		assert.True(t, authenticated.(bool))

		userID, exists := c.Get("user_id")
		assert.True(t, exists)
		assert.Equal(t, "1", userID)

		c.JSON(200, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Body.String(), "success")
}

func TestLightweightAuthMiddleware_NoHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tokenValidator := auth.NewJWTTokenProvider("test-secret", "test-issuer")

	router := gin.New()
	router.Use(auth.LightweightAuthMiddleware(tokenValidator))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusUnauthorized, resp.Code)
	assert.Contains(t, resp.Body.String(), "Authorization header required")
}

func TestLightweightAuthMiddleware_ValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tokenValidator := auth.NewJWTTokenProvider("test-secret", "test-issuer")

	// Generate a valid token for testing
	token, _ := tokenValidator.GenerateAccessToken(auth.UserID("user123"))

	var contextUserID string
	var contextAuth bool

	router := gin.New()
	router.Use(auth.LightweightAuthMiddleware(tokenValidator))
	router.GET("/test", func(c *gin.Context) {
		userID, _, _, authenticated := auth.GetUserContextFromMiddleware(c)
		contextUserID = string(userID)
		contextAuth = authenticated
		c.JSON(200, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "user123", contextUserID)
	assert.Equal(t, true, contextAuth)
}

func TestAdminRequired_NotAuthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(auth.AdminRequired())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusUnauthorized, resp.Code)
	assert.Contains(t, resp.Body.String(), "Authentication required")
}

func TestAdminRequired_NotAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("authenticated", true)
		c.Set("user_id", "user123")
		c.Set("user_role", "user")
		c.Next()
	})
	router.Use(auth.AdminRequired())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusForbidden, resp.Code)
	assert.Contains(t, resp.Body.String(), "Admin access required")
}

func TestAdminRequired_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("authenticated", true)
		c.Set("user_id", "admin123")
		c.Set("user_role", "admin")
		c.Next()
	})
	router.Use(auth.AdminRequired())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Body.String(), "success")
}

func TestGetUserContextFromMiddleware_Authenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		c.Set("authenticated", true)
		c.Set("user_id", "user123")
		c.Set("user_email", "test@example.com")
		c.Set("user_role", "admin")

		userID, email, role, authenticated := auth.GetUserContextFromMiddleware(c)
		c.JSON(200, gin.H{
			"user_id":       userID,
			"email":         email,
			"role":          role,
			"authenticated": authenticated,
		})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Body.String(), "user123")
	assert.Contains(t, resp.Body.String(), "test@example.com")
	assert.Contains(t, resp.Body.String(), "admin")
	assert.Contains(t, resp.Body.String(), "true")
}

func TestGetUserContextFromMiddleware_NotAuthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		userID, email, role, authenticated := auth.GetUserContextFromMiddleware(c)
		c.JSON(200, gin.H{
			"user_id":       userID,
			"email":         email,
			"role":          role,
			"authenticated": authenticated,
		})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Body.String(), `"user_id":""`)
	assert.Contains(t, resp.Body.String(), `"email":""`)
	assert.Contains(t, resp.Body.String(), `"role":""`)
	assert.Contains(t, resp.Body.String(), `"authenticated":false`)
}

func TestIsAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should return true for admin user", func(t *testing.T) {
		router := gin.New()
		router.GET("/test", func(c *gin.Context) {
			c.Set("user_role", "admin")
			result := auth.IsAdmin(c)
			c.JSON(200, gin.H{"is_admin": result})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Contains(t, resp.Body.String(), `"is_admin":true`)
	})

	t.Run("should return false for non-admin user", func(t *testing.T) {
		router := gin.New()
		router.GET("/test", func(c *gin.Context) {
			c.Set("user_role", "user")
			result := auth.IsAdmin(c)
			c.JSON(200, gin.H{"is_admin": result})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Contains(t, resp.Body.String(), `"is_admin":false`)
	})
}

func TestIsAuthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should return true for authenticated user", func(t *testing.T) {
		router := gin.New()
		router.GET("/test", func(c *gin.Context) {
			c.Set("authenticated", true)
			result := auth.IsAuthenticated(c)
			c.JSON(200, gin.H{"is_authenticated": result})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Contains(t, resp.Body.String(), `"is_authenticated":true`)
	})

	t.Run("should return false for unauthenticated user", func(t *testing.T) {
		router := gin.New()
		router.GET("/test", func(c *gin.Context) {
			result := auth.IsAuthenticated(c)
			c.JSON(200, gin.H{"is_authenticated": result})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Contains(t, resp.Body.String(), `"is_authenticated":false`)
	})
}
