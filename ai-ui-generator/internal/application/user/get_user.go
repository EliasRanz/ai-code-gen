package user

import (
	"context"
	"fmt"

	"github.com/EliasRanz/ai-code-gen/ai-ui-generator/internal/domain/common"
	"github.com/EliasRanz/ai-code-gen/ai-ui-generator/internal/domain/user"
)

// GetUserUseCase handles user retrieval business logic
type GetUserUseCase struct {
	userRepo user.Repository
}

// NewGetUserUseCase creates a new instance of GetUserUseCase
func NewGetUserUseCase(userRepo user.Repository) *GetUserUseCase {
	return &GetUserUseCase{
		userRepo: userRepo,
	}
}

// GetUserRequest represents the input for getting a user
type GetUserRequest struct {
	UserID common.UserID `validate:"required"`
}

// GetUserResponse represents the output of user retrieval
type GetUserResponse struct {
	User *user.User `json:"user"`
}

// Execute performs the get user use case
func (uc *GetUserUseCase) Execute(ctx context.Context, req GetUserRequest) (*GetUserResponse, error) {
	if req.UserID.IsEmpty() {
		return nil, common.NewValidationError("user ID is required", nil)
	}

	u, err := uc.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		if common.IsNotFoundError(err) {
			return nil, common.NewNotFoundError("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &GetUserResponse{
		User: &u,
	}, nil
}
