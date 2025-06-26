package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	pb "github.com/EliasRanz/ai-code-gen/ai-ui-generator/api/proto/user"
)

// CreateProjectHandler creates a new project
func CreateProjectHandler(c *gin.Context) {
	log.Info().Msg("Create project request")

	var req struct {
		Name        string   `json:"name" binding:"required"`
		Description string   `json:"description"`
		UserID      string   `json:"user_id" binding:"required"`
		Tags        []string `json:"tags"`
		Config      string   `json:"config"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if grpcClient == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gRPC client not initialized"})
		return
	}

	grpcReq := &pb.CreateProjectRequest{
		Name:        req.Name,
		Description: req.Description,
		UserId:      req.UserID,
		Tags:        req.Tags,
		Config:      req.Config,
	}

	resp, err := grpcClient.CreateProject(grpcReq)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create project via gRPC")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create project"})
		return
	}

	if resp.Error != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": resp.Error})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"project": createProjectResponseJSON(resp.Project),
	})
}

// GetProjectHandler retrieves a project by ID
func GetProjectHandler(c *gin.Context) {
	projectID := c.Param("id")
	log.Info().Str("project_id", projectID).Msg("Get project request")

	if grpcClient == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gRPC client not initialized"})
		return
	}
	resp, err := grpcClient.GetProject(projectID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get project via gRPC")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get project"})
		return
	}
	if resp.Error != "" {
		if resp.Error == "Project not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": resp.Error})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": resp.Error})
		return
	}
	if resp.Project == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"project": gin.H{
			"id":          resp.Project.Id,
			"name":        resp.Project.Name,
			"description": resp.Project.Description,
			"user_id":     resp.Project.UserId,
			"status":      resp.Project.Status.String(),
			"tags":        resp.Project.Tags,
			"config":      resp.Project.Config,
			"created_at":  resp.Project.CreatedAt,
			"updated_at":  resp.Project.UpdatedAt,
		},
	})
}

// UpdateProjectHandler updates a project
func UpdateProjectHandler(c *gin.Context) {
	projectID := c.Param("id")
	log.Info().Str("project_id", projectID).Msg("Update project request")

	var req struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Status      string   `json:"status"`
		Tags        []string `json:"tags"`
		Config      string   `json:"config"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if grpcClient == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gRPC client not initialized"})
		return
	}
	// Convert status string to enum
	statusEnum := convertStatusStringToEnum(req.Status)
	grpcReq := &pb.UpdateProjectRequest{
		Id:          projectID,
		Name:        req.Name,
		Description: req.Description,
		Status:      statusEnum,
		Tags:        req.Tags,
		Config:      req.Config,
	}
	resp, err := grpcClient.UpdateProject(grpcReq)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update project via gRPC")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update project"})
		return
	}
	if resp.Error != "" {
		if resp.Error == "Project not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": resp.Error})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": resp.Error})
		return
	}
	if resp.Project == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"project": gin.H{
			"id":          resp.Project.Id,
			"name":        resp.Project.Name,
			"description": resp.Project.Description,
			"user_id":     resp.Project.UserId,
			"status":      resp.Project.Status.String(),
			"tags":        resp.Project.Tags,
			"config":      resp.Project.Config,
			"created_at":  resp.Project.CreatedAt,
			"updated_at":  resp.Project.UpdatedAt,
		},
	})
}

// DeleteProjectHandler deletes a project
func DeleteProjectHandler(c *gin.Context) {
	projectID := c.Param("id")
	log.Info().Str("project_id", projectID).Msg("Delete project request")

	if grpcClient == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gRPC client not initialized"})
		return
	}
	resp, err := grpcClient.DeleteProject(projectID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete project via gRPC")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete project"})
		return
	}
	if resp.Error != "" {
		if resp.Error == "Project not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": resp.Error})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": resp.Error})
		return
	}
	if !resp.Success {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Project deleted successfully"})
}

// ListProjectsHandler lists projects with pagination
func ListProjectsHandler(c *gin.Context) {
	log.Info().Msg("List projects request")

	page := ParseInt32(c.Query("page"), 1)
	limit := ParseInt32(c.Query("limit"), 10)
	search := c.Query("search")
	statusEnum := convertStatusStringToEnum(c.Query("status"))

	if grpcClient == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gRPC client not initialized"})
		return
	}
	resp, err := grpcClient.ListProjects(page, limit, search, statusEnum)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list projects via gRPC")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list projects"})
		return
	}
	if resp.Error != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": resp.Error})
		return
	}
	projects := make([]gin.H, len(resp.Projects))
	for i, p := range resp.Projects {
		projects[i] = gin.H{
			"id":          p.Id,
			"name":        p.Name,
			"description": p.Description,
			"user_id":     p.UserId,
			"status":      p.Status.String(),
			"tags":        p.Tags,
			"config":      p.Config,
			"created_at":  p.CreatedAt,
			"updated_at":  p.UpdatedAt,
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"projects": projects,
		"total":    resp.Total,
		"page":     page,
		"limit":    limit,
	})
}

// ListUserProjectsHandler lists projects for a specific user
func ListUserProjectsHandler(c *gin.Context) {
	userID := c.Param("user_id")
	log.Info().Str("user_id", userID).Msg("List user projects request")

	page := ParseInt32(c.Query("page"), 1)
	limit := ParseInt32(c.Query("limit"), 10)
	statusEnum := convertStatusStringToEnum(c.Query("status"))

	if grpcClient == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gRPC client not initialized"})
		return
	}
	resp, err := grpcClient.ListUserProjects(userID, page, limit, statusEnum)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list user projects via gRPC")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list user projects"})
		return
	}
	if resp.Error != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": resp.Error})
		return
	}
	projects := make([]gin.H, len(resp.Projects))
	for i, p := range resp.Projects {
		projects[i] = gin.H{
			"id":          p.Id,
			"name":        p.Name,
			"description": p.Description,
			"user_id":     p.UserId,
			"status":      p.Status.String(),
			"tags":        p.Tags,
			"config":      p.Config,
			"created_at":  p.CreatedAt,
			"updated_at":  p.UpdatedAt,
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"projects": projects,
		"total":    resp.Total,
		"page":     page,
		"limit":    limit,
	})
}