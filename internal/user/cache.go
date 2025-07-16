package user

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/EliasRanz/ai-code-gen/internal/cache"
)

// CacheManager handles user-specific caching operations
type CacheManager struct {
	provider cache.CacheProvider
	config   CacheConfig
}

// CacheConfig holds user cache configuration
type CacheConfig struct {
	TTL              time.Duration `json:"ttl"`
	UserKeyPrefix    string        `json:"user_key_prefix"`
	ProjectKeyPrefix string        `json:"project_key_prefix"`
	SessionKeyPrefix string        `json:"session_key_prefix"`
}

// DefaultCacheConfig returns default user cache configuration
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		TTL:              10 * time.Minute,
		UserKeyPrefix:    "user:profile:",
		ProjectKeyPrefix: "user:project:",
		SessionKeyPrefix: "user:session:",
	}
}

// NewCacheManager creates a new user cache manager
func NewCacheManager(provider cache.CacheProvider, config CacheConfig) *CacheManager {
	return &CacheManager{
		provider: provider,
		config:   config,
	}
}

// CacheUserProfile stores user profile data in cache
func (cm *CacheManager) CacheUserProfile(ctx context.Context, userID string, profile *CachedUserProfile) error {
	if userID == "" {
		return fmt.Errorf("user ID cannot be empty")
	}
	if profile == nil {
		return fmt.Errorf("user profile cannot be nil")
	}

	profile.CachedAt = time.Now()
	key := cm.GenerateKey("user", userID)

	return cm.SetJSON(ctx, key, profile, cm.config.TTL)
}

// GetUserProfile retrieves user profile from cache
func (cm *CacheManager) GetUserProfile(ctx context.Context, userID string) (*CachedUserProfile, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}

	key := cm.GenerateKey("user", userID)

	var profile CachedUserProfile
	if err := cm.GetJSON(ctx, key, &profile); err != nil {
		return nil, err
	}

	// Check if we got data
	if profile.UserID == "" {
		return nil, nil // Cache miss
	}

	return &profile, nil
}

// InvalidateUserProfile removes user profile from cache
func (cm *CacheManager) InvalidateUserProfile(ctx context.Context, userID string) error {
	if userID == "" {
		return fmt.Errorf("user ID cannot be empty")
	}

	key := cm.GenerateKey("user", userID)
	return cm.provider.Delete(ctx, key)
}

// CacheUserProjects stores user projects list in cache
func (cm *CacheManager) CacheUserProjects(ctx context.Context, userID string, projects []ProjectSummary) error {
	if userID == "" {
		return fmt.Errorf("user ID cannot be empty")
	}

	key := cm.GenerateKey("projects", userID)

	data := ProjectsCache{
		UserID:   userID,
		Projects: projects,
		CachedAt: time.Now(),
	}

	return cm.SetJSON(ctx, key, data, cm.config.TTL)
}

// GetUserProjects retrieves user projects from cache
func (cm *CacheManager) GetUserProjects(ctx context.Context, userID string) ([]ProjectSummary, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}

	key := cm.GenerateKey("projects", userID)

	var data ProjectsCache
	if err := cm.GetJSON(ctx, key, &data); err != nil {
		return nil, err
	}

	// Check if we got data
	if data.UserID == "" {
		return nil, nil // Cache miss
	}

	return data.Projects, nil
}

// InvalidateUserProjects removes user projects from cache
func (cm *CacheManager) InvalidateUserProjects(ctx context.Context, userID string) error {
	if userID == "" {
		return fmt.Errorf("user ID cannot be empty")
	}

	key := cm.GenerateKey("projects", userID)
	return cm.provider.Delete(ctx, key)
}

// CacheProject stores individual project data in cache
func (cm *CacheManager) CacheProject(ctx context.Context, projectID string, project *CachedProject) error {
	if projectID == "" {
		return fmt.Errorf("project ID cannot be empty")
	}
	if project == nil {
		return fmt.Errorf("project cannot be nil")
	}

	project.CachedAt = time.Now()
	key := cm.GenerateKey("project", projectID)

	return cm.SetJSON(ctx, key, project, cm.config.TTL)
}

// GetProject retrieves project from cache
func (cm *CacheManager) GetProject(ctx context.Context, projectID string) (*CachedProject, error) {
	if projectID == "" {
		return nil, fmt.Errorf("project ID cannot be empty")
	}

	key := cm.GenerateKey("project", projectID)

	var project CachedProject
	if err := cm.GetJSON(ctx, key, &project); err != nil {
		return nil, err
	}

	// Check if we got data
	if project.ProjectID == "" {
		return nil, nil // Cache miss
	}

	return &project, nil
}

// InvalidateProject removes project from cache
func (cm *CacheManager) InvalidateProject(ctx context.Context, projectID string) error {
	if projectID == "" {
		return fmt.Errorf("project ID cannot be empty")
	}

	key := cm.GenerateKey("project", projectID)
	return cm.provider.Delete(ctx, key)
}

// CacheUserSessions stores user chat sessions in cache
func (cm *CacheManager) CacheUserSessions(ctx context.Context, userID string, sessions []ChatSessionSummary) error {
	if userID == "" {
		return fmt.Errorf("user ID cannot be empty")
	}

	key := cm.GenerateKey("sessions", userID)

	data := SessionsCache{
		UserID:   userID,
		Sessions: sessions,
		CachedAt: time.Now(),
	}

	return cm.SetJSON(ctx, key, data, cm.config.TTL)
}

// GetUserSessions retrieves user chat sessions from cache
func (cm *CacheManager) GetUserSessions(ctx context.Context, userID string) ([]ChatSessionSummary, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}

	key := cm.GenerateKey("sessions", userID)

	var data SessionsCache
	if err := cm.GetJSON(ctx, key, &data); err != nil {
		return nil, err
	}

	// Check if we got data
	if data.UserID == "" {
		return nil, nil // Cache miss
	}

	return data.Sessions, nil
}

// InvalidateUserSessions removes user sessions from cache
func (cm *CacheManager) InvalidateUserSessions(ctx context.Context, userID string) error {
	if userID == "" {
		return fmt.Errorf("user ID cannot be empty")
	}

	key := cm.GenerateKey("sessions", userID)
	return cm.provider.Delete(ctx, key)
}

// GenerateKey creates cache keys with proper prefixes
func (cm *CacheManager) GenerateKey(keyType string, identifiers ...string) string {
	switch keyType {
	case "user":
		if len(identifiers) >= 1 {
			return fmt.Sprintf("%s%s", cm.config.UserKeyPrefix, identifiers[0])
		}
	case "project":
		if len(identifiers) >= 1 {
			return fmt.Sprintf("%s%s", cm.config.ProjectKeyPrefix, identifiers[0])
		}
	case "projects":
		if len(identifiers) >= 1 {
			return fmt.Sprintf("user:projects:%s", identifiers[0])
		}
	case "sessions":
		if len(identifiers) >= 1 {
			return fmt.Sprintf("%ssessions:%s", cm.config.SessionKeyPrefix, identifiers[0])
		}
	}
	return fmt.Sprintf("user:unknown:%s", identifiers[0])
}

// InvalidateByPattern removes all keys matching a pattern
func (cm *CacheManager) InvalidateByPattern(ctx context.Context, pattern string) error {
	return cm.provider.DeleteByPattern(ctx, pattern)
}

// InvalidateByUser removes all cached data for a specific user
func (cm *CacheManager) InvalidateByUser(ctx context.Context, userID string) error {
	patterns := []string{
		fmt.Sprintf("user:*:%s:*", userID),
		fmt.Sprintf("user:*:%s", userID),
	}

	for _, pattern := range patterns {
		if err := cm.provider.DeleteByPattern(ctx, pattern); err != nil {
			return err
		}
	}

	return nil
}

// HealthCheck verifies cache connectivity
func (cm *CacheManager) HealthCheck(ctx context.Context) error {
	return cm.provider.HealthCheck(ctx)
}

// GetJSON retrieves and unmarshals JSON data from cache
func (cm *CacheManager) GetJSON(ctx context.Context, key string, target interface{}) error {
	data, err := cm.provider.Get(ctx, key)
	if err != nil {
		return err
	}

	if data == "" {
		return nil // Cache miss
	}

	return json.Unmarshal([]byte(data), target)
}

// SetJSON marshals and stores JSON data in cache
func (cm *CacheManager) SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return cm.provider.Set(ctx, key, string(data), ttl)
}

// Cache data structures

// CachedUserProfile represents cached user profile data with caching metadata
type CachedUserProfile struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	CachedAt  time.Time `json:"cached_at"`
}

// CachedProject represents cached project data with caching metadata
type CachedProject struct {
	ProjectID   string    `json:"project_id"`
	UserID      string    `json:"user_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Settings    string    `json:"settings"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	CachedAt    time.Time `json:"cached_at"`
}

// ProjectSummary represents a summarized project for lists
type ProjectSummary struct {
	ProjectID   string    `json:"project_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ProjectsCache represents cached projects list
type ProjectsCache struct {
	UserID   string           `json:"user_id"`
	Projects []ProjectSummary `json:"projects"`
	CachedAt time.Time        `json:"cached_at"`
}

// ChatSessionSummary represents a summarized chat session
type ChatSessionSummary struct {
	SessionID string    `json:"session_id"`
	ProjectID string    `json:"project_id"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SessionsCache represents cached sessions list
type SessionsCache struct {
	UserID   string               `json:"user_id"`
	Sessions []ChatSessionSummary `json:"sessions"`
	CachedAt time.Time            `json:"cached_at"`
}
