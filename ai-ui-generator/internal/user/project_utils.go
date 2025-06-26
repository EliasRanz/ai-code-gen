package user

import (
	"github.com/gin-gonic/gin"
	pb "github.com/EliasRanz/ai-code-gen/ai-ui-generator/api/proto/user"
)

// convertStatusStringToEnum converts a status string to protobuf enum
func convertStatusStringToEnum(statusStr string) pb.ProjectStatus {
	if statusStr == "" {
		return pb.ProjectStatus_PROJECT_STATUS_DRAFT
	}
	
	switch statusStr {
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

// createProjectResponseJSON creates a standardized JSON response for a project
func createProjectResponseJSON(project *pb.Project) gin.H {
	return gin.H{
		"id":          project.Id,
		"name":        project.Name,
		"description": project.Description,
		"user_id":     project.UserId,
		"status":      project.Status.String(),
		"tags":        project.Tags,
		"config":      project.Config,
		"created_at":  project.CreatedAt,
		"updated_at":  project.UpdatedAt,
	}
}

// createProjectsResponseJSON creates a standardized JSON response for multiple projects
func createProjectsResponseJSON(projects []*pb.Project) []gin.H {
	result := make([]gin.H, len(projects))
	for i, p := range projects {
		result[i] = createProjectResponseJSON(p)
	}
	return result
}
