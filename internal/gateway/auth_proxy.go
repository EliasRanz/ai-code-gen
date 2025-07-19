package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/EliasRanz/ai-code-gen/internal/cache"
)

// AuthProxyMiddleware implements auth service integration with caching
type AuthProxyMiddleware struct {
	authServiceURL string
	authCache      *cache.AuthCache
	name           string
	order          int
}

// NewAuthProxyMiddleware creates a new auth proxy middleware
func NewAuthProxyMiddleware(authServiceURL string, authCache *cache.AuthCache) *AuthProxyMiddleware {
	return &AuthProxyMiddleware{
		authServiceURL: authServiceURL,
		authCache:      authCache,
		name:           "auth-proxy",
		order:          100,
	}
}

// Process implements middleware processing with auth validation
func (a *AuthProxyMiddleware) Process(ctx Context, next Next) error {
	token, err := a.extractBearerToken(ctx)
	if err != nil {
		return a.handleAuthError(ctx, http.StatusUnauthorized, err.Error())
	}

	userContext, err := a.validateToken(ctx.Request().Context(), token)
	if err != nil {
		a.logTokenValidationFailure(token, err)
		return a.handleAuthError(ctx, http.StatusUnauthorized, "Authentication failed")
	}

	a.setUserContext(ctx, userContext)
	a.logAuthSuccess(userContext.UserID, userContext.Role, "")

	return next()
}

// ValidateToken validates a token and returns user context
func (a *AuthProxyMiddleware) ValidateToken(ctx Context) (*UserContext, error) {
	token, err := a.extractBearerToken(ctx)
	if err != nil {
		return nil, err
	}

	return a.validateToken(ctx.Request().Context(), token)
}

// CheckPermissions checks if user has required permissions
func (a *AuthProxyMiddleware) CheckPermissions(ctx Context, permissions []string) error {
	if len(permissions) == 0 {
		return nil
	}

	userRole, exists := ctx.Get("user_role")
	if !exists {
		return fmt.Errorf("user role not found in context")
	}

	role, ok := userRole.(string)
	if !ok {
		return fmt.Errorf("invalid user role type")
	}

	return a.checkRolePermissions(role, permissions)
}

// GetConfig returns middleware configuration
func (a *AuthProxyMiddleware) GetConfig() MiddlewareConfig {
	return &BasicMiddlewareConfig{
		name:    a.name,
		enabled: true,
		parameters: map[string]interface{}{
			"auth_service_url": a.authServiceURL,
			"cache_enabled":    a.authCache != nil,
		},
	}
}

// GetName returns middleware name
func (a *AuthProxyMiddleware) GetName() string {
	return a.name
}

// GetOrder returns middleware execution order
func (a *AuthProxyMiddleware) GetOrder() int {
	return a.order
}

// HealthCheck validates middleware health
func (a *AuthProxyMiddleware) HealthCheck() error {
	if a.authServiceURL == "" {
		return fmt.Errorf("auth service URL not configured")
	}
	return nil
}

// ValidateConfig validates middleware configuration
func (a *AuthProxyMiddleware) ValidateConfig() error {
	if a.authServiceURL == "" {
		return fmt.Errorf("auth service URL is required")
	}
	return nil
}

// extractBearerToken extracts Bearer token from request
func (a *AuthProxyMiddleware) extractBearerToken(ctx Context) (string, error) {
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("Authorization header required")
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", fmt.Errorf("Invalid authorization header format")
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" {
		return "", fmt.Errorf("Token not provided")
	}

	return token, nil
}

// validateToken validates token via cache or auth service
func (a *AuthProxyMiddleware) validateToken(ctx context.Context, token string) (*UserContext, error) {
	// Try cache first if available
	if a.authCache != nil {
		userContext, err := a.getCachedUserContext(ctx, token)
		if err != nil {
			log.Debug().Err(err).Msg("Cache error, falling back to auth service")
		}
		if userContext != nil {
			return userContext, nil
		}
	}

	// Validate with auth service
	userContext, err := a.validateWithAuthService(token)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if a.authCache != nil {
		if err := a.setCachedUserContext(ctx, token, userContext); err != nil {
			log.Debug().Err(err).Msg("Failed to cache auth result")
		}
	}

	return userContext, nil
}

// validateWithAuthService calls auth service for validation
func (a *AuthProxyMiddleware) validateWithAuthService(token string) (*UserContext, error) {
	requestBody := map[string]string{"access_token": token}
	return a.makeAuthServiceCall("/api/auth/validate", requestBody)
}

// makeAuthServiceCall makes HTTP call to auth service
func (a *AuthProxyMiddleware) makeAuthServiceCall(endpoint string, requestBody map[string]string) (*UserContext, error) {
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s%s", a.authServiceURL, endpoint)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("auth service request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("auth service returned status %d", resp.StatusCode)
	}

	var userContext UserContext
	if err := json.NewDecoder(resp.Body).Decode(&userContext); err != nil {
		return nil, fmt.Errorf("failed to decode auth service response: %w", err)
	}

	return &userContext, nil
}

// getCachedUserContext retrieves from cache
func (a *AuthProxyMiddleware) getCachedUserContext(ctx context.Context, token string) (*UserContext, error) {
	tokenHash := cache.HashToken(token)
	cachedContext, err := a.authCache.GetUserContext(ctx, tokenHash)
	if err != nil {
		return nil, err
	}

	if cachedContext == nil {
		return nil, nil
	}

	return &UserContext{
		UserID: cachedContext.UserID,
		Email:  cachedContext.Email,
		Role:   cachedContext.Role,
		Active: true,
	}, nil
}

// setCachedUserContext stores in cache
func (a *AuthProxyMiddleware) setCachedUserContext(ctx context.Context, token string, userContext *UserContext) error {
	tokenHash := cache.HashToken(token)
	cacheContext := &cache.UserContext{
		UserID: userContext.UserID,
		Email:  userContext.Email,
		Role:   userContext.Role,
	}
	return a.authCache.SetUserContext(ctx, tokenHash, cacheContext)
}

// checkRolePermissions validates user role against required permissions
func (a *AuthProxyMiddleware) checkRolePermissions(userRole string, permissions []string) error {
	// Admin can access everything
	if userRole == "admin" {
		return nil
	}

	// Check if user role matches any required permission
	for _, permission := range permissions {
		if userRole == permission {
			return nil
		}
	}

	return fmt.Errorf("insufficient permissions: requires one of %v, got %s", permissions, userRole)
}

// handleAuthError creates auth error response
func (a *AuthProxyMiddleware) handleAuthError(ctx Context, statusCode int, message string) error {
	if ginCtx, ok := ctx.(*GinContextWrapper); ok {
		ginCtx.JSON(statusCode, gin.H{"error": message})
		ginCtx.Abort()
	}
	return fmt.Errorf("auth error: %s", message)
}

// setUserContext sets user context in request
func (a *AuthProxyMiddleware) setUserContext(ctx Context, userContext *UserContext) {
	ctx.Set("user_id", userContext.UserID)
	ctx.Set("user_email", userContext.Email)
	ctx.Set("user_role", userContext.Role)
	ctx.Set("authenticated", true)
}

// logTokenValidationFailure logs validation failures
func (a *AuthProxyMiddleware) logTokenValidationFailure(token string, err error) {
	tokenPrefix := token
	if len(token) > 10 {
		tokenPrefix = token[:10]
	}

	log.Debug().
		Str("token_prefix", tokenPrefix).
		Err(err).
		Msg("Auth proxy: Token validation failed")
}

// logAuthSuccess logs successful authentication
func (a *AuthProxyMiddleware) logAuthSuccess(userID, userRole, requiredRole string) {
	logger := log.Debug().
		Str("user_id", userID).
		Str("user_role", userRole)

	if requiredRole != "" {
		logger = logger.Str("required_role", requiredRole)
		logger.Msg("Auth proxy: User authenticated and authorized successfully")
	} else {
		logger.Msg("Auth proxy: User authenticated successfully")
	}
}
