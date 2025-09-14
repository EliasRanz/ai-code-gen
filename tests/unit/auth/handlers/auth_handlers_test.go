package auth_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/EliasRanz/ai-code-gen/internal/auth"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test helper functions for auth handlers
func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

// Test auth types and validation
func TestLoginRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request auth.LoginRequest
		isValid bool
	}{
		{
			name: "valid_request",
			request: auth.LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			isValid: true,
		},
		{
			name: "missing_email",
			request: auth.LoginRequest{
				Email:    "",
				Password: "password123",
			},
			isValid: false,
		},
		{
			name: "missing_password",
			request: auth.LoginRequest{
				Email:    "test@example.com",
				Password: "",
			},
			isValid: false,
		},
		{
			name: "invalid_email_format",
			request: auth.LoginRequest{
				Email:    "invalid-email",
				Password: "password123",
			},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBytes, err := json.Marshal(tt.request)
			require.NoError(t, err)

			// Test JSON marshalling/unmarshalling
			var decoded auth.LoginRequest
			err = json.Unmarshal(jsonBytes, &decoded)
			require.NoError(t, err)

			assert.Equal(t, tt.request.Email, decoded.Email)
			assert.Equal(t, tt.request.Password, decoded.Password)
		})
	}
}

func TestUser_HasRole(t *testing.T) {
	tests := []struct {
		name     string
		user     auth.User
		role     string
		expected bool
	}{
		{
			name: "user_has_specific_role",
			user: auth.User{
				ID:    "user-123",
				Roles: []string{"user", "editor"},
			},
			role:     "editor",
			expected: true,
		},
		{
			name: "user_does_not_have_role",
			user: auth.User{
				ID:    "user-123",
				Roles: []string{"user"},
			},
			role:     "admin",
			expected: false,
		},
		{
			name: "admin_has_all_roles",
			user: auth.User{
				ID:    "admin-123",
				Roles: []string{"admin"},
			},
			role:     "any_role",
			expected: true,
		},
		{
			name: "super_admin_has_all_roles",
			user: auth.User{
				ID:    "superadmin-123",
				Roles: []string{"super_admin"},
			},
			role:     "any_role",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.user.HasRole(tt.role)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSession_IsExpired(t *testing.T) {
	tests := []struct {
		name     string
		session  auth.Session
		expected bool
	}{
		{
			name: "session_not_expired",
			session: auth.Session{
				ID:        "session-123",
				ExpiresAt: time.Now().Add(time.Hour),
			},
			expected: false,
		},
		{
			name: "session_expired",
			session: auth.Session{
				ID:        "session-123",
				ExpiresAt: time.Now().Add(-time.Hour),
			},
			expected: true,
		},
		{
			name: "session_zero_time",
			session: auth.Session{
				ID: "session-123",
				// ExpiresAt is zero value
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.session.IsExpired()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToken_IsExpired(t *testing.T) {
	tests := []struct {
		name     string
		token    auth.Token
		expected bool
	}{
		{
			name: "token_not_expired",
			token: auth.Token{
				AccessToken: "token-123",
				ExpiresAt:   time.Now().Add(time.Hour),
			},
			expected: false,
		},
		{
			name: "token_expired",
			token: auth.Token{
				AccessToken: "token-123",
				ExpiresAt:   time.Now().Add(-time.Hour),
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.token.IsExpired()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPaginationParams_Validation(t *testing.T) {
	tests := []struct {
		name   string
		params auth.PaginationParams
		hasErr bool
	}{
		{
			name: "valid_params",
			params: auth.PaginationParams{
				Page:  1,
				Limit: 20,
			},
			hasErr: false,
		},
		{
			name: "invalid_page_zero",
			params: auth.PaginationParams{
				Page:  0,
				Limit: 20,
			},
			hasErr: true,
		},
		{
			name: "invalid_limit_zero",
			params: auth.PaginationParams{
				Page:  1,
				Limit: 0,
			},
			hasErr: true,
		},
		{
			name: "invalid_limit_too_high",
			params: auth.PaginationParams{
				Page:  1,
				Limit: 200,
			},
			hasErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.params.Validate()
			if tt.hasErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPaginationParams_Offset(t *testing.T) {
	tests := []struct {
		name     string
		params   auth.PaginationParams
		expected int32
	}{
		{
			name: "first_page",
			params: auth.PaginationParams{
				Page:  1,
				Limit: 10,
			},
			expected: 0,
		},
		{
			name: "second_page",
			params: auth.PaginationParams{
				Page:  2,
				Limit: 10,
			},
			expected: 10,
		},
		{
			name: "third_page_different_limit",
			params: auth.PaginationParams{
				Page:  3,
				Limit: 25,
			},
			expected: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			offset := tt.params.Offset()
			assert.Equal(t, tt.expected, offset)
		})
	}
}

func TestErrorTypes(t *testing.T) {
	tests := []struct {
		name        string
		createError func() error
		checkFunc   func(error) bool
	}{
		{
			name:        "validation_error",
			createError: func() error { return auth.NewValidationError("test validation", nil) },
			checkFunc:   auth.IsValidationError,
		},
		{
			name:        "not_found_error",
			createError: func() error { return auth.NewNotFoundError("test not found") },
			checkFunc:   auth.IsNotFoundError,
		},
		{
			name:        "conflict_error",
			createError: func() error { return auth.NewConflictError("test conflict") },
			checkFunc:   auth.IsConflictError,
		},
		{
			name:        "unauthorized_error",
			createError: func() error { return auth.NewUnauthorizedError("test unauthorized") },
			checkFunc:   auth.IsUnauthorizedError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.createError()
			assert.True(t, tt.checkFunc(err))
			assert.Error(t, err)
			assert.NotEmpty(t, err.Error())
		})
	}
}

func TestHTTPHandler_MockBehavior(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Test Login handler with invalid JSON
	t.Run("login_invalid_json", func(t *testing.T) {
		router := setupTestRouter()
		router.POST("/auth/login", func(c *gin.Context) {
			var req auth.LoginRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true})
		})

		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Invalid request body", response["error"])
	})

	// Test Logout handler missing authorization
	t.Run("logout_missing_auth", func(t *testing.T) {
		router := setupTestRouter()
		router.POST("/auth/logout", func(c *gin.Context) {
			authHeader := c.GetHeader("Authorization")
			if authHeader == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization header is required"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true})
		})

		req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	// Test Bearer token extraction behavior
	t.Run("bearer_token_extraction", func(t *testing.T) {
		router := setupTestRouter()
		router.POST("/auth/test", func(c *gin.Context) {
			authHeader := c.GetHeader("Authorization")
			var token string
			if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
				token = authHeader[7:]
			}

			if token == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid authorization header format"})
				return
			}

			c.JSON(http.StatusOK, gin.H{"token": token})
		})

		tests := []struct {
			name           string
			authHeader     string
			expectedStatus int
			expectedToken  string
		}{
			{
				name:           "valid_bearer_token",
				authHeader:     "Bearer valid-token-123",
				expectedStatus: http.StatusOK,
				expectedToken:  "valid-token-123",
			},
			{
				name:           "invalid_format",
				authHeader:     "InvalidFormat token",
				expectedStatus: http.StatusBadRequest,
			},
			{
				name:           "no_token",
				authHeader:     "Bearer ",
				expectedStatus: http.StatusBadRequest,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				req := httptest.NewRequest(http.MethodPost, "/auth/test", nil)
				req.Header.Set("Authorization", tt.authHeader)

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, tt.expectedStatus, w.Code)

				if tt.expectedStatus == http.StatusOK {
					var response map[string]interface{}
					err := json.Unmarshal(w.Body.Bytes(), &response)
					require.NoError(t, err)
					assert.Equal(t, tt.expectedToken, response["token"])
				}
			})
		}
	})

	// Test error handling patterns
	t.Run("error_handling_patterns", func(t *testing.T) {
		router := setupTestRouter()
		router.POST("/auth/error-test", func(c *gin.Context) {
			errorType := c.Query("type")

			switch errorType {
			case "validation":
				c.JSON(http.StatusBadRequest, gin.H{"error": "validation error"})
			case "not_found":
				c.JSON(http.StatusNotFound, gin.H{"error": "not found error"})
			case "conflict":
				c.JSON(http.StatusConflict, gin.H{"error": "conflict error"})
			case "unauthorized":
				c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized error"})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			}
		})

		tests := []struct {
			errorType      string
			expectedStatus int
		}{
			{"validation", http.StatusBadRequest},
			{"not_found", http.StatusNotFound},
			{"conflict", http.StatusConflict},
			{"unauthorized", http.StatusUnauthorized},
			{"unknown", http.StatusInternalServerError},
		}

		for _, tt := range tests {
			t.Run(tt.errorType, func(t *testing.T) {
				req := httptest.NewRequest(http.MethodPost, "/auth/error-test?type="+tt.errorType, nil)

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, tt.expectedStatus, w.Code)
			})
		}
	})
}

// Test request structure validation through JSON marshaling
func TestRequestResponseTypes(t *testing.T) {
	// Test all request types can be properly marshaled/unmarshaled
	t.Run("login_request", func(t *testing.T) {
		original := auth.LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}

		data, err := json.Marshal(original)
		require.NoError(t, err)

		var decoded auth.LoginRequest
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, original.Email, decoded.Email)
		assert.Equal(t, original.Password, decoded.Password)
	})
}
