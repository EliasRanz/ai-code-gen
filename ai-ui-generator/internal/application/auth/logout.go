package auth

import (
	"context"
	"fmt"

	"github.com/ai-code-gen/ai-ui-generator/internal/domain/auth"
	"github.com/ai-code-gen/ai-ui-generator/internal/domain/common"
)

// LogoutUseCase handles user logout business logic
type LogoutUseCase struct {
	sessionRepo auth.SessionRepository
}

// NewLogoutUseCase creates a new instance of LogoutUseCase
func NewLogoutUseCase(sessionRepo auth.SessionRepository) *LogoutUseCase {
	return &LogoutUseCase{
		sessionRepo: sessionRepo,
	}
}

// LogoutRequest represents the input for user logout
type LogoutRequest struct {
	AccessToken string `validate:"required"`
}

// LogoutResponse represents the output of user logout
type LogoutResponse struct {
	Success bool `json:"success"`
}

// Execute performs the logout use case
func (uc *LogoutUseCase) Execute(ctx context.Context, req LogoutRequest) (*LogoutResponse, error) {
	if req.AccessToken == "" {
		return nil, common.NewValidationError("access token is required", nil)
	}

	// Get session by access token
	session, err := uc.sessionRepo.GetByAccessToken(ctx, req.AccessToken)
	if err != nil {
		if common.IsNotFoundError(err) {
			// Session not found, consider it already logged out
			return &LogoutResponse{Success: true}, nil
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Delete session
	if err := uc.sessionRepo.Delete(ctx, session.ID); err != nil {
		return nil, fmt.Errorf("failed to delete session: %w", err)
	}

	return &LogoutResponse{Success: true}, nil
}
