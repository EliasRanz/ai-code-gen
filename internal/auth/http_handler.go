// Package auth provides HTTP handler for authentication operations
package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/EliasRanz/ai-code-gen/internal/infrastructure/observability"
	"github.com/EliasRanz/ai-code-gen/internal/utilities"
)

// HTTPHandler handles HTTP requests for authentication operations
type HTTPHandler struct {
	loginUC          *LoginUseCase
	logoutUC         *LogoutUseCase
	refreshTokenUC   *RefreshTokenUseCase
	validateTokenUC  *ValidateTokenService
	checkRoleUC      *CheckRoleService
	getSessionUC     *GetSessionService
	getUserContextUC *GetUserContextUseCase
	logger           observability.Logger
}

// NewHTTPHandler creates a new auth HTTP handler
func NewHTTPHandler(
	loginUC *LoginUseCase,
	logoutUC *LogoutUseCase,
	refreshTokenUC *RefreshTokenUseCase,
	validateTokenUC *ValidateTokenService,
	checkRoleUC *CheckRoleService,
	getSessionUC *GetSessionService,
	getUserContextUC *GetUserContextUseCase,
	logger observability.Logger,
) *HTTPHandler {
	return &HTTPHandler{
		loginUC:          loginUC,
		logoutUC:         logoutUC,
		refreshTokenUC:   refreshTokenUC,
		validateTokenUC:  validateTokenUC,
		checkRoleUC:      checkRoleUC,
		getSessionUC:     getSessionUC,
		getUserContextUC: getUserContextUC,
		logger:           logger,
	}
}

// RegisterRoutes implements utilities.HTTPHandler interface
func (h *HTTPHandler) RegisterRoutes(router utilities.Router) error {
	if router == nil {
		return utilities.NewValidationError("router cannot be nil", nil)
	}

	// Register auth routes
	apiGroup := router.Group("/api")
	h.registerAuthRoutes(apiGroup)

	return nil
}

// HealthCheck implements utilities.HTTPHandler interface
func (h *HTTPHandler) HealthCheck() error {
	// Check if all required dependencies are available
	if h.loginUC == nil || h.logoutUC == nil || h.refreshTokenUC == nil ||
		h.validateTokenUC == nil || h.checkRoleUC == nil || h.getSessionUC == nil ||
		h.getUserContextUC == nil {
		return utilities.NewValidationError("auth handler dependencies not properly initialized", nil)
	}
	return nil
}

// ValidateRoutes implements utilities.HTTPHandler interface
func (h *HTTPHandler) ValidateRoutes() error {
	// Validate that all required use cases are available
	if h.loginUC == nil {
		return utilities.NewValidationError("login use case is required", nil)
	}
	if h.logoutUC == nil {
		return utilities.NewValidationError("logout use case is required", nil)
	}
	if h.refreshTokenUC == nil {
		return utilities.NewValidationError("refresh token use case is required", nil)
	}
	if h.validateTokenUC == nil {
		return utilities.NewValidationError("validate token use case is required", nil)
	}
	return nil
}

// registerAuthRoutes registers all authentication routes
func (h *HTTPHandler) registerAuthRoutes(rg utilities.RouterGroup) {
	authGroup := rg.Group("/auth")

	// Authentication endpoints
	authGroup.POST("/login", h.adaptHandlerFunc(h.Login))
	authGroup.POST("/logout", h.adaptHandlerFunc(h.Logout))
	authGroup.POST("/refresh", h.adaptHandlerFunc(h.RefreshToken))

	// Centralized auth endpoints
	authGroup.POST("/validate", h.adaptHandlerFunc(h.ValidateToken))
	authGroup.POST("/check-role", h.adaptHandlerFunc(h.CheckRole))
	authGroup.GET("/session", h.adaptHandlerFunc(h.GetSession))
	authGroup.GET("/user/:id", h.adaptHandlerFunc(h.GetUserContext))
}

// adaptHandlerFunc adapts gin handler to utilities.HandlerFunc
func (h *HTTPHandler) adaptHandlerFunc(ginHandler gin.HandlerFunc) utilities.HandlerFunc {
	return func(ctx utilities.Context) {
		// For now, we'll handle the adaptation in the router layer
		// This is a placeholder implementation
	}
}

// Login handles POST /auth/login
func (h *HTTPHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid login request", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	resp, err := h.loginUC.Execute(c.Request.Context(), req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	h.logger.Info("User logged in successfully", map[string]interface{}{
		"user_id": resp.User.ID,
		"email":   resp.User.Email,
	})

	c.JSON(http.StatusOK, resp)
}

// Logout handles POST /auth/logout
func (h *HTTPHandler) Logout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization header is required"})
		return
	}

	accessToken := h.extractBearerToken(authHeader)
	if accessToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid authorization header format"})
		return
	}

	req := LogoutRequest{
		AccessToken: accessToken,
	}

	resp, err := h.logoutUC.Execute(c.Request.Context(), req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	h.logger.Info("User logged out successfully")
	c.JSON(http.StatusOK, resp)
}

// RefreshToken handles POST /auth/refresh
func (h *HTTPHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid refresh token request", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	resp, err := h.refreshTokenUC.Execute(c.Request.Context(), req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	h.logger.Info("Token refreshed successfully")
	c.JSON(http.StatusOK, resp)
}

// ValidateToken handles POST /auth/validate
func (h *HTTPHandler) ValidateToken(c *gin.Context) {
	var req ValidateTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid validate token request", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	resp, err := h.validateTokenUC.Execute(c.Request.Context(), req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	h.logger.Info("Token validation completed", map[string]interface{}{
		"valid": resp.Valid,
	})

	c.JSON(http.StatusOK, resp)
}

// CheckRole handles POST /auth/check-role
func (h *HTTPHandler) CheckRole(c *gin.Context) {
	var req CheckRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid check role request", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	resp, err := h.checkRoleUC.Execute(c.Request.Context(), req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	h.logger.Info("Role check completed", map[string]interface{}{
		"user_id":    req.UserID,
		"authorized": resp.Authorized,
	})

	c.JSON(http.StatusOK, resp)
}

// GetSession handles GET /auth/session
func (h *HTTPHandler) GetSession(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization header is required"})
		return
	}

	accessToken := h.extractBearerToken(authHeader)
	if accessToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid authorization header format"})
		return
	}

	h.logger.Info("Session info requested")
	c.JSON(http.StatusOK, gin.H{"message": "Session info endpoint - skeleton implementation"})
}

// GetUserContext handles GET /auth/user/:id
func (h *HTTPHandler) GetUserContext(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	h.logger.Info("User context requested", map[string]interface{}{
		"user_id": userID,
	})

	c.JSON(http.StatusOK, gin.H{"message": "User context endpoint - skeleton implementation"})
}

// extractBearerToken extracts token from Authorization header
func (h *HTTPHandler) extractBearerToken(authHeader string) string {
	if len(authHeader) > 7 && strings.HasPrefix(authHeader, "Bearer ") {
		return authHeader[7:]
	}
	return ""
}

// handleError handles different types of domain errors
func (h *HTTPHandler) handleError(c *gin.Context, err error) {
	h.logger.Error("Auth request failed", err, map[string]interface{}{
		"path":   c.Request.URL.Path,
		"method": c.Request.Method,
	})

	if IsValidationError(err) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if IsNotFoundError(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if IsConflictError(err) {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	// Unauthorized errors
	if err.Error() == "unauthorized" ||
		err.Error() == "invalid credentials" ||
		err.Error() == "user account is inactive" ||
		err.Error() == "invalid refresh token" ||
		err.Error() == "refresh token expired" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Default to internal server error
	c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
}
