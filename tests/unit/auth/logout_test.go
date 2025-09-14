package auth_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/EliasRanz/ai-code-gen/internal/auth"
)

func TestLogoutUseCase_Execute(t *testing.T) {
	tests := []struct {
		name        string
		request     auth.LogoutRequest
		setupMocks  func(*MockSessionRepository)
		expectedErr string
	}{
		{
			name: "successful logout",
			request: auth.LogoutRequest{
				AccessToken: "valid_token",
			},
			setupMocks: func(sessionRepo *MockSessionRepository) {
				session := auth.Session{
					ID:          auth.SessionID("session123"),
					UserID:      auth.UserID("user123"),
					AccessToken: "valid_token",
				}
				sessionRepo.On("GetByAccessToken", mock.Anything, "valid_token").Return(session, nil)
				sessionRepo.On("Delete", mock.Anything, auth.SessionID("session123")).Return(nil)
			},
		},
		{
			name: "empty access token",
			request: auth.LogoutRequest{
				AccessToken: "",
			},
			setupMocks:  func(sessionRepo *MockSessionRepository) {},
			expectedErr: "access token is required",
		},
		{
			name: "session not found - should succeed",
			request: auth.LogoutRequest{
				AccessToken: "non_existent_token",
			},
			setupMocks: func(sessionRepo *MockSessionRepository) {
				sessionRepo.On("GetByAccessToken", mock.Anything, "non_existent_token").Return(auth.Session{}, auth.NewNotFoundError("session not found"))
			},
		},
		{
			name: "database error on get session",
			request: auth.LogoutRequest{
				AccessToken: "valid_token",
			},
			setupMocks: func(sessionRepo *MockSessionRepository) {
				sessionRepo.On("GetByAccessToken", mock.Anything, "valid_token").Return(auth.Session{}, assert.AnError)
			},
			expectedErr: "failed to get session",
		},
		{
			name: "database error on delete session",
			request: auth.LogoutRequest{
				AccessToken: "valid_token",
			},
			setupMocks: func(sessionRepo *MockSessionRepository) {
				session := auth.Session{
					ID:          auth.SessionID("session123"),
					UserID:      auth.UserID("user123"),
					AccessToken: "valid_token",
				}
				sessionRepo.On("GetByAccessToken", mock.Anything, "valid_token").Return(session, nil)
				sessionRepo.On("Delete", mock.Anything, auth.SessionID("session123")).Return(assert.AnError)
			},
			expectedErr: "failed to delete session",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			sessionRepo := &MockSessionRepository{}
			tt.setupMocks(sessionRepo)

			// Create use case
			uc := auth.NewLogoutUseCase(sessionRepo)

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
				assert.True(t, result.Success)
			}

			// Verify mocks
			sessionRepo.AssertExpectations(t)
		})
	}
}
