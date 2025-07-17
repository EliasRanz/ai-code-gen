package user

import (
	"context"
	"fmt"
	"time"

	"github.com/EliasRanz/ai-code-gen/internal/utilities"
	"gorm.io/gorm"
)

// ProjectModel is the GORM model for a project
type ProjectModel struct {
	ID          utilities.ProjectID `gorm:"type:uuid;primary_key"`
	UserID      utilities.UserID    `gorm:"type:uuid"`
	Name        string              `gorm:"size:255;not null"`
	Description string              `gorm:"type:text"`
	Status      ProjectStatus       `gorm:"size:50"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (pm *ProjectModel) TableName() string {
	return "projects"
}

// FromProject converts a domain Project to a GORM ProjectModel
func (pm *ProjectModel) FromProject(p Project) {
	pm.ID = p.ID
	pm.UserID = p.UserID
	pm.Name = p.Name
	pm.Description = p.Description
	pm.Status = p.Status
	pm.CreatedAt = p.CreatedAt
	pm.UpdatedAt = p.UpdatedAt
}

// ToProject converts a GORM ProjectModel to a domain Project
func (pm *ProjectModel) ToProject() Project {
	return Project{
		ID:          pm.ID,
		UserID:      pm.UserID,
		Name:        pm.Name,
		Description: pm.Description,
		Status:      pm.Status,
		Timestamps: utilities.Timestamps{
			CreatedAt: pm.CreatedAt,
			UpdatedAt: pm.UpdatedAt,
		},
	}
}

// PostgreSQLProjectRepository implements ProjectRepository interface using GORM and template method pattern
type PostgreSQLProjectRepository struct {
	*utilities.BaseRepository
	db *gorm.DB
}

// NewPostgreSQLProjectRepository creates a new PostgreSQL project repository
func NewPostgreSQLProjectRepository(db *gorm.DB) (ProjectRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection is required")
	}

	// Auto-migrate the schema
	if err := db.AutoMigrate(&ProjectModel{}); err != nil {
		return nil, fmt.Errorf("failed to migrate project schema: %w", err)
	}

	logger := &utilities.ZerologAdapter{}
	metrics := &utilities.NoOpMetrics{}

	return &PostgreSQLProjectRepository{
		BaseRepository: utilities.NewBaseRepository(nil, logger, metrics),
		db:             db,
	}, nil
}

// Create creates a new project - implements ProjectRepository interface
func (r *PostgreSQLProjectRepository) Create(ctx context.Context, project Project) error {
	if err := r.validateProject(project); err != nil {
		return fmt.Errorf("project validation failed: %w", err)
	}

	projectModel := &ProjectModel{}
	projectModel.FromProject(project)

	if err := r.db.WithContext(ctx).Create(projectModel).Error; err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	return nil
}

// GetByID retrieves a project by ID - implements ProjectRepository interface
func (r *PostgreSQLProjectRepository) GetByID(ctx context.Context, id utilities.ProjectID) (Project, error) {
	var projectModel ProjectModel

	if err := r.db.WithContext(ctx).First(&projectModel, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return Project{}, fmt.Errorf("project not found")
		}
		return Project{}, fmt.Errorf("failed to get project by id: %w", err)
	}

	return projectModel.ToProject(), nil
}

// Update updates a project - implements ProjectRepository interface
func (r *PostgreSQLProjectRepository) Update(ctx context.Context, project Project) error {
	if err := r.validateProject(project); err != nil {
		return fmt.Errorf("project validation failed: %w", err)
	}

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

// Delete deletes a project - implements ProjectRepository interface
func (r *PostgreSQLProjectRepository) Delete(ctx context.Context, id utilities.ProjectID) error {
	result := r.db.WithContext(ctx).Delete(&ProjectModel{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete project: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("project not found")
	}

	return nil
}

// List lists projects with pagination - implements ProjectRepository interface
func (r *PostgreSQLProjectRepository) List(ctx context.Context, params utilities.PaginationParams, search string, status ProjectStatus) ([]Project, error) {
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

	projects := make([]Project, len(projectModels))
	for i, projectModel := range projectModels {
		projects[i] = projectModel.ToProject()
	}
	return projects, nil
}

// ListByUserID lists projects by user ID with pagination - implements ProjectRepository interface
func (r *PostgreSQLProjectRepository) ListByUserID(ctx context.Context, userID utilities.UserID, params utilities.PaginationParams) ([]Project, error) {
	var projectModels []ProjectModel
	query := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Limit(int(params.Limit)).Offset(int(params.Offset()))

	if err := query.Find(&projectModels).Error; err != nil {
		return nil, fmt.Errorf("failed to list projects by user: %w", err)
	}

	projects := make([]Project, len(projectModels))
	for i, projectModel := range projectModels {
		projects[i] = projectModel.ToProject()
	}
	return projects, nil
}

// === Additional Repository Interface Pattern Methods ===

// CreateEntity creates a new entity - implements utilities.Repository interface
func (r *PostgreSQLProjectRepository) CreateEntity(ctx context.Context, entity interface{}) error {
	project, ok := entity.(Project)
	if !ok {
		return fmt.Errorf("invalid entity type, expected Project")
	}
	return r.Create(ctx, project)
}

// GetByIDEntity retrieves an entity by ID - implements utilities.Repository interface
func (r *PostgreSQLProjectRepository) GetByIDEntity(ctx context.Context, id string, entity interface{}) error {
	projectPtr, ok := entity.(*Project)
	if !ok {
		return fmt.Errorf("invalid entity type, expected *Project")
	}

	projectID := utilities.ProjectID(id)
	project, err := r.GetByID(ctx, projectID)
	if err != nil {
		return err
	}

	*projectPtr = project
	return nil
}

// UpdateEntity updates an entity - implements utilities.Repository interface
func (r *PostgreSQLProjectRepository) UpdateEntity(ctx context.Context, entity interface{}) error {
	project, ok := entity.(Project)
	if !ok {
		return fmt.Errorf("invalid entity type, expected Project")
	}
	return r.Update(ctx, project)
}

// DeleteEntity deletes an entity - implements utilities.Repository interface
func (r *PostgreSQLProjectRepository) DeleteEntity(ctx context.Context, id string) error {
	projectID := utilities.ProjectID(id)
	return r.Delete(ctx, projectID)
}

// ListEntities lists entities with filtering - implements utilities.Repository interface
func (r *PostgreSQLProjectRepository) ListEntities(ctx context.Context, filter utilities.QueryFilter, entities interface{}) error {
	projectsPtr, ok := entities.(*[]Project)
	if !ok {
		return fmt.Errorf("invalid entity type, expected *[]Project")
	}

	var projectModels []ProjectModel
	query := r.db.WithContext(ctx).Order("created_at DESC")

	// Apply filters
	if filter.Search != "" {
		query = query.Where("name ILIKE ?", "%"+filter.Search+"%")
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	if err := query.Find(&projectModels).Error; err != nil {
		return fmt.Errorf("failed to list projects: %w", err)
	}

	projects := make([]Project, len(projectModels))
	for i, projectModel := range projectModels {
		projects[i] = projectModel.ToProject()
	}

	*projectsPtr = projects
	return nil
}

// CountEntities counts entities with filtering - implements utilities.Repository interface
func (r *PostgreSQLProjectRepository) CountEntities(ctx context.Context, filter utilities.QueryFilter) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&ProjectModel{})

	// Apply filters
	if filter.Search != "" {
		query = query.Where("name ILIKE ?", "%"+filter.Search+"%")
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count projects: %w", err)
	}

	return count, nil
}

// BeginTx begins a transaction
func (r *PostgreSQLProjectRepository) BeginTx(ctx context.Context) (utilities.Transaction, error) {
	return nil, fmt.Errorf("transaction not implemented for GORM repository")
}

// HealthCheck checks repository health
func (r *PostgreSQLProjectRepository) HealthCheck(ctx context.Context) error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}

// Close closes the repository
func (r *PostgreSQLProjectRepository) Close() error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	return sqlDB.Close()
}

// GetByUserID retrieves projects by user ID - specific to ProjectRepository
func (r *PostgreSQLProjectRepository) GetByUserID(ctx context.Context, userID utilities.UserID) (interface{}, error) {
	var projectModels []ProjectModel

	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&projectModels).Error; err != nil {
		return nil, fmt.Errorf("failed to list projects by user: %w", err)
	}

	projects := make([]Project, len(projectModels))
	for i, projectModel := range projectModels {
		projects[i] = projectModel.ToProject()
	}

	return projects, nil
}

// GetByStatus retrieves projects by status - specific to ProjectRepository
func (r *PostgreSQLProjectRepository) GetByStatus(ctx context.Context, status string) (interface{}, error) {
	var projectModels []ProjectModel

	if err := r.db.WithContext(ctx).Where("status = ?", status).Order("created_at DESC").Find(&projectModels).Error; err != nil {
		return nil, fmt.Errorf("failed to list projects by status: %w", err)
	}

	projects := make([]Project, len(projectModels))
	for i, projectModel := range projectModels {
		projects[i] = projectModel.ToProject()
	}

	return projects, nil
}

// validateProject validates project data
func (r *PostgreSQLProjectRepository) validateProject(project Project) error {
	if project.ID == "" {
		return fmt.Errorf("project ID is required")
	}
	if project.UserID == "" {
		return fmt.Errorf("user ID is required")
	}
	if project.Name == "" {
		return fmt.Errorf("project name is required")
	}
	if len(project.Name) > 255 {
		return fmt.Errorf("project name too long (max 255 characters)")
	}
	return nil
}

// BeforeOperation hook for project-specific pre-operation logic
func (r *PostgreSQLProjectRepository) BeforeOperation(ctx context.Context, operation utilities.OperationType) error {
	// Call base implementation first
	if err := r.BaseRepository.BeforeOperation(ctx, operation); err != nil {
		return err
	}

	// Project-specific pre-operation logic
	return nil
}

// OnError hook for project-specific error handling
func (r *PostgreSQLProjectRepository) OnError(ctx context.Context, operation utilities.OperationType, err error) error {
	// Call base implementation first
	r.BaseRepository.OnError(ctx, operation, err)

	// Transform database-specific errors to domain errors
	return r.transformError(err)
}

// transformError transforms database errors to domain errors
func (r *PostgreSQLProjectRepository) transformError(err error) error {
	if err == nil {
		return nil
	}

	errStr := err.Error()
	switch {
	case errStr == gorm.ErrRecordNotFound.Error():
		return fmt.Errorf("project not found")
	case contains(errStr, "duplicate"):
		return fmt.Errorf("project already exists")
	case contains(errStr, "foreign key"):
		return fmt.Errorf("invalid user reference")
	default:
		return err
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && indexOf(s, substr) >= 0
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
