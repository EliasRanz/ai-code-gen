package user

import (
	"context"
	"strconv"
	"time"

	pb "github.com/EliasRanz/ai-code-gen/ai-ui-generator/api/proto/user"
)

// GRPCClient wraps the gRPC service for HTTP handlers
type GRPCClient struct {
	server *GRPCServer
}

// NewGRPCClient creates a new gRPC client wrapper
func NewGRPCClient(server *GRPCServer) *GRPCClient {
	return &GRPCClient{
		server: server,
	}
}

// Internal helper to convert protobuf User to domain User
func convertPBUserToDomain(pbUser *pb.User) *User {
	return &User{
		ID:        pbUser.Id,
		Email:     pbUser.Email,
		Name:      pbUser.Name,
		AvatarURL: pbUser.AvatarUrl,
		CreatedAt: time.Unix(pbUser.CreatedAt, 0),
		UpdatedAt: time.Unix(pbUser.UpdatedAt, 0),
	}
}

// CreateUser creates a new user via gRPC
func (c *GRPCClient) CreateUser(req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	return c.server.CreateUser(context.Background(), req)
}

// GetUser retrieves a user by ID via gRPC
func (c *GRPCClient) GetUser(userID string) (*pb.GetUserResponse, error) {
	req := &pb.GetUserRequest{Id: userID}
	return c.server.GetUser(context.Background(), req)
}

// UpdateUser updates a user via gRPC
func (c *GRPCClient) UpdateUser(req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	return c.server.UpdateUser(context.Background(), req)
}

// DeleteUser deletes a user via gRPC
func (c *GRPCClient) DeleteUser(userID string) (*pb.DeleteUserResponse, error) {
	req := &pb.DeleteUserRequest{Id: userID}
	return c.server.DeleteUser(context.Background(), req)
}

// ListUsers lists users via gRPC
func (c *GRPCClient) ListUsers(page, limit int32, search string) (*pb.ListUsersResponse, error) {
	req := &pb.ListUsersRequest{
		Page:   page,
		Limit:  limit,
		Search: search,
	}
	return c.server.ListUsers(context.Background(), req)
}

// CreateProject creates a new project via gRPC
func (c *GRPCClient) CreateProject(req *pb.CreateProjectRequest) (*pb.CreateProjectResponse, error) {
	return c.server.CreateProject(context.Background(), req)
}

// GetProject retrieves a project by ID via gRPC
func (c *GRPCClient) GetProject(projectID string) (*pb.GetProjectResponse, error) {
	req := &pb.GetProjectRequest{Id: projectID}
	return c.server.GetProject(context.Background(), req)
}

// UpdateProject updates a project via gRPC
func (c *GRPCClient) UpdateProject(req *pb.UpdateProjectRequest) (*pb.UpdateProjectResponse, error) {
	return c.server.UpdateProject(context.Background(), req)
}

// DeleteProject deletes a project via gRPC
func (c *GRPCClient) DeleteProject(projectID string) (*pb.DeleteProjectResponse, error) {
	req := &pb.DeleteProjectRequest{Id: projectID}
	return c.server.DeleteProject(context.Background(), req)
}

// ListProjects lists projects via gRPC
func (c *GRPCClient) ListProjects(page, limit int32, search string, status pb.ProjectStatus) (*pb.ListProjectsResponse, error) {
	req := &pb.ListProjectsRequest{
		Page:   page,
		Limit:  limit,
		Search: search,
		Status: status,
	}
	return c.server.ListProjects(context.Background(), req)
}

// ListUserProjects lists projects for a specific user via gRPC
func (c *GRPCClient) ListUserProjects(userID string, page, limit int32, status pb.ProjectStatus) (*pb.ListUserProjectsResponse, error) {
	req := &pb.ListUserProjectsRequest{
		UserId: userID,
		Page:   page,
		Limit:  limit,
		Status: status,
	}
	return c.server.ListUserProjects(context.Background(), req)
}

// Helper functions for parameter parsing

// ParseInt32 safely parses a string to int32
func ParseInt32(s string, defaultValue int32) int32 {
	if s == "" {
		return defaultValue
	}
	if val, err := strconv.ParseInt(s, 10, 32); err == nil {
		return int32(val)
	}
	return defaultValue
}

// ParseProjectStatus parses a string to ProjectStatus
func ParseProjectStatus(s string) pb.ProjectStatus {
	switch s {
	case "draft":
		return pb.ProjectStatus_PROJECT_STATUS_DRAFT
	case "active":
		return pb.ProjectStatus_PROJECT_STATUS_ACTIVE
	case "completed":
		return pb.ProjectStatus_PROJECT_STATUS_COMPLETED
	case "archived":
		return pb.ProjectStatus_PROJECT_STATUS_ARCHIVED
	default:
		return pb.ProjectStatus_PROJECT_STATUS_DRAFT
	}
}
