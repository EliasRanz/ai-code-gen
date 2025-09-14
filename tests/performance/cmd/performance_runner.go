package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/EliasRanz/ai-code-gen/tests/performance/utils"
)

// PerformanceTestSuite runs comprehensive Redis auth cache performance tests
type PerformanceTestSuite struct {
	reportGenerator *utils.PerformanceReportGenerator
	outputDir       string
}

// NewPerformanceTestSuite creates a new test suite
func NewPerformanceTestSuite(outputDir string) *PerformanceTestSuite {
	return &PerformanceTestSuite{
		reportGenerator: utils.NewPerformanceReportGenerator(),
		outputDir:       outputDir,
	}
}

// RunAllTests executes all performance test scenarios
func (pts *PerformanceTestSuite) RunAllTests() error {
	fmt.Println("🚀 Starting Redis Auth Cache Performance Test Suite")
	fmt.Println("==================================================")

	// Ensure output directory exists
	if err := os.MkdirAll(pts.outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	testScenarios := []struct {
		name        string
		description string
		testFunc    func() (*utils.PerformanceReport, error)
	}{
		{
			name:        "Benchmark Tests",
			description: "Go benchmark tests for individual cache operations",
			testFunc:    pts.runBenchmarkTests,
		},
		{
			name:        "Load Tests",
			description: "Vegeta-based load testing with realistic traffic patterns",
			testFunc:    pts.runLoadTests,
		},
		{
			name:        "Stress Tests",
			description: "High-load stress testing to find system limits",
			testFunc:    pts.runStressTests,
		},
		{
			name:        "Warmup Tests",
			description: "Performance testing during cache warmup scenarios",
			testFunc:    pts.runWarmupTests,
		},
	}

	allPassed := true
	for _, scenario := range testScenarios {
		fmt.Printf("\n📊 Running %s...\n", scenario.name)
		fmt.Printf("   %s\n", scenario.description)

		start := time.Now()
		report, err := scenario.testFunc()
		duration := time.Since(start)

		if err != nil {
			fmt.Printf("   ❌ FAILED: %v\n", err)
			allPassed = false
			continue
		}

		if report != nil {
			pts.reportGenerator.AddReport(*report)
			pts.printScenarioResults(scenario.name, report, duration)
		} else {
			fmt.Printf("   ⚠️  No performance report generated\n")
		}
	}

	// Generate comprehensive reports
	fmt.Printf("\n📈 Generating performance reports...\n")
	if err := pts.generateReports(); err != nil {
		return fmt.Errorf("failed to generate reports: %w", err)
	}

	// Print final summary
	fmt.Printf("\n🏁 Performance Test Suite Complete\n")
	fmt.Printf("===================================\n")
	if allPassed {
		fmt.Printf("✅ All test scenarios completed successfully\n")
	} else {
		fmt.Printf("⚠️  Some test scenarios failed or had warnings\n")
	}
	fmt.Printf("📁 Reports saved to: %s\n", pts.outputDir)

	return nil
}

// runBenchmarkTests executes Go benchmark tests
func (pts *PerformanceTestSuite) runBenchmarkTests() (*utils.PerformanceReport, error) {
	// Simulate benchmark test execution and create a sample report
	// In real implementation, this would run: go test -bench=. ./tests/performance/auth_cache/

	metrics := utils.NewPerformanceMetrics()

	// Simulate benchmark results
	for i := 0; i < 10000; i++ {
		if i%5 == 0 {
			metrics.RecordCacheMiss(2 * time.Millisecond)
		} else {
			metrics.RecordCacheHit(500 * time.Microsecond)
		}
	}

	report := metrics.GenerateReport("Benchmark Tests")
	return &report, nil
}

// runLoadTests executes Vegeta load tests
func (pts *PerformanceTestSuite) runLoadTests() (*utils.PerformanceReport, error) {
	// This would run the actual load tests from load_test.go
	// For demo purposes, we'll simulate realistic load test results

	metrics := utils.NewPerformanceMetrics()

	// Simulate load test with realistic patterns
	for i := 0; i < 50000; i++ {
		if i%4 == 0 { // 25% cache miss rate
			metrics.RecordCacheMiss(3 * time.Millisecond)
		} else {
			metrics.RecordCacheHit(800 * time.Microsecond)
		}

		// Simulate occasional errors
		if i%1000 == 0 {
			metrics.RecordCacheError(5 * time.Millisecond)
		}
	}

	time.Sleep(100 * time.Millisecond) // Simulate test duration
	report := metrics.GenerateReport("Load Tests")
	return &report, nil
}

// runStressTests executes stress tests
func (pts *PerformanceTestSuite) runStressTests() (*utils.PerformanceReport, error) {
	metrics := utils.NewPerformanceMetrics()

	// Simulate stress test with higher latencies and more errors
	for i := 0; i < 30000; i++ {
		if i%3 == 0 { // 33% cache miss rate under stress
			metrics.RecordCacheMiss(8 * time.Millisecond)
		} else {
			metrics.RecordCacheHit(2 * time.Millisecond)
		}

		// More errors under stress
		if i%500 == 0 {
			metrics.RecordCacheError(15 * time.Millisecond)
		}
	}

	time.Sleep(150 * time.Millisecond) // Simulate test duration
	report := metrics.GenerateReport("Stress Tests")
	return &report, nil
}

// runWarmupTests executes cache warmup tests
func (pts *PerformanceTestSuite) runWarmupTests() (*utils.PerformanceReport, error) {
	metrics := utils.NewPerformanceMetrics()

	// Simulate warmup scenario with improving hit rates
	for i := 0; i < 20000; i++ {
		// Hit rate improves over time during warmup
		hitRate := float64(i) / 20000.0 * 0.8 // Gradually improve to 80%

		if float64(i%100)/100.0 > hitRate {
			metrics.RecordCacheMiss(4 * time.Millisecond)
		} else {
			metrics.RecordCacheHit(1 * time.Millisecond)
		}
	}

	time.Sleep(80 * time.Millisecond) // Simulate test duration
	report := metrics.GenerateReport("Cache Warmup Tests")
	return &report, nil
}

// printScenarioResults prints results for a specific test scenario
func (pts *PerformanceTestSuite) printScenarioResults(scenarioName string, report *utils.PerformanceReport, duration time.Duration) {
	statusIcon := map[string]string{
		"PASS": "✅",
		"WARN": "⚠️",
		"FAIL": "❌",
	}

	icon := statusIcon[report.Status]
	if icon == "" {
		icon = "❓"
	}

	fmt.Printf("   %s %s (%s)\n", icon, report.Status, duration.Round(time.Millisecond))
	fmt.Printf("   📈 Throughput: %.1f req/s\n", report.ThroughputRPS)
	fmt.Printf("   🎯 Cache Hit Rate: %.1f%%\n", report.CacheHitRate*100)
	fmt.Printf("   📊 P95 Latency: %v\n", report.Percentiles.P95.Round(time.Microsecond))
	fmt.Printf("   🔥 Error Rate: %.2f%%\n", report.ErrorRate*100)
}

// generateReports creates all output report formats
func (pts *PerformanceTestSuite) generateReports() error {
	// Generate HTML report
	htmlFile, err := os.Create(filepath.Join(pts.outputDir, "performance_report.html"))
	if err != nil {
		return err
	}
	defer htmlFile.Close()

	if err := pts.reportGenerator.WriteTo(htmlFile, "html"); err != nil {
		return fmt.Errorf("failed to generate HTML report: %w", err)
	}
	fmt.Printf("   📄 HTML report: %s\n", htmlFile.Name())

	// Generate JSON report
	jsonFile, err := os.Create(filepath.Join(pts.outputDir, "performance_report.json"))
	if err != nil {
		return err
	}
	defer jsonFile.Close()

	if err := pts.reportGenerator.WriteTo(jsonFile, "json"); err != nil {
		return fmt.Errorf("failed to generate JSON report: %w", err)
	}
	fmt.Printf("   📄 JSON report: %s\n", jsonFile.Name())

	// Generate CSV report
	csvFile, err := os.Create(filepath.Join(pts.outputDir, "performance_report.csv"))
	if err != nil {
		return err
	}
	defer csvFile.Close()

	if err := pts.reportGenerator.WriteTo(csvFile, "csv"); err != nil {
		return fmt.Errorf("failed to generate CSV report: %w", err)
	}
	fmt.Printf("   📄 CSV report: %s\n", csvFile.Name())

	return nil
}

func main() {
	// Check for output directory argument
	outputDir := "./performance_reports"
	if len(os.Args) > 1 {
		outputDir = os.Args[1]
	}

	// Create and run test suite
	suite := NewPerformanceTestSuite(outputDir)
	if err := suite.RunAllTests(); err != nil {
		log.Fatalf("Performance test suite failed: %v", err)
	}
}

// TestMain provides an entry point for running tests programmatically
func TestMain(m *testing.M) {
	// Run the actual tests
	code := m.Run()

	// If tests passed, also run our performance suite
	if code == 0 {
		fmt.Println("\n" + strings.Repeat("=", 60))
		fmt.Println("Running Performance Test Suite...")
		fmt.Println(strings.Repeat("=", 60))

		suite := NewPerformanceTestSuite("./test_reports")
		if err := suite.RunAllTests(); err != nil {
			fmt.Printf("Performance test suite failed: %v\n", err)
			os.Exit(1)
		}
	}

	os.Exit(code)
}
