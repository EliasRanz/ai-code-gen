package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/ai-code-gen/ai-ui-generator/internal/domain/auth"
	"github.com/ai-code-gen/ai-ui-generator/internal/domain/common"
	"github.com/ai-code-gen/ai-ui-generator/internal/domain/user"
)

// LoginUseCase handles user authentication business logic
type LoginUseCase struct {
	userRepo       user.Repository
	sessionRepo    auth.SessionRepository
	passwordHasher user.PasswordHasher
	tokenProvider  auth.TokenProvider
}

// NewLoginUseCase creates a new instance of LoginUseCase
func NewLoginUseCase(
	userRepo user.Repository,
	sessionRepo auth.SessionRepository,
	passwordHasher user.PasswordHasher,
	tokenProvider auth.TokenProvider,
) *LoginUseCase {
	return &LoginUseCase{
		userRepo:       userRepo,
		sessionRepo:    sessionRepo,
		passwordHasher: passwordHasher,
		tokenProvider:  tokenProvider,
	}
}

// LoginRequest represents the input for user login
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents the output of user login
type LoginResponse struct {
	User         *user.User    `json:"user"`
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
	ExpiresAt    time.Time     `json:"expires_at"`
	Session      *auth.Session `json:"session"`
}

// Execute performs the login use case
func (uc *LoginUseCase) Execute(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	// Get user by email
	u, err := uc.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if common.IsNotFoundError(err) {
			return nil, common.NewUnauthorizedError("invalid credentials")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Check if user is active
	if !u.Active {
		return nil, common.NewUnauthorizedError("user account is inactive")
	}

	// Verify password using the password hasher
	if !u.VerifyPassword(uc.passwordHasher, req.Password) {
		return nil, common.NewUnauthorizedError("invalid credentials")
	}

	// Generate tokens
	accessToken, err := uc.tokenProvider.GenerateAccessToken(u.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := uc.tokenProvider.GenerateRefreshToken(u.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Create session
	session := &auth.Session{
		UserID:       u.ID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(24 * time.Hour), // 24 hours
		Status:       auth.StatusActive,
	}

	if err := uc.sessionRepo.Create(ctx, *session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &LoginResponse{
		User:         &u,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    session.ExpiresAt,
		Session:      session,
	}, nil
}
