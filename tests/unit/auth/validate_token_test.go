package authtest

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	appAuth "github.com/EliasRanz/ai-code-gen/internal/auth"
)

// TestValidateTokenUseCase_Execute tests the token validation use case
func TestValidateTokenUseCase_Execute(t *testing.T) {
	ctx := context.Background()
	userID := appAuth.UserID("test-user-id")
	validToken := "valid-jwt-token"

	testUser := appAuth.User{
		ID:       userID,
		Email:    "test@example.com",
		Username: "testuser",
		Name:     "Test User",
		Active:   true,
		Roles:    []string{"user", "admin"},
	}

	t.Run("successful token validation", func(t *testing.T) {
		// Setup mocks
		tokenProvider := new(MockTokenProvider)
		userRepo := new(MockUserRepository)

		useCase := appAuth.NewValidateToken(tokenProvider, userRepo)

		// Setup expectations
		tokenProvider.On("ValidateAccessToken", validToken).Return(userID, nil)
		userRepo.On("GetByID", ctx, userID).Return(testUser, nil)

		req := appAuth.ValidateTokenRequest{
			AccessToken: validToken,
		}

		// Execute use case
		resp, err := useCase.Execute(ctx, req)

		// Assertions
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.True(t, resp.Valid)
		assert.Empty(t, resp.Error)

		// Verify user context
		assert.NotNil(t, resp.UserContext)
		assert.Equal(t, userID, resp.UserContext.UserID)
		assert.Equal(t, "test@example.com", resp.UserContext.Email)
		assert.Equal(t, "testuser", resp.UserContext.Username)
		assert.Equal(t, "Test User", resp.UserContext.Name)
		assert.True(t, resp.UserContext.Active)
		assert.Equal(t, "user", resp.UserContext.Role) // First role becomes primary
		assert.Equal(t, []string{"user", "admin"}, resp.UserContext.Roles)
		assert.Empty(t, resp.UserContext.Permissions) // Empty for now

		// Verify mocks were called
		tokenProvider.AssertExpectations(t)
		userRepo.AssertExpectations(t)
	})

	t.Run("empty token should return invalid", func(t *testing.T) {
		tokenProvider := new(MockTokenProvider)
		userRepo := new(MockUserRepository)

		useCase := appAuth.NewValidateToken(tokenProvider, userRepo)

		req := appAuth.ValidateTokenRequest{
			AccessToken: "",
		}

		resp, err := useCase.Execute(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.False(t, resp.Valid)
		assert.Equal(t, "access token is required", resp.Error)
		assert.Nil(t, resp.UserContext)
	})

	t.Run("invalid token should return invalid", func(t *testing.T) {
		tokenProvider := new(MockTokenProvider)
		userRepo := new(MockUserRepository)

		useCase := appAuth.NewValidateToken(tokenProvider, userRepo)

		tokenProvider.On("ValidateAccessToken", "invalid-token").Return(appAuth.UserID(""), assert.AnError)

		req := appAuth.ValidateTokenRequest{
			AccessToken: "invalid-token",
		}

		resp, err := useCase.Execute(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.False(t, resp.Valid)
		assert.Equal(t, "invalid or expired token", resp.Error)
		assert.Nil(t, resp.UserContext)

		tokenProvider.AssertExpectations(t)
	})

	t.Run("user not found should return invalid", func(t *testing.T) {
		tokenProvider := new(MockTokenProvider)
		userRepo := new(MockUserRepository)

		useCase := appAuth.NewValidateToken(tokenProvider, userRepo)

		tokenProvider.On("ValidateAccessToken", validToken).Return(userID, nil)
		userRepo.On("GetByID", ctx, userID).Return(appAuth.User{}, appAuth.ErrNotFound)

		req := appAuth.ValidateTokenRequest{
			AccessToken: validToken,
		}

		resp, err := useCase.Execute(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.False(t, resp.Valid)
		assert.Equal(t, "user not found", resp.Error)
		assert.Nil(t, resp.UserContext)

		tokenProvider.AssertExpectations(t)
		userRepo.AssertExpectations(t)
	})

	t.Run("inactive user should return invalid", func(t *testing.T) {
		tokenProvider := new(MockTokenProvider)
		userRepo := new(MockUserRepository)

		useCase := appAuth.NewValidateToken(tokenProvider, userRepo)

		inactiveUser := testUser
		inactiveUser.Active = false

		tokenProvider.On("ValidateAccessToken", validToken).Return(userID, nil)
		userRepo.On("GetByID", ctx, userID).Return(inactiveUser, nil)

		req := appAuth.ValidateTokenRequest{
			AccessToken: validToken,
		}

		resp, err := useCase.Execute(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.False(t, resp.Valid)
		assert.Equal(t, "user account is inactive", resp.Error)
		assert.Nil(t, resp.UserContext)

		tokenProvider.AssertExpectations(t)
		userRepo.AssertExpectations(t)
	})

	t.Run("user with no roles should default to user role", func(t *testing.T) {
		tokenProvider := new(MockTokenProvider)
		userRepo := new(MockUserRepository)

		useCase := appAuth.NewValidateToken(tokenProvider, userRepo)

		userWithoutRoles := testUser
		userWithoutRoles.Roles = []string{}

		tokenProvider.On("ValidateAccessToken", validToken).Return(userID, nil)
		userRepo.On("GetByID", ctx, userID).Return(userWithoutRoles, nil)

		req := appAuth.ValidateTokenRequest{
			AccessToken: validToken,
		}

		resp, err := useCase.Execute(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.True(t, resp.Valid)
		assert.Equal(t, "user", resp.UserContext.Role) // Default role
		assert.Empty(t, resp.UserContext.Roles)

		tokenProvider.AssertExpectations(t)
		userRepo.AssertExpectations(t)
	})
}
