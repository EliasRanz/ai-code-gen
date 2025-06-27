package authtest

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/EliasRanz/ai-code-gen/internal/auth"
	"github.com/EliasRanz/ai-code-gen/internal/user"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestJWTMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		setupAuth      func() *auth.Service
		authHeader     string
		expectedStatus int
		expectedError  string
		expectContext  bool
	}{
		{
			name: "valid token",
			setupAuth: func() *auth.Service {
				userRepo := new(MockUserRepository)
				tokenManager := auth.NewTokenManager("test-secret", "test-issuer")
				service := auth.NewService(userRepo, tokenManager)

				testUser := &user.User{
					ID:       "user-123",
					Email:    "user@example.com",
					Roles:    []string{"user"},
					IsActive: true,
				}

				userRepo.On("GetByID", "user-123").Return(testUser, nil)
				return service
			},
			authHeader:     "", // Will be set dynamically
			expectedStatus: http.StatusOK,
			expectContext:  true,
		},
		{
			name: "missing authorization header",
			setupAuth: func() *auth.Service {
				return auth.NewService(new(MockUserRepository), auth.NewTokenManager("test-secret", "test-issuer"))
			},
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Authorization header required",
		},
		{
			name: "invalid header format",
			setupAuth: func() *auth.Service {
				return auth.NewService(new(MockUserRepository), auth.NewTokenManager("test-secret", "test-issuer"))
			},
			authHeader:     "InvalidFormat token",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid authorization header format",
		},
		{
			name: "empty token",
			setupAuth: func() *auth.Service {
				return auth.NewService(new(MockUserRepository), auth.NewTokenManager("test-secret", "test-issuer"))
			},
			authHeader:     "Bearer ",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Token required",
		},
		{
			name: "invalid token",
			setupAuth: func() *auth.Service {
				return auth.NewService(new(MockUserRepository), auth.NewTokenManager("test-secret", "test-issuer"))
			},
			authHeader:     "Bearer invalid-token",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid or expired token",
		},
		{
			name: "valid token but inactive user",
			setupAuth: func() *auth.Service {
				userRepo := new(MockUserRepository)
				tokenManager := auth.NewTokenManager("test-secret", "test-issuer")
				service := auth.NewService(userRepo, tokenManager)

				// Mock ValidateToken to return user ID, but user is inactive
				inactiveUser := &user.User{
					ID:       "inactive-123",
					Email:    "inactive@example.com",
					Roles:    []string{"user"},
					IsActive: false,
				}

				userRepo.On("GetByID", "inactive-123").Return(inactiveUser, nil)
				return service
			},
			authHeader:     "", // Will be set dynamically with inactive user
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid or expired token", // This will be caught by ValidateToken
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authService := tt.setupAuth()

			// Create a test endpoint
			r := gin.New()
			r.Use(auth.JWTMiddleware(authService))
			r.GET("/protected", func(c *gin.Context) {
				// Check if context was set properly
				if tt.expectContext {
					userID, exists := auth.GetUserID(c)
					assert.True(t, exists)
					assert.NotEmpty(t, userID)

					email, exists := auth.GetUserEmail(c)
					assert.True(t, exists)
					assert.NotEmpty(t, email)

					roles, exists := auth.GetUserRoles(c)
					assert.True(t, exists)
					assert.NotEmpty(t, roles)

					user, exists := auth.GetUser(c)
					assert.True(t, exists)
					assert.NotNil(t, user)
				}
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			// Create request
			req := httptest.NewRequest("GET", "/protected", nil)

			// Set authorization header
			authHeader := tt.authHeader
			if tt.name == "valid token" {
				// Generate a valid token for this test
				token, err := authService.TokenManager.GenerateToken("user-123", 15*time.Minute)
				assert.NoError(t, err)
				authHeader = "Bearer " + token
			} else if tt.name == "valid token but inactive user" {
				// Generate a valid token for inactive user
				token, err := authService.TokenManager.GenerateToken("inactive-123", 15*time.Minute)
				assert.NoError(t, err)
				authHeader = "Bearer " + token
			}

			if authHeader != "" {
				req.Header.Set("Authorization", authHeader)
			}

			// Record response
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			// Assertions
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			}
		})
	}
}

func TestRequireAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userRoles      []string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "admin user",
			userRoles:      []string{"admin", "user"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "non-admin user",
			userRoles:      []string{"user"},
			expectedStatus: http.StatusForbidden,
			expectedError:  "Admin access required",
		},
		{
			name:           "no roles",
			userRoles:      []string{},
			expectedStatus: http.StatusForbidden,
			expectedError:  "Admin access required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()

			// Mock the authentication middleware
			r.Use(func(c *gin.Context) {
				c.Set("user_id", "test-user")
				c.Set("user_email", "test@example.com")
				c.Set("user_roles", tt.userRoles)
				c.Next()
			})

			r.Use(auth.RequireAdmin())
			r.GET("/admin", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "admin access granted"})
			})

			req := httptest.NewRequest("GET", "/admin", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			}
		})
	}
}

func TestHelperFunctions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/test", func(c *gin.Context) {
		// Set test context
		testUser := &user.User{
			ID:    "test-123",
			Email: "test@example.com",
			Roles: []string{"admin", "user"},
		}

		c.Set("user_id", testUser.ID)
		c.Set("user_email", testUser.Email)
		c.Set("user_roles", testUser.Roles)
		c.Set("user", testUser)

		// Test helper functions
		userID, exists := auth.GetUserID(c)
		assert.True(t, exists)
		assert.Equal(t, "test-123", userID)

		email, exists := auth.GetUserEmail(c)
		assert.True(t, exists)
		assert.Equal(t, "test@example.com", email)

		roles, exists := auth.GetUserRoles(c)
		assert.True(t, exists)
		assert.Equal(t, []string{"admin", "user"}, roles)

		user, exists := auth.GetUser(c)
		assert.True(t, exists)
		assert.Equal(t, testUser, user)

		// Test role checking
		assert.True(t, auth.HasRole(c, "admin"))
		assert.True(t, auth.HasRole(c, "user"))
		assert.False(t, auth.HasRole(c, "moderator"))
		assert.True(t, auth.IsAdmin(c))

		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
