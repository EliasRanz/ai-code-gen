// Package http provides HTTP interface adapters
package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ai-code-gen/ai-ui-generator/internal/application/auth"
	"github.com/ai-code-gen/ai-ui-generator/internal/domain/common"
	"github.com/ai-code-gen/ai-ui-generator/internal/infrastructure/observability"
)

// AuthHandler handles HTTP requests for authentication operations
type AuthHandler struct {
	loginUC        *auth.LoginUseCase
	logoutUC       *auth.LogoutUseCase
	refreshTokenUC *auth.RefreshTokenUseCase
	logger         observability.Logger
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(
	loginUC *auth.LoginUseCase,
	logoutUC *auth.LogoutUseCase,
	refreshTokenUC *auth.RefreshTokenUseCase,
	logger observability.Logger,
) *AuthHandler {
	return &AuthHandler{
		loginUC:        loginUC,
		logoutUC:       logoutUC,
		refreshTokenUC: refreshTokenUC,
		logger:         logger,
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

// handleError handles different types of domain errors
func (h *AuthHandler) handleError(c *gin.Context, err error) {
	h.logger.Error("Auth request failed", err, map[string]interface{}{
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
