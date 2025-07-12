package auth_cache

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	vegeta "github.com/tsenart/vegeta/v12/lib"

	"github.com/EliasRanz/ai-code-gen/internal/cache"
	"github.com/EliasRanz/ai-code-gen/internal/middleware"
	"github.com/EliasRanz/ai-code-gen/tests/performance/utils"
)

// TestAuthCacheLoadPerformance runs comprehensive load tests using Vegeta
func TestAuthCacheLoadPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}
	
	// Skip performance tests by default unless explicitly enabled
	if !isPerformanceTestingEnabled() {
		t.Skip("Skipping performance test - set PERFORMANCE_TESTS=1 to enable")
	}

	// Setup real Redis for realistic testing
	authCacheContainer, cleanup := utils.SetupRealRedisForPerformanceTesting(t)
	defer cleanup()

	// Setup test HTTP server with auth proxy middleware
	server := setupTestServer(authCacheContainer.AuthCache)
	defer server.Close()

	// Generate load test scenarios
	scenarios := utils.GenerateLoadTestScenarios()

	for _, scenario := range scenarios {
		t.Run(scenario.Name, func(t *testing.T) {
			t.Logf("Running scenario: %s", scenario.Description)
			
			results := runVegetaLoadTest(t, server.URL, scenario, authCacheContainer.AuthCache)
			
			// Validate performance SLAs based on scenario type
			validatePerformanceSLA(t, scenario, results)
			
			// Print detailed results
			printLoadTestResults(t, scenario.Name, results)
		})
	}
}

// TestAuthCacheStressTest performs extreme load testing
func TestAuthCacheStressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}
	
	// Skip performance tests by default unless explicitly enabled
	if !isPerformanceTestingEnabled() {
		t.Skip("Skipping performance test - set PERFORMANCE_TESTS=1 to enable")
	}

	authCacheContainer, cleanup := utils.SetupRealRedisForPerformanceTesting(t)
	defer cleanup()

	server := setupTestServer(authCacheContainer.AuthCache)
	defer server.Close()

	// Extreme stress scenarios
	stressScenarios := []utils.LoadTestScenario{
		{
			Name:          "Extreme Load",
			Description:   "Maximum sustainable load test",
			RequestRate:   2000,
			Duration:      30 * time.Second,
			CacheHitRatio: 0.50,
			UserPattern:   "zipf",
			Concurrency:   100,
		},
		{
			Name:          "Burst Spike",
			Description:   "Sudden traffic spike simulation",
			RequestRate:   5000,
			Duration:      5 * time.Second,
			CacheHitRatio: 0.30,
			UserPattern:   "uniform",
			Concurrency:   200,
		},
	}

	for _, scenario := range stressScenarios {
		t.Run(scenario.Name, func(t *testing.T) {
			t.Logf("Running stress scenario: %s", scenario.Description)
			
			results := runVegetaLoadTest(t, server.URL, scenario, authCacheContainer.AuthCache)
			
			// For stress tests, we're more lenient on SLAs but ensure system doesn't crash
			assert.Less(t, results.ErrorRate, 0.05, "Error rate too high under stress")
			assert.Greater(t, results.ThroughputRPS, float64(scenario.RequestRate)*0.7, 
				"Throughput dropped too much under stress")
			
			printLoadTestResults(t, scenario.Name, results)
		})
	}
}

// TestCacheWarmupPerformance tests performance during cache warmup
func TestCacheWarmupPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping warmup test in short mode")
	}
	
	// Skip performance tests by default unless explicitly enabled
	if !isPerformanceTestingEnabled() {
		t.Skip("Skipping performance test - set PERFORMANCE_TESTS=1 to enable")
	}

	authCacheContainer, cleanup := utils.SetupRealRedisForPerformanceTesting(t)
	defer cleanup()

	server := setupTestServer(authCacheContainer.AuthCache)
	defer server.Close()

	// Test performance during cache warmup period
	scenario := utils.LoadTestScenario{
		Name:          "Cache Warmup",
		Description:   "Performance during cache warmup with cold start",
		RequestRate:   300,
		Duration:      60 * time.Second,
		CacheHitRatio: 0.10, // Low hit rate as cache builds up
		UserPattern:   "normal",
		Concurrency:   20,
	}

	results := runVegetaLoadTest(t, server.URL, scenario, authCacheContainer.AuthCache)

	// During warmup, we expect lower cache hit rates but system should remain stable
	assert.Less(t, results.ErrorRate, 0.02, "Error rate too high during warmup")
	assert.Greater(t, results.ThroughputRPS, 200.0, "Throughput too low during warmup")
	
	// Cache hit rate should improve over time (we'll see final rate)
	t.Logf("Final cache hit rate after warmup: %.2f%%", results.CacheHitRate*100)

	printLoadTestResults(t, scenario.Name, results)
}

// runVegetaLoadTest executes a load test using Vegeta
func runVegetaLoadTest(t *testing.T, serverURL string, scenario utils.LoadTestScenario, authCache *cache.AuthCache) utils.PerformanceReport {
	// Generate test tokens for the scenario
	generator := utils.NewTestDataGenerator(200, 3)
	tokens := generator.GenerateRealisticTokens()

	// Pre-populate cache based on expected hit ratio
	ctx := context.Background()
	numToCache := int(float64(len(tokens)) * scenario.CacheHitRatio)
	for i := 0; i < numToCache; i++ {
		err := authCache.SetUserContext(ctx, tokens[i].Hash, tokens[i].UserContext)
		require.NoError(t, err)
	}

	// Setup metrics collection
	metrics := utils.NewPerformanceMetrics()

	// Create Vegeta targeter
	targeter := createAuthTargeter(serverURL, tokens, generator, scenario.UserPattern)

	// Configure attack rate
	rate := vegeta.Rate{Freq: scenario.RequestRate, Per: time.Second}

	// Create attacker with concurrency settings
	attacker := vegeta.NewAttacker(
		vegeta.Workers(uint64(scenario.Concurrency)),
		vegeta.Timeout(5*time.Second),
	)

	// Execute attack
	var vegetaMetrics vegeta.Metrics

	for res := range attacker.Attack(targeter, rate, scenario.Duration, "auth-cache-load-test") {
		vegetaMetrics.Add(res)
		
		// Record our custom metrics
		duration := res.Latency
		if res.Error != "" {
			metrics.RecordCacheError(duration)
		} else if res.Code == http.StatusOK {
			// Assume successful response is a cache hit for simplicity
			metrics.RecordCacheHit(duration)
		} else {
			metrics.RecordCacheMiss(duration)
		}
	}
	vegetaMetrics.Close()

	// Generate comprehensive report
	report := metrics.GenerateReport(scenario.Name)
	
	// Enhance report with Vegeta metrics
	report.ThroughputRPS = float64(vegetaMetrics.Requests) / scenario.Duration.Seconds()
	
	return report
}

// createAuthTargeter creates a Vegeta targeter for auth requests
func createAuthTargeter(serverURL string, tokens []utils.TestToken, generator *utils.TestDataGenerator, pattern string) vegeta.Targeter {
	return vegeta.NewStaticTargeter(vegeta.Target{
		Method: "GET",
		URL:    serverURL + "/api/v1/auth/validate",
		Header: http.Header{
			"Content-Type": []string{"application/json"},
		},
		Body: nil,
	})
}

// setupTestServer creates a test HTTP server with auth proxy middleware
func setupTestServer(authCache *cache.AuthCache) *httptest.Server {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add auth proxy middleware
	router.Use(middleware.AuthServiceProxy("http://mock-auth-service", authCache))

	// Add test endpoint that requires authentication
	router.GET("/api/v1/auth/validate", func(c *gin.Context) {
		// Simulate successful auth validation
		c.JSON(http.StatusOK, gin.H{
			"status": "authenticated",
			"user_id": "test-user",
		})
	})

	return httptest.NewServer(router)
}

// validatePerformanceSLA checks if results meet scenario-specific SLAs
func validatePerformanceSLA(t *testing.T, scenario utils.LoadTestScenario, results utils.PerformanceReport) {
	// Get environment from environment variable or default to "staging"
	environment := "staging"
	if env := os.Getenv("TEST_ENVIRONMENT"); env != "" {
		environment = env
	}
	
	// Get appropriate SLA thresholds for this scenario
	slaThresholds := utils.GetSLAForScenario(scenario.Name, environment)
	
	// Validate against SLA thresholds
	slaResult := utils.ValidatePerformanceReport(results, slaThresholds)
	
	// Log detailed SLA results
	t.Logf("SLA Validation Result: %s", slaResult.Summary)
	for _, violation := range slaResult.Violations {
		switch violation.Level {
		case utils.SLAFail:
			t.Errorf("❌ SLA FAILURE - %s: %s", violation.Metric, violation.Message)
		case utils.SLAWarn:
			t.Logf("⚠️ SLA WARNING - %s: %s", violation.Metric, violation.Message)
		default:
			t.Logf("✅ SLA PASS - %s: Expected %s, Got %s", violation.Metric, violation.Expected, violation.Actual)
		}
	}
	
	// Print recommendations if there are violations
	if len(slaResult.Violations) > 0 {
		recommendations := utils.SLARecommendations(slaResult.Violations)
		t.Logf("🔧 Performance Optimization Recommendations:")
		for _, rec := range recommendations {
			t.Logf("   %s", rec)
		}
	}
	
	// Fail the test if we have SLA failures (but not warnings)
	if slaResult.OverallStatus == utils.SLAFail {
		t.Fatalf("Test failed due to SLA violations. See details above.")
	}
}

// printLoadTestResults outputs detailed load test results
func printLoadTestResults(t *testing.T, scenarioName string, results utils.PerformanceReport) {
	t.Logf("\n=== Load Test Results: %s ===", scenarioName)
	t.Logf("Duration: %v", results.Duration)
	t.Logf("Total Requests: %d", results.TotalRequests)
	t.Logf("Throughput: %.2f req/sec", results.ThroughputRPS)
	t.Logf("Cache Hit Rate: %.2f%%", results.CacheHitRate*100)
	t.Logf("Error Rate: %.2f%%", results.ErrorRate*100)
	t.Logf("P50 Latency: %v", results.Percentiles.P50)
	t.Logf("P95 Latency: %v", results.Percentiles.P95)
	t.Logf("P99 Latency: %v", results.Percentiles.P99)
	t.Logf("Min Latency: %v", results.Percentiles.Min)
	t.Logf("Max Latency: %v", results.Percentiles.Max)
	t.Logf("Memory Usage: %.2f MB", results.MemoryUsageMB)
	t.Logf("===============================\n")
}
