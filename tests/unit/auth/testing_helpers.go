package authtest

import (
	"context"

	"github.com/EliasRanz/ai-code-gen/internal/auth"
	"github.com/stretchr/testify/mock"
)

// MockTokenProvider for testing
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

// MockUserRepository for testing
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetByID(ctx context.Context, id auth.UserID) (auth.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return auth.User{}, args.Error(1)
	}
	return args.Get(0).(auth.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (auth.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return auth.User{}, args.Error(1)
	}
	return args.Get(0).(auth.User), args.Error(1)
}

func (m *MockUserRepository) Create(ctx context.Context, u auth.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

func (m *MockUserRepository) Update(ctx context.Context, u auth.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id auth.UserID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, params auth.PaginationParams, search string) ([]auth.User, error) {
	args := m.Called(ctx, params, search)
	return args.Get(0).([]auth.User), args.Error(1)
}

func (m *MockUserRepository) Count(ctx context.Context, search string) (int, error) {
	args := m.Called(ctx, search)
	return args.Get(0).(int), args.Error(1)
}

// MockSessionRepository for testing
type MockSessionRepository struct {
	mock.Mock
}

func (m *MockSessionRepository) Create(ctx context.Context, session auth.Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockSessionRepository) GetByID(ctx context.Context, id auth.SessionID) (auth.Session, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return auth.Session{}, args.Error(1)
	}
	return args.Get(0).(auth.Session), args.Error(1)
}

func (m *MockSessionRepository) GetByRefreshToken(ctx context.Context, refreshToken string) (auth.Session, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return auth.Session{}, args.Error(1)
	}
	return args.Get(0).(auth.Session), args.Error(1)
}

func (m *MockSessionRepository) Update(ctx context.Context, session auth.Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockSessionRepository) Delete(ctx context.Context, id auth.SessionID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSessionRepository) DeleteByUserID(ctx context.Context, userID auth.UserID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockSessionRepository) GetByAccessToken(ctx context.Context, accessToken string) (auth.Session, error) {
	args := m.Called(ctx, accessToken)
	if args.Get(0) == nil {
		return auth.Session{}, args.Error(1)
	}
	return args.Get(0).(auth.Session), args.Error(1)
}

func (m *MockSessionRepository) CleanExpired(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
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
