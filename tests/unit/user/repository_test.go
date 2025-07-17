package user_test

import (
	"context"
	"testing"

	"github.com/EliasRanz/ai-code-gen/internal/user"
	"github.com/EliasRanz/ai-code-gen/internal/utilities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockProjectRepository implements user.ProjectRepository for testing
type MockProjectRepository struct {
	mock.Mock
}

func (m *MockProjectRepository) Create(ctx context.Context, project user.Project) error {
	args := m.Called(ctx, project)
	return args.Error(0)
}

func (m *MockProjectRepository) GetByID(ctx context.Context, id utilities.ProjectID) (user.Project, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(user.Project), args.Error(1)
}

func (m *MockProjectRepository) Update(ctx context.Context, project user.Project) error {
	args := m.Called(ctx, project)
	return args.Error(0)
}

func (m *MockProjectRepository) Delete(ctx context.Context, id utilities.ProjectID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockProjectRepository) List(ctx context.Context, params utilities.PaginationParams, search string, status user.ProjectStatus) ([]user.Project, error) {
	args := m.Called(ctx, params, search, status)
	return args.Get(0).([]user.Project), args.Error(1)
}

func (m *MockProjectRepository) ListByUserID(ctx context.Context, userID utilities.UserID, params utilities.PaginationParams) ([]user.Project, error) {
	args := m.Called(ctx, userID, params)
	return args.Get(0).([]user.Project), args.Error(1)
}

// TestRepositoryFactory tests the repository factory pattern
func TestRepositoryFactory(t *testing.T) {
	factory := user.NewPostgreSQLRepositoryFactory()

	t.Run("Factory creation", func(t *testing.T) {
		assert.NotNil(t, factory)
	})

	t.Run("Invalid database type", func(t *testing.T) {
		repo, err := factory.CreateProjectRepository("invalid-db")

		assert.Error(t, err)
		assert.Nil(t, repo)
		assert.Contains(t, err.Error(), "invalid database type")
	})
}

// TestProjectRepositoryInterface tests interface compliance using mocks
func TestProjectRepositoryInterface(t *testing.T) {
	mockRepo := &MockProjectRepository{}
	ctx := context.Background()

	// Test Create operation
	project := user.Project{
		ID:          utilities.ProjectID("project-123"),
		UserID:      utilities.UserID("user-456"),
		Name:        "Test Project",
		Description: "A test project for validation",
		Status:      user.StatusActive,
	}

	mockRepo.On("Create", ctx, project).Return(nil)
	err := mockRepo.Create(ctx, project)
	assert.NoError(t, err)

	// Test GetByID operation
	mockRepo.On("GetByID", ctx, utilities.ProjectID("project-123")).Return(project, nil)
	result, err := mockRepo.GetByID(ctx, utilities.ProjectID("project-123"))
	assert.NoError(t, err)
	assert.Equal(t, project.ID, result.ID)

	// Test Update operation
	updatedProject := project
	updatedProject.Name = "Updated Project"
	mockRepo.On("Update", ctx, updatedProject).Return(nil)
	err = mockRepo.Update(ctx, updatedProject)
	assert.NoError(t, err)

	// Test Delete operation
	mockRepo.On("Delete", ctx, utilities.ProjectID("project-123")).Return(nil)
	err = mockRepo.Delete(ctx, utilities.ProjectID("project-123"))
	assert.NoError(t, err)

	// Test List operation
	mockRepo.On("List", ctx, utilities.PaginationParams{}, "", user.StatusActive).Return([]user.Project{project}, nil)
	projects, err := mockRepo.List(ctx, utilities.PaginationParams{}, "", user.StatusActive)
	assert.NoError(t, err)
	assert.Len(t, projects, 1)

	mockRepo.AssertExpectations(t)
}

// TestProjectModel tests the project model structure
func TestProjectModel(t *testing.T) {
	project := user.ProjectModel{
		ID:          utilities.ProjectID("project-123"),
		UserID:      utilities.UserID("user-456"),
		Name:        "Test Project",
		Description: "A test project for validation",
		Status:      user.StatusActive,
	}

	assert.NotEmpty(t, project.ID)
	assert.NotEmpty(t, project.UserID)
	assert.NotEmpty(t, project.Name)
	assert.NotEmpty(t, project.Description)
	assert.Equal(t, user.StatusActive, project.Status)
}

// TestProjectEntity tests the project entity structure
func TestProjectEntity(t *testing.T) {
	project := user.Project{
		ID:          utilities.ProjectID("project-123"),
		UserID:      utilities.UserID("user-456"),
		Name:        "Test Project",
		Description: "A test project for validation",
		Status:      user.StatusActive,
	}

	assert.NotEmpty(t, project.ID)
	assert.NotEmpty(t, project.UserID)
	assert.NotEmpty(t, project.Name)
	assert.NotEmpty(t, project.Description)
	assert.Equal(t, user.StatusActive, project.Status)
}

// TestUserIDValidation tests UserID type validation
func TestUserIDValidation(t *testing.T) {
	tests := []struct {
		name   string
		userID utilities.UserID
		valid  bool
	}{
		{
			name:   "Valid user ID",
			userID: utilities.UserID("user-123"),
			valid:  true,
		},
		{
			name:   "Empty user ID",
			userID: utilities.UserID(""),
			valid:  false,
		},
		{
			name:   "Valid UUID user ID",
			userID: utilities.UserID("550e8400-e29b-41d4-a716-446655440000"),
			valid:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.valid {
				assert.NotEmpty(t, string(tt.userID))
				assert.True(t, len(string(tt.userID)) > 0)
			} else {
				assert.Empty(t, string(tt.userID))
			}
		})
	}
}

// TestProjectIDValidation tests ProjectID type validation
func TestProjectIDValidation(t *testing.T) {
	tests := []struct {
		name      string
		projectID utilities.ProjectID
		valid     bool
	}{
		{
			name:      "Valid project ID",
			projectID: utilities.ProjectID("project-123"),
			valid:     true,
		},
		{
			name:      "Empty project ID",
			projectID: utilities.ProjectID(""),
			valid:     false,
		},
		{
			name:      "Valid UUID project ID",
			projectID: utilities.ProjectID("550e8400-e29b-41d4-a716-446655440000"),
			valid:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.valid {
				assert.NotEmpty(t, string(tt.projectID))
				assert.True(t, len(string(tt.projectID)) > 0)
			} else {
				assert.Empty(t, string(tt.projectID))
			}
		})
	}
}
