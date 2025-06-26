package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/EliasRanz/ai-code-gen/ai-ui-generator/internal/user"
	"golang.org/x/crypto/bcrypt"
)

// MockUserRepository for testing
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetByID(id string) (*user.User, error) {
	args := m.Called(id)
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(email string) (*user.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserRepository) Create(u *user.User) error {
	args := m.Called(u)
	return args.Error(0)
}

func (m *MockUserRepository) Update(id string, updates map[string]interface{}) (*user.User, error) {
	args := m.Called(id, updates)
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepository) List(limit, offset int) ([]*user.User, error) {
	args := m.Called(limit, offset)
	return args.Get(0).([]*user.User), args.Error(1)
}

// MockPasswordHasher for testing
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

func TestService_Login(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		password    string
		setupMocks  func(*MockUserRepository, *MockPasswordHasher)
		wantErr     bool
		expectedErr string
	}{
		{
			name:     "successful login",
			email:    "user@example.com",
			password: "password123",
			setupMocks: func(userRepo *MockUserRepository, hasher *MockPasswordHasher) {
				hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
				testUser := &user.User{
					ID:           "user-123",
					Email:        "user@example.com",
					PasswordHash: string(hash),
					IsActive:     true,
				}
				userRepo.On("GetByEmail", "user@example.com").Return(testUser, nil)
				hasher.On("Verify", "password123", string(hash)).Return(true)
			},
			wantErr: false,
		},
		{
			name:     "empty email",
			email:    "",
			password: "password123",
			setupMocks: func(userRepo *MockUserRepository, hasher *MockPasswordHasher) {
				// No mocks needed for validation failure
			},
			wantErr:     true,
			expectedErr: "email cannot be empty",
		},
		{
			name:     "empty password",
			email:    "user@example.com",
			password: "",
			setupMocks: func(userRepo *MockUserRepository, hasher *MockPasswordHasher) {
				// No mocks needed for validation failure
			},
			wantErr:     true,
			expectedErr: "password cannot be empty",
		},
		{
			name:     "user not found",
			email:    "nonexistent@example.com",
			password: "password123",
			setupMocks: func(userRepo *MockUserRepository, hasher *MockPasswordHasher) {
				userRepo.On("GetByEmail", "nonexistent@example.com").Return(nil, nil)
			},
			wantErr:     true,
			expectedErr: "invalid email or password",
		},
		{
			name:     "inactive user",
			email:    "inactive@example.com",
			password: "password123",
			setupMocks: func(userRepo *MockUserRepository, hasher *MockPasswordHasher) {
				testUser := &user.User{
					ID:           "user-123",
					Email:        "inactive@example.com",
					PasswordHash: "hash",
					IsActive:     false,
				}
				userRepo.On("GetByEmail", "inactive@example.com").Return(testUser, nil)
			},
			wantErr:     true,
			expectedErr: "user account is not active",
		},
		{
			name:     "invalid password",
			email:    "user@example.com",
			password: "wrongpassword",
			setupMocks: func(userRepo *MockUserRepository, hasher *MockPasswordHasher) {
				testUser := &user.User{
					ID:           "user-123",
					Email:        "user@example.com",
					PasswordHash: "hash",
					IsActive:     true,
				}
				userRepo.On("GetByEmail", "user@example.com").Return(testUser, nil)
				hasher.On("Verify", "wrongpassword", "hash").Return(false)
			},
			wantErr:     true,
			expectedErr: "invalid email or password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := new(MockUserRepository)
			hasher := new(MockPasswordHasher)
			tokenManager := NewTokenManager("test-secret", "test-issuer")
			
			service := NewServiceWithPasswordHasher(userRepo, tokenManager, hasher)
			
			tt.setupMocks(userRepo, hasher)
			
			token, err := service.Login(tt.email, tt.password)
			
			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != "" {
					assert.Contains(t, err.Error(), tt.expectedErr)
				}
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
			}
			
			userRepo.AssertExpectations(t)
			hasher.AssertExpectations(t)
		})
	}
}

func TestService_ValidateToken(t *testing.T) {
	tests := []struct {
		name        string
		token       string
		setupMocks  func(*MockUserRepository)
		wantUserID  string
		wantErr     bool
		expectedErr string
	}{
		{
			name:        "empty token",
			token:       "",
			setupMocks:  func(userRepo *MockUserRepository) {},
			wantErr:     true,
			expectedErr: "token cannot be empty",
		},
		{
			name:        "invalid token format",
			token:       "invalid-token",
			setupMocks:  func(userRepo *MockUserRepository) {},
			wantErr:     true,
			expectedErr: "invalid token",
		},
		{
			name: "valid token with active user",
			setupMocks: func(userRepo *MockUserRepository) {
				testUser := &user.User{
					ID:       "user-123",
					Email:    "user@example.com",
					IsActive: true,
				}
				userRepo.On("GetByID", "user-123").Return(testUser, nil)
			},
			wantUserID: "user-123",
			wantErr:    false,
		},
		{
			name: "valid token with inactive user",
			setupMocks: func(userRepo *MockUserRepository) {
				testUser := &user.User{
					ID:       "user-123",
					Email:    "user@example.com",
					IsActive: false,
				}
				userRepo.On("GetByID", "user-123").Return(testUser, nil)
			},
			wantErr:     true,
			expectedErr: "user account is not active",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := new(MockUserRepository)
			tokenManager := NewTokenManager("test-secret", "test-issuer")
			service := NewService(userRepo, tokenManager)
			
			// Generate a valid token for tests that need it
			if tt.name == "valid token with active user" || tt.name == "valid token with inactive user" {
				validToken, err := tokenManager.GenerateToken("user-123", 15*time.Minute)
				assert.NoError(t, err)
				tt.token = validToken
			}
			
			tt.setupMocks(userRepo)
			
			userID, err := service.ValidateToken(tt.token)
			
			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != "" {
					assert.Contains(t, err.Error(), tt.expectedErr)
				}
				assert.Empty(t, userID)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantUserID, userID)
			}
			
			userRepo.AssertExpectations(t)
		})
	}
}
