package user

import (
	"fmt"
	"gorm.io/gorm"
)

// ProjectRepository defines the interface for project data access
type ProjectRepository interface {
	GetByID(id string) (*Project, error)
	Create(project *Project) error
	Update(id string, updates map[string]interface{}) (*Project, error)
	Delete(id string) error
	List(limit, offset int) ([]*Project, error)
	ListByUserID(userID string, limit, offset int) ([]*Project, error)
}

// GormProjectRepository implements ProjectRepository using GORM
type GormProjectRepository struct {
	db *gorm.DB
}

// NewGormProjectRepository creates a new GORM project repository
func NewGormProjectRepository(db *gorm.DB) *GormProjectRepository {
	return &GormProjectRepository{db: db}
}

// GetByID retrieves a project by ID
func (r *GormProjectRepository) GetByID(id string) (*Project, error) {
	var projectModel ProjectModel
	if err := r.db.First(&projectModel, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("project not found")
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	return projectModel.ToProject(), nil
}

// Create creates a new project
func (r *GormProjectRepository) Create(project *Project) error {
	projectModel := &ProjectModel{}
	if err := projectModel.FromProject(project); err != nil {
		return fmt.Errorf("failed to convert project: %w", err)
	}
	
	if err := r.db.Create(projectModel).Error; err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}
	
	// Update the original project with generated values
	*project = *projectModel.ToProject()
	return nil
}

// Update updates a project
func (r *GormProjectRepository) Update(id string, updates map[string]interface{}) (*Project, error) {
	if len(updates) == 0 {
		return r.GetByID(id) // No updates to perform
	}

	result := r.db.Model(&ProjectModel{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to update project: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("project not found")
	}
	
	return r.GetByID(id)
}

// Delete deletes a project
func (r *GormProjectRepository) Delete(id string) error {
	result := r.db.Delete(&ProjectModel{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete project: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return fmt.Errorf("project not found")
	}
	
	return nil
}

// List lists projects with pagination
func (r *GormProjectRepository) List(limit, offset int) ([]*Project, error) {
	var projectModels []ProjectModel
	if err := r.db.Order("created_at DESC").Limit(limit).Offset(offset).Find(&projectModels).Error; err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}
	
	projects := make([]*Project, len(projectModels))
	for i, projectModel := range projectModels {
		projects[i] = projectModel.ToProject()
	}
	
	return projects, nil
}

// ListByUserID lists projects for a specific user with pagination
func (r *GormProjectRepository) ListByUserID(userID string, limit, offset int) ([]*Project, error) {
	var projectModels []ProjectModel
	if err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Limit(limit).Offset(offset).Find(&projectModels).Error; err != nil {
		return nil, fmt.Errorf("failed to list projects by user ID: %w", err)
	}
	
	projects := make([]*Project, len(projectModels))
	for i, projectModel := range projectModels {
		projects[i] = projectModel.ToProject()
	}
	
	return projects, nil
}
