package utils

import (
	"fmt"
	"time"
)

// SLAThresholds defines performance SLA thresholds for different scenarios
type SLAThresholds struct {
	// Latency thresholds
	P50Latency  time.Duration `json:"p50_latency"`
	P95Latency  time.Duration `json:"p95_latency"`
	P99Latency  time.Duration `json:"p99_latency"`
	P999Latency time.Duration `json:"p999_latency"`

	// Performance thresholds
	MinThroughputRPS float64 `json:"min_throughput_rps"`
	MinCacheHitRate  float64 `json:"min_cache_hit_rate"`
	MaxErrorRate     float64 `json:"max_error_rate"`

	// Resource thresholds
	MaxMemoryUsageMB float64 `json:"max_memory_usage_mb"`

	// Labels for reporting
	ScenarioName string `json:"scenario_name"`
	Environment  string `json:"environment"` // "production", "staging", "development"
}

// SLALevel represents the severity of SLA violation
type SLALevel int

const (
	SLAPass SLALevel = iota
	SLAWarn
	SLAFail
)

func (s SLALevel) String() string {
	switch s {
	case SLAPass:
		return "PASS"
	case SLAWarn:
		return "WARN"
	case SLAFail:
		return "FAIL"
	default:
		return "UNKNOWN"
	}
}

// SLAViolation represents a specific SLA violation
type SLAViolation struct {
	Metric   string   `json:"metric"`
	Expected string   `json:"expected"`
	Actual   string   `json:"actual"`
	Level    SLALevel `json:"level"`
	Message  string   `json:"message"`
}

// SLAValidationResult contains the complete SLA validation results
type SLAValidationResult struct {
	OverallStatus SLALevel       `json:"overall_status"`
	Violations    []SLAViolation `json:"violations"`
	Summary       string         `json:"summary"`
}

// DefaultSLAConfigurations returns industry-standard SLA configurations
func DefaultSLAConfigurations() map[string]SLAThresholds {
	return map[string]SLAThresholds{
		// Production environment - strict SLAs
		"production_baseline": {
			P50Latency:       1 * time.Millisecond,
			P95Latency:       5 * time.Millisecond,
			P99Latency:       15 * time.Millisecond,
			P999Latency:      35 * time.Millisecond,
			MinThroughputRPS: 500,
			MinCacheHitRate:  0.85,  // 85%
			MaxErrorRate:     0.001, // 0.1%
			MaxMemoryUsageMB: 100,
			ScenarioName:     "Production Baseline",
			Environment:      "production",
		},

		"production_peak": {
			P50Latency:       2 * time.Millisecond,
			P95Latency:       10 * time.Millisecond,
			P99Latency:       25 * time.Millisecond,
			P999Latency:      50 * time.Millisecond,
			MinThroughputRPS: 1000,
			MinCacheHitRate:  0.80,  // 80%
			MaxErrorRate:     0.005, // 0.5%
			MaxMemoryUsageMB: 150,
			ScenarioName:     "Production Peak Traffic",
			Environment:      "production",
		},

		"production_burst": {
			P50Latency:       3 * time.Millisecond,
			P95Latency:       20 * time.Millisecond,
			P99Latency:       40 * time.Millisecond,
			P999Latency:      80 * time.Millisecond,
			MinThroughputRPS: 2000,
			MinCacheHitRate:  0.70, // 70%
			MaxErrorRate:     0.02, // 2%
			MaxMemoryUsageMB: 200,
			ScenarioName:     "Production Burst Load",
			Environment:      "production",
		},

		// Staging environment - moderate SLAs
		"staging_baseline": {
			P50Latency:       2 * time.Millisecond,
			P95Latency:       10 * time.Millisecond,
			P99Latency:       25 * time.Millisecond,
			P999Latency:      50 * time.Millisecond,
			MinThroughputRPS: 300,
			MinCacheHitRate:  0.75, // 75%
			MaxErrorRate:     0.01, // 1%
			MaxMemoryUsageMB: 150,
			ScenarioName:     "Staging Baseline",
			Environment:      "staging",
		},

		// Development environment - relaxed SLAs
		"development_baseline": {
			P50Latency:       5 * time.Millisecond,
			P95Latency:       20 * time.Millisecond,
			P99Latency:       50 * time.Millisecond,
			P999Latency:      100 * time.Millisecond,
			MinThroughputRPS: 100,
			MinCacheHitRate:  0.60, // 60%
			MaxErrorRate:     0.05, // 5%
			MaxMemoryUsageMB: 200,
			ScenarioName:     "Development Baseline",
			Environment:      "development",
		},

		// Special scenarios
		"cache_warmup": {
			P50Latency:       5 * time.Millisecond,
			P95Latency:       25 * time.Millisecond,
			P99Latency:       60 * time.Millisecond,
			P999Latency:      120 * time.Millisecond,
			MinThroughputRPS: 200,
			MinCacheHitRate:  0.40, // 40% - lower during warmup
			MaxErrorRate:     0.03, // 3%
			MaxMemoryUsageMB: 250,
			ScenarioName:     "Cache Warmup",
			Environment:      "production",
		},

		"stress_test": {
			P50Latency:       10 * time.Millisecond,
			P95Latency:       50 * time.Millisecond,
			P99Latency:       100 * time.Millisecond,
			P999Latency:      200 * time.Millisecond,
			MinThroughputRPS: 1500,
			MinCacheHitRate:  0.50, // 50% - degraded under stress
			MaxErrorRate:     0.10, // 10% - acceptable under extreme stress
			MaxMemoryUsageMB: 500,
			ScenarioName:     "Stress Test",
			Environment:      "testing",
		},
	}
}

// ValidatePerformanceReport validates a performance report against SLA thresholds
func ValidatePerformanceReport(report PerformanceReport, thresholds SLAThresholds) SLAValidationResult {
	violations := []SLAViolation{}
	overallStatus := SLAPass

	// Validate P95 Latency (most critical metric)
	if report.Percentiles.P95 > thresholds.P95Latency {
		level := determineSeverity(
			float64(report.Percentiles.P95),
			float64(thresholds.P95Latency),
			1.2, // 20% tolerance for WARN
			1.5, // 50% tolerance for FAIL
		)
		violations = append(violations, SLAViolation{
			Metric:   "P95 Latency",
			Expected: fmt.Sprintf("< %v", thresholds.P95Latency),
			Actual:   fmt.Sprintf("%v", report.Percentiles.P95),
			Level:    level,
			Message:  fmt.Sprintf("P95 latency %v exceeds threshold %v", report.Percentiles.P95, thresholds.P95Latency),
		})
		if level > overallStatus {
			overallStatus = level
		}
	}

	// Validate P99 Latency
	if report.Percentiles.P99 > thresholds.P99Latency {
		level := determineSeverity(
			float64(report.Percentiles.P99),
			float64(thresholds.P99Latency),
			1.3, // 30% tolerance for WARN
			2.0, // 100% tolerance for FAIL
		)
		violations = append(violations, SLAViolation{
			Metric:   "P99 Latency",
			Expected: fmt.Sprintf("< %v", thresholds.P99Latency),
			Actual:   fmt.Sprintf("%v", report.Percentiles.P99),
			Level:    level,
			Message:  fmt.Sprintf("P99 latency %v exceeds threshold %v", report.Percentiles.P99, thresholds.P99Latency),
		})
		if level > overallStatus {
			overallStatus = level
		}
	}

	// Validate Throughput
	if report.ThroughputRPS < thresholds.MinThroughputRPS {
		level := determineSeverity(
			report.ThroughputRPS,
			thresholds.MinThroughputRPS,
			0.8, // 20% below for WARN
			0.5, // 50% below for FAIL
		)
		violations = append(violations, SLAViolation{
			Metric:   "Throughput",
			Expected: fmt.Sprintf(">= %.1f req/s", thresholds.MinThroughputRPS),
			Actual:   fmt.Sprintf("%.1f req/s", report.ThroughputRPS),
			Level:    level,
			Message:  fmt.Sprintf("Throughput %.1f req/s below threshold %.1f req/s", report.ThroughputRPS, thresholds.MinThroughputRPS),
		})
		if level > overallStatus {
			overallStatus = level
		}
	}

	// Validate Cache Hit Rate
	if report.CacheHitRate < thresholds.MinCacheHitRate {
		level := determineSeverity(
			report.CacheHitRate,
			thresholds.MinCacheHitRate,
			0.9, // 10% below for WARN
			0.7, // 30% below for FAIL
		)
		violations = append(violations, SLAViolation{
			Metric:   "Cache Hit Rate",
			Expected: fmt.Sprintf(">= %.1f%%", thresholds.MinCacheHitRate*100),
			Actual:   fmt.Sprintf("%.1f%%", report.CacheHitRate*100),
			Level:    level,
			Message:  fmt.Sprintf("Cache hit rate %.1f%% below threshold %.1f%%", report.CacheHitRate*100, thresholds.MinCacheHitRate*100),
		})
		if level > overallStatus {
			overallStatus = level
		}
	}

	// Validate Error Rate
	if report.ErrorRate > thresholds.MaxErrorRate {
		level := determineSeverity(
			report.ErrorRate,
			thresholds.MaxErrorRate,
			2.0, // 100% above for WARN
			5.0, // 400% above for FAIL
		)
		violations = append(violations, SLAViolation{
			Metric:   "Error Rate",
			Expected: fmt.Sprintf("<= %.2f%%", thresholds.MaxErrorRate*100),
			Actual:   fmt.Sprintf("%.2f%%", report.ErrorRate*100),
			Level:    level,
			Message:  fmt.Sprintf("Error rate %.2f%% exceeds threshold %.2f%%", report.ErrorRate*100, thresholds.MaxErrorRate*100),
		})
		if level > overallStatus {
			overallStatus = level
		}
	}

	// Validate Memory Usage
	if report.MemoryUsageMB > thresholds.MaxMemoryUsageMB {
		level := determineSeverity(
			report.MemoryUsageMB,
			thresholds.MaxMemoryUsageMB,
			1.2, // 20% above for WARN
			1.5, // 50% above for FAIL
		)
		violations = append(violations, SLAViolation{
			Metric:   "Memory Usage",
			Expected: fmt.Sprintf("<= %.1f MB", thresholds.MaxMemoryUsageMB),
			Actual:   fmt.Sprintf("%.1f MB", report.MemoryUsageMB),
			Level:    level,
			Message:  fmt.Sprintf("Memory usage %.1f MB exceeds threshold %.1f MB", report.MemoryUsageMB, thresholds.MaxMemoryUsageMB),
		})
		if level > overallStatus {
			overallStatus = level
		}
	}

	// Generate summary
	summary := generateSLASummary(overallStatus, violations, thresholds)

	return SLAValidationResult{
		OverallStatus: overallStatus,
		Violations:    violations,
		Summary:       summary,
	}
}

// determineSeverity calculates violation severity based on tolerance levels
func determineSeverity(actual, threshold, warnTolerance, failTolerance float64) SLALevel {
	if actual <= threshold {
		return SLAPass
	}

	ratio := actual / threshold
	if ratio <= warnTolerance {
		return SLAPass
	} else if ratio <= failTolerance {
		return SLAWarn
	} else {
		return SLAFail
	}
}

// generateSLASummary creates a human-readable summary of SLA validation
func generateSLASummary(status SLALevel, violations []SLAViolation, thresholds SLAThresholds) string {
	if len(violations) == 0 {
		return fmt.Sprintf("✅ All SLAs met for %s (%s environment)",
			thresholds.ScenarioName, thresholds.Environment)
	}

	failCount := 0
	warnCount := 0
	for _, v := range violations {
		switch v.Level {
		case SLAFail:
			failCount++
		case SLAWarn:
			warnCount++
		}
	}

	switch status {
	case SLAWarn:
		return fmt.Sprintf("⚠️ SLA warnings detected for %s: %d warnings out of %d metrics",
			thresholds.ScenarioName, warnCount, len(violations))
	case SLAFail:
		return fmt.Sprintf("❌ SLA failures detected for %s: %d failures, %d warnings out of %d metrics",
			thresholds.ScenarioName, failCount, warnCount, len(violations))
	default:
		return fmt.Sprintf("✅ All SLAs met for %s", thresholds.ScenarioName)
	}
}

// GetSLAForScenario returns appropriate SLA thresholds for a test scenario
func GetSLAForScenario(scenarioName, environment string) SLAThresholds {
	configs := DefaultSLAConfigurations()

	// Try exact match first
	key := fmt.Sprintf("%s_%s", environment, scenarioName)
	if config, exists := configs[key]; exists {
		return config
	}

	// Try scenario-specific matches
	for configKey, config := range configs {
		if config.ScenarioName == scenarioName || configKey == scenarioName {
			return config
		}
	}

	// Fallback to environment default
	switch environment {
	case "production":
		return configs["production_baseline"]
	case "staging":
		return configs["staging_baseline"]
	case "development":
		return configs["development_baseline"]
	default:
		return configs["development_baseline"] // Most permissive as fallback
	}
}

// SLARecommendations provides optimization recommendations based on violations
func SLARecommendations(violations []SLAViolation) []string {
	recommendations := []string{}

	hasLatencyIssues := false
	hasThroughputIssues := false
	hasCacheIssues := false
	hasErrorIssues := false

	for _, v := range violations {
		switch v.Metric {
		case "P95 Latency", "P99 Latency":
			hasLatencyIssues = true
		case "Throughput":
			hasThroughputIssues = true
		case "Cache Hit Rate":
			hasCacheIssues = true
		case "Error Rate":
			hasErrorIssues = true
		}
	}

	if hasLatencyIssues {
		recommendations = append(recommendations,
			"🚀 Optimize Redis connection pooling and consider Redis cluster for better latency",
			"⚡ Review Redis configuration (memory policy, persistence settings)",
			"🔍 Analyze slow queries and optimize data structures")
	}

	if hasThroughputIssues {
		recommendations = append(recommendations,
			"📈 Scale Redis horizontally with read replicas",
			"🔧 Optimize connection pool size and concurrent connections",
			"⚖️ Consider load balancing across multiple Redis instances")
	}

	if hasCacheIssues {
		recommendations = append(recommendations,
			"🎯 Review cache TTL settings and eviction policies",
			"📊 Analyze access patterns and implement cache warming strategies",
			"🔄 Consider implementing cache preloading for hot data")
	}

	if hasErrorIssues {
		recommendations = append(recommendations,
			"🛡️ Implement circuit breaker pattern for Redis failures",
			"🔍 Monitor Redis health and connection stability",
			"⚠️ Add fallback mechanisms for auth validation")
	}

	return recommendations
}
