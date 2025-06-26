// Package auth contains tests for authentication use cases
package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/EliasRanz/ai-code-gen/ai-ui-generator/internal/domain/auth"
	"github.com/EliasRanz/ai-code-gen/ai-ui-generator/internal/domain/common"
	"github.com/EliasRanz/ai-code-gen/ai-ui-generator/internal/domain/user"
)

// Mock implementations for testing
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, u user.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id common.UserID) (user.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(user.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (user.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(user.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, u user.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id common.UserID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, params common.PaginationParams, search string) ([]user.User, error) {
	args := m.Called(ctx, params, search)
	return args.Get(0).([]user.User), args.Error(1)
}

func (m *MockUserRepository) Count(ctx context.Context, search string) (int, error) {
	args := m.Called(ctx, search)
	return args.Int(0), args.Error(1)
}

type MockSessionRepository struct {
	mock.Mock
}

func (m *MockSessionRepository) Create(ctx context.Context, session auth.Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockSessionRepository) GetByRefreshToken(ctx context.Context, refreshToken string) (auth.Session, error) {
	args := m.Called(ctx, refreshToken)
	return args.Get(0).(auth.Session), args.Error(1)
}

func (m *MockSessionRepository) GetByAccessToken(ctx context.Context, accessToken string) (auth.Session, error) {
	args := m.Called(ctx, accessToken)
	return args.Get(0).(auth.Session), args.Error(1)
}

func (m *MockSessionRepository) Update(ctx context.Context, session auth.Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockSessionRepository) Delete(ctx context.Context, sessionID common.SessionID) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func (m *MockSessionRepository) DeleteByUserID(ctx context.Context, userID common.UserID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockSessionRepository) CleanExpired(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

type MockPasswordHasher struct {
	mock.Mock
}

func (m *MockPasswordHasher) Hash(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

func (m *MockPasswordHasher) Verify(password, hash string) bool {
	args := m.Called(password, hash)
	return args.Bool(0)
}

type MockTokenProvider struct {
	mock.Mock
}

func (m *MockTokenProvider) GenerateAccessToken(userID common.UserID) (string, error) {
	args := m.Called(userID)
	return args.String(0), args.Error(1)
}

func (m *MockTokenProvider) GenerateRefreshToken(userID common.UserID) (string, error) {
	args := m.Called(userID)
	return args.String(0), args.Error(1)
}

func (m *MockTokenProvider) ValidateAccessToken(token string) (common.UserID, error) {
	args := m.Called(token)
	return args.Get(0).(common.UserID), args.Error(1)
}

func (m *MockTokenProvider) ValidateRefreshToken(token string) (common.UserID, error) {
	args := m.Called(token)
	return args.Get(0).(common.UserID), args.Error(1)
}

func TestLoginUseCase_Execute(t *testing.T) {
	ctx := context.Background()
	userID := common.UserID("test-user-id")
	email := "test@example.com"
	password := "testpassword123"
	passwordHash := "hashed_password"

	testUser := user.User{
		ID:           userID,
		Email:        email,
		Username:     "testuser",
		Name:         "Test User",
		PasswordHash: passwordHash,
		Active:       true,
		Status:       user.StatusActiveUser,
		Role:         user.RoleUser,
	}

	t.Run("successful login", func(t *testing.T) {
		userRepo := new(MockUserRepository)
		sessionRepo := new(MockSessionRepository)
		passwordHasher := new(MockPasswordHasher)
		tokenProvider := new(MockTokenProvider)

		useCase := NewLoginUseCase(userRepo, sessionRepo, passwordHasher, tokenProvider)

		// Setup expectations
		userRepo.On("GetByEmail", ctx, email).Return(testUser, nil)
		passwordHasher.On("Verify", password, passwordHash).Return(true)
		tokenProvider.On("GenerateAccessToken", userID).Return("access_token", nil)
		tokenProvider.On("GenerateRefreshToken", userID).Return("refresh_token", nil)
		sessionRepo.On("Create", ctx, mock.AnythingOfType("auth.Session")).Return(nil)

		// Execute
		request := LoginRequest{
			Email:    email,
			Password: password,
		}
		response, err := useCase.Execute(ctx, request)

		// Assertions
		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, &testUser, response.User)
		assert.Equal(t, "access_token", response.AccessToken)
		assert.Equal(t, "refresh_token", response.RefreshToken)
		assert.NotNil(t, response.Session)

		// Verify all expectations were met
		userRepo.AssertExpectations(t)
		sessionRepo.AssertExpectations(t)
		passwordHasher.AssertExpectations(t)
		tokenProvider.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		userRepo := new(MockUserRepository)
		sessionRepo := new(MockSessionRepository)
		passwordHasher := new(MockPasswordHasher)
		tokenProvider := new(MockTokenProvider)

		useCase := NewLoginUseCase(userRepo, sessionRepo, passwordHasher, tokenProvider)

		userRepo.On("GetByEmail", ctx, email).Return(user.User{}, common.NewNotFoundError("user not found"))

		request := LoginRequest{
			Email:    email,
			Password: password,
		}
		response, err := useCase.Execute(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "invalid credentials")

		userRepo.AssertExpectations(t)
	})

	t.Run("invalid password", func(t *testing.T) {
		userRepo := new(MockUserRepository)
		sessionRepo := new(MockSessionRepository)
		passwordHasher := new(MockPasswordHasher)
		tokenProvider := new(MockTokenProvider)

		useCase := NewLoginUseCase(userRepo, sessionRepo, passwordHasher, tokenProvider)

		userRepo.On("GetByEmail", ctx, email).Return(testUser, nil)
		passwordHasher.On("Verify", "wrongpassword", passwordHash).Return(false)

		request := LoginRequest{
			Email:    email,
			Password: "wrongpassword",
		}
		response, err := useCase.Execute(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "invalid credentials")

		userRepo.AssertExpectations(t)
		passwordHasher.AssertExpectations(t)
	})

	t.Run("inactive user", func(t *testing.T) {
		userRepo := new(MockUserRepository)
		sessionRepo := new(MockSessionRepository)
		passwordHasher := new(MockPasswordHasher)
		tokenProvider := new(MockTokenProvider)

		useCase := NewLoginUseCase(userRepo, sessionRepo, passwordHasher, tokenProvider)

		inactiveUser := testUser
		inactiveUser.Active = false

		userRepo.On("GetByEmail", ctx, email).Return(inactiveUser, nil)

		request := LoginRequest{
			Email:    email,
			Password: password,
		}
		response, err := useCase.Execute(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "user account is inactive")

		userRepo.AssertExpectations(t)
	})

	t.Run("token generation failure", func(t *testing.T) {
		userRepo := new(MockUserRepository)
		sessionRepo := new(MockSessionRepository)
		passwordHasher := new(MockPasswordHasher)
		tokenProvider := new(MockTokenProvider)

		useCase := NewLoginUseCase(userRepo, sessionRepo, passwordHasher, tokenProvider)

		userRepo.On("GetByEmail", ctx, email).Return(testUser, nil)
		passwordHasher.On("Verify", password, passwordHash).Return(true)
		tokenProvider.On("GenerateAccessToken", userID).Return("", assert.AnError)

		request := LoginRequest{
			Email:    email,
			Password: password,
		}
		response, err := useCase.Execute(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "failed to generate access token")

		userRepo.AssertExpectations(t)
		passwordHasher.AssertExpectations(t)
		tokenProvider.AssertExpectations(t)
	})
}
