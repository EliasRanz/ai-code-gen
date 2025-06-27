package user

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	pb "github.com/EliasRanz/ai-code-gen/api/proto/user"
)

// GRPCServer implements the UserService gRPC interface
type GRPCServer struct {
	pb.UnimplementedUserServiceServer
	service *Service
}

// NewGRPCServer creates a new gRPC server instance
func NewGRPCServer(service *Service) *GRPCServer {
	return &GRPCServer{
		service: service,
	}
}

// User management methods

// CreateUser creates a new user
func (s *GRPCServer) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	log.Info().
		Str("email", req.Email).
		Str("name", req.Name).
		Msg("gRPC CreateUser called")

	// Validate request
	if req.Email == "" {
		return &pb.CreateUserResponse{
			User:  nil,
			Error: "email is required",
		}, nil
	}
	if req.Name == "" {
		return &pb.CreateUserResponse{
			User:  nil,
			Error: "name is required",
		}, nil
	}

	// Create domain user object
	user := &User{
		ID:        generateID(),
		Email:     req.Email,
		Name:      req.Name,
		AvatarURL: req.AvatarUrl,
		Roles:     req.Roles,
	}

	// Create user using service
	err := s.service.CreateUser(user)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create user")
		return &pb.CreateUserResponse{
			User:  nil,
			Error: fmt.Sprintf("failed to create user: %v", err),
		}, nil
	}

	// Convert to protobuf and return
	pbUser := convertDomainUserToPB(user)
	return &pb.CreateUserResponse{
		User:  pbUser,
		Error: "",
	}, nil
}

// GetUser retrieves a user by ID
func (s *GRPCServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	log.Info().
		Str("user_id", req.Id).
		Msg("gRPC GetUser called")

	// Validate request
	if req.Id == "" {
		return &pb.GetUserResponse{
			User:  nil,
			Error: "user ID is required",
		}, nil
	}

	// Get user from service
	user, err := s.service.GetUser(req.Id)
	if err != nil {
		log.Error().Err(err).Str("user_id", req.Id).Msg("Failed to get user")
		return &pb.GetUserResponse{
			User:  nil,
			Error: fmt.Sprintf("failed to get user: %v", err),
		}, nil
	}

	// Convert to protobuf and return
	pbUser := convertDomainUserToPB(user)
	return &pb.GetUserResponse{
		User:  pbUser,
		Error: "",
	}, nil
}

// UpdateUser updates an existing user
func (s *GRPCServer) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	log.Info().
		Str("user_id", req.Id).
		Str("name", req.Name).
		Msg("gRPC UpdateUser called")

	// Validate request
	if req.Id == "" {
		return &pb.UpdateUserResponse{
			User:  nil,
			Error: "user ID is required",
		}, nil
	}

	// Build updates map
	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.AvatarUrl != "" {
		updates["avatar_url"] = req.AvatarUrl
	}
	if req.Roles != nil {
		updates["roles"] = req.Roles
	}

	// Update user using service
	user, err := s.service.UpdateUser(req.Id, updates)
	if err != nil {
		log.Error().Err(err).Str("user_id", req.Id).Msg("Failed to update user")
		return &pb.UpdateUserResponse{
			User:  nil,
			Error: fmt.Sprintf("failed to update user: %v", err),
		}, nil
	}

	// Convert to protobuf and return
	pbUser := convertDomainUserToPB(user)
	return &pb.UpdateUserResponse{
		User:  pbUser,
		Error: "",
	}, nil
}

// DeleteUser deletes a user
func (s *GRPCServer) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	log.Info().
		Str("user_id", req.Id).
		Msg("gRPC DeleteUser called")

	// Validate request
	if req.Id == "" {
		return &pb.DeleteUserResponse{
			Success: false,
			Error:   "user ID is required",
		}, nil
	}

	// Delete user using service
	err := s.service.DeleteUser(req.Id)
	if err != nil {
		log.Error().Err(err).Str("user_id", req.Id).Msg("Failed to delete user")
		return &pb.DeleteUserResponse{
			Success: false,
			Error:   fmt.Sprintf("failed to delete user: %v", err),
		}, nil
	}

	return &pb.DeleteUserResponse{
		Success: true,
		Error:   "",
	}, nil
}

// ListUsers lists users with pagination
func (s *GRPCServer) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	log.Info().
		Int32("page", req.Page).
		Int32("limit", req.Limit).
		Str("search", req.Search).
		Msg("gRPC ListUsers called")

	// Set default pagination values
	page := req.Page
	if page < 1 {
		page = 1
	}
	limit := req.Limit
	if limit <= 0 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}

	// Calculate offset
	offset := (page - 1) * limit

	// Get users from service
	users, err := s.service.ListUsers(int(limit), int(offset))
	if err != nil {
		log.Error().Err(err).Msg("Failed to list users")
		return &pb.ListUsersResponse{
			Users: nil,
			Total: 0,
			Error: fmt.Sprintf("failed to list users: %v", err),
		}, nil
	}

	// Convert to protobuf
	pbUsers := make([]*pb.User, len(users))
	for i, user := range users {
		pbUsers[i] = convertDomainUserToPB(user)
	}

	return &pb.ListUsersResponse{
		Users: pbUsers,
		Total: int32(len(pbUsers)),
		Error: "",
	}, nil
}

// Project management methods

// CreateProject creates a new project
func (s *GRPCServer) CreateProject(ctx context.Context, req *pb.CreateProjectRequest) (*pb.CreateProjectResponse, error) {
	log.Info().
		Str("name", req.Name).
		Str("user_id", req.UserId).
		Msg("gRPC CreateProject called")

	// Validate request
	if req.Name == "" {
		return &pb.CreateProjectResponse{
			Project: nil,
			Error:   "project name is required",
		}, nil
	}
	if req.UserId == "" {
		return &pb.CreateProjectResponse{
			Project: nil,
			Error:   "user ID is required",
		}, nil
	}

	// Parse config JSON if provided
	var config map[string]interface{}
	if req.Config != "" {
		if err := json.Unmarshal([]byte(req.Config), &config); err != nil {
			return &pb.CreateProjectResponse{
				Project: nil,
				Error:   fmt.Sprintf("invalid config JSON: %v", err),
			}, nil
		}
	}

	// Create domain project object
	project := &Project{
		ID:          generateID(),
		Name:        req.Name,
		Description: req.Description,
		UserID:      req.UserId,
		Status:      "draft", // Default status
		Tags:        req.Tags,
		Config:      config,
		IsPublic:    false, // Default to private
	}

	// Create project using service
	err := s.service.CreateProject(project)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create project")
		return &pb.CreateProjectResponse{
			Project: nil,
			Error:   fmt.Sprintf("failed to create project: %v", err),
		}, nil
	}

	// Convert to protobuf and return
	pbProject := convertDomainProjectToPB(project)
	return &pb.CreateProjectResponse{
		Project: pbProject,
		Error:   "",
	}, nil
}

// GetProject retrieves a project by ID
func (s *GRPCServer) GetProject(ctx context.Context, req *pb.GetProjectRequest) (*pb.GetProjectResponse, error) {
	log.Info().
		Str("project_id", req.Id).
		Msg("gRPC GetProject called")

	// Validate request
	if req.Id == "" {
		return &pb.GetProjectResponse{
			Project: nil,
			Error:   "project ID is required",
		}, nil
	}

	// Get project from service
	project, err := s.service.GetProject(req.Id)
	if err != nil {
		log.Error().Err(err).Str("project_id", req.Id).Msg("Failed to get project")
		return &pb.GetProjectResponse{
			Project: nil,
			Error:   fmt.Sprintf("failed to get project: %v", err),
		}, nil
	}

	// Convert to protobuf and return
	pbProject := convertDomainProjectToPB(project)
	return &pb.GetProjectResponse{
		Project: pbProject,
		Error:   "",
	}, nil
}

// UpdateProject updates an existing project
func (s *GRPCServer) UpdateProject(ctx context.Context, req *pb.UpdateProjectRequest) (*pb.UpdateProjectResponse, error) {
	log.Info().
		Str("project_id", req.Id).
		Str("name", req.Name).
		Msg("gRPC UpdateProject called")

	// Validate request
	if req.Id == "" {
		return &pb.UpdateProjectResponse{
			Project: nil,
			Error:   "project ID is required",
		}, nil
	}

	// Build updates map
	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Status != 0 { // 0 is the default/unspecified value
		updates["status"] = convertPBStatusToDomain(req.Status)
	}
	if req.Tags != nil {
		updates["tags"] = req.Tags
	}
	if req.Config != "" {
		var config map[string]interface{}
		if err := json.Unmarshal([]byte(req.Config), &config); err != nil {
			return &pb.UpdateProjectResponse{
				Project: nil,
				Error:   fmt.Sprintf("invalid config JSON: %v", err),
			}, nil
		}
		updates["config"] = config
	}

	// Update project using service
	project, err := s.service.UpdateProject(req.Id, updates)
	if err != nil {
		log.Error().Err(err).Str("project_id", req.Id).Msg("Failed to update project")
		return &pb.UpdateProjectResponse{
			Project: nil,
			Error:   fmt.Sprintf("failed to update project: %v", err),
		}, nil
	}

	// Convert to protobuf and return
	pbProject := convertDomainProjectToPB(project)
	return &pb.UpdateProjectResponse{
		Project: pbProject,
		Error:   "",
	}, nil
}

// DeleteProject deletes a project
func (s *GRPCServer) DeleteProject(ctx context.Context, req *pb.DeleteProjectRequest) (*pb.DeleteProjectResponse, error) {
	log.Info().
		Str("project_id", req.Id).
		Msg("gRPC DeleteProject called")

	// Validate request
	if req.Id == "" {
		return &pb.DeleteProjectResponse{
			Success: false,
			Error:   "project ID is required",
		}, nil
	}

	// Delete project using service
	err := s.service.DeleteProject(req.Id)
	if err != nil {
		log.Error().Err(err).Str("project_id", req.Id).Msg("Failed to delete project")
		return &pb.DeleteProjectResponse{
			Success: false,
			Error:   fmt.Sprintf("failed to delete project: %v", err),
		}, nil
	}

	return &pb.DeleteProjectResponse{
		Success: true,
		Error:   "",
	}, nil
}

// ListProjects lists projects with pagination
func (s *GRPCServer) ListProjects(ctx context.Context, req *pb.ListProjectsRequest) (*pb.ListProjectsResponse, error) {
	log.Info().
		Int32("page", req.Page).
		Int32("limit", req.Limit).
		Str("search", req.Search).
		Msg("gRPC ListProjects called")

	// Set default pagination values
	page := req.Page
	if page < 1 {
		page = 1
	}
	limit := req.Limit
	if limit <= 0 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}

	// Calculate offset
	offset := (page - 1) * limit

	// Get projects from service
	projects, err := s.service.ListProjects(int(limit), int(offset))
	if err != nil {
		log.Error().Err(err).Msg("Failed to list projects")
		return &pb.ListProjectsResponse{
			Projects: nil,
			Total:    0,
			Error:    fmt.Sprintf("failed to list projects: %v", err),
		}, nil
	}

	// Convert to protobuf
	pbProjects := make([]*pb.Project, len(projects))
	for i, project := range projects {
		pbProjects[i] = convertDomainProjectToPB(project)
	}

	return &pb.ListProjectsResponse{
		Projects: pbProjects,
		Total:    int32(len(pbProjects)),
		Error:    "",
	}, nil
}

// ListUserProjects lists projects for a specific user
func (s *GRPCServer) ListUserProjects(ctx context.Context, req *pb.ListUserProjectsRequest) (*pb.ListUserProjectsResponse, error) {
	log.Info().
		Str("user_id", req.UserId).
		Int32("page", req.Page).
		Int32("limit", req.Limit).
		Msg("gRPC ListUserProjects called")

	// Validate request
	if req.UserId == "" {
		return &pb.ListUserProjectsResponse{
			Projects: nil,
			Total:    0,
			Error:    "user ID is required",
		}, nil
	}

	// Set default pagination values
	page := req.Page
	if page < 1 {
		page = 1
	}
	limit := req.Limit
	if limit <= 0 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}

	// Calculate offset
	offset := (page - 1) * limit

	// Get user projects from service
	projects, err := s.service.ListUserProjects(req.UserId, int(limit), int(offset))
	if err != nil {
		log.Error().Err(err).Str("user_id", req.UserId).Msg("Failed to list user projects")
		return &pb.ListUserProjectsResponse{
			Projects: nil,
			Total:    0,
			Error:    fmt.Sprintf("failed to list user projects: %v", err),
		}, nil
	}

	// Convert to protobuf
	pbProjects := make([]*pb.Project, len(projects))
	for i, project := range projects {
		pbProjects[i] = convertDomainProjectToPB(project)
	}

	return &pb.ListUserProjectsResponse{
		Projects: pbProjects,
		Total:    int32(len(pbProjects)),
		Error:    "",
	}, nil
}

// Helper functions

// convertDomainUserToPB converts a domain User to protobuf User
func convertDomainUserToPB(user *User) *pb.User {
	return &pb.User{
		Id:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		AvatarUrl: user.AvatarURL,
		Roles:     user.Roles,
		CreatedAt: user.CreatedAt.Unix(),
		UpdatedAt: user.UpdatedAt.Unix(),
	}
}

// convertDomainProjectToPB converts a domain Project to protobuf Project
func convertDomainProjectToPB(project *Project) *pb.Project {
	// Convert config map to JSON string
	configJSON := "{}"
	if project.Config != nil {
		if jsonData, err := json.Marshal(project.Config); err == nil {
			configJSON = string(jsonData)
		}
	}
	
	// Convert status string to protobuf enum
	status := pb.ProjectStatus_PROJECT_STATUS_DRAFT
	switch project.Status {
	case "active":
		status = pb.ProjectStatus_PROJECT_STATUS_ACTIVE
	case "completed":
		status = pb.ProjectStatus_PROJECT_STATUS_COMPLETED
	case "archived":
		status = pb.ProjectStatus_PROJECT_STATUS_ARCHIVED
	case "draft":
		status = pb.ProjectStatus_PROJECT_STATUS_DRAFT
	}

	return &pb.Project{
		Id:          project.ID,
		Name:        project.Name,
		Description: project.Description,
		UserId:      project.UserID,
		Status:      status,
		Tags:        project.Tags,
		Config:      configJSON,
		CreatedAt:   project.CreatedAt.Unix(),
		UpdatedAt:   project.UpdatedAt.Unix(),
	}
}

// convertPBStatusToDomain converts protobuf ProjectStatus to domain string
func convertPBStatusToDomain(status pb.ProjectStatus) string {
	switch status {
	case pb.ProjectStatus_PROJECT_STATUS_ACTIVE:
		return "active"
	case pb.ProjectStatus_PROJECT_STATUS_COMPLETED:
		return "completed"
	case pb.ProjectStatus_PROJECT_STATUS_ARCHIVED:
		return "archived"
	default:
		return "draft"
	}
}

// generateID generates a proper UUID
func generateID() string {
	return uuid.New().String()
}
