package authtest

import (
	"github.com/EliasRanz/ai-code-gen/ai-ui-generator/internal/auth"
	"github.com/EliasRanz/ai-code-gen/ai-ui-generator/internal/user"
	"github.com/stretchr/testify/mock"
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

// Helper functions
func CreateTestTokenManager() *auth.TokenManager {
	return auth.NewTokenManager("test-secret", "test-issuer")
}

func CreateTestService() *auth.Service {
	tokenManager := CreateTestTokenManager()
	userRepo := &MockUserRepository{}
	return auth.NewService(userRepo, tokenManager)
}

func CreateTestHandler() *auth.Handler {
	service := CreateTestService()
	return auth.NewHandler(service)
}
