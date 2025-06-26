package fixtures

import (
	pb "github.com/EliasRanz/ai-code-gen/ai-ui-generator/api/proto/user"
)

// MockUser returns a mock user for testing
func MockUser() *pb.User {
	return &pb.User{
		Id:        "test-user-id",
		Email:     "test@example.com",
		Name:      "Test User",
		AvatarUrl: "https://example.com/avatar.jpg",
		Roles:     []string{"user"},
		CreatedAt: 1672531200, // 2023-01-01T00:00:00Z as unix timestamp
		UpdatedAt: 1672531200, // 2023-01-01T00:00:00Z as unix timestamp
	}
}

// MockProject returns a mock project for testing
func MockProject() *pb.Project {
	return &pb.Project{
		Id:          "test-project-id",
		UserId:      "test-user-id",
		Name:        "Test Project",
		Description: "A test project",
		Status:      pb.ProjectStatus_PROJECT_STATUS_ACTIVE,
		CreatedAt:   1672531200, // 2023-01-01T00:00:00Z as unix timestamp
		UpdatedAt:   1672531200, // 2023-01-01T00:00:00Z as unix timestamp
	}
}
