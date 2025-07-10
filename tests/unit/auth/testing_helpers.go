package authtest

import (
	"context"

	"github.com/EliasRanz/ai-code-gen/internal/application/auth"
	auth_infra "github.com/EliasRanz/ai-code-gen/internal/auth"
	domain_auth "github.com/EliasRanz/ai-code-gen/internal/domain/auth"
	"github.com/EliasRanz/ai-code-gen/internal/domain/common"
	"github.com/EliasRanz/ai-code-gen/internal/domain/user"
	"github.com/EliasRanz/ai-code-gen/internal/infrastructure/observability"
	http_iface "github.com/EliasRanz/ai-code-gen/internal/interfaces/http"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository for testing
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetByID(ctx context.Context, id common.UserID) (user.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return user.User{}, args.Error(1)
	}
	return args.Get(0).(user.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (user.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return user.User{}, args.Error(1)
	}
	return args.Get(0).(user.User), args.Error(1)
}

func (m *MockUserRepository) Create(ctx context.Context, u user.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
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
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]user.User), args.Error(1)
}

func (m *MockUserRepository) Count(ctx context.Context, search string) (int, error) {
	args := m.Called(ctx, search)
	return args.Int(0), args.Error(1)
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
func CreateTestTokenManager() *auth_infra.TokenManager {
	return auth_infra.NewTokenManager("test-secret", "test-issuer")
}

func CreateTestLoginUseCase(userRepo user.Repository, sessionRepo domain_auth.SessionRepository) *auth.LoginUseCase {
	tokenManager := CreateTestTokenManager()
	return auth.NewLoginUseCase(userRepo, sessionRepo, &MockPasswordHasher{}, tokenManager)
}

func CreateTestLogoutUseCase(sessionRepo domain_auth.SessionRepository) *auth.LogoutUseCase {
	return auth.NewLogoutUseCase(sessionRepo)
}

func CreateTestRefreshTokenUseCase(userRepo user.Repository, sessionRepo domain_auth.SessionRepository) *auth.RefreshTokenUseCase {
	tokenManager := CreateTestTokenManager()
	return auth.NewRefreshTokenUseCase(sessionRepo, tokenManager, userRepo)
}

func CreateTestHandler(userRepo user.Repository, sessionRepo domain_auth.SessionRepository) *http_iface.AuthHandler {
	loginUC := CreateTestLoginUseCase(userRepo, sessionRepo)
	logoutUC := CreateTestLogoutUseCase(sessionRepo)
	refreshTokenUC := CreateTestRefreshTokenUseCase(userRepo, sessionRepo)
	logger := observability.NewLogger("debug", "console")
	return http_iface.NewAuthHandler(loginUC, logoutUC, refreshTokenUC, logger)
}
