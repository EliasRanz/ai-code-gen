package auth_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	appAuth "github.com/EliasRanz/ai-code-gen/internal/auth"
)

func TestCheckRoleUseCase_Execute(t *testing.T) {
	t.Run("successful role authorization for exact match", func(t *testing.T) {
		// Setup
		userRepo := new(MockUserRepository)
		useCase := appAuth.NewCheckRole(userRepo)

		userData := appAuth.User{
			ID:       appAuth.UserID("test-user-id"),
			Email:    "test@example.com",
			Username: "testuser",
			Name:     "Test User",
			Roles:    []string{"editor", "viewer"},
			Active:   true,
		}

		userRepo.On("GetByID", mock.Anything, appAuth.UserID("test-user-id")).Return(userData, nil)

		// Execute
		req := appAuth.CheckRoleRequest{
			UserID:       appAuth.UserID("test-user-id"),
			RequiredRole: "editor",
		}
		resp, err := useCase.Execute(context.Background(), req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.True(t, resp.Authorized)
		assert.Equal(t, []string{"editor", "viewer"}, resp.UserRoles)
		assert.Empty(t, resp.Reason)

		userRepo.AssertExpectations(t)
	})

	t.Run("successful role authorization for admin user", func(t *testing.T) {
		// Setup
		userRepo := new(MockUserRepository)
		useCase := appAuth.NewCheckRole(userRepo)

		userData := appAuth.User{
			ID:       appAuth.UserID("admin-user-id"),
			Email:    "admin@example.com",
			Username: "admin",
			Name:     "Admin User",
			Roles:    []string{"admin"},
			Active:   true,
		}

		userRepo.On("GetByID", mock.Anything, appAuth.UserID("admin-user-id")).Return(userData, nil)

		// Execute
		req := appAuth.CheckRoleRequest{
			UserID:       appAuth.UserID("admin-user-id"),
			RequiredRole: "any_role",
		}
		resp, err := useCase.Execute(context.Background(), req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.True(t, resp.Authorized)
		assert.Equal(t, []string{"admin"}, resp.UserRoles)
		assert.Empty(t, resp.Reason)

		userRepo.AssertExpectations(t)
	})

	t.Run("denied access for missing role", func(t *testing.T) {
		// Setup
		userRepo := new(MockUserRepository)
		useCase := appAuth.NewCheckRole(userRepo)

		userData := appAuth.User{
			ID:       appAuth.UserID("test-user-id"),
			Email:    "test@example.com",
			Username: "testuser",
			Name:     "Test User",
			Roles:    []string{"viewer"},
			Active:   true,
		}

		userRepo.On("GetByID", mock.Anything, appAuth.UserID("test-user-id")).Return(userData, nil)

		// Execute
		req := appAuth.CheckRoleRequest{
			UserID:       appAuth.UserID("test-user-id"),
			RequiredRole: "admin",
		}
		resp, err := useCase.Execute(context.Background(), req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.False(t, resp.Authorized)
		assert.Equal(t, []string{"viewer"}, resp.UserRoles)
		assert.Equal(t, "user does not have required role: admin", resp.Reason)

		userRepo.AssertExpectations(t)
	})

	t.Run("should handle empty user ID", func(t *testing.T) {
		// Setup
		userRepo := new(MockUserRepository)
		useCase := appAuth.NewCheckRole(userRepo)

		// Execute
		req := appAuth.CheckRoleRequest{
			UserID:       appAuth.UserID(""),
			RequiredRole: "admin",
		}
		resp, err := useCase.Execute(context.Background(), req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.False(t, resp.Authorized)
		assert.Empty(t, resp.UserRoles)
		assert.Equal(t, "user ID is required", resp.Reason)

		// Verify no repository calls were made
		userRepo.AssertNotCalled(t, "GetByID")
	})

	t.Run("should handle user not found", func(t *testing.T) {
		// Setup
		userRepo := new(MockUserRepository)
		useCase := appAuth.NewCheckRole(userRepo)

		notFoundErr := appAuth.NewNotFoundError("user not found")
		userRepo.On("GetByID", mock.Anything, appAuth.UserID("test-user-id")).Return(appAuth.User{}, notFoundErr)

		// Execute
		req := appAuth.CheckRoleRequest{
			UserID:       appAuth.UserID("test-user-id"),
			RequiredRole: "admin",
		}
		resp, err := useCase.Execute(context.Background(), req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.False(t, resp.Authorized)
		assert.Empty(t, resp.UserRoles)
		assert.Equal(t, "user not found", resp.Reason)

		userRepo.AssertExpectations(t)
	})

	t.Run("should handle inactive user", func(t *testing.T) {
		// Setup
		userRepo := new(MockUserRepository)
		useCase := appAuth.NewCheckRole(userRepo)

		userData := appAuth.User{
			ID:       appAuth.UserID("test-user-id"),
			Email:    "test@example.com",
			Username: "testuser",
			Name:     "Test User",
			Roles:    []string{"admin"},
			Active:   false, // Inactive user
		}

		userRepo.On("GetByID", mock.Anything, appAuth.UserID("test-user-id")).Return(userData, nil)

		// Execute
		req := appAuth.CheckRoleRequest{
			UserID:       appAuth.UserID("test-user-id"),
			RequiredRole: "admin",
		}
		resp, err := useCase.Execute(context.Background(), req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.False(t, resp.Authorized)
		assert.Equal(t, []string{"admin"}, resp.UserRoles)
		assert.Equal(t, "user account is inactive", resp.Reason)

		userRepo.AssertExpectations(t)
	})

	t.Run("should handle database error", func(t *testing.T) {
		// Setup
		userRepo := new(MockUserRepository)
		useCase := appAuth.NewCheckRole(userRepo)

		dbErr := errors.New("database connection error")
		userRepo.On("GetByID", mock.Anything, appAuth.UserID("test-user-id")).Return(appAuth.User{}, dbErr)

		// Execute
		req := appAuth.CheckRoleRequest{
			UserID:       appAuth.UserID("test-user-id"),
			RequiredRole: "admin",
		}
		resp, err := useCase.Execute(context.Background(), req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.False(t, resp.Authorized)
		assert.Empty(t, resp.UserRoles)
		assert.Equal(t, "failed to retrieve user information", resp.Reason)

		userRepo.AssertExpectations(t)
	})

	t.Run("should handle empty required role", func(t *testing.T) {
		// Setup
		userRepo := new(MockUserRepository)
		useCase := appAuth.NewCheckRole(userRepo)

		userData := appAuth.User{
			ID:       appAuth.UserID("super-admin-id"),
			Email:    "superadmin@example.com",
			Username: "superadmin",
			Name:     "Super Admin",
			Roles:    []string{"super_admin"},
			Active:   true,
		}

		userRepo.On("GetByID", mock.Anything, appAuth.UserID("super-admin-id")).Return(userData, nil)

		// Execute
		req := appAuth.CheckRoleRequest{
			UserID:       appAuth.UserID("super-admin-id"),
			RequiredRole: "",
		}
		resp, err := useCase.Execute(context.Background(), req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.True(t, resp.Authorized) // Empty role should allow access for valid user
		assert.Equal(t, []string{"super_admin"}, resp.UserRoles)
		assert.Empty(t, resp.Reason)

		userRepo.AssertExpectations(t)
	})
}
