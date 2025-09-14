package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/EliasRanz/ai-code-gen/internal/auth"
)

func TestRefreshTokenUseCase_Execute(t *testing.T) {
	tests := []struct {
		name        string
		request     auth.RefreshTokenRequest
		setupMocks  func(*MockSessionRepository, *MockTokenProvider, *MockUserRepository)
		expectedErr string
	}{
		{
			name: "successful token refresh",
			request: auth.RefreshTokenRequest{
				RefreshToken: "valid_refresh_token",
			},
			setupMocks: func(sessionRepo *MockSessionRepository, tokenProvider *MockTokenProvider, userRepo *MockUserRepository) {
				session := auth.Session{
					ID:           auth.SessionID("session123"),
					UserID:       auth.UserID("user123"),
					AccessToken:  "old_access_token",
					RefreshToken: "valid_refresh_token",
					ExpiresAt:    time.Now().Add(1 * time.Hour), // Not expired
					Status:       auth.StatusActive,
				}
				user := auth.User{
					ID:     auth.UserID("user123"),
					Active: true,
				}

				sessionRepo.On("GetByRefreshToken", mock.Anything, "valid_refresh_token").Return(session, nil)
				userRepo.On("GetByID", mock.Anything, auth.UserID("user123")).Return(user, nil)
				tokenProvider.On("GenerateAccessToken", auth.UserID("user123")).Return("new_access_token", nil)
				tokenProvider.On("GenerateRefreshToken", auth.UserID("user123")).Return("new_refresh_token", nil)
				sessionRepo.On("Update", mock.Anything, mock.MatchedBy(func(s auth.Session) bool {
					return s.AccessToken == "new_access_token" && s.RefreshToken == "new_refresh_token"
				})).Return(nil)
			},
		},
		{
			name: "empty refresh token",
			request: auth.RefreshTokenRequest{
				RefreshToken: "",
			},
			setupMocks: func(sessionRepo *MockSessionRepository, tokenProvider *MockTokenProvider, userRepo *MockUserRepository) {
			},
			expectedErr: "refresh token is required",
		},
		{
			name: "invalid refresh token",
			request: auth.RefreshTokenRequest{
				RefreshToken: "invalid_token",
			},
			setupMocks: func(sessionRepo *MockSessionRepository, tokenProvider *MockTokenProvider, userRepo *MockUserRepository) {
				sessionRepo.On("GetByRefreshToken", mock.Anything, "invalid_token").Return(auth.Session{}, auth.NewNotFoundError("session not found"))
			},
			expectedErr: "invalid refresh token",
		},
		{
			name: "expired session",
			request: auth.RefreshTokenRequest{
				RefreshToken: "expired_token",
			},
			setupMocks: func(sessionRepo *MockSessionRepository, tokenProvider *MockTokenProvider, userRepo *MockUserRepository) {
				expiredSession := auth.Session{
					ID:           auth.SessionID("session123"),
					UserID:       auth.UserID("user123"),
					RefreshToken: "expired_token",
					ExpiresAt:    time.Now().Add(-1 * time.Hour), // Expired
				}
				sessionRepo.On("GetByRefreshToken", mock.Anything, "expired_token").Return(expiredSession, nil)
				sessionRepo.On("Delete", mock.Anything, auth.SessionID("session123")).Return(nil)
			},
			expectedErr: "refresh token expired",
		},
		{
			name: "user not found",
			request: auth.RefreshTokenRequest{
				RefreshToken: "valid_refresh_token",
			},
			setupMocks: func(sessionRepo *MockSessionRepository, tokenProvider *MockTokenProvider, userRepo *MockUserRepository) {
				session := auth.Session{
					ID:           auth.SessionID("session123"),
					UserID:       auth.UserID("user123"),
					RefreshToken: "valid_refresh_token",
					ExpiresAt:    time.Now().Add(1 * time.Hour),
				}
				sessionRepo.On("GetByRefreshToken", mock.Anything, "valid_refresh_token").Return(session, nil)
				userRepo.On("GetByID", mock.Anything, auth.UserID("user123")).Return(auth.User{}, auth.NewNotFoundError("user not found"))
				sessionRepo.On("Delete", mock.Anything, auth.SessionID("session123")).Return(nil)
			},
			expectedErr: "user not found",
		},
		{
			name: "inactive user",
			request: auth.RefreshTokenRequest{
				RefreshToken: "valid_refresh_token",
			},
			setupMocks: func(sessionRepo *MockSessionRepository, tokenProvider *MockTokenProvider, userRepo *MockUserRepository) {
				session := auth.Session{
					ID:           auth.SessionID("session123"),
					UserID:       auth.UserID("user123"),
					RefreshToken: "valid_refresh_token",
					ExpiresAt:    time.Now().Add(1 * time.Hour),
				}
				user := auth.User{
					ID:     auth.UserID("user123"),
					Active: false,
				}
				sessionRepo.On("GetByRefreshToken", mock.Anything, "valid_refresh_token").Return(session, nil)
				userRepo.On("GetByID", mock.Anything, auth.UserID("user123")).Return(user, nil)
				sessionRepo.On("Delete", mock.Anything, auth.SessionID("session123")).Return(nil)
			},
			expectedErr: "user account is inactive",
		},
		{
			name: "access token generation failure",
			request: auth.RefreshTokenRequest{
				RefreshToken: "valid_refresh_token",
			},
			setupMocks: func(sessionRepo *MockSessionRepository, tokenProvider *MockTokenProvider, userRepo *MockUserRepository) {
				session := auth.Session{
					ID:           auth.SessionID("session123"),
					UserID:       auth.UserID("user123"),
					RefreshToken: "valid_refresh_token",
					ExpiresAt:    time.Now().Add(1 * time.Hour),
				}
				user := auth.User{
					ID:     auth.UserID("user123"),
					Active: true,
				}
				sessionRepo.On("GetByRefreshToken", mock.Anything, "valid_refresh_token").Return(session, nil)
				userRepo.On("GetByID", mock.Anything, auth.UserID("user123")).Return(user, nil)
				tokenProvider.On("GenerateAccessToken", auth.UserID("user123")).Return("", assert.AnError)
			},
			expectedErr: "failed to generate access token",
		},
		{
			name: "refresh token generation failure",
			request: auth.RefreshTokenRequest{
				RefreshToken: "valid_refresh_token",
			},
			setupMocks: func(sessionRepo *MockSessionRepository, tokenProvider *MockTokenProvider, userRepo *MockUserRepository) {
				session := auth.Session{
					ID:           auth.SessionID("session123"),
					UserID:       auth.UserID("user123"),
					RefreshToken: "valid_refresh_token",
					ExpiresAt:    time.Now().Add(1 * time.Hour),
				}
				user := auth.User{
					ID:     auth.UserID("user123"),
					Active: true,
				}
				sessionRepo.On("GetByRefreshToken", mock.Anything, "valid_refresh_token").Return(session, nil)
				userRepo.On("GetByID", mock.Anything, auth.UserID("user123")).Return(user, nil)
				tokenProvider.On("GenerateAccessToken", auth.UserID("user123")).Return("new_access_token", nil)
				tokenProvider.On("GenerateRefreshToken", auth.UserID("user123")).Return("", assert.AnError)
			},
			expectedErr: "failed to generate refresh token",
		},
		{
			name: "session update failure",
			request: auth.RefreshTokenRequest{
				RefreshToken: "valid_refresh_token",
			},
			setupMocks: func(sessionRepo *MockSessionRepository, tokenProvider *MockTokenProvider, userRepo *MockUserRepository) {
				session := auth.Session{
					ID:           auth.SessionID("session123"),
					UserID:       auth.UserID("user123"),
					RefreshToken: "valid_refresh_token",
					ExpiresAt:    time.Now().Add(1 * time.Hour),
				}
				user := auth.User{
					ID:     auth.UserID("user123"),
					Active: true,
				}
				sessionRepo.On("GetByRefreshToken", mock.Anything, "valid_refresh_token").Return(session, nil)
				userRepo.On("GetByID", mock.Anything, auth.UserID("user123")).Return(user, nil)
				tokenProvider.On("GenerateAccessToken", auth.UserID("user123")).Return("new_access_token", nil)
				tokenProvider.On("GenerateRefreshToken", auth.UserID("user123")).Return("new_refresh_token", nil)
				sessionRepo.On("Update", mock.Anything, mock.AnythingOfType("auth.Session")).Return(assert.AnError)
			},
			expectedErr: "failed to update session",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			sessionRepo := &MockSessionRepository{}
			tokenProvider := &MockTokenProvider{}
			userRepo := &MockUserRepository{}

			tt.setupMocks(sessionRepo, tokenProvider, userRepo)

			// Create use case
			uc := auth.NewRefreshTokenUseCase(sessionRepo, tokenProvider, userRepo)

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
				assert.NotEmpty(t, result.AccessToken)
				assert.NotEmpty(t, result.RefreshToken)
				assert.False(t, result.ExpiresAt.IsZero())
			}

			// Verify mocks
			sessionRepo.AssertExpectations(t)
			tokenProvider.AssertExpectations(t)
			userRepo.AssertExpectations(t)
		})
	}
}
