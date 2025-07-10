package grpc

import (
	"context"
	"fmt"

	"github.com/EliasRanz/ai-code-gen/api/proto/user"
	applicationuser "github.com/EliasRanz/ai-code-gen/internal/application/user"
	"github.com/EliasRanz/ai-code-gen/internal/domain/common"
	domainUser "github.com/EliasRanz/ai-code-gen/internal/domain/user"
	"github.com/rs/zerolog/log"
)

// UserServer implements the UserService gRPC interface
type UserServer struct {
	user.UnimplementedUserServiceServer
	createUserUC *applicationuser.CreateUserUseCase
	getUserUC    *applicationuser.GetUserUseCase
	updateUserUC *applicationuser.UpdateUserUseCase
	deleteUserUC *applicationuser.DeleteUserUseCase
	listUsersUC  *applicationuser.ListUsersUseCase
}

// NewUserServer creates a new gRPC server instance
func NewUserServer(
	createUserUC *applicationuser.CreateUserUseCase,
	getUserUC *applicationuser.GetUserUseCase,
	updateUserUC *applicationuser.UpdateUserUseCase,
	deleteUserUC *applicationuser.DeleteUserUseCase,
	listUsersUC *applicationuser.ListUsersUseCase,
) *UserServer {
	return &UserServer{
		createUserUC: createUserUC,
		getUserUC:    getUserUC,
		updateUserUC: updateUserUC,
		deleteUserUC: deleteUserUC,
		listUsersUC:  listUsersUC,
	}
}

func convertDomainUserToPB(u *domainUser.User) *user.User {
	return &user.User{
		Id:        fmt.Sprint(u.ID),
		Email:     u.Email,
		Name:      u.Username,
		AvatarUrl: u.AvatarURL,
		Roles:     u.Roles,
		CreatedAt: u.CreatedAt.Unix(),
		UpdatedAt: u.UpdatedAt.Unix(),
	}
}

// CreateUser creates a new user
func (s *UserServer) CreateUser(ctx context.Context, req *user.CreateUserRequest) (*user.CreateUserResponse, error) {
	log.Info().
		Str("email", req.Email).
		Str("name", req.Name).
		Msg("gRPC CreateUser called")

	// Validate request
	if req.Email == "" {
		return nil, fmt.Errorf("email is required")
	}
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	// Create domain user object
	createReq := applicationuser.CreateUserRequest{
		Email:     req.Email,
		Name:      req.Name,
		AvatarURL: req.AvatarUrl,
		Roles:     req.Roles,
	}

	// Create user using service
	resp, err := s.createUserUC.Execute(ctx, createReq)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create user")
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Convert to protobuf and return
	pbUser := convertDomainUserToPB(resp.User)
	return &user.CreateUserResponse{
		User: pbUser,
	}, nil
}

// GetUser retrieves a user by ID
func (s *UserServer) GetUser(ctx context.Context, req *user.GetUserRequest) (*user.GetUserResponse, error) {
	log.Info().
		Str("user_id", req.Id).
		Msg("gRPC GetUser called")

	// Validate request
	if req.Id == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	// Get user using service
	getReq := applicationuser.GetUserRequest{
		UserID: common.UserID(req.Id),
	}
	resp, err := s.getUserUC.Execute(ctx, getReq)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user")
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Convert to protobuf and return
	pbUser := convertDomainUserToPB(resp.User)
	return &user.GetUserResponse{
		User: pbUser,
	}, nil
}

// UpdateUser updates an existing user
func (s *UserServer) UpdateUser(ctx context.Context, req *user.UpdateUserRequest) (*user.UpdateUserResponse, error) {
	log.Info().
		Str("user_id", req.Id).
		Msg("gRPC UpdateUser called")

	// Validate request
	if req.Id == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	// Create update request DTO
	updateReq := applicationuser.UpdateUserRequest{
		UserID:    common.UserID(req.Id),
		Name:      &req.Name,
		AvatarURL: &req.AvatarUrl,
		Roles:     &req.Roles,
	}

	// Update user using service
	resp, err := s.updateUserUC.Execute(ctx, updateReq)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update user")
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Convert to protobuf and return
	pbUser := convertDomainUserToPB(resp.User)
	return &user.UpdateUserResponse{
		User: pbUser,
	}, nil
}

// DeleteUser deletes a user by ID
func (s *UserServer) DeleteUser(ctx context.Context, req *user.DeleteUserRequest) (*user.DeleteUserResponse, error) {
	log.Info().
		Str("user_id", req.Id).
		Msg("gRPC DeleteUser called")

	// Validate request
	if req.Id == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	// Delete user using service
	deleteReq := applicationuser.DeleteUserRequest{
		UserID: common.UserID(req.Id),
	}
	_, err := s.deleteUserUC.Execute(ctx, deleteReq)
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete user")
		return nil, fmt.Errorf("failed to delete user: %w", err)
	}

	return &user.DeleteUserResponse{}, nil
}

// ListUsers lists all users
func (s *UserServer) ListUsers(ctx context.Context, req *user.ListUsersRequest) (*user.ListUsersResponse, error) {
	log.Info().Msg("gRPC ListUsers called")

	// List users using service
	listReq := applicationuser.ListUsersRequest{
		Page:   req.Page,
		Limit:  req.Limit,
		Search: req.Search,
	}
	resp, err := s.listUsersUC.Execute(ctx, listReq)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list users")
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	// Convert to protobuf and return
	pbUsers := make([]*user.User, len(resp.Users))
	for i, u := range resp.Users {
		pbUsers[i] = convertDomainUserToPB(&u)
	}

	return &user.ListUsersResponse{
		Users: pbUsers,
		Total: int32(resp.TotalCount),
	}, nil
}
