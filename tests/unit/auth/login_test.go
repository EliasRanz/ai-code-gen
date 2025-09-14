package auth_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/EliasRanz/ai-code-gen/internal/auth"
)

func TestLoginUseCase_Execute(t *testing.T) {
	tests := []struct {
		name        string
		request     auth.LoginRequest
		setupMocks  func(*MockUserRepository, *MockSessionRepository, *MockPasswordHasher, *MockTokenProvider)
		expectedErr string
		expectNil   bool
	}{
		{
			name: "successful login",
			request: auth.LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMocks: func(userRepo *MockUserRepository, sessionRepo *MockSessionRepository, hasher *MockPasswordHasher, tokenProvider *MockTokenProvider) {
				user := auth.User{
					ID:       auth.UserID("user123"),
					Email:    "test@example.com",
					Password: "hashedpassword",
					Active:   true,
					Roles:    []string{"user"},
				}
				userRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(user, nil)
				hasher.On("Verify", "password123", "hashedpassword").Return(true)
				tokenProvider.On("GenerateAccessToken", auth.UserID("user123")).Return("access_token", nil)
				tokenProvider.On("GenerateRefreshToken", auth.UserID("user123")).Return("refresh_token", nil)
				sessionRepo.On("Create", mock.Anything, mock.MatchedBy(func(s auth.Session) bool {
					return s.UserID == auth.UserID("user123") && s.AccessToken == "access_token" && s.RefreshToken == "refresh_token"
				})).Return(nil)
			},
		},
		{
			name: "user not found",
			request: auth.LoginRequest{
				Email:    "notfound@example.com",
				Password: "password123",
			},
			setupMocks: func(userRepo *MockUserRepository, sessionRepo *MockSessionRepository, hasher *MockPasswordHasher, tokenProvider *MockTokenProvider) {
				userRepo.On("GetByEmail", mock.Anything, "notfound@example.com").Return(auth.User{}, auth.NewNotFoundError("user not found"))
			},
			expectedErr: "invalid credentials",
		},
		{
			name: "inactive user",
			request: auth.LoginRequest{
				Email:    "inactive@example.com",
				Password: "password123",
			},
			setupMocks: func(userRepo *MockUserRepository, sessionRepo *MockSessionRepository, hasher *MockPasswordHasher, tokenProvider *MockTokenProvider) {
				user := auth.User{
					ID:       auth.UserID("user123"),
					Email:    "inactive@example.com",
					Password: "hashedpassword",
					Active:   false,
					Roles:    []string{"user"},
				}
				userRepo.On("GetByEmail", mock.Anything, "inactive@example.com").Return(user, nil)
			},
			expectedErr: "user account is inactive",
		},
		{
			name: "invalid password",
			request: auth.LoginRequest{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			setupMocks: func(userRepo *MockUserRepository, sessionRepo *MockSessionRepository, hasher *MockPasswordHasher, tokenProvider *MockTokenProvider) {
				user := auth.User{
					ID:       auth.UserID("user123"),
					Email:    "test@example.com",
					Password: "hashedpassword",
					Active:   true,
					Roles:    []string{"user"},
				}
				userRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(user, nil)
				hasher.On("Verify", "wrongpassword", "hashedpassword").Return(false)
			},
			expectedErr: "invalid credentials",
		},
		{
			name: "access token generation failure",
			request: auth.LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMocks: func(userRepo *MockUserRepository, sessionRepo *MockSessionRepository, hasher *MockPasswordHasher, tokenProvider *MockTokenProvider) {
				user := auth.User{
					ID:       auth.UserID("user123"),
					Email:    "test@example.com",
					Password: "hashedpassword",
					Active:   true,
					Roles:    []string{"user"},
				}
				userRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(user, nil)
				hasher.On("Verify", "password123", "hashedpassword").Return(true)
				tokenProvider.On("GenerateAccessToken", auth.UserID("user123")).Return("", assert.AnError)
			},
			expectedErr: "failed to generate access token",
		},
		{
			name: "refresh token generation failure",
			request: auth.LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMocks: func(userRepo *MockUserRepository, sessionRepo *MockSessionRepository, hasher *MockPasswordHasher, tokenProvider *MockTokenProvider) {
				user := auth.User{
					ID:       auth.UserID("user123"),
					Email:    "test@example.com",
					Password: "hashedpassword",
					Active:   true,
					Roles:    []string{"user"},
				}
				userRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(user, nil)
				hasher.On("Verify", "password123", "hashedpassword").Return(true)
				tokenProvider.On("GenerateAccessToken", auth.UserID("user123")).Return("access_token", nil)
				tokenProvider.On("GenerateRefreshToken", auth.UserID("user123")).Return("", assert.AnError)
			},
			expectedErr: "failed to generate refresh token",
		},
		{
			name: "session creation failure",
			request: auth.LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMocks: func(userRepo *MockUserRepository, sessionRepo *MockSessionRepository, hasher *MockPasswordHasher, tokenProvider *MockTokenProvider) {
				user := auth.User{
					ID:       auth.UserID("user123"),
					Email:    "test@example.com",
					Password: "hashedpassword",
					Active:   true,
					Roles:    []string{"user"},
				}
				userRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(user, nil)
				hasher.On("Verify", "password123", "hashedpassword").Return(true)
				tokenProvider.On("GenerateAccessToken", auth.UserID("user123")).Return("access_token", nil)
				tokenProvider.On("GenerateRefreshToken", auth.UserID("user123")).Return("refresh_token", nil)
				sessionRepo.On("Create", mock.Anything, mock.AnythingOfType("auth.Session")).Return(assert.AnError)
			},
			expectedErr: "failed to create session",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			userRepo := &MockUserRepository{}
			sessionRepo := &MockSessionRepository{}
			hasher := &MockPasswordHasher{}
			tokenProvider := &MockTokenProvider{}

			tt.setupMocks(userRepo, sessionRepo, hasher, tokenProvider)

			// Create use case
			uc := auth.NewLoginUseCase(userRepo, sessionRepo, hasher, tokenProvider)

			// Execute
			result, err := uc.Execute(context.Background(), tt.request)

			// Assertions
			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.request.Email, result.User.Email)
				assert.NotEmpty(t, result.AccessToken)
				assert.NotEmpty(t, result.RefreshToken)
				assert.False(t, result.ExpiresAt.IsZero())
				assert.NotNil(t, result.Session)
			}

			// Verify mocks
			userRepo.AssertExpectations(t)
			sessionRepo.AssertExpectations(t)
			hasher.AssertExpectations(t)
			tokenProvider.AssertExpectations(t)
		})
	}
}
