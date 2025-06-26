package user

import (
	"context"
	"fmt"

	"github.com/ai-code-gen/ai-ui-generator/internal/domain/common"
	"github.com/ai-code-gen/ai-ui-generator/internal/domain/user"
)

// DeleteUserUseCase handles user deletion business logic
type DeleteUserUseCase struct {
	userRepo user.Repository
	notifier user.NotificationService
}

// NewDeleteUserUseCase creates a new instance of DeleteUserUseCase
func NewDeleteUserUseCase(
	userRepo user.Repository,
	notifier user.NotificationService,
) *DeleteUserUseCase {
	return &DeleteUserUseCase{
		userRepo: userRepo,
		notifier: notifier,
	}
}

// DeleteUserRequest represents the input for deleting a user
type DeleteUserRequest struct {
	UserID common.UserID `validate:"required"`
}

// DeleteUserResponse represents the output of user deletion
type DeleteUserResponse struct {
	Success bool `json:"success"`
}

// Execute performs the delete user use case
func (uc *DeleteUserUseCase) Execute(ctx context.Context, req DeleteUserRequest) (*DeleteUserResponse, error) {
	if req.UserID.IsEmpty() {
		return nil, common.NewValidationError("user ID is required", nil)
	}

	// Check if user exists
	_, err := uc.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		if common.IsNotFoundError(err) {
			return nil, common.NewNotFoundError("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Delete user
	if err := uc.userRepo.Delete(ctx, req.UserID); err != nil {
		return nil, fmt.Errorf("failed to delete user: %w", err)
	}

	// Send notification (async, don't block on failure)
	go func() {
		if err := uc.notifier.NotifyUserDeleted(context.Background(), req.UserID); err != nil {
			// Log error but don't fail the use case
			// Logger would be injected in production
		}
	}()

	return &DeleteUserResponse{
		Success: true,
	}, nil
}
