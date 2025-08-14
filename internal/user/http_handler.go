// Package user provides HTTP handler for user operations
package user

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/EliasRanz/ai-code-gen/internal/observability"
	"github.com/EliasRanz/ai-code-gen/internal/utilities"
)

// HTTPHandler handles HTTP requests for user operations
type HTTPHandler struct {
	userCreator   *UserCreator
	userRetriever *UserRetriever
	userUpdater   *UserUpdater
	userLister    *UserLister
	userDeleter   *UserDeleter
	logger        observability.Logger
}

// NewHTTPHandler creates a new user HTTP handler
func NewHTTPHandler(
	userCreator *UserCreator,
	userRetriever *UserRetriever,
	userUpdater *UserUpdater,
	userLister *UserLister,
	userDeleter *UserDeleter,
	logger observability.Logger,
) *HTTPHandler {
	return &HTTPHandler{
		userCreator:   userCreator,
		userRetriever: userRetriever,
		userUpdater:   userUpdater,
		userLister:    userLister,
		userDeleter:   userDeleter,
		logger:        logger,
	}
}

// RegisterRoutes implements utilities.HTTPHandler interface
func (h *HTTPHandler) RegisterRoutes(router utilities.Router) error {
	if router == nil {
		return utilities.NewValidationError("router cannot be nil", nil)
	}

	// Register user routes
	apiGroup := router.Group("/api")
	h.registerUserRoutes(apiGroup)

	return nil
}

// HealthCheck implements utilities.HTTPHandler interface
func (h *HTTPHandler) HealthCheck() error {
	// Check if all required dependencies are available
	if h.userCreator == nil || h.userRetriever == nil ||
		h.userUpdater == nil || h.userLister == nil ||
		h.userDeleter == nil {
		return utilities.NewValidationError("user handler dependencies not properly initialized", nil)
	}
	return nil
}

// ValidateRoutes implements utilities.HTTPHandler interface
func (h *HTTPHandler) ValidateRoutes() error {
	// Validate that all required use cases are available
	if h.userCreator == nil {
		return utilities.NewValidationError("user creator use case is required", nil)
	}
	if h.userRetriever == nil {
		return utilities.NewValidationError("user retriever use case is required", nil)
	}
	if h.userUpdater == nil {
		return utilities.NewValidationError("user updater use case is required", nil)
	}
	if h.userLister == nil {
		return utilities.NewValidationError("user lister use case is required", nil)
	}
	if h.userDeleter == nil {
		return utilities.NewValidationError("user deleter use case is required", nil)
	}
	return nil
}

// registerUserRoutes registers all user-related routes
func (h *HTTPHandler) registerUserRoutes(rg utilities.RouterGroup) {
	userGroup := rg.Group("/users")
	userGroup.POST("", h.adaptHandlerFunc(h.CreateUser))
	userGroup.GET("/:id", h.adaptHandlerFunc(h.GetUser))
	userGroup.PUT("/:id", h.adaptHandlerFunc(h.UpdateUser))
	userGroup.DELETE("/:id", h.adaptHandlerFunc(h.DeleteUser))
	userGroup.GET("", h.adaptHandlerFunc(h.ListUsers))
}

// adaptHandlerFunc adapts gin handler to utilities.HandlerFunc
func (h *HTTPHandler) adaptHandlerFunc(ginHandler gin.HandlerFunc) utilities.HandlerFunc {
	return func(ctx utilities.Context) {
		// For now, we'll handle the adaptation in the router layer
		// This is a placeholder implementation
	}
}

// CreateUser handles POST /users
func (h *HTTPHandler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
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
func (h *HTTPHandler) GetUser(c *gin.Context) {
	userIDStr := c.Param("id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	req := GetUserRequest{
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
func (h *HTTPHandler) UpdateUser(c *gin.Context) {
	userIDStr := c.Param("id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	var req UpdateUserRequest
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
func (h *HTTPHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 32)
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "20"), 10, 32)
	search := c.Query("search")

	req := ListUsersRequest{
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
func (h *HTTPHandler) DeleteUser(c *gin.Context) {
	userIDStr := c.Param("id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	req := DeleteUserRequest{
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
func (h *HTTPHandler) handleError(c *gin.Context, err error) {
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
