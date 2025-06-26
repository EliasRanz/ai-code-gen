package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/EliasRanz/ai-code-gen/ai-ui-generator/internal/domain/auth"
	"github.com/EliasRanz/ai-code-gen/ai-ui-generator/internal/domain/common"
	"github.com/EliasRanz/ai-code-gen/ai-ui-generator/internal/domain/user"
)

// RefreshTokenUseCase handles token refresh business logic
type RefreshTokenUseCase struct {
	sessionRepo   auth.SessionRepository
	tokenProvider auth.TokenProvider
	userRepo      user.Repository
}

// NewRefreshTokenUseCase creates a new instance of RefreshTokenUseCase
func NewRefreshTokenUseCase(
	sessionRepo auth.SessionRepository,
	tokenProvider auth.TokenProvider,
	userRepo user.Repository,
) *RefreshTokenUseCase {
	return &RefreshTokenUseCase{
		sessionRepo:   sessionRepo,
		tokenProvider: tokenProvider,
		userRepo:      userRepo,
	}
}

// RefreshTokenRequest represents the input for token refresh
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// RefreshTokenResponse represents the output of token refresh
type RefreshTokenResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// Execute performs the refresh token use case
func (uc *RefreshTokenUseCase) Execute(ctx context.Context, req RefreshTokenRequest) (*RefreshTokenResponse, error) {
	if req.RefreshToken == "" {
		return nil, common.NewValidationError("refresh token is required", nil)
	}

	// Get session by refresh token
	session, err := uc.sessionRepo.GetByRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		if common.IsNotFoundError(err) {
			return nil, common.NewUnauthorizedError("invalid refresh token")
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Check if session is expired
	if session.IsExpired() {
		// Delete expired session
		_ = uc.sessionRepo.Delete(ctx, session.ID)
		return nil, common.NewUnauthorizedError("refresh token expired")
	}

	// Verify user still exists and is active
	u, err := uc.userRepo.GetByID(ctx, session.UserID)
	if err != nil {
		if common.IsNotFoundError(err) {
			// Delete session for non-existent user
			_ = uc.sessionRepo.Delete(ctx, session.ID)
			return nil, common.NewUnauthorizedError("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if !u.Active {
		// Delete session for inactive user
		_ = uc.sessionRepo.Delete(ctx, session.ID)
		return nil, common.NewUnauthorizedError("user account is inactive")
	}

	// Generate new tokens
	newAccessToken, err := uc.tokenProvider.GenerateAccessToken(session.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken, err := uc.tokenProvider.GenerateRefreshToken(session.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Update session with new tokens
	session.AccessToken = newAccessToken
	session.RefreshToken = newRefreshToken
	session.ExpiresAt = time.Now().Add(24 * time.Hour) // Extend expiration

	if err := uc.sessionRepo.Update(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to update session: %w", err)
	}

	return &RefreshTokenResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		ExpiresAt:    session.ExpiresAt,
	}, nil
}
