package auth

import (
	"context"
)

// GetUserContextUseCase handles user context retrieval by user ID
type GetUserContextUseCase struct {
	userRepo UserRepository
}

// NewGetUserContextUseCase creates a new instance of GetUserContextUseCase
func NewGetUserContextUseCase(userRepo UserRepository) *GetUserContextUseCase {
	return &GetUserContextUseCase{
		userRepo: userRepo,
	}
}

// GetUserContextRequest represents the input for user context retrieval
type GetUserContextRequest struct {
	UserID UserID `json:"user_id" validate:"required"`
}

// GetUserContextResponse represents the output of user context retrieval
type GetUserContextResponse struct {
	UserContext *UserContextData `json:"user_context"`
}

// Execute retrieves user context by user ID
func (uc *GetUserContextUseCase) Execute(ctx context.Context, req GetUserContextRequest) (*GetUserContextResponse, error) {
	user, err := uc.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	userContext := &UserContextData{
		UserID:   user.ID,
		Email:    user.Email,
		Username: user.Username,
		Name:     user.Name,
		Active:   user.Active,
		Roles:    user.Roles,
	}

	// Set primary role
	if len(user.Roles) > 0 {
		userContext.Role = user.Roles[0]
	} else {
		userContext.Role = "user"
	}

	userContext.Permissions = []string{} // Will be implemented later

	return &GetUserContextResponse{
		UserContext: userContext,
	}, nil
}
