package user

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	pb "github.com/EliasRanz/ai-code-gen/api/proto/user"
)

// parseIntParam parses an integer parameter from the query string
func parseIntParam(c *gin.Context, key string, defaultValue int) int {
	value := c.Query(key)
	if value == "" {
		return defaultValue
	}
	if intValue, err := strconv.Atoi(value); err == nil {
		return intValue
	}
	return defaultValue
}

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
		c.JSON(http.StatusBadRequest, gin.H{"error": resp.Error})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// ListUsersHandler lists users with pagination and search
func ListUsersHandler(c *gin.Context) {
	log.Info().Msg("List users request")

	page := parseIntParam(c, "page", 1)
	limit := parseIntParam(c, "limit", 10)
	search := c.Query("search")

	if grpcClient == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gRPC client not initialized"})
		return
	}

	resp, err := grpcClient.ListUsers(int32(page), int32(limit), search)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list users via gRPC")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list users"})
		return
	}

	if resp.Error != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": resp.Error})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users":       resp.Users,
		"total":       resp.Total,
		"page":        page,
		"limit":       limit,
		"total_pages": (resp.Total + int32(limit) - 1) / int32(limit),
	})
}
