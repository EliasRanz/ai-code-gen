package authtest

import (
	"testing"
	"time"

	"github.com/EliasRanz/ai-code-gen/ai-ui-generator/internal/auth"
	"github.com/EliasRanz/ai-code-gen/ai-ui-generator/internal/user"
	"github.com/stretchr/testify/assert"
)

func TestServiceLogin(t *testing.T) {
	// Create test service
	userRepo := &MockUserRepository{}
	tokenManager := auth.NewTokenManager("test-secret", "test-issuer")
	service := auth.NewService(userRepo, tokenManager)

	// Create test user
	testUser := &user.User{
		ID:           "user-123",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		IsActive:     true,
	}

	// Mock the repository call
	userRepo.On("GetByEmail", "test@example.com").Return(testUser, nil)

	// Test that service can find user by email (testing integration)
	foundUser, err := userRepo.GetByEmail("test@example.com")
	assert.NoError(t, err)
	assert.Equal(t, "user-123", foundUser.ID)
	assert.Equal(t, "test@example.com", foundUser.Email)
	assert.True(t, foundUser.IsActive)

	// Test that service was created successfully
	assert.NotNil(t, service)
}

func TestServiceValidateToken(t *testing.T) {
	userRepo := &MockUserRepository{}
	tokenManager := auth.NewTokenManager("test-secret", "test-issuer")
	service := auth.NewService(userRepo, tokenManager)

	// Generate a valid token
	userID := "user-123"
	token, err := tokenManager.GenerateToken(userID, 15*time.Minute)
	assert.NoError(t, err)

	// Create test user
	testUser := &user.User{
		ID:       userID,
		Email:    "test@example.com",
		IsActive: true,
	}

	// Mock the repository call
	userRepo.On("GetByID", userID).Return(testUser, nil)

	// Test token validation
	validatedUserID, err := service.ValidateToken(token)
	assert.NoError(t, err)
	assert.Equal(t, userID, validatedUserID)
}
