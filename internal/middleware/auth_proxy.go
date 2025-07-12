package middleware

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

// UserContext represents the user context returned by auth service
type UserContext struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	Active bool   `json:"active"`
}

// AuthServiceProxy creates middleware that validates tokens via auth service with optional caching
func AuthServiceProxy(authServiceURL string, authCache *cache.AuthCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := extractBearerToken(c)
		if err != nil {
			handleAuthError(c, http.StatusUnauthorized, err.Error())
			return
		}

		// Try cache first (if available)
		userContext, cacheErr := getCachedUserContext(c.Request.Context(), authCache, token)
		if cacheErr != nil {
			log.Debug().Err(cacheErr).Msg("Cache error, falling back to auth service")
		}

		// Cache miss or error - validate with auth service
		if userContext == nil {
			userContext, err = validateWithAuthService(authServiceURL, token)
			if err != nil {
				logTokenValidationFailure(token, err)
				handleAuthError(c, http.StatusUnauthorized, "Authentication failed")
				return
			}

			// Cache the result for future requests (if cache available)
			if authCache != nil {
				if err := setCachedUserContext(c.Request.Context(), authCache, token, userContext); err != nil {
					log.Debug().Err(err).Msg("Failed to cache auth result")
				}
			}
		}

		setUserContext(c, userContext)
		logAuthSuccess(userContext.UserID, userContext.Role, "")
		c.Next()
	}
}

// AuthServiceRoleProxy creates middleware that validates tokens and checks roles via auth service with caching
func AuthServiceRoleProxy(authServiceURL string, authCache *cache.AuthCache, requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := extractBearerToken(c)
		if err != nil {
			handleAuthError(c, http.StatusUnauthorized, err.Error())
			return
		}

		// Try cache first (if available)
		userContext, cacheErr := getCachedUserContext(c.Request.Context(), authCache, token)
		if cacheErr != nil {
			log.Debug().Err(cacheErr).Msg("Cache error, falling back to auth service")
		}

		// Cache miss or error - validate with auth service
		if userContext == nil {
			if err := checkRoleWithAuthService(authServiceURL, token, requiredRole); err != nil {
				logRoleCheckFailure(token, requiredRole, err)
				handleAuthError(c, http.StatusForbidden, "Insufficient permissions")
				return
			}

			userContext, err = validateWithAuthService(authServiceURL, token)
			if err != nil {
				logTokenValidationFailure(token, err)
				handleAuthError(c, http.StatusUnauthorized, "Authentication failed")
				return
			}

			// Cache the result for future requests (if cache available)
			if authCache != nil {
				if err := setCachedUserContext(c.Request.Context(), authCache, token, userContext); err != nil {
					log.Debug().Err(err).Msg("Failed to cache auth result")
				}
			}
		} else {
			// Check role for cached user
			if !hasRequiredRole(userContext, requiredRole) {
				logRoleCheckFailure(token, requiredRole, fmt.Errorf("user role %s insufficient for %s", userContext.Role, requiredRole))
				handleAuthError(c, http.StatusForbidden, "Insufficient permissions")
				return
			}
		}

		setUserContext(c, userContext)
		logAuthSuccess(userContext.UserID, userContext.Role, requiredRole)
		c.Next()
	}
}

// extractBearerToken extracts and validates the Bearer token from request
func extractBearerToken(c *gin.Context) (string, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		log.Debug().Msg("Auth proxy: No authorization header provided")
		return "", fmt.Errorf("Authorization header required")
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		log.Debug().Msg("Auth proxy: Invalid authorization header format")
		return "", fmt.Errorf("Invalid authorization header format")
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" {
		log.Debug().Msg("Auth proxy: Empty token provided")
		return "", fmt.Errorf("Token not provided")
	}

	return token, nil
}

// validateWithAuthService calls the centralized auth service to validate a token
func validateWithAuthService(authServiceURL, token string) (*UserContext, error) {
	requestBody := map[string]string{"access_token": token}
	return makeAuthServiceCall(authServiceURL, "/api/auth/validate", requestBody)
}

// checkRoleWithAuthService calls the centralized auth service to check user role
func checkRoleWithAuthService(authServiceURL, token, requiredRole string) error {
	requestBody := map[string]string{
		"access_token":  token,
		"required_role": requiredRole,
	}

	_, err := makeAuthServiceCall(authServiceURL, "/api/auth/check-role", requestBody)
	return err
}

// makeAuthServiceCall makes an HTTP call to the auth service
func makeAuthServiceCall(authServiceURL, endpoint string, requestBody map[string]string) (*UserContext, error) {
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s%s", authServiceURL, endpoint)
	req, err := createHTTPRequest(url, jsonBody)
	if err != nil {
		return nil, err
	}

	resp, err := executeHTTPRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("auth service returned status %d", resp.StatusCode)
	}

	return parseAuthResponse(resp, endpoint)
}

// createHTTPRequest creates an HTTP request with proper headers
func createHTTPRequest(url string, jsonBody []byte) (*http.Request, error) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

// executeHTTPRequest executes the HTTP request with timeout
func executeHTTPRequest(req *http.Request) (*http.Response, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("auth service request failed: %w", err)
	}
	return resp, nil
}

// parseAuthResponse parses the auth service response
func parseAuthResponse(resp *http.Response, endpoint string) (*UserContext, error) {
	// For role check endpoint, we don't need to parse response
	if endpoint == "/api/auth/check-role" {
		return nil, nil
	}

	// Parse response for validate endpoint
	var userContext UserContext
	if err := json.NewDecoder(resp.Body).Decode(&userContext); err != nil {
		return nil, fmt.Errorf("failed to decode auth service response: %w", err)
	}

	return &userContext, nil
}

// handleAuthError responds with auth error and aborts request
func handleAuthError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, gin.H{"error": message})
	c.Abort()
}

// setUserContext sets validated user context in gin context
func setUserContext(c *gin.Context, userContext *UserContext) {
	c.Set("user_id", userContext.UserID)
	c.Set("user_email", userContext.Email)
	c.Set("user_role", userContext.Role)
	c.Set("authenticated", true)
}

// getCachedUserContext retrieves user context from cache
func getCachedUserContext(ctx context.Context, authCache *cache.AuthCache, token string) (*UserContext, error) {
	if authCache == nil {
		return nil, nil // No cache configured
	}

	tokenHash := cache.HashToken(token)
	cachedContext, err := authCache.GetUserContext(ctx, tokenHash)
	if err != nil {
		return nil, fmt.Errorf("cache retrieval error: %w", err)
	}

	if cachedContext == nil {
		return nil, nil // Cache miss
	}

	// Convert cache.UserContext to middleware.UserContext
	return &UserContext{
		UserID: cachedContext.UserID,
		Email:  cachedContext.Email,
		Role:   cachedContext.Role,
		Active: true, // Cached users are considered active
	}, nil
}

// setCachedUserContext stores user context in cache
func setCachedUserContext(ctx context.Context, authCache *cache.AuthCache, token string, userContext *UserContext) error {
	if authCache == nil {
		return nil // No cache configured
	}

	tokenHash := cache.HashToken(token)
	
	// Convert middleware.UserContext to cache.UserContext
	cacheContext := &cache.UserContext{
		UserID: userContext.UserID,
		Email:  userContext.Email,
		Role:   userContext.Role,
	}

	return authCache.SetUserContext(ctx, tokenHash, cacheContext)
}

// hasRequiredRole checks if user has the required role
func hasRequiredRole(userContext *UserContext, requiredRole string) bool {
	if requiredRole == "" {
		return true // No specific role required
	}

	// Admin users can access any role
	if userContext.Role == "admin" {
		return true
	}

	// Check exact role match
	return userContext.Role == requiredRole
}

// InvalidateUserCache invalidates cached auth results for a user
func InvalidateUserCache(ctx context.Context, authCache *cache.AuthCache, token string) error {
	if authCache == nil {
		return nil // No cache configured
	}

	tokenHash := cache.HashToken(token)
	return authCache.InvalidateUserContext(ctx, tokenHash)
}

// logTokenValidationFailure logs token validation failures with safe token prefix
func logTokenValidationFailure(token string, err error) {
	log.Debug().
		Str("token_prefix", token[:min(10, len(token))]).
		Err(err).
		Msg("Auth proxy: Token validation failed")
}

// logRoleCheckFailure logs role check failures with safe token prefix
func logRoleCheckFailure(token, requiredRole string, err error) {
	log.Debug().
		Str("token_prefix", token[:min(10, len(token))]).
		Str("required_role", requiredRole).
		Err(err).
		Msg("Auth proxy: Role check failed")
}

// logAuthSuccess logs successful authentication
func logAuthSuccess(userID, userRole, requiredRole string) {
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
