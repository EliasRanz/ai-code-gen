package user

import (
	pb "github.com/EliasRanz/ai-code-gen/ai-ui-generator/api/proto/user"
)

// UserGRPCClient defines the interface for gRPC client methods used by handlers
// This allows for mocking in tests
//
//go:generate mockgen -destination=mock_grpc_client.go -package=user . UserGRPCClient
type UserGRPCClient interface {
	UpdateUser(req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error)
	CreateUser(req *pb.CreateUserRequest) (*pb.CreateUserResponse, error)
	GetUser(userID string) (*pb.GetUserResponse, error)
	ListUsers(page, limit int32, search string) (*pb.ListUsersResponse, error)
	DeleteUser(userID string) (*pb.DeleteUserResponse, error)
	CreateProject(req *pb.CreateProjectRequest) (*pb.CreateProjectResponse, error)
	GetProject(projectID string) (*pb.GetProjectResponse, error)
	UpdateProject(req *pb.UpdateProjectRequest) (*pb.UpdateProjectResponse, error)
	DeleteProject(projectID string) (*pb.DeleteProjectResponse, error)
	ListProjects(page, limit int32, search string, status pb.ProjectStatus) (*pb.ListProjectsResponse, error)
	ListUserProjects(userID string, page, limit int32, status pb.ProjectStatus) (*pb.ListUserProjectsResponse, error)
}

// Global gRPC client - will be set by the main function
var grpcClient UserGRPCClient

// SetGRPCClient sets the global gRPC client for handlers
func SetGRPCClient(client UserGRPCClient) {
	grpcClient = client
}
