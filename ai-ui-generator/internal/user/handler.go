package user

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

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
