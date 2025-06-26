package ai

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimiter manages rate limiting for AI requests
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     r,
		burst:    b,
	}
}

// GetLimiter returns the rate limiter for a given key (e.g., user ID or IP)
func (rl *RateLimiter) GetLimiter(key string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[key]
	if !exists {
		limiter = rate.NewLimiter(rl.rate, rl.burst)
		rl.limiters[key] = limiter
	}

	return limiter
}

// RateLimitMiddleware creates a middleware for rate limiting
func (rl *RateLimiter) RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Use user_id if available, otherwise fall back to IP
		key := c.Query("user_id")
		if key == "" {
			key = c.ClientIP()
		}

		limiter := rl.GetLimiter(key)
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
				"retry_after": time.Second / time.Duration(rl.rate),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// QuotaManager manages usage quotas
type QuotaManager struct {
	quotas map[string]*UserQuota
	mu     sync.RWMutex
}

// UserQuota tracks usage for a user
type UserQuota struct {
	UserID       string
	DailyLimit   int
	UsedToday    int
	LastResetDay time.Time
}

// NewQuotaManager creates a new quota manager
func NewQuotaManager() *QuotaManager {
	return &QuotaManager{
		quotas: make(map[string]*UserQuota),
	}
}

// CheckQuota checks if user has quota remaining
func (qm *QuotaManager) CheckQuota(userID string, dailyLimit int) bool {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	quota, exists := qm.quotas[userID]
	if !exists {
		quota = &UserQuota{
			UserID:       userID,
			DailyLimit:   dailyLimit,
			UsedToday:    0,
			LastResetDay: time.Now().Truncate(24 * time.Hour),
		}
		qm.quotas[userID] = quota
	}

	// Reset if it's a new day
	today := time.Now().Truncate(24 * time.Hour)
	if quota.LastResetDay.Before(today) {
		quota.UsedToday = 0
		quota.LastResetDay = today
	}

	return quota.UsedToday < quota.DailyLimit
}

// UseQuota increments the usage count
func (qm *QuotaManager) UseQuota(userID string) {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	if quota, exists := qm.quotas[userID]; exists {
		quota.UsedToday++
	}
}

// GetQuotaStatus returns current quota status
func (qm *QuotaManager) GetQuotaStatus(userID string) *UserQuota {
	qm.mu.RLock()
	defer qm.mu.RUnlock()

	quota, exists := qm.quotas[userID]
	if !exists {
		return nil
	}

	// Make a copy to avoid race conditions
	return &UserQuota{
		UserID:       quota.UserID,
		DailyLimit:   quota.DailyLimit,
		UsedToday:    quota.UsedToday,
		LastResetDay: quota.LastResetDay,
	}
}

// QuotaMiddleware creates a middleware for quota checking
func (qm *QuotaManager) QuotaMiddleware(defaultDailyLimit int) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Query("user_id")
		if userID == "" {
			// Skip quota check for anonymous users or use IP-based limiting
			c.Next()
			return
		}

		if !qm.CheckQuota(userID, defaultDailyLimit) {
			quota := qm.GetQuotaStatus(userID)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "daily quota exceeded",
				"daily_limit": quota.DailyLimit,
				"used_today":  quota.UsedToday,
				"reset_time":  quota.LastResetDay.Add(24 * time.Hour),
			})
			c.Abort()
			return
		}

		// Store quota manager in context for later use
		c.Set("quota_manager", qm)
		c.Set("user_id", userID)
		c.Next()
	}
}
