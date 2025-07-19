package utilities

import (
	"context"

	pb_auth "github.com/EliasRanz/ai-code-gen/api/proto/auth"
	pb_user "github.com/EliasRanz/ai-code-gen/api/proto/user"
)

// UserGRPCService defines the interface for user gRPC service operations
type UserGRPCService interface {
	GRPCService
	GetUser(ctx context.Context, req *pb_user.GetUserRequest) (*pb_user.GetUserResponse, error)
	UpdateUser(ctx context.Context, req *pb_user.UpdateUserRequest) (*pb_user.UpdateUserResponse, error)
	ListUsers(ctx context.Context, req *pb_user.ListUsersRequest) (*pb_user.ListUsersResponse, error)
	CreateUser(ctx context.Context, req *pb_user.CreateUserRequest) (*pb_user.CreateUserResponse, error)
	DeleteUser(ctx context.Context, req *pb_user.DeleteUserRequest) (*pb_user.DeleteUserResponse, error)
}

// AuthGRPCService defines the interface for auth gRPC service operations
type AuthGRPCService interface {
	GRPCService
	ValidateToken(ctx context.Context, req *pb_auth.ValidateTokenRequest) (*pb_auth.ValidateTokenResponse, error)
	RefreshToken(ctx context.Context, req *pb_auth.RefreshTokenRequest) (*pb_auth.RefreshTokenResponse, error)
	Login(ctx context.Context, req *pb_auth.LoginRequest) (*pb_auth.LoginResponse, error)
	Logout(ctx context.Context, req *pb_auth.LogoutRequest) (*pb_auth.LogoutResponse, error)
}
