package ai

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"

	"github.com/EliasRanz/ai-code-gen/internal/ai"
)

func TestRateLimiter(t *testing.T) {
	rl := ai.NewRateLimiter(rate.Limit(1), 1) // 1 request per second, burst of 1

	limiter1 := rl.GetLimiter("user1")
	limiter2 := rl.GetLimiter("user2")

	// Different users should get different limiters
	if limiter1 == limiter2 {
		t.Errorf("Expected different limiters for different users")
	}

	// Same user should get same limiter
	limiter1Again := rl.GetLimiter("user1")
	if limiter1 != limiter1Again {
		t.Errorf("Expected same limiter for same user")
	}
}

func TestRateLimitMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rl := ai.NewRateLimiter(rate.Limit(1), 1) // Very restrictive for testing

	r := gin.New()
	r.Use(rl.RateLimitMiddleware())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// First request should succeed
	req1, _ := http.NewRequest("GET", "/test?user_id=user1", nil)
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)
	assert.Equal(t, 200, w1.Code)

	// Second request should be rate limited
	req2, _ := http.NewRequest("GET", "/test?user_id=user1", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	assert.Equal(t, 429, w2.Code)
	assert.Contains(t, w2.Body.String(), "rate limit exceeded")
}

func TestQuotaManager(t *testing.T) {
	qm := ai.NewQuotaManager()
	userID := "user1"
	dailyLimit := 5

	// Should have quota initially
	assert.True(t, qm.CheckQuota(userID, dailyLimit))

	// Use quota multiple times
	for i := 0; i < dailyLimit; i++ {
		qm.UseQuota(userID)
	}

	// Should exceed quota now
	assert.False(t, qm.CheckQuota(userID, dailyLimit))

	// Check status
	status := qm.GetQuotaStatus(userID)
	assert.Equal(t, userID, status.UserID)
	assert.Equal(t, dailyLimit, status.DailyLimit)
	assert.Equal(t, dailyLimit, status.UsedToday)
}

func TestQuotaMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	qm := ai.NewQuotaManager()

	r := gin.New()
	r.Use(qm.QuotaMiddleware(1)) // Daily limit of 1 for testing
	r.GET("/test", func(c *gin.Context) {
		// Use quota in the handler
		if qmCtx, exists := c.Get("quota_manager"); exists {
			if quotaManager, ok := qmCtx.(*ai.QuotaManager); ok {
				if userID, exists := c.Get("user_id"); exists {
					if uid, ok := userID.(string); ok {
						quotaManager.UseQuota(uid)
					}
				}
			}
		}
		c.JSON(200, gin.H{"status": "ok"})
	})

	// First request should succeed
	req1, _ := http.NewRequest("GET", "/test?user_id=user1", nil)
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)
	assert.Equal(t, 200, w1.Code)

	// Second request should be quota limited
	req2, _ := http.NewRequest("GET", "/test?user_id=user1", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	assert.Equal(t, 429, w2.Code)
	assert.Contains(t, w2.Body.String(), "daily quota exceeded")
}

func TestGetQuotaEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := ai.NewService(&mockLLM{})
	h := ai.NewHandler(svc)
	r := gin.New()
	group := r.Group("")

	// Register routes without middleware for simpler testing
	ai := group.Group("/ai")
	ai.GET("/quota", h.GetQuota)

	// Test with user_id
	req, _ := http.NewRequest("GET", "/ai/quota?user_id=user1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "user1")
	assert.Contains(t, w.Body.String(), "daily_limit")

	// Test without user_id
	req2, _ := http.NewRequest("GET", "/ai/quota", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	assert.Equal(t, 400, w2.Code)
	assert.Contains(t, w2.Body.String(), "user_id required")
}
