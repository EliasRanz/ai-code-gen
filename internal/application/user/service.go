package user

import (
	"context"

	"github.com/EliasRanz/ai-code-gen/internal/domain/common"
	"github.com/EliasRanz/ai-code-gen/internal/domain/user"
)

// UpdateUserDTO defines the data transfer object for updating a user.
type UpdateUserDTO struct {
	Username  *string
	AvatarURL *string
	Roles     []string
}

// UpdateProjectDTO defines the data transfer object for updating a project.
type UpdateProjectDTO struct {
	Name        *string
	Description *string
}

// Service provides user and project business logic.
type Service struct {
	userRepo    user.Repository
	projectRepo user.ProjectRepository
}

// NewService creates a new user application service.
func NewService(userRepo user.Repository, projectRepo user.ProjectRepository) *Service {
	return &Service{
		userRepo:    userRepo,
		projectRepo: projectRepo,
	}
}

// GetUser retrieves a user by ID.
func (s *Service) GetUser(ctx context.Context, id common.UserID) (user.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

// CreateUser creates a new user.
func (s *Service) CreateUser(ctx context.Context, u user.User) error {
	return s.userRepo.Create(ctx, u)
}

// UpdateUser updates a user.
func (s *Service) UpdateUser(ctx context.Context, id common.UserID, dto UpdateUserDTO) (*user.User, error) {
	u, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if dto.Username != nil {
		u.Username = *dto.Username
	}
	if dto.AvatarURL != nil {
		u.AvatarURL = *dto.AvatarURL
	}
	if dto.Roles != nil {
		u.Roles = dto.Roles
	}

	err = s.userRepo.Update(ctx, u)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// DeleteUser deletes a user.
func (s *Service) DeleteUser(ctx context.Context, id common.UserID) error {
	return s.userRepo.Delete(ctx, id)
}

// ListUsers lists all users.
func (s *Service) ListUsers(ctx context.Context) ([]user.User, error) {
	// Note: Pagination and search parameters are not yet implemented here.
	return s.userRepo.List(ctx, common.PaginationParams{}, "")
}

// GetProject retrieves a project by ID.
func (s *Service) GetProject(ctx context.Context, id common.ProjectID) (user.Project, error) {
	return s.projectRepo.GetByID(ctx, id)
}

// CreateProject creates a new project.
func (s *Service) CreateProject(ctx context.Context, p user.Project) error {
	return s.projectRepo.Create(ctx, p)
}

// UpdateProject updates a project.
func (s *Service) UpdateProject(ctx context.Context, id common.ProjectID, dto UpdateProjectDTO) (*user.Project, error) {
	p, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if dto.Name != nil {
		p.Name = *dto.Name
	}
	if dto.Description != nil {
		p.Description = *dto.Description
	}

	err = s.projectRepo.Update(ctx, p)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// DeleteProject deletes a project.
func (s *Service) DeleteProject(ctx context.Context, id common.ProjectID) error {
	return s.projectRepo.Delete(ctx, id)
}

// ListUserProjects lists projects for a specific user.
func (s *Service) ListUserProjects(ctx context.Context, userID common.UserID) ([]user.Project, error) {
	// Note: Pagination parameters are not yet implemented here.
	return s.projectRepo.ListByUserID(ctx, userID, common.PaginationParams{})
}
