package user

import (
	"context"
	"testing"
	"time"

	"github.com/EliasRanz/ai-code-gen/internal/cache"
	"github.com/EliasRanz/ai-code-gen/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserCacheManager(t *testing.T) {
	ctx := context.Background()

	// Create memory provider for testing
	cacheConfig := cache.DefaultCacheConfig()
	provider, err := cache.NewMemoryProvider(cacheConfig)
	require.NoError(t, err)
	defer provider.Close()

	// Create user cache manager
	userConfig := user.DefaultCacheConfig()
	cm := user.NewCacheManager(provider, userConfig)

	t.Run("User Profile Caching", func(t *testing.T) {
		userID := "user123"
		profile := &user.CachedUserProfile{
			UserID:    userID,
			Email:     "test@example.com",
			Name:      "Test User",
			Role:      "user",
			CreatedAt: time.Now().Add(-24 * time.Hour),
			UpdatedAt: time.Now(),
		}

		// Test Cache User Profile
		err := cm.CacheUserProfile(ctx, userID, profile)
		assert.NoError(t, err)

		// Test Get User Profile
		cached, err := cm.GetUserProfile(ctx, userID)
		assert.NoError(t, err)
		require.NotNil(t, cached)
		assert.Equal(t, profile.UserID, cached.UserID)
		assert.Equal(t, profile.Email, cached.Email)
		assert.Equal(t, profile.Name, cached.Name)
		assert.Equal(t, profile.Role, cached.Role)
		assert.False(t, cached.CachedAt.IsZero())

		// Test Invalidate User Profile
		err = cm.InvalidateUserProfile(ctx, userID)
		assert.NoError(t, err)

		// Verify deletion
		cached, err = cm.GetUserProfile(ctx, userID)
		assert.NoError(t, err)
		assert.Nil(t, cached)
	})

	t.Run("User Projects Caching", func(t *testing.T) {
		userID := "user123"
		projects := []user.ProjectSummary{
			{
				ProjectID:   "proj1",
				Name:        "Project 1",
				Description: "Test project 1",
				Status:      "active",
				CreatedAt:   time.Now().Add(-48 * time.Hour),
				UpdatedAt:   time.Now().Add(-24 * time.Hour),
			},
			{
				ProjectID:   "proj2",
				Name:        "Project 2",
				Description: "Test project 2",
				Status:      "inactive",
				CreatedAt:   time.Now().Add(-72 * time.Hour),
				UpdatedAt:   time.Now().Add(-12 * time.Hour),
			},
		}

		// Test Cache User Projects
		err := cm.CacheUserProjects(ctx, userID, projects)
		assert.NoError(t, err)

		// Test Get User Projects
		cached, err := cm.GetUserProjects(ctx, userID)
		assert.NoError(t, err)
		require.NotNil(t, cached)
		assert.Len(t, cached, 2)
		assert.Equal(t, projects[0].ProjectID, cached[0].ProjectID)
		assert.Equal(t, projects[1].Name, cached[1].Name)

		// Test Invalidate User Projects
		err = cm.InvalidateUserProjects(ctx, userID)
		assert.NoError(t, err)

		// Verify deletion
		cached, err = cm.GetUserProjects(ctx, userID)
		assert.NoError(t, err)
		assert.Nil(t, cached)
	})

	t.Run("Project Caching", func(t *testing.T) {
		projectID := "proj123"
		project := &user.CachedProject{
			ProjectID:   projectID,
			UserID:      "user123",
			Name:        "Test Project",
			Description: "A test project",
			Status:      "active",
			Settings:    `{"theme": "dark"}`,
			CreatedAt:   time.Now().Add(-48 * time.Hour),
			UpdatedAt:   time.Now(),
		}

		// Test Cache Project
		err := cm.CacheProject(ctx, projectID, project)
		assert.NoError(t, err)

		// Test Get Project
		cached, err := cm.GetProject(ctx, projectID)
		assert.NoError(t, err)
		require.NotNil(t, cached)
		assert.Equal(t, project.ProjectID, cached.ProjectID)
		assert.Equal(t, project.UserID, cached.UserID)
		assert.Equal(t, project.Name, cached.Name)
		assert.Equal(t, project.Settings, cached.Settings)
		assert.False(t, cached.CachedAt.IsZero())

		// Test Invalidate Project
		err = cm.InvalidateProject(ctx, projectID)
		assert.NoError(t, err)

		// Verify deletion
		cached, err = cm.GetProject(ctx, projectID)
		assert.NoError(t, err)
		assert.Nil(t, cached)
	})

	t.Run("User Sessions Caching", func(t *testing.T) {
		userID := "user123"
		sessions := []user.ChatSessionSummary{
			{
				SessionID: "session1",
				ProjectID: "proj1",
				Title:     "Chat Session 1",
				Status:    "active",
				CreatedAt: time.Now().Add(-2 * time.Hour),
				UpdatedAt: time.Now().Add(-1 * time.Hour),
			},
			{
				SessionID: "session2",
				ProjectID: "proj2",
				Title:     "Chat Session 2",
				Status:    "completed",
				CreatedAt: time.Now().Add(-4 * time.Hour),
				UpdatedAt: time.Now().Add(-30 * time.Minute),
			},
		}

		// Test Cache User Sessions
		err := cm.CacheUserSessions(ctx, userID, sessions)
		assert.NoError(t, err)

		// Test Get User Sessions
		cached, err := cm.GetUserSessions(ctx, userID)
		assert.NoError(t, err)
		require.NotNil(t, cached)
		assert.Len(t, cached, 2)
		assert.Equal(t, sessions[0].SessionID, cached[0].SessionID)
		assert.Equal(t, sessions[1].Title, cached[1].Title)

		// Test Invalidate User Sessions
		err = cm.InvalidateUserSessions(ctx, userID)
		assert.NoError(t, err)

		// Verify deletion
		cached, err = cm.GetUserSessions(ctx, userID)
		assert.NoError(t, err)
		assert.Nil(t, cached)
	})

	t.Run("Key Generation", func(t *testing.T) {
		// Test user key generation
		userKey := cm.GenerateKey("user", "user123")
		assert.Equal(t, "user:profile:user123", userKey)

		// Test project key generation
		projectKey := cm.GenerateKey("project", "proj123")
		assert.Equal(t, "user:project:proj123", projectKey)

		// Test projects key generation
		projectsKey := cm.GenerateKey("projects", "user123")
		assert.Equal(t, "user:projects:user123", projectsKey)

		// Test sessions key generation
		sessionsKey := cm.GenerateKey("sessions", "user123")
		assert.Equal(t, "user:session:sessions:user123", sessionsKey)

		// Test unknown key type
		unknownKey := cm.GenerateKey("unknown", "test")
		assert.Equal(t, "user:unknown:test", unknownKey)
	})

	t.Run("JSON Operations", func(t *testing.T) {
		key := "test:json:key"
		data := map[string]interface{}{
			"field1": "value1",
			"field2": 42,
			"field3": true,
		}

		// Test SetJSON
		err := cm.SetJSON(ctx, key, data, time.Minute)
		assert.NoError(t, err)

		// Test GetJSON
		var result map[string]interface{}
		err = cm.GetJSON(ctx, key, &result)
		assert.NoError(t, err)
		assert.Equal(t, data["field1"], result["field1"])
		assert.Equal(t, float64(42), result["field2"]) // JSON numbers are float64
		assert.Equal(t, data["field3"], result["field3"])
	})

	t.Run("Input Validation", func(t *testing.T) {
		// Empty user ID
		err := cm.CacheUserProfile(ctx, "", &user.CachedUserProfile{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user ID cannot be empty")

		_, err = cm.GetUserProfile(ctx, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user ID cannot be empty")

		// Nil user profile
		err = cm.CacheUserProfile(ctx, "user", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user profile cannot be nil")

		// Empty project ID
		err = cm.CacheProject(ctx, "", &user.CachedProject{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "project ID cannot be empty")

		// Nil project
		err = cm.CacheProject(ctx, "proj", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "project cannot be nil")
	})

	t.Run("Health Check", func(t *testing.T) {
		err := cm.HealthCheck(ctx)
		assert.NoError(t, err)
	})
}

func TestUserCacheConfig(t *testing.T) {
	t.Run("Default Config", func(t *testing.T) {
		config := user.DefaultCacheConfig()
		assert.Equal(t, 10*time.Minute, config.TTL)
		assert.Equal(t, "user:profile:", config.UserKeyPrefix)
		assert.Equal(t, "user:project:", config.ProjectKeyPrefix)
		assert.Equal(t, "user:session:", config.SessionKeyPrefix)
	})
}

func BenchmarkUserCacheManager(b *testing.B) {
	ctx := context.Background()

	cacheConfig := cache.DefaultCacheConfig()
	provider, _ := cache.NewMemoryProvider(cacheConfig)
	defer provider.Close()

	userConfig := user.DefaultCacheConfig()
	cm := user.NewCacheManager(provider, userConfig)

	profile := &user.CachedUserProfile{
		UserID: "user123",
		Email:  "test@example.com",
		Name:   "Test User",
		Role:   "user",
	}

	b.Run("CacheUserProfile", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			userID := "user_" + string(rune(i))
			cm.CacheUserProfile(ctx, userID, profile)
		}
	})

	b.Run("GetUserProfile", func(b *testing.B) {
		userID := "bench_user"
		cm.CacheUserProfile(ctx, userID, profile)
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			cm.GetUserProfile(ctx, userID)
		}
	})

	projects := []user.ProjectSummary{
		{ProjectID: "proj1", Name: "Project 1", Status: "active"},
		{ProjectID: "proj2", Name: "Project 2", Status: "inactive"},
	}

	b.Run("CacheUserProjects", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			userID := "user_" + string(rune(i))
			cm.CacheUserProjects(ctx, userID, projects)
		}
	})
}
