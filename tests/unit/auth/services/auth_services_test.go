package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/EliasRanz/ai-code-gen/internal/auth"
)

// Mock repositories and providers
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetByID(ctx context.Context, id auth.UserID) (auth.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(auth.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (auth.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(auth.User), args.Error(1)
}

func (m *MockUserRepository) Create(ctx context.Context, user auth.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Update(ctx context.Context, user auth.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id auth.UserID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockSessionRepository struct {
	mock.Mock
}

func (m *MockSessionRepository) Create(ctx context.Context, session auth.Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockSessionRepository) GetByAccessToken(ctx context.Context, token string) (auth.Session, error) {
	args := m.Called(ctx, token)
	return args.Get(0).(auth.Session), args.Error(1)
}

func (m *MockSessionRepository) GetByRefreshToken(ctx context.Context, token string) (auth.Session, error) {
	args := m.Called(ctx, token)
	return args.Get(0).(auth.Session), args.Error(1)
}

func (m *MockSessionRepository) Update(ctx context.Context, session auth.Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockSessionRepository) Delete(ctx context.Context, sessionID auth.SessionID) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func (m *MockSessionRepository) DeleteByUserID(ctx context.Context, userID auth.UserID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockSessionRepository) CleanExpired(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

type MockTokenProvider struct {
	mock.Mock
}

func (m *MockTokenProvider) GenerateAccessToken(userID auth.UserID) (string, error) {
	args := m.Called(userID)
	return args.String(0), args.Error(1)
}

func (m *MockTokenProvider) GenerateRefreshToken(userID auth.UserID) (string, error) {
	args := m.Called(userID)
	return args.String(0), args.Error(1)
}

func (m *MockTokenProvider) ValidateAccessToken(token string) (auth.UserID, error) {
	args := m.Called(token)
	return args.Get(0).(auth.UserID), args.Error(1)
}

func (m *MockTokenProvider) ValidateRefreshToken(token string) (auth.UserID, error) {
	args := m.Called(token)
	return args.Get(0).(auth.UserID), args.Error(1)
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

// Test Login Use Case
func TestLoginUseCase_Execute(t *testing.T) {
	tests := []struct {
		name         string
		request      auth.LoginRequest
		setupMocks   func(*MockUserRepository, *MockSessionRepository, *MockTokenProvider, *MockPasswordHasher)
		expectError  bool
		expectedUser auth.UserID
	}{
		{
			name: "successful login with valid credentials",
			request: auth.LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMocks: func(userRepo *MockUserRepository, sessionRepo *MockSessionRepository, tokenProvider *MockTokenProvider, hasher *MockPasswordHasher) {
				user := auth.User{
					ID:       auth.UserID("user123"),
					Email:    "test@example.com",
					Password: "hashedPassword",
					Active:   true,
					Roles:    []string{"user"},
				}

				userRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(user, nil)
				hasher.On("Verify", "password123", "hashedPassword").Return(true)
				tokenProvider.On("GenerateAccessToken", auth.UserID("user123")).Return("access_token_123", nil)
				tokenProvider.On("GenerateRefreshToken", auth.UserID("user123")).Return("refresh_token_123", nil)
				sessionRepo.On("Create", mock.Anything, mock.AnythingOfType("auth.Session")).Return(nil)
			},
			expectError:  false,
			expectedUser: auth.UserID("user123"),
		},
		{
			name: "login failure with non-existent user",
			request: auth.LoginRequest{
				Email:    "nonexistent@example.com",
				Password: "password123",
			},
			setupMocks: func(userRepo *MockUserRepository, sessionRepo *MockSessionRepository, tokenProvider *MockTokenProvider, hasher *MockPasswordHasher) {
				userRepo.On("GetByEmail", mock.Anything, "nonexistent@example.com").Return(auth.User{}, auth.NewNotFoundError("user not found"))
			},
			expectError: true,
		},
		{
			name: "login failure with inactive user",
			request: auth.LoginRequest{
				Email:    "inactive@example.com",
				Password: "password123",
			},
			setupMocks: func(userRepo *MockUserRepository, sessionRepo *MockSessionRepository, tokenProvider *MockTokenProvider, hasher *MockPasswordHasher) {
				user := auth.User{
					ID:       auth.UserID("user456"),
					Email:    "inactive@example.com",
					Password: "hashedPassword",
					Active:   false,
					Roles:    []string{"user"},
				}

				userRepo.On("GetByEmail", mock.Anything, "inactive@example.com").Return(user, nil)
			},
			expectError: true,
		},
		{
			name: "login failure with invalid password",
			request: auth.LoginRequest{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			setupMocks: func(userRepo *MockUserRepository, sessionRepo *MockSessionRepository, tokenProvider *MockTokenProvider, hasher *MockPasswordHasher) {
				user := auth.User{
					ID:       auth.UserID("user123"),
					Email:    "test@example.com",
					Password: "hashedPassword",
					Active:   true,
					Roles:    []string{"user"},
				}

				userRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(user, nil)
				hasher.On("Verify", "wrongpassword", "hashedPassword").Return(false)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			userRepo := &MockUserRepository{}
			sessionRepo := &MockSessionRepository{}
			tokenProvider := &MockTokenProvider{}
			hasher := &MockPasswordHasher{}

			tt.setupMocks(userRepo, sessionRepo, tokenProvider, hasher)

			// Create use case
			loginUC := auth.NewLoginUseCase(userRepo, sessionRepo, hasher, tokenProvider)

			// Execute
			response, err := loginUC.Execute(context.Background(), tt.request)

			// Assert results
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.Equal(t, tt.expectedUser, response.User.ID)
				assert.NotEmpty(t, response.AccessToken)
				assert.NotEmpty(t, response.RefreshToken)
			}

			// Verify mock expectations
			userRepo.AssertExpectations(t)
			sessionRepo.AssertExpectations(t)
			tokenProvider.AssertExpectations(t)
			hasher.AssertExpectations(t)
		})
	}
}

// Test Validate Token Use Case
func TestValidateTokenService_Execute(t *testing.T) {
	tests := []struct {
		name         string
		request      auth.ValidateTokenRequest
		setupMocks   func(*MockUserRepository, *MockTokenProvider)
		expectError  bool
		expectValid  bool
		expectedUser auth.UserID
	}{
		{
			name: "successful token validation",
			request: auth.ValidateTokenRequest{
				AccessToken: "valid_token_123",
			},
			setupMocks: func(userRepo *MockUserRepository, tokenProvider *MockTokenProvider) {
				user := auth.User{
					ID:     auth.UserID("user123"),
					Email:  "test@example.com",
					Active: true,
					Roles:  []string{"user"},
				}

				tokenProvider.On("ValidateAccessToken", "valid_token_123").Return(auth.UserID("user123"), nil)
				userRepo.On("GetByID", mock.Anything, auth.UserID("user123")).Return(user, nil)
			},
			expectError:  false,
			expectValid:  true,
			expectedUser: auth.UserID("user123"),
		},
		{
			name: "token validation failure with invalid token",
			request: auth.ValidateTokenRequest{
				AccessToken: "invalid_token",
			},
			setupMocks: func(userRepo *MockUserRepository, tokenProvider *MockTokenProvider) {
				tokenProvider.On("ValidateAccessToken", "invalid_token").Return(auth.UserID(""), assert.AnError)
			},
			expectError: false,
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			userRepo := &MockUserRepository{}
			tokenProvider := &MockTokenProvider{}

			tt.setupMocks(userRepo, tokenProvider)

			// Create use case
			validateUC := auth.NewValidateToken(tokenProvider, userRepo)

			// Execute
			response, err := validateUC.Execute(context.Background(), tt.request)

			// Assert results
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.Equal(t, tt.expectValid, response.Valid)

				if tt.expectValid && tt.expectedUser != "" {
					assert.NotNil(t, response.UserContext)
					assert.Equal(t, tt.expectedUser, response.UserContext.UserID)
				}
			}

			// Verify mock expectations
			userRepo.AssertExpectations(t)
			tokenProvider.AssertExpectations(t)
		})
	}
}

// Test Logout Use Case
func TestLogoutUseCase_Execute(t *testing.T) {
	tests := []struct {
		name        string
		request     auth.LogoutRequest
		setupMocks  func(*MockSessionRepository)
		expectError bool
	}{
		{
			name: "successful logout",
			request: auth.LogoutRequest{
				AccessToken: "valid_access_token",
			},
			setupMocks: func(sessionRepo *MockSessionRepository) {
				session := auth.Session{
					ID:           auth.SessionID("session123"),
					UserID:       auth.UserID("user123"),
					AccessToken:  "valid_access_token",
					RefreshToken: "refresh_token",
					ExpiresAt:    time.Now().Add(time.Hour),
					Status:       auth.StatusActive,
				}

				sessionRepo.On("GetByAccessToken", mock.Anything, "valid_access_token").Return(session, nil)
				sessionRepo.On("Delete", mock.Anything, auth.SessionID("session123")).Return(nil)
			},
			expectError: false,
		},
		{
			name: "logout with non-existent session should succeed",
			request: auth.LogoutRequest{
				AccessToken: "non_existent_token",
			},
			setupMocks: func(sessionRepo *MockSessionRepository) {
				sessionRepo.On("GetByAccessToken", mock.Anything, "non_existent_token").Return(auth.Session{}, auth.NewNotFoundError("session not found"))
			},
			expectError: false,
		},
		{
			name: "logout with empty access token",
			request: auth.LogoutRequest{
				AccessToken: "",
			},
			setupMocks: func(sessionRepo *MockSessionRepository) {
				// No mocks needed
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			sessionRepo := &MockSessionRepository{}
			tt.setupMocks(sessionRepo)

			// Create use case
			logoutUC := auth.NewLogoutUseCase(sessionRepo)

			// Execute
			response, err := logoutUC.Execute(context.Background(), tt.request)

			// Assert results
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.True(t, response.Success)
			}

			// Verify mock expectations
			sessionRepo.AssertExpectations(t)
		})
	}
}

// Test Check Role Service
func TestCheckRoleService_Execute(t *testing.T) {
	tests := []struct {
		name         string
		request      auth.CheckRoleRequest
		setupMocks   func(*MockUserRepository)
		expectError  bool
		expectAccess bool
	}{
		{
			name: "successful role check - exact match",
			request: auth.CheckRoleRequest{
				UserID:       auth.UserID("user123"),
				RequiredRole: "user",
			},
			setupMocks: func(userRepo *MockUserRepository) {
				user := auth.User{
					ID:     auth.UserID("user123"),
					Email:  "test@example.com",
					Active: true,
					Roles:  []string{"user"},
				}

				userRepo.On("GetByID", mock.Anything, auth.UserID("user123")).Return(user, nil)
			},
			expectError:  false,
			expectAccess: true,
		},
		{
			name: "successful role check - admin has all roles",
			request: auth.CheckRoleRequest{
				UserID:       auth.UserID("admin123"),
				RequiredRole: "user",
			},
			setupMocks: func(userRepo *MockUserRepository) {
				user := auth.User{
					ID:     auth.UserID("admin123"),
					Email:  "admin@example.com",
					Active: true,
					Roles:  []string{"admin"},
				}

				userRepo.On("GetByID", mock.Anything, auth.UserID("admin123")).Return(user, nil)
			},
			expectError:  false,
			expectAccess: true,
		},
		{
			name: "role check denial - insufficient privileges",
			request: auth.CheckRoleRequest{
				UserID:       auth.UserID("user456"),
				RequiredRole: "admin",
			},
			setupMocks: func(userRepo *MockUserRepository) {
				user := auth.User{
					ID:     auth.UserID("user456"),
					Email:  "user@example.com",
					Active: true,
					Roles:  []string{"user"},
				}

				userRepo.On("GetByID", mock.Anything, auth.UserID("user456")).Return(user, nil)
			},
			expectError:  false,
			expectAccess: false,
		},
		{
			name: "role check failure - user not found",
			request: auth.CheckRoleRequest{
				UserID:       auth.UserID("nonexistent"),
				RequiredRole: "user",
			},
			setupMocks: func(userRepo *MockUserRepository) {
				userRepo.On("GetByID", mock.Anything, auth.UserID("nonexistent")).Return(auth.User{}, auth.NewNotFoundError("user not found"))
			},
			expectError:  false, // CheckRoleService returns a response with Authorized: false, not an error
			expectAccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			userRepo := &MockUserRepository{}
			tt.setupMocks(userRepo)

			// Create use case
			checkRoleUC := auth.NewCheckRole(userRepo)

			// Execute
			response, err := checkRoleUC.Execute(context.Background(), tt.request)

			// Assert results
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.Equal(t, tt.expectAccess, response.Authorized)
			}

			// Verify mock expectations
			userRepo.AssertExpectations(t)
		})
	}
}

// Test Get Session Service
func TestGetSessionService_Execute(t *testing.T) {
	tests := []struct {
		name        string
		request     auth.GetSessionRequest
		setupMocks  func(*MockSessionRepository, *MockTokenProvider)
		expectError bool
	}{
		{
			name: "successful session retrieval",
			request: auth.GetSessionRequest{
				AccessToken: "valid_access_token",
			},
			setupMocks: func(sessionRepo *MockSessionRepository, tokenProvider *MockTokenProvider) {
				session := auth.Session{
					ID:           auth.SessionID("session123"),
					UserID:       auth.UserID("user123"),
					AccessToken:  "valid_access_token",
					RefreshToken: "refresh_token",
					ExpiresAt:    time.Now().Add(time.Hour),
					Status:       auth.StatusActive,
					CreatedAt:    time.Now().Add(-time.Hour),
				}

				tokenProvider.On("ValidateAccessToken", "valid_access_token").Return(auth.UserID("user123"), nil)
				sessionRepo.On("GetByAccessToken", mock.Anything, "valid_access_token").Return(session, nil)
			},
			expectError: false,
		},
		{
			name: "session retrieval failure - session not found",
			request: auth.GetSessionRequest{
				AccessToken: "invalid_access_token",
			},
			setupMocks: func(sessionRepo *MockSessionRepository, tokenProvider *MockTokenProvider) {
				tokenProvider.On("ValidateAccessToken", "invalid_access_token").Return(auth.UserID(""), assert.AnError)
			},
			expectError: true,
		},
		{
			name: "session retrieval failure - empty access token",
			request: auth.GetSessionRequest{
				AccessToken: "",
			},
			setupMocks: func(sessionRepo *MockSessionRepository, tokenProvider *MockTokenProvider) {
				// No mocks needed
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			sessionRepo := &MockSessionRepository{}
			tokenProvider := &MockTokenProvider{}
			tt.setupMocks(sessionRepo, tokenProvider)

			// Create use case
			getSessionUC := auth.NewGetSession(sessionRepo, tokenProvider)

			// Execute
			response, err := getSessionUC.Execute(context.Background(), tt.request)

			// Assert results
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.Equal(t, auth.UserID("user123"), response.UserID)
			}

			// Verify mock expectations
			sessionRepo.AssertExpectations(t)
			tokenProvider.AssertExpectations(t)
		})
	}
}
