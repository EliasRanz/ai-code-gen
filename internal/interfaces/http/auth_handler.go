// Package http provides HTTP interface adapters
package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/EliasRanz/ai-code-gen/internal/auth"
	"github.com/EliasRanz/ai-code-gen/internal/infrastructure/observability"
)

// AuthHandler handles HTTP requests for authentication operations
type AuthHandler struct {
	loginUC          *auth.LoginUseCase
	logoutUC         *auth.LogoutUseCase
	refreshTokenUC   *auth.RefreshTokenUseCase
	validateTokenUC  *auth.ValidateTokenService
	checkRoleUC      *auth.CheckRoleService
	getSessionUC     *auth.GetSessionService
	getUserContextUC *auth.GetUserContextUseCase
	logger           observability.Logger
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(
	loginUC *auth.LoginUseCase,
	logoutUC *auth.LogoutUseCase,
	refreshTokenUC *auth.RefreshTokenUseCase,
	validateTokenUC *auth.ValidateTokenService,
	checkRoleUC *auth.CheckRoleService,
	getSessionUC *auth.GetSessionService,
	getUserContextUC *auth.GetUserContextUseCase,
	logger observability.Logger,
) *AuthHandler {
	return &AuthHandler{
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

// RegisterRoutes registers all authentication routes
func (h *AuthHandler) RegisterRoutes(router *gin.RouterGroup) {
	authGroup := router.Group("/auth")
	{
		// Existing authentication endpoints
		authGroup.POST("/login", h.Login)
		authGroup.POST("/logout", h.Logout)
		authGroup.POST("/refresh", h.RefreshToken)

		// New centralized auth endpoints
		authGroup.POST("/validate", h.ValidateToken)
		authGroup.POST("/check-role", h.CheckRole)
		authGroup.GET("/session", h.GetSession)
		authGroup.GET("/user/:id", h.GetUserContext)
	}
}

// Login handles POST /auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req auth.LoginRequest
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
func (h *AuthHandler) Logout(c *gin.Context) {
	// Extract access token from Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization header is required"})
		return
	}

	// Remove "Bearer " prefix
	accessToken := ""
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		accessToken = authHeader[7:]
	}

	if accessToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid authorization header format"})
		return
	}

	req := auth.LogoutRequest{
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
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req auth.RefreshTokenRequest
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
func (h *AuthHandler) ValidateToken(c *gin.Context) {
	// Extract token from request body
	var req auth.ValidateTokenRequest
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
func (h *AuthHandler) CheckRole(c *gin.Context) {
	var req auth.CheckRoleRequest
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
func (h *AuthHandler) GetSession(c *gin.Context) {
	// Extract token from Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization header is required"})
		return
	}

	// Remove "Bearer " prefix
	accessToken := ""
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		accessToken = authHeader[7:]
	}

	if accessToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid authorization header format"})
		return
	}

	// req := auth.GetSessionRequest{
	//     AccessToken: accessToken,
	// }

	// TODO: Implement session retrieval logic
	// resp, err := h.getSessionUC.Execute(c.Request.Context(), req)
	// if err != nil {
	//     h.handleError(c, err)
	//     return
	// }

	h.logger.Info("Session info requested")

	// Skeleton response
	c.JSON(http.StatusOK, gin.H{"message": "Session info endpoint - skeleton implementation"})
}

// GetUserContext handles GET /auth/user/:id
func (h *AuthHandler) GetUserContext(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	// req := auth.GetUserContextRequest{
	//     UserID: common.UserID(userID),
	// }

	// TODO: Implement user context retrieval logic
	// resp, err := h.getUserContextUC.Execute(c.Request.Context(), req)
	// if err != nil {
	//     h.handleError(c, err)
	//     return
	// }

	h.logger.Info("User context requested", map[string]interface{}{
		"user_id": userID,
	})

	// Skeleton response
	c.JSON(http.StatusOK, gin.H{"message": "User context endpoint - skeleton implementation"})
}

// handleError handles different types of domain errors
func (h *AuthHandler) handleError(c *gin.Context, err error) {
	h.logger.Error("Auth request failed", err, map[string]interface{}{
		"path":   c.Request.URL.Path,
		"method": c.Request.Method,
	})

	if auth.IsValidationError(err) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if auth.IsNotFoundError(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if auth.IsConflictError(err) {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	// Unauthorized errors (invalid credentials, expired tokens, etc.)
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
