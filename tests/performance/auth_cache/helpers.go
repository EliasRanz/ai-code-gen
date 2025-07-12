package auth_cache

import "os"

// isPerformanceTestingEnabled checks if performance tests should run
// Performance tests are disabled by default and require explicit enabling
// Set PERFORMANCE_TESTS=1 or PERFORMANCE_TESTS=true to enable them
func isPerformanceTestingEnabled() bool {
	value := os.Getenv("PERFORMANCE_TESTS")
	return value == "1" || value == "true"
}
