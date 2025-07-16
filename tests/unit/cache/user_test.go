package cache

import (
	"context"
	"testing"
	"time"

	"github.com/EliasRanz/ai-code-gen/internal/cache"
	"github.com/EliasRanz/ai-code-gen/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUserCacheIntegration tests user service cache integration with core infrastructure
func TestUserCacheIntegration(t *testing.T) {
	ctx := context.Background()

	// Create cache service for user
	serviceConfig := cache.DefaultServiceConfig()
	service, err := cache.NewService(serviceConfig)
	require.NoError(t, err)
	defer service.Close()

	// Create user cache manager using the service provider
	userConfig := user.DefaultCacheConfig()
	provider := service.GetProvider()
	cm := user.NewCacheManager(provider, userConfig)

	t.Run("User Profile Caching", func(t *testing.T) {
		userID := "user123"
		profile := &user.CachedUserProfile{
			UserID: userID,
			Email:  "user@test.com",
			Name:   "Test User",
			Role:   "user",
		}

		// Test user profile caching
		err := cm.CacheUserProfile(ctx, userID, profile)
		assert.NoError(t, err)

		cachedProfile, err := cm.GetUserProfile(ctx, userID)
		assert.NoError(t, err)
		require.NotNil(t, cachedProfile)
		assert.Equal(t, profile.UserID, cachedProfile.UserID)
		assert.Equal(t, profile.Email, cachedProfile.Email)
		assert.Equal(t, profile.Name, cachedProfile.Name)
		assert.Equal(t, profile.Role, cachedProfile.Role)

		// Test profile invalidation
		err = cm.InvalidateUserProfile(ctx, userID)
		assert.NoError(t, err)

		cachedProfile, err = cm.GetUserProfile(ctx, userID)
		assert.NoError(t, err)
		assert.Nil(t, cachedProfile)
	})

	t.Run("User Projects Caching", func(t *testing.T) {
		userID := "user456"
		projects := []user.ProjectSummary{
			{
				ProjectID:   "proj1",
				Name:        "Project 1",
				Description: "Test Project 1",
				Status:      "active",
			},
			{
				ProjectID:   "proj2",
				Name:        "Project 2",
				Description: "Test Project 2",
				Status:      "active",
			},
		}

		// Test projects caching
		err := cm.CacheUserProjects(ctx, userID, projects)
		assert.NoError(t, err)

		cachedProjects, err := cm.GetUserProjects(ctx, userID)
		assert.NoError(t, err)
		require.NotNil(t, cachedProjects)
		assert.Len(t, cachedProjects, 2)
		assert.Equal(t, projects[0].ProjectID, cachedProjects[0].ProjectID)
		assert.Equal(t, projects[1].Name, cachedProjects[1].Name)

		// Test projects invalidation
		err = cm.InvalidateUserProjects(ctx, userID)
		assert.NoError(t, err)

		cachedProjects, err = cm.GetUserProjects(ctx, userID)
		assert.NoError(t, err)
		assert.Nil(t, cachedProjects)
	})

	t.Run("User Chat Sessions Caching", func(t *testing.T) {
		userID := "user789"
		sessions := []user.ChatSessionSummary{
			{
				SessionID: "session1",
				ProjectID: "proj1",
				Title:     "Chat Session 1",
				Status:    "active",
			},
			{
				SessionID: "session2",
				ProjectID: "proj2",
				Title:     "Chat Session 2",
				Status:    "active",
			},
		}

		// Test chat sessions caching
		err := cm.CacheUserSessions(ctx, userID, sessions)
		assert.NoError(t, err)

		cachedSessions, err := cm.GetUserSessions(ctx, userID)
		assert.NoError(t, err)
		require.NotNil(t, cachedSessions)
		assert.Len(t, cachedSessions, 2)
		assert.Equal(t, sessions[0].SessionID, cachedSessions[0].SessionID)
		assert.Equal(t, sessions[1].Title, cachedSessions[1].Title)

		// Test sessions invalidation
		err = cm.InvalidateUserSessions(ctx, userID)
		assert.NoError(t, err)

		cachedSessions, err = cm.GetUserSessions(ctx, userID)
		assert.NoError(t, err)
		assert.Nil(t, cachedSessions)
	})

	t.Run("User Cache Health Check", func(t *testing.T) {
		// Test that user cache works with infrastructure health checks
		err := service.HealthCheck(ctx)
		assert.NoError(t, err)

		// Test provider is accessible through cache manager
		assert.NotNil(t, provider)
		err = provider.HealthCheck(ctx)
		assert.NoError(t, err)
	})
}

// TestUserCacheConfiguration tests user cache configuration integration
func TestUserCacheConfiguration(t *testing.T) {
	t.Run("Default User Configuration", func(t *testing.T) {
		config := user.DefaultCacheConfig()
		assert.Equal(t, 10*time.Minute, config.TTL)
		assert.Equal(t, "user:profile:", config.UserKeyPrefix)
		assert.Equal(t, "user:project:", config.ProjectKeyPrefix)
		assert.Equal(t, "user:session:", config.SessionKeyPrefix)
	})

	t.Run("Custom User Configuration", func(t *testing.T) {
		serviceConfig := cache.DefaultServiceConfig()
		service, err := cache.NewService(serviceConfig)
		require.NoError(t, err)
		defer service.Close()

		customConfig := user.CacheConfig{
			TTL:              15 * time.Minute,
			UserKeyPrefix:    "custom:user:profile:",
			ProjectKeyPrefix: "custom:user:projects:",
			SessionKeyPrefix: "custom:user:sessions:",
		}

		provider := service.GetProvider()
		cm := user.NewCacheManager(provider, customConfig)

		// Test that custom configuration is used
		ctx := context.Background()
		userID := "custom_user"
		profile := &user.CachedUserProfile{
			UserID: userID,
			Email:  "custom@test.com",
			Name:   "Custom User",
			Role:   "user",
		}

		err = cm.CacheUserProfile(ctx, userID, profile)
		assert.NoError(t, err)

		cached, err := cm.GetUserProfile(ctx, userID)
		assert.NoError(t, err)
		assert.NotNil(t, cached)
		assert.Equal(t, profile.Name, cached.Name)
	})
}

// BenchmarkUserCacheOperations benchmarks user-specific cache operations
func BenchmarkUserCacheOperations(b *testing.B) {
	ctx := context.Background()
	serviceConfig := cache.DefaultServiceConfig()
	service, _ := cache.NewService(serviceConfig)
	defer service.Close()

	userConfig := user.DefaultCacheConfig()
	provider := service.GetProvider()
	cm := user.NewCacheManager(provider, userConfig)

	profile := &user.CachedUserProfile{
		UserID: "bench_user",
		Email:  "bench@test.com",
		Name:   "Bench User",
		Role:   "user",
	}

	b.Run("UserCacheProfile", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			userID := "bench_user_" + string(rune(i))
			cm.CacheUserProfile(ctx, userID, profile)
		}
	})

	b.Run("UserGetProfile", func(b *testing.B) {
		userID := "bench_get_user"
		cm.CacheUserProfile(ctx, userID, profile)
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			cm.GetUserProfile(ctx, userID)
		}
	})

	b.Run("UserCacheProjects", func(b *testing.B) {
		projects := []user.ProjectSummary{
			{
				ProjectID:   "bench_proj1",
				Name:        "Bench Project 1",
				Description: "Bench Project",
				Status:      "active",
			},
		}

		for i := 0; i < b.N; i++ {
			userID := "bench_projects_user_" + string(rune(i))
			cm.CacheUserProjects(ctx, userID, projects)
		}
	})
}
