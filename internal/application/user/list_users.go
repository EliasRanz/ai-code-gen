package user

import (
	"context"
	"fmt"

	"github.com/EliasRanz/ai-code-gen/internal/domain/common"
	"github.com/EliasRanz/ai-code-gen/internal/domain/user"
)

// ListUsersUseCase handles user listing business logic
type ListUsersUseCase struct {
	userRepo user.Repository
}

// NewListUsersUseCase creates a new instance of ListUsersUseCase
func NewListUsersUseCase(userRepo user.Repository) *ListUsersUseCase {
	return &ListUsersUseCase{
		userRepo: userRepo,
	}
}

// ListUsersRequest represents the input for listing users
type ListUsersRequest struct {
	Page   int32  `json:"page" validate:"min=1"`
	Limit  int32  `json:"limit" validate:"min=1,max=100"`
	Search string `json:"search"`
}

// ListUsersResponse represents the output of user listing
type ListUsersResponse struct {
	Users      []user.User `json:"users"`
	TotalCount int         `json:"total_count"`
	Page       int32       `json:"page"`
	Limit      int32       `json:"limit"`
}

// Execute performs the list users use case
func (uc *ListUsersUseCase) Execute(ctx context.Context, req ListUsersRequest) (*ListUsersResponse, error) {
	// Set defaults
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	params := common.PaginationParams{
		Page:  req.Page,
		Limit: req.Limit,
	}

	users, err := uc.userRepo.List(ctx, params, req.Search)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	// Get total count for pagination
	totalCount, err := uc.userRepo.Count(ctx, req.Search)
	if err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}

	return &ListUsersResponse{
		Users:      users,
		TotalCount: totalCount,
		Page:       req.Page,
		Limit:      req.Limit,
	}, nil
}
