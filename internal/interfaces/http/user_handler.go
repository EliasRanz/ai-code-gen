// Package http provides HTTP interface adapters
package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/EliasRanz/ai-code-gen/internal/infrastructure/observability"
	userDomain "github.com/EliasRanz/ai-code-gen/internal/user"
	"github.com/EliasRanz/ai-code-gen/internal/utilities"
)

// UserHandler handles HTTP requests for user operations
type UserHandler struct {
	userCreator   *userDomain.UserCreator
	userRetriever *userDomain.UserRetriever
	userUpdater   *userDomain.UserUpdater
	userLister    *userDomain.UserLister
	userDeleter   *userDomain.UserDeleter
	logger        observability.Logger
}

// NewUserHandler creates a new user handler
func NewUserHandler(
	userCreator *userDomain.UserCreator,
	userRetriever *userDomain.UserRetriever,
	userUpdater *userDomain.UserUpdater,
	userLister *userDomain.UserLister,
	userDeleter *userDomain.UserDeleter,
	logger observability.Logger,
) *UserHandler {
	return &UserHandler{
		userCreator:   userCreator,
		userRetriever: userRetriever,
		userUpdater:   userUpdater,
		userLister:    userLister,
		userDeleter:   userDeleter,
		logger:        logger,
	}
}

// RegisterRoutes registers all user-related routes
func (h *UserHandler) RegisterRoutes(router *gin.Engine) *gin.Engine {
	// Register all user and project routes using the handler
	apiGroup := router.Group("/api")
	h.RegisterUserRoutes(apiGroup)

	return router
}

// RegisterUserRoutes registers all user-related routes
func (h *UserHandler) RegisterUserRoutes(rg *gin.RouterGroup) {
	userGroup := rg.Group("/users")
	{
		userGroup.POST("", h.CreateUser)
		userGroup.GET("/:id", h.GetUser)
		userGroup.PUT("/:id", h.UpdateUser)
		userGroup.DELETE("/:id", h.DeleteUser)
		userGroup.GET("", h.ListUsers)
	}
}

// CreateUser handles POST /users
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req userDomain.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	resp, err := h.userCreator.Execute(c.Request.Context(), req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	h.logger.Info("User created successfully", map[string]interface{}{
		"user_id": resp.User.ID,
		"email":   resp.User.Email,
	})

	c.JSON(http.StatusCreated, resp)
}

// GetUser handles GET /users/:id
func (h *UserHandler) GetUser(c *gin.Context) {
	userIDStr := c.Param("id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	req := userDomain.GetUserRequest{
		UserID: utilities.UserID(userIDStr),
	}

	resp, err := h.userRetriever.Execute(c.Request.Context(), req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// UpdateUser handles PUT /users/:id
func (h *UserHandler) UpdateUser(c *gin.Context) {
	userIDStr := c.Param("id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	var req userDomain.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	req.UserID = utilities.UserID(userIDStr)

	resp, err := h.userUpdater.Execute(c.Request.Context(), req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	h.logger.Info("User updated successfully", map[string]interface{}{
		"user_id": resp.User.ID,
	})

	c.JSON(http.StatusOK, resp)
}

// ListUsers handles GET /users
func (h *UserHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 32)
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "20"), 10, 32)
	search := c.Query("search")

	req := userDomain.ListUsersRequest{
		Page:   int32(page),
		Limit:  int32(limit),
		Search: search,
	}

	resp, err := h.userLister.Execute(c.Request.Context(), req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// DeleteUser handles DELETE /users/:id
func (h *UserHandler) DeleteUser(c *gin.Context) {
	userIDStr := c.Param("id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	req := userDomain.DeleteUserRequest{
		UserID: utilities.UserID(userIDStr),
	}

	resp, err := h.userDeleter.Execute(c.Request.Context(), req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	h.logger.Info("User deleted successfully", map[string]interface{}{
		"user_id": userIDStr,
	})

	c.JSON(http.StatusOK, resp)
}

// handleError handles different types of domain errors
func (h *UserHandler) handleError(c *gin.Context, err error) {
	h.logger.Error("Request failed", err, map[string]interface{}{
		"path":   c.Request.URL.Path,
		"method": c.Request.Method,
	})

	if utilities.IsValidationError(err) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if utilities.IsNotFoundError(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if utilities.IsConflictError(err) {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	// Default to internal server error
	c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
}
