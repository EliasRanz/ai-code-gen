// Package http provides HTTP interface adapters
package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/ai-code-gen/ai-ui-generator/internal/application/user"
	"github.com/ai-code-gen/ai-ui-generator/internal/domain/common"
	"github.com/ai-code-gen/ai-ui-generator/internal/infrastructure/observability"
)

// UserHandler handles HTTP requests for user operations
type UserHandler struct {
	createUserUC *user.CreateUserUseCase
	getUserUC    *user.GetUserUseCase
	updateUserUC *user.UpdateUserUseCase
	listUsersUC  *user.ListUsersUseCase
	deleteUserUC *user.DeleteUserUseCase
	logger       observability.Logger
}

// NewUserHandler creates a new user handler
func NewUserHandler(
	createUserUC *user.CreateUserUseCase,
	getUserUC *user.GetUserUseCase,
	updateUserUC *user.UpdateUserUseCase,
	listUsersUC *user.ListUsersUseCase,
	deleteUserUC *user.DeleteUserUseCase,
	logger observability.Logger,
) *UserHandler {
	return &UserHandler{
		createUserUC: createUserUC,
		getUserUC:    getUserUC,
		updateUserUC: updateUserUC,
		listUsersUC:  listUsersUC,
		deleteUserUC: deleteUserUC,
		logger:       logger,
	}
}

// CreateUser handles POST /users
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req user.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	resp, err := h.createUserUC.Execute(c.Request.Context(), req)
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

	req := user.GetUserRequest{
		UserID: common.UserID(userIDStr),
	}

	resp, err := h.getUserUC.Execute(c.Request.Context(), req)
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

	var req user.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	req.UserID = common.UserID(userIDStr)

	resp, err := h.updateUserUC.Execute(c.Request.Context(), req)
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

	req := user.ListUsersRequest{
		Page:   int32(page),
		Limit:  int32(limit),
		Search: search,
	}

	resp, err := h.listUsersUC.Execute(c.Request.Context(), req)
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

	req := user.DeleteUserRequest{
		UserID: common.UserID(userIDStr),
	}

	resp, err := h.deleteUserUC.Execute(c.Request.Context(), req)
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

	if common.IsValidationError(err) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if common.IsNotFoundError(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if common.IsConflictError(err) {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	// Default to internal server error
	c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
}
