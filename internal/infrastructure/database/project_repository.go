package database

import (
	"context"
	"fmt"
	"time"

	"github.com/EliasRanz/ai-code-gen/internal/domain/common"
	"github.com/EliasRanz/ai-code-gen/internal/domain/user"
	"gorm.io/gorm"
)

// ProjectModel is the GORM model for a project
type ProjectModel struct {
	ID          common.ProjectID   `gorm:"type:uuid;primary_key"`
	UserID      common.UserID      `gorm:"type:uuid"`
	Name        string             `gorm:"size:255;not null"`
	Description string             `gorm:"type:text"`
	Status      user.ProjectStatus `gorm:"size:50"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (pm *ProjectModel) TableName() string {
	return "projects"
}

// FromProject converts a domain Project to a GORM ProjectModel
func (pm *ProjectModel) FromProject(p user.Project) {
	pm.ID = p.ID
	pm.UserID = p.UserID
	pm.Name = p.Name
	pm.Description = p.Description
	pm.Status = p.Status
	pm.CreatedAt = p.CreatedAt
	pm.UpdatedAt = p.UpdatedAt
}

// ToProject converts a GORM ProjectModel to a domain Project
func (pm *ProjectModel) ToProject() user.Project {
	return user.Project{
		ID:          pm.ID,
		UserID:      pm.UserID,
		Name:        pm.Name,
		Description: pm.Description,
		Status:      pm.Status,
		Timestamps: common.Timestamps{
			CreatedAt: pm.CreatedAt,
			UpdatedAt: pm.UpdatedAt,
		},
	}
}

// PostgreSQLProjectRepository implements the user.ProjectRepository interface using GORM
type PostgreSQLProjectRepository struct {
	db *gorm.DB
}

// NewPostgreSQLProjectRepository creates a new PostgreSQL project repository
func NewPostgreSQLProjectRepository(db *gorm.DB) (user.ProjectRepository, error) {
	// Auto-migrate the schema
	if err := db.AutoMigrate(&ProjectModel{}); err != nil {
		return nil, fmt.Errorf("failed to migrate project schema: %w", err)
	}
	return &PostgreSQLProjectRepository{db: db}, nil
}

// Create creates a new project
func (r *PostgreSQLProjectRepository) Create(ctx context.Context, project user.Project) error {
	projectModel := &ProjectModel{}
	projectModel.FromProject(project)

	if err := r.db.WithContext(ctx).Create(projectModel).Error; err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	return nil
}

// Update updates a project
func (r *PostgreSQLProjectRepository) Update(ctx context.Context, project user.Project) error {
	projectModel := &ProjectModel{}
	projectModel.FromProject(project)

	result := r.db.WithContext(ctx).Model(&ProjectModel{}).Where("id = ?", project.ID).Updates(projectModel)
	if result.Error != nil {
		return fmt.Errorf("failed to update project: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("project not found")
	}

	return nil
}

// Delete deletes a project
func (r *PostgreSQLProjectRepository) Delete(ctx context.Context, id common.ProjectID) error {
	result := r.db.WithContext(ctx).Delete(&ProjectModel{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete project: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("project not found")
	}

	return nil
}

// List lists projects with pagination
func (r *PostgreSQLProjectRepository) List(ctx context.Context, params common.PaginationParams, search string, status user.ProjectStatus) ([]user.Project, error) {
	var projectModels []ProjectModel
	query := r.db.WithContext(ctx).Order("created_at DESC").Limit(int(params.Limit)).Offset(int(params.Offset()))

	if search != "" {
		query = query.Where("name ILIKE ?", "%"+search+"%")
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Find(&projectModels).Error; err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	projects := make([]user.Project, len(projectModels))
	for i, projectModel := range projectModels {
		projects[i] = projectModel.ToProject()
	}
	return projects, nil
}

// ListByUserID lists projects by user ID with pagination
func (r *PostgreSQLProjectRepository) ListByUserID(ctx context.Context, userID common.UserID, params common.PaginationParams) ([]user.Project, error) {
	var projectModels []ProjectModel
	query := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Limit(int(params.Limit)).Offset(int(params.Offset()))

	if err := query.Find(&projectModels).Error; err != nil {
		return nil, fmt.Errorf("failed to list projects by user: %w", err)
	}

	projects := make([]user.Project, len(projectModels))
	for i, projectModel := range projectModels {
		projects[i] = projectModel.ToProject()
	}
	return projects, nil
}

// GetByID retrieves a project by ID
func (r *PostgreSQLProjectRepository) GetByID(ctx context.Context, id common.ProjectID) (user.Project, error) {
	var projectModel ProjectModel
	if err := r.db.WithContext(ctx).First(&projectModel, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return user.Project{}, fmt.Errorf("project not found")
		}
		return user.Project{}, fmt.Errorf("failed to get project by id: %w", err)
	}
	return projectModel.ToProject(), nil
}
