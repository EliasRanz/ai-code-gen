package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	appAuth "github.com/EliasRanz/ai-code-gen/internal/auth"
)

func TestGetSessionUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	t.Run("successful session retrieval", func(t *testing.T) {
		// Setup
		sessionRepo := new(MockSessionRepository)
		tokenProvider := new(MockTokenProvider)
		useCase := appAuth.NewGetSession(sessionRepo, tokenProvider)

		// Mock data
		userID := appAuth.UserID("user_456")
		sessionID := appAuth.SessionID("session_123")
		accessToken := "valid_token_123"

		tokenProvider.On("ValidateAccessToken", accessToken).Return(userID, nil)

		session := appAuth.Session{
			ID:          sessionID,
			UserID:      userID,
			AccessToken: accessToken,
			Status:      appAuth.StatusActive,
			ExpiresAt:   time.Now().Add(time.Hour),
			CreatedAt:   time.Now().Add(-time.Hour),
		}
		sessionRepo.On("GetByAccessToken", ctx, accessToken).Return(session, nil)

		// Execute
		req := appAuth.GetSessionRequest{
			AccessToken: accessToken,
		}
		resp, err := useCase.Execute(ctx, req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, sessionID, resp.SessionID)
		assert.Equal(t, userID, resp.UserID)
		assert.Equal(t, "active", resp.Status)

		// Verify mocks
		tokenProvider.AssertExpectations(t)
		sessionRepo.AssertExpectations(t)
	})

	t.Run("should handle empty access token", func(t *testing.T) {
		// Setup
		sessionRepo := new(MockSessionRepository)
		tokenProvider := new(MockTokenProvider)
		useCase := appAuth.NewGetSession(sessionRepo, tokenProvider)

		// Execute
		req := appAuth.GetSessionRequest{
			AccessToken: "",
		}
		resp, err := useCase.Execute(ctx, req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "access token is required")

		// Verify no mock calls were made
		tokenProvider.AssertNotCalled(t, "ValidateAccessToken")
		sessionRepo.AssertNotCalled(t, "GetByAccessToken")
	})

	t.Run("should handle invalid token", func(t *testing.T) {
		// Setup
		sessionRepo := new(MockSessionRepository)
		tokenProvider := new(MockTokenProvider)
		useCase := appAuth.NewGetSession(sessionRepo, tokenProvider)

		accessToken := "invalid_token"
		tokenProvider.On("ValidateAccessToken", accessToken).Return(appAuth.UserID(""), errors.New("invalid token"))

		// Execute
		req := appAuth.GetSessionRequest{
			AccessToken: accessToken,
		}
		resp, err := useCase.Execute(ctx, req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, resp)

		// Verify mocks
		tokenProvider.AssertExpectations(t)
		sessionRepo.AssertNotCalled(t, "GetByAccessToken")
	})

	t.Run("should handle session not found", func(t *testing.T) {
		// Setup
		sessionRepo := new(MockSessionRepository)
		tokenProvider := new(MockTokenProvider)
		useCase := appAuth.NewGetSession(sessionRepo, tokenProvider)

		userID := appAuth.UserID("user_123")
		accessToken := "valid_but_missing_session"

		tokenProvider.On("ValidateAccessToken", accessToken).Return(userID, nil)
		sessionRepo.On("GetByAccessToken", ctx, accessToken).Return(appAuth.Session{}, appAuth.NewNotFoundError("session not found"))

		// Execute
		req := appAuth.GetSessionRequest{
			AccessToken: accessToken,
		}
		resp, err := useCase.Execute(ctx, req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, resp)

		// Verify mocks
		tokenProvider.AssertExpectations(t)
		sessionRepo.AssertExpectations(t)
	})

	t.Run("should handle expired session", func(t *testing.T) {
		// Setup
		sessionRepo := new(MockSessionRepository)
		tokenProvider := new(MockTokenProvider)
		useCase := appAuth.NewGetSession(sessionRepo, tokenProvider)

		userID := appAuth.UserID("user_789")
		sessionID := appAuth.SessionID("expired_session")
		accessToken := "expired_token"

		tokenProvider.On("ValidateAccessToken", accessToken).Return(userID, nil)

		expiredSession := appAuth.Session{
			ID:          sessionID,
			UserID:      userID,
			AccessToken: accessToken,
			Status:      appAuth.StatusExpired,
			ExpiresAt:   time.Now().Add(-time.Hour), // Expired
			CreatedAt:   time.Now().Add(-2 * time.Hour),
		}
		sessionRepo.On("GetByAccessToken", ctx, accessToken).Return(expiredSession, nil)

		// Execute
		req := appAuth.GetSessionRequest{
			AccessToken: accessToken,
		}
		resp, err := useCase.Execute(ctx, req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, sessionID, resp.SessionID)
		assert.Equal(t, userID, resp.UserID)
		assert.Equal(t, "expired", resp.Status)

		// Verify mocks
		tokenProvider.AssertExpectations(t)
		sessionRepo.AssertExpectations(t)
	})

	t.Run("should handle database error", func(t *testing.T) {
		// Setup
		sessionRepo := new(MockSessionRepository)
		tokenProvider := new(MockTokenProvider)
		useCase := appAuth.NewGetSession(sessionRepo, tokenProvider)

		userID := appAuth.UserID("user_123")
		accessToken := "repo_error_token"

		tokenProvider.On("ValidateAccessToken", accessToken).Return(userID, nil)
		sessionRepo.On("GetByAccessToken", ctx, accessToken).Return(appAuth.Session{}, errors.New("database error"))

		// Execute
		req := appAuth.GetSessionRequest{
			AccessToken: accessToken,
		}
		resp, err := useCase.Execute(ctx, req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, resp)

		// Verify mocks
		tokenProvider.AssertExpectations(t)
		sessionRepo.AssertExpectations(t)
	})
}
