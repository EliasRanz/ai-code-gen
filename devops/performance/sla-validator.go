package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/EliasRanz/ai-code-gen/tests/performance/utils"
)

// SLA Configuration Validator - DevOps tool for performance SLA management
func main() {
	fmt.Println("🎯 Redis Auth Cache SLA Configuration Validator")
	fmt.Println(strings.Repeat("=", 60))

	// Parse command line arguments
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "list":
			listAvailableSLAs()
		case "validate":
			if len(os.Args) < 3 {
				fmt.Println("Usage: sla-validator validate <environment>")
				os.Exit(1)
			}
			validateEnvironmentSLAs(os.Args[2])
		case "compare":
			compareEnvironmentSLAs()
		case "recommend":
			if len(os.Args) < 3 {
				fmt.Println("Usage: sla-validator recommend <scenario>")
				os.Exit(1)
			}
			recommendSLAConfiguration(os.Args[2])
		default:
			showUsage()
		}
	} else {
		// Default: Show comprehensive overview
		showSLAOverview()
	}
}

// titleCase provides simple title case conversion
func titleCase(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(string(s[0])) + strings.ToLower(s[1:])
}

func showUsage() {
	fmt.Println("📋 SLA Configuration Validator - DevOps Tool")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Printf("  %s [command] [args]\n", os.Args[0])
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  list                     - List all available SLA configurations")
	fmt.Println("  validate <environment>   - Validate SLAs for specific environment")
	fmt.Println("  compare                  - Compare SLAs across environments")
	fmt.Println("  recommend <scenario>     - Get SLA recommendations for scenario")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Printf("  %s list\n", os.Args[0])
	fmt.Printf("  %s validate production\n", os.Args[0])
	fmt.Printf("  %s recommend baseline\n", os.Args[0])
}

func showSLAOverview() {
	fmt.Println("📊 SLA Configuration Overview")
	fmt.Println("")

	// Show current environment detection
	environment := detectEnvironment()
	fmt.Printf("🌍 Detected Environment: %s\n", environment)
	fmt.Println("")

	// Show SLA validation examples
	demonstrateSLAValidation()

	// Show environment comparison
	fmt.Println("\n" + strings.Repeat("-", 60))
	compareEnvironmentSLAs()
}

func listAvailableSLAs() {
	fmt.Println("📋 Available SLA Configurations:")
	fmt.Println("")

	configs := utils.DefaultSLAConfigurations()

	// Group by environment
	environments := map[string][]utils.SLAThresholds{
		"production":  {},
		"staging":     {},
		"development": {},
		"testing":     {},
	}

	for _, config := range configs {
		environments[config.Environment] = append(environments[config.Environment], config)
	}

	for env, configs := range environments {
		if len(configs) == 0 {
			continue
		}

		fmt.Printf("🔹 %s Environment:\n", titleCase(env))
		for _, config := range configs {
			fmt.Printf("  📊 %s\n", config.ScenarioName)
			fmt.Printf("     P95 Latency: < %v\n", config.P95Latency)
			fmt.Printf("     Cache Hit Rate: > %.1f%%\n", config.MinCacheHitRate*100)
			fmt.Printf("     Error Rate: < %.2f%%\n", config.MaxErrorRate*100)
			fmt.Printf("     Min Throughput: > %.0f req/s\n", config.MinThroughputRPS)
			fmt.Println()
		}
	}
}

func validateEnvironmentSLAs(environment string) {
	fmt.Printf("🔍 Validating SLAs for %s environment\n", environment)
	fmt.Println("")

	// Get baseline SLA for environment
	sla := utils.GetSLAForScenario("baseline", environment)

	fmt.Printf("📊 %s SLA Thresholds:\n", sla.ScenarioName)
	fmt.Printf("   P95 Latency: < %v\n", sla.P95Latency)
	fmt.Printf("   P99 Latency: < %v\n", sla.P99Latency)
	fmt.Printf("   Cache Hit Rate: > %.1f%%\n", sla.MinCacheHitRate*100)
	fmt.Printf("   Error Rate: < %.3f%%\n", sla.MaxErrorRate*100)
	fmt.Printf("   Min Throughput: > %.0f req/s\n", sla.MinThroughputRPS)
	fmt.Printf("   Max Memory: < %.0f MB\n", sla.MaxMemoryUsageMB)
	fmt.Println("")

	// Simulate validation with hypothetical performance data
	testReport := generateTestPerformanceReport(environment)
	result := utils.ValidatePerformanceReport(testReport, sla)

	fmt.Printf("✅ Validation Result: %s\n", result.Summary)

	if len(result.Violations) > 0 {
		fmt.Println("\n📋 SLA Violations Analysis:")
		for _, violation := range result.Violations {
			statusIcon := map[utils.SLALevel]string{
				utils.SLAPass: "✅",
				utils.SLAWarn: "⚠️",
				utils.SLAFail: "❌",
			}

			icon := statusIcon[violation.Level]
			fmt.Printf("   %s %s: Expected %s, Simulated %s\n",
				icon, violation.Metric, violation.Expected, violation.Actual)
		}

		// Show recommendations
		recommendations := utils.SLARecommendations(result.Violations)
		if len(recommendations) > 0 {
			fmt.Println("\n💡 Optimization Recommendations:")
			for _, rec := range recommendations {
				fmt.Printf("   %s\n", rec)
			}
		}
	}
}

func compareEnvironmentSLAs() {
	fmt.Println("🔄 SLA Comparison Across Environments")
	fmt.Println("")

	environments := []string{"production", "staging", "development"}

	fmt.Printf("%-20s %-15s %-15s %-15s %-15s\n", "Environment", "P95 Latency", "Cache Hit Rate", "Error Rate", "Throughput")
	fmt.Println(strings.Repeat("-", 80))

	for _, env := range environments {
		sla := utils.GetSLAForScenario("baseline", env)
		fmt.Printf("%-20s %-15v %-15.1f%% %-15.3f%% %-15.0f req/s\n",
			titleCase(env),
			sla.P95Latency,
			sla.MinCacheHitRate*100,
			sla.MaxErrorRate*100,
			sla.MinThroughputRPS)
	}

	fmt.Println("")
	fmt.Println("📈 Observations:")
	fmt.Println("   • Production has strictest latency requirements (5ms vs 20ms dev)")
	fmt.Println("   • Cache hit rate expectations decrease in lower environments")
	fmt.Println("   • Error tolerance increases significantly in development")
	fmt.Println("   • Throughput requirements scale with environment criticality")
}

func recommendSLAConfiguration(scenario string) {
	fmt.Printf("💡 SLA Recommendations for '%s' scenario\n", scenario)
	fmt.Println("")

	// Get SLA configurations for this scenario across environments
	environments := []string{"production", "staging", "development"}

	fmt.Printf("Scenario: %s\n", scenario)
	fmt.Println(strings.Repeat("-", 40))

	for _, env := range environments {
		sla := utils.GetSLAForScenario(scenario, env)
		fmt.Printf("\n🔹 %s Environment:\n", titleCase(env))
		fmt.Printf("   Use Case: %s\n", getUseCaseDescription(env))
		fmt.Printf("   P95 Latency: < %v\n", sla.P95Latency)
		fmt.Printf("   Cache Hit Rate: > %.1f%%\n", sla.MinCacheHitRate*100)
		fmt.Printf("   Error Rate: < %.3f%%\n", sla.MaxErrorRate*100)
		fmt.Printf("   Min Throughput: > %.0f req/s\n", sla.MinThroughputRPS)
	}

	fmt.Println("\n🎯 Implementation Guidance:")
	fmt.Println("   1. Start with development SLAs for initial testing")
	fmt.Println("   2. Validate with staging SLAs before production deployment")
	fmt.Println("   3. Monitor production SLAs continuously with alerting")
	fmt.Println("   4. Adjust thresholds based on actual traffic patterns")
}

func demonstrateSLAValidation() {
	fmt.Println("🧪 SLA Validation Examples:")

	examples := []struct {
		name   string
		status string
		icon   string
	}{
		{"High-performing system", "All SLAs met", "✅"},
		{"System with warnings", "Performance degradation detected", "⚠️"},
		{"System with failures", "Critical SLA violations", "❌"},
	}

	for _, example := range examples {
		fmt.Printf("\n%s %s: %s\n", example.icon, example.name, example.status)
	}
}

func detectEnvironment() string {
	if env := os.Getenv("TEST_ENVIRONMENT"); env != "" {
		return env
	}
	if env := os.Getenv("ENVIRONMENT"); env != "" {
		return env
	}
	if env := os.Getenv("NODE_ENV"); env != "" {
		return env
	}
	return "staging" // Default
}

func getUseCaseDescription(environment string) string {
	switch environment {
	case "production":
		return "Live user traffic, strict availability requirements"
	case "staging":
		return "Pre-production validation, moderate requirements"
	case "development":
		return "Development testing, relaxed requirements"
	default:
		return "General testing environment"
	}
}

func generateTestPerformanceReport(environment string) utils.PerformanceReport {
	// Generate realistic test data based on environment
	var report utils.PerformanceReport

	switch environment {
	case "production":
		report = utils.PerformanceReport{
			TestName:      "Production Baseline Simulation",
			TotalRequests: 10000,
			ThroughputRPS: 750,
			CacheHitRate:  0.87,   // 87%
			ErrorRate:     0.0008, // 0.08%
			Percentiles: utils.PerformancePercentiles{
				P95: 4 * time.Millisecond,
				P99: 12 * time.Millisecond,
			},
			MemoryUsageMB: 85,
			Timestamp:     time.Now(),
		}
	case "staging":
		report = utils.PerformanceReport{
			TestName:      "Staging Environment Simulation",
			TotalRequests: 5000,
			ThroughputRPS: 400,
			CacheHitRate:  0.78,  // 78%
			ErrorRate:     0.005, // 0.5%
			Percentiles: utils.PerformancePercentiles{
				P95: 8 * time.Millisecond,
				P99: 18 * time.Millisecond,
			},
			MemoryUsageMB: 95,
			Timestamp:     time.Now(),
		}
	default: // development
		report = utils.PerformanceReport{
			TestName:      "Development Environment Simulation",
			TotalRequests: 1000,
			ThroughputRPS: 150,
			CacheHitRate:  0.65, // 65%
			ErrorRate:     0.02, // 2%
			Percentiles: utils.PerformancePercentiles{
				P95: 15 * time.Millisecond,
				P99: 35 * time.Millisecond,
			},
			MemoryUsageMB: 120,
			Timestamp:     time.Now(),
		}
	}

	return report
}
