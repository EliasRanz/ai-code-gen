package user

import (
	"context"
	"fmt"

	"github.com/EliasRanz/ai-code-gen/api/proto/user"
	"github.com/EliasRanz/ai-code-gen/internal/utilities"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

// UserGRPCServer implements the UserService gRPC interface and utilities.UserGRPCService
type UserGRPCServer struct {
	user.UnimplementedUserServiceServer
	userCreator   *UserCreator
	userRetriever *UserRetriever
	userUpdater   *UserUpdater
	userDeleter   *UserDeleter
	userLister    *UserLister
	serviceInfo   utilities.ServiceInfo
}

// NewUserGRPCServer creates a new gRPC server instance
func NewUserGRPCServer(
	userCreator *UserCreator,
	userRetriever *UserRetriever,
	userUpdater *UserUpdater,
	userDeleter *UserDeleter,
	userLister *UserLister,
) *UserGRPCServer {
	return &UserGRPCServer{
		userCreator:   userCreator,
		userRetriever: userRetriever,
		userUpdater:   userUpdater,
		userDeleter:   userDeleter,
		userLister:    userLister,
		serviceInfo: utilities.ServiceInfo{
			Name:        "user-service",
			Version:     "1.0.0",
			Description: "User management gRPC service",
			Endpoints: []string{
				"GetUser", "CreateUser", "UpdateUser", "DeleteUser", "ListUsers",
			},
		},
	}
}

// RegisterService registers the service with gRPC server
func (s *UserGRPCServer) RegisterService(server *grpc.Server) error {
	user.RegisterUserServiceServer(server, s)
	log.Info().
		Str("service", s.serviceInfo.Name).
		Str("version", s.serviceInfo.Version).
		Msg("User gRPC service registered")
	return nil
}

// GetServiceInfo returns service metadata
func (s *UserGRPCServer) GetServiceInfo() utilities.ServiceInfo {
	return s.serviceInfo
}

// GetInterceptors returns list of interceptors for this service
func (s *UserGRPCServer) GetInterceptors() []grpc.UnaryServerInterceptor {
	return []grpc.UnaryServerInterceptor{
		s.loggingInterceptor(),
		s.validationInterceptor(),
	}
}

// HealthCheck validates service health
func (s *UserGRPCServer) HealthCheck(ctx context.Context) error {
	if s.userRetriever == nil {
		return fmt.Errorf("user retriever not initialized")
	}
	return nil
}

// ValidateService validates service configuration
func (s *UserGRPCServer) ValidateService() error {
	if s.userCreator == nil || s.userRetriever == nil ||
		s.userUpdater == nil || s.userDeleter == nil || s.userLister == nil {
		return fmt.Errorf("one or more user service dependencies not initialized")
	}
	return nil
}

// Start initializes the service
func (s *UserGRPCServer) Start(ctx context.Context) error {
	return s.ValidateService()
}

// Stop gracefully shuts down the service
func (s *UserGRPCServer) Stop(ctx context.Context) error {
	log.Info().Msg("User gRPC service stopping")
	return nil
}

// loggingInterceptor provides request/response logging
func (s *UserGRPCServer) loggingInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (interface{}, error) {
		log.Info().
			Str("method", info.FullMethod).
			Msg("gRPC request received")

		resp, err := handler(ctx, req)
		if err != nil {
			log.Error().Err(err).
				Str("method", info.FullMethod).
				Msg("gRPC request failed")
		}
		return resp, err
	}
}

// validationInterceptor provides request validation
func (s *UserGRPCServer) validationInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (interface{}, error) {
		// Basic validation can be added here
		return handler(ctx, req)
	}
}

// convertDomainUserToPB converts domain user to protobuf
func convertDomainUserToPB(u *User) *user.User {
	return &user.User{
		Id:        fmt.Sprint(u.ID),
		Email:     u.Email,
		Name:      u.Name,
		AvatarUrl: u.AvatarURL,
		Roles:     u.Roles,
		CreatedAt: u.CreatedAt.Unix(),
		UpdatedAt: u.UpdatedAt.Unix(),
	}
}

// CreateUser creates a new user
func (s *UserGRPCServer) CreateUser(ctx context.Context,
	req *user.CreateUserRequest) (*user.CreateUserResponse, error) {
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
	createReq := CreateUserRequest{
		Email:     req.Email,
		Name:      req.Name,
		AvatarURL: req.AvatarUrl,
		Roles:     req.Roles,
	}

	// Create user using service
	resp, err := s.userCreator.Execute(ctx, createReq)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create user")
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Convert to protobuf and return
	pbUser := convertDomainUserToPB(resp.User)
	return &user.CreateUserResponse{User: pbUser}, nil
}

// GetUser retrieves a user by ID
func (s *UserGRPCServer) GetUser(ctx context.Context,
	req *user.GetUserRequest) (*user.GetUserResponse, error) {
	log.Info().
		Str("user_id", req.Id).
		Msg("gRPC GetUser called")

	// Validate request
	if req.Id == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	// Get user using service
	getReq := GetUserRequest{
		UserID: utilities.UserID(req.Id),
	}
	resp, err := s.userRetriever.Execute(ctx, getReq)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user")
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Convert to protobuf and return
	pbUser := convertDomainUserToPB(resp.User)
	return &user.GetUserResponse{User: pbUser}, nil
}

// UpdateUser updates an existing user
func (s *UserGRPCServer) UpdateUser(ctx context.Context,
	req *user.UpdateUserRequest) (*user.UpdateUserResponse, error) {
	log.Info().
		Str("user_id", req.Id).
		Msg("gRPC UpdateUser called")

	// Validate request
	if req.Id == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	// Create update request DTO
	var role *Role
	if len(req.Roles) > 0 {
		r := Role(req.Roles[0])
		role = &r
	}

	updateReq := UpdateUserRequest{
		UserID:    utilities.UserID(req.Id),
		Name:      &req.Name,
		AvatarURL: &req.AvatarUrl,
		Role:      role,
	}

	// Update user using service
	resp, err := s.userUpdater.Execute(ctx, updateReq)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update user")
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Convert to protobuf and return
	pbUser := convertDomainUserToPB(resp.User)
	return &user.UpdateUserResponse{User: pbUser}, nil
}

// DeleteUser deletes a user by ID
func (s *UserGRPCServer) DeleteUser(ctx context.Context,
	req *user.DeleteUserRequest) (*user.DeleteUserResponse, error) {
	log.Info().
		Str("user_id", req.Id).
		Msg("gRPC DeleteUser called")

	// Validate request
	if req.Id == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	// Delete user using service
	deleteReq := DeleteUserRequest{
		UserID: utilities.UserID(req.Id),
	}
	_, err := s.userDeleter.Execute(ctx, deleteReq)
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete user")
		return nil, fmt.Errorf("failed to delete user: %w", err)
	}

	return &user.DeleteUserResponse{}, nil
}

// ListUsers lists all users
func (s *UserGRPCServer) ListUsers(ctx context.Context,
	req *user.ListUsersRequest) (*user.ListUsersResponse, error) {
	log.Info().Msg("gRPC ListUsers called")

	// List users using service
	listReq := ListUsersRequest{
		Page:   req.Page,
		Limit:  req.Limit,
		Search: req.Search,
	}
	resp, err := s.userLister.Execute(ctx, listReq)
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
		Total: int32(resp.Total),
	}, nil
}
