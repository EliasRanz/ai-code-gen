package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	mathrand "math/rand"
	"time"

	"github.com/EliasRanz/ai-code-gen/internal/cache"
)

// TestDataGenerator creates realistic test data for performance testing
type TestDataGenerator struct {
	UserCount     int
	TokensPerUser int
	Roles         []string
}

// TestToken represents a test authentication token with expected cache behavior
type TestToken struct {
	Hash        string
	UserContext *cache.UserContext
	HitProbability float64 // Probability this token will be a cache hit
}

// LoadTestScenario defines a load testing scenario
type LoadTestScenario struct {
	Name          string
	Description   string
	RequestRate   int           // Requests per second
	Duration      time.Duration
	CacheHitRatio float64       // Expected cache hit ratio (0.0 to 1.0)
	UserPattern   string        // "uniform", "normal", "zipf" - distribution of user requests
	Concurrency   int           // Number of concurrent workers
}

// NewTestDataGenerator creates a new test data generator
func NewTestDataGenerator(userCount, tokensPerUser int) *TestDataGenerator {
	return &TestDataGenerator{
		UserCount:     userCount,
		TokensPerUser: tokensPerUser,
		Roles:         []string{"user", "admin", "moderator", "premium", "basic"},
	}
}

// GenerateTestTokens creates a set of test tokens with realistic distributions
func (tdg *TestDataGenerator) GenerateTestTokens() []TestToken {
	tokens := make([]TestToken, 0, tdg.UserCount*tdg.TokensPerUser)
	
	for userID := 0; userID < tdg.UserCount; userID++ {
		for tokenID := 0; tokenID < tdg.TokensPerUser; tokenID++ {
			token := TestToken{
				Hash: tdg.generateTokenHash(userID, tokenID),
				UserContext: &cache.UserContext{
					UserID: fmt.Sprintf("user-%d", userID),
					Email:  fmt.Sprintf("user%d@test.com", userID),
					Role:   tdg.selectRandomRole(),
				},
				HitProbability: tdg.calculateHitProbability(userID, tokenID),
			}
			tokens = append(tokens, token)
		}
	}
	
	return tokens
}

// GenerateRealisticTokens creates tokens that follow real-world usage patterns
func (tdg *TestDataGenerator) GenerateRealisticTokens() []TestToken {
	tokens := make([]TestToken, 0, tdg.UserCount*tdg.TokensPerUser)
	
	// 80/20 rule: 20% of users generate 80% of requests
	heavyUsers := int(float64(tdg.UserCount) * 0.2)
	lightUsers := tdg.UserCount - heavyUsers
	
	// Heavy users (high hit probability)
	for userID := 0; userID < heavyUsers; userID++ {
		for tokenID := 0; tokenID < tdg.TokensPerUser*4; tokenID++ { // More tokens for heavy users
			tokens = append(tokens, TestToken{
				Hash: tdg.generateTokenHash(userID, tokenID),
				UserContext: &cache.UserContext{
					UserID: fmt.Sprintf("heavy-user-%d", userID),
					Email:  fmt.Sprintf("heavy%d@test.com", userID),
					Role:   tdg.selectWeightedRole(),
				},
				HitProbability: 0.85 + (0.10 * mathrand.Float64()), // 85-95% hit rate
			})
		}
	}
	
	// Light users (lower hit probability)
	for userID := 0; userID < lightUsers; userID++ {
		tokens = append(tokens, TestToken{
			Hash: tdg.generateTokenHash(userID+heavyUsers, 0),
			UserContext: &cache.UserContext{
				UserID: fmt.Sprintf("light-user-%d", userID),
				Email:  fmt.Sprintf("light%d@test.com", userID),
				Role:   "user", // Most light users are basic users
			},
			HitProbability: 0.20 + (0.30 * mathrand.Float64()), // 20-50% hit rate
		})
	}
	
	return tokens
}

// GenerateLoadTestScenarios creates common performance testing scenarios
func GenerateLoadTestScenarios() []LoadTestScenario {
	return []LoadTestScenario{
		{
			Name:          "Baseline Load",
			Description:   "Normal application load during business hours",
			RequestRate:   100,
			Duration:      30 * time.Second,
			CacheHitRatio: 0.80,
			UserPattern:   "normal",
			Concurrency:   10,
		},
		{
			Name:          "Peak Traffic",
			Description:   "High traffic during peak usage periods",
			RequestRate:   500,
			Duration:      60 * time.Second,
			CacheHitRatio: 0.75,
			UserPattern:   "zipf",
			Concurrency:   25,
		},
		{
			Name:          "Burst Load",
			Description:   "Sudden traffic spike (viral content, marketing campaign)",
			RequestRate:   1000,
			Duration:      10 * time.Second,
			CacheHitRatio: 0.60,
			UserPattern:   "uniform",
			Concurrency:   50,
		},
		{
			Name:          "Sustained High Load",
			Description:   "Extended period of high traffic",
			RequestRate:   750,
			Duration:      120 * time.Second,
			CacheHitRatio: 0.70,
			UserPattern:   "zipf",
			Concurrency:   35,
		},
		{
			Name:          "Cache Cold Start",
			Description:   "Performance when cache is empty (system restart)",
			RequestRate:   200,
			Duration:      45 * time.Second,
			CacheHitRatio: 0.20, // Low hit rate as cache builds up
			UserPattern:   "normal",
			Concurrency:   15,
		},
		{
			Name:          "Memory Pressure",
			Description:   "High memory usage scenario with cache evictions",
			RequestRate:   300,
			Duration:      90 * time.Second,
			CacheHitRatio: 0.65,
			UserPattern:   "uniform",
			Concurrency:   20,
		},
	}
}

// SelectTokenForRequest chooses a token based on realistic usage patterns
func (tdg *TestDataGenerator) SelectTokenForRequest(tokens []TestToken, pattern string) TestToken {
	switch pattern {
	case "uniform":
		return tdg.selectUniformToken(tokens)
	case "normal":
		return tdg.selectNormalDistributionToken(tokens)
	case "zipf":
		return tdg.selectZipfToken(tokens)
	default:
		return tdg.selectUniformToken(tokens)
	}
}

// generateTokenHash creates a realistic token hash
func (tdg *TestDataGenerator) generateTokenHash(userID, tokenID int) string {
	// Generate a hash similar to what cache.HashToken would produce
	_ = fmt.Sprintf("jwt-token-user-%d-session-%d-%d", userID, tokenID, time.Now().UnixNano())
	
	// Create SHA256-like hash
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// selectRandomRole picks a random role with weighted distribution
func (tdg *TestDataGenerator) selectRandomRole() string {
	weights := []float64{0.70, 0.15, 0.08, 0.05, 0.02} // user, admin, moderator, premium, basic
	
	r := mathrand.Float64()
	cumulative := 0.0
	
	for i, weight := range weights {
		cumulative += weight
		if r <= cumulative {
			return tdg.Roles[i]
		}
	}
	
	return tdg.Roles[0] // fallback to "user"
}

// selectWeightedRole picks roles for heavy users (more likely to be premium/admin)
func (tdg *TestDataGenerator) selectWeightedRole() string {
	weights := []float64{0.40, 0.25, 0.15, 0.15, 0.05} // Different distribution for heavy users
	
	r := mathrand.Float64()
	cumulative := 0.0
	
	for i, weight := range weights {
		cumulative += weight
		if r <= cumulative {
			return tdg.Roles[i]
		}
	}
	
	return tdg.Roles[0]
}

// calculateHitProbability determines cache hit probability based on user and token patterns
func (tdg *TestDataGenerator) calculateHitProbability(userID, tokenID int) float64 {
	// Recent tokens are more likely to be cache hits
	recencyFactor := math.Max(0.1, 1.0-float64(tokenID)/float64(tdg.TokensPerUser))
	
	// Some users are more active (higher base hit rate)
	userFactor := 0.5 + 0.4*mathrand.Float64()
	
	return math.Min(0.95, recencyFactor*userFactor)
}

// selectUniformToken randomly selects a token with uniform distribution
func (tdg *TestDataGenerator) selectUniformToken(tokens []TestToken) TestToken {
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(tokens))))
	return tokens[n.Int64()]
}

// selectNormalDistributionToken selects token with normal distribution (most requests to middle users)
func (tdg *TestDataGenerator) selectNormalDistributionToken(tokens []TestToken) TestToken {
	// Approximate normal distribution using sum of uniform random numbers
	sum := 0.0
	for i := 0; i < 12; i++ {
		sum += mathrand.Float64()
	}
	normalized := (sum - 6.0) / 6.0 // Mean=0, roughly stddev=1
	
	// Map to token index
	center := len(tokens) / 2
	index := center + int(normalized*float64(len(tokens))/6)
	
	// Clamp to valid range
	if index < 0 {
		index = 0
	}
	if index >= len(tokens) {
		index = len(tokens) - 1
	}
	
	return tokens[index]
}

// selectZipfToken selects token following Zipf distribution (80/20 rule)
func (tdg *TestDataGenerator) selectZipfToken(tokens []TestToken) TestToken {
	// Simple Zipf approximation: 80% of requests go to first 20% of tokens
	r := mathrand.Float64()
	
	var index int
	if r < 0.8 {
		// 80% of requests go to first 20% of tokens
		first20Percent := len(tokens) / 5
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(first20Percent)))
		index = int(n.Int64())
	} else {
		// 20% of requests go to remaining 80% of tokens
		remaining := len(tokens) - (len(tokens) / 5)
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(remaining)))
		index = (len(tokens) / 5) + int(n.Int64())
	}
	
	return tokens[index]
}
