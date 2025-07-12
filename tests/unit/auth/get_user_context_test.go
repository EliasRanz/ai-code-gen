package authtest

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	appAuth "github.com/EliasRanz/ai-code-gen/internal/auth"
)

func TestGetUserContextUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	t.Run("successful user context retrieval", func(t *testing.T) {
		// Setup
		userRepo := new(MockUserRepository)
		useCase := appAuth.NewGetUserContextUseCase(userRepo)

		userData := appAuth.User{
			ID:       appAuth.UserID("test-user-id"),
			Email:    "test@example.com",
			Username: "testuser",
			Name:     "Test User",
			Active:   true,
			Roles:    []string{"user"},
		}

		userRepo.On("GetByID", ctx, appAuth.UserID("test-user-id")).Return(userData, nil)

		req := appAuth.GetUserContextRequest{
			UserID: appAuth.UserID("test-user-id"),
		}

		// Execute
		result, err := useCase.Execute(ctx, req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.UserContext)
		assert.Equal(t, "test@example.com", result.UserContext.Email)
		assert.Equal(t, "testuser", result.UserContext.Username)
		assert.Equal(t, "Test User", result.UserContext.Name)
		assert.True(t, result.UserContext.Active)
		assert.Equal(t, []string{"user"}, result.UserContext.Roles)

		userRepo.AssertExpectations(t)
	})

	t.Run("should handle empty user ID", func(t *testing.T) {
		// Setup
		userRepo := new(MockUserRepository)
		useCase := appAuth.NewGetUserContextUseCase(userRepo)

		// Empty user ID will be passed to repository, which should return an error
		userRepo.On("GetByID", ctx, appAuth.UserID("")).Return(appAuth.User{}, errors.New("user not found"))

		req := appAuth.GetUserContextRequest{
			UserID: "",
		}

		// Execute
		result, err := useCase.Execute(ctx, req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "user not found")

		userRepo.AssertExpectations(t)
	})

	t.Run("should handle user not found", func(t *testing.T) {
		// Setup
		userRepo := new(MockUserRepository)
		useCase := appAuth.NewGetUserContextUseCase(userRepo)

		userRepo.On("GetByID", ctx, appAuth.UserID("nonexistent")).Return(appAuth.User{}, errors.New("user not found"))

		req := appAuth.GetUserContextRequest{
			UserID: appAuth.UserID("nonexistent"),
		}

		// Execute
		result, err := useCase.Execute(ctx, req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "user not found")

		userRepo.AssertExpectations(t)
	})
}
