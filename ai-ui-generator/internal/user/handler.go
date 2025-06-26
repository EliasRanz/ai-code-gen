package user

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	pb "github.com/EliasRanz/ai-code-gen/ai-ui-generator/api/proto/user"
)

// UserGRPCClient defines the interface for gRPC client methods used by handlers
// This allows for mocking in tests
//go:generate mockgen -destination=mock_grpc_client.go -package=user . UserGRPCClient

// UserGRPCClient is an interface for gRPC client
// Only methods used by handlers are included
// (expand as needed for more handlers)
type UserGRPCClient interface {
	UpdateUser(req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error)
	CreateUser(req *pb.CreateUserRequest) (*pb.CreateUserResponse, error)
	GetUser(userID string) (*pb.GetUserResponse, error)
	ListUsers(page, limit int32, search string) (*pb.ListUsersResponse, error)
	CreateProject(req *pb.CreateProjectRequest) (*pb.CreateProjectResponse, error)
	DeleteUser(userID string) (*pb.DeleteUserResponse, error)
	GetProject(projectID string) (*pb.GetProjectResponse, error)
	UpdateProject(req *pb.UpdateProjectRequest) (*pb.UpdateProjectResponse, error)
	DeleteProject(projectID string) (*pb.DeleteProjectResponse, error)
	ListProjects(page, limit int32, search string, status pb.ProjectStatus) (*pb.ListProjectsResponse, error)
	ListUserProjects(userID string, page, limit int32, status pb.ProjectStatus) (*pb.ListUserProjectsResponse, error)
	// ...add more as needed
}

// Global gRPC client - will be set by the main function
var grpcClient UserGRPCClient

// SetGRPCClient sets the global gRPC client for handlers
func SetGRPCClient(client UserGRPCClient) {
	grpcClient = client
}

// User HTTP handlers that call gRPC service

// CreateUserHandler handles user creation
func CreateUserHandler(c *gin.Context) {
	log.Info().Msg("Create user request")

	var req struct {
		Email     string   `json:"email" binding:"required"`
		Name      string   `json:"name" binding:"required"`
		AvatarURL string   `json:"avatar_url"`
		Roles     []string `json:"roles"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if grpcClient == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gRPC client not initialized"})
		return
	}

	grpcReq := &pb.CreateUserRequest{
		Email:     req.Email,
		Name:      req.Name,
		AvatarUrl: req.AvatarURL,
		Roles:     req.Roles,
	}

	resp, err := grpcClient.CreateUser(grpcReq)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create user via gRPC")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	if resp.Error != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": resp.Error})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"user": gin.H{
			"id":         resp.User.Id,
			"email":      resp.User.Email,
			"name":       resp.User.Name,
			"avatar_url": resp.User.AvatarUrl,
			"roles":      resp.User.Roles,
			"created_at": resp.User.CreatedAt,
			"updated_at": resp.User.UpdatedAt,
		},
	})
}

// GetUserHandler retrieves a user by ID
func GetUserHandler(c *gin.Context) {
	userID := c.Param("id")
	log.Info().Str("user_id", userID).Msg("Get user request")

	if grpcClient == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gRPC client not initialized"})
		return
	}

	resp, err := grpcClient.GetUser(userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user via gRPC")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}

	if resp.Error != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": resp.Error})
		return
	}

	if resp.User == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":         resp.User.Id,
			"email":      resp.User.Email,
			"name":       resp.User.Name,
			"avatar_url": resp.User.AvatarUrl,
			"roles":      resp.User.Roles,
			"created_at": resp.User.CreatedAt,
			"updated_at": resp.User.UpdatedAt,
		},
	})
}

// UpdateUserHandler updates a user
func UpdateUserHandler(c *gin.Context) {
	userID := c.Param("id")
	log.Info().Str("user_id", userID).Msg("Update user request")

	var req struct {
		Name      string   `json:"name"`
		AvatarURL string   `json:"avatar_url"`
		Roles     []string `json:"roles"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if grpcClient == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gRPC client not initialized"})
		return
	}
	grpcReq := &pb.UpdateUserRequest{
		Id:        userID,
		Name:      req.Name,
		AvatarUrl: req.AvatarURL,
		Roles:     req.Roles,
	}
	resp, err := grpcClient.UpdateUser(grpcReq)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update user via gRPC")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}
	if resp.Error != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": resp.Error})
		return
	}
	if resp.User == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":         resp.User.Id,
			"email":      resp.User.Email,
			"name":       resp.User.Name,
			"avatar_url": resp.User.AvatarUrl,
			"roles":      resp.User.Roles,
			"created_at": resp.User.CreatedAt,
			"updated_at": resp.User.UpdatedAt,
		},
	})
}

// DeleteUserHandler deletes a user
func DeleteUserHandler(c *gin.Context) {
	userID := c.Param("id")
	log.Info().Str("user_id", userID).Msg("Delete user request")

	if grpcClient == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gRPC client not initialized"})
		return
	}
	resp, err := grpcClient.DeleteUser(userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete user via gRPC")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}
	if resp.Error != "" {
		if resp.Error == "User not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": resp.Error})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": resp.Error})
		return
	}
	if !resp.Success {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// ListUsersHandler lists users with pagination
func ListUsersHandler(c *gin.Context) {
	log.Info().Msg("List users request")

	// Parse query parameters
	page := ParseInt32(c.Query("page"), 1)
	limit := ParseInt32(c.Query("limit"), 10)
	search := c.Query("search")

	if grpcClient == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gRPC client not initialized"})
		return
	}

	resp, err := grpcClient.ListUsers(page, limit, search)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list users via gRPC")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list users"})
		return
	}

	if resp.Error != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": resp.Error})
		return
	}

	users := make([]gin.H, len(resp.Users))
	for i, user := range resp.Users {
		users[i] = gin.H{
			"id":         user.Id,
			"email":      user.Email,
			"name":       user.Name,
			"avatar_url": user.AvatarUrl,
			"roles":      user.Roles,
			"created_at": user.CreatedAt,
			"updated_at": user.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"total": resp.Total,
		"page":  page,
		"limit": limit,
	})
}

// GetUserProfileHandler gets user profile
func GetUserProfileHandler(c *gin.Context) {
	userID := c.Param("id")
	log.Info().Str("user_id", userID).Msg("Get user profile request")

	if grpcClient == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gRPC client not initialized"})
		return
	}
	resp, err := grpcClient.GetUser(userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user via gRPC")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}
	if resp.Error != "" {
		if resp.Error == "User not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": resp.Error})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": resp.Error})
		return
	}
	if resp.User == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":         resp.User.Id,
			"email":      resp.User.Email,
			"name":       resp.User.Name,
			"avatar_url": resp.User.AvatarUrl,
			"roles":      resp.User.Roles,
			"created_at": resp.User.CreatedAt,
			"updated_at": resp.User.UpdatedAt,
		},
	})
}

// UpdateUserProfileHandler updates user profile
func UpdateUserProfileHandler(c *gin.Context) {
	userID := c.Param("id")
	log.Info().Str("user_id", userID).Msg("Update user profile request")

	var req struct {
		Name      string `json:"name"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if grpcClient == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gRPC client not initialized"})
		return
	}
	grpcReq := &pb.UpdateUserRequest{
		Id:        userID,
		Name:      req.Name,
		AvatarUrl: req.AvatarURL,
	}
	resp, err := grpcClient.UpdateUser(grpcReq)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update user profile via gRPC")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user profile"})
		return
	}
	if resp.Error != "" {
		if resp.Error == "User not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": resp.Error})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": resp.Error})
		return
	}
	if resp.User == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":         resp.User.Id,
			"email":      resp.User.Email,
			"name":       resp.User.Name,
			"avatar_url": resp.User.AvatarUrl,
			"roles":      resp.User.Roles,
			"created_at": resp.User.CreatedAt,
			"updated_at": resp.User.UpdatedAt,
		},
	})
}

// Project handlers are now in project_handlers.go

// Handler holds the dependencies for HTTP handlers
type Handler struct {
	service *Service
}

// NewHandler creates a new HTTP handler
func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// Legacy handler struct for compatibility
type LegacyHandler struct {
	service *Service
}

// NewLegacyHandler creates a new legacy user handler
func NewLegacyHandler(service *Service) *LegacyHandler {
	return &LegacyHandler{
		service: service,
	}
}

// Legacy methods that delegate to the new handlers
func (h *LegacyHandler) GetUser(c *gin.Context) {
	GetUserHandler(c)
}

func (h *LegacyHandler) UpdateUser(c *gin.Context) {
	UpdateUserHandler(c)
}

func (h *LegacyHandler) DeleteUser(c *gin.Context) {
	DeleteUserHandler(c)
}

func (h *LegacyHandler) ListUsers(c *gin.Context) {
	ListUsersHandler(c)
}

// AdminListUsersHandler handles admin listing of all users
func (h *Handler) AdminListUsersHandler(c *gin.Context) {
	log.Info().Msg("Admin list users request")

	// Check if user has admin role
	roles, exists := c.Get("roles")
	if !exists {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "admin access required",
		})
		return
	}

	if !containsRole(roles, "admin") {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "admin access required",
		})
		return
	}

	// For now, return a placeholder response with admin access confirmed
	c.JSON(http.StatusOK, gin.H{
		"message": "Admin list users endpoint - placeholder",
		"users":   []interface{}{},
	})
}

// AdminListProjectsHandler handles admin listing of all projects
func (h *Handler) AdminListProjectsHandler(c *gin.Context) {
	log.Info().Msg("Admin list projects request")

	// Check if user has admin role
	roles, exists := c.Get("roles")
	if !exists {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "admin access required",
		})
		return
	}

	if !containsRole(roles, "admin") {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "admin access required",
		})
		return
	}

	// For now, return a placeholder response with admin access confirmed
	c.JSON(http.StatusOK, gin.H{
		"message":  "Admin list projects endpoint - placeholder",
		"projects": []interface{}{},
	})
}

// GetStatsHandler handles getting system statistics
func (h *Handler) GetStatsHandler(c *gin.Context) {
	log.Info().Msg("Get stats request")

	// For now, just return a placeholder response
	c.JSON(http.StatusOK, gin.H{
		"message": "Stats endpoint - placeholder",
		"stats": gin.H{
			"total_users":    0,
			"total_projects": 0,
		},
	})
}

// containsRole checks if roles contains the given role
func containsRole(roles interface{}, role string) bool {
	if roleList, ok := roles.([]string); ok {
		for _, r := range roleList {
			if r == role {
				return true
			}
		}
	}
	return false
}

// parsePaginationParams parses page and limit from query params, with defaults
func parsePaginationParams(c *gin.Context) (int, int) {
	page := 1
	limit := 20
	if p := c.Query("page"); p != "" {
		fmt.Sscanf(p, "%d", &page)
		if page < 1 {
			page = 1
		}
	}
	if l := c.Query("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
		if limit < 1 || limit > 100 {
			limit = 20
		}
	}
	return page, limit
}
