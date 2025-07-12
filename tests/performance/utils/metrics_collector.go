package utils

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/VictoriaMetrics/metrics"
	"gonum.org/v1/gonum/stat"
)

// PerformanceMetrics tracks comprehensive performance data
type PerformanceMetrics struct {
	CacheHits       *metrics.Counter
	CacheMisses     *metrics.Counter
	CacheErrors     *metrics.Counter
	AuthServiceCalls *metrics.Counter
	TotalRequests   *metrics.Counter

	// Latency tracking
	ResponseTimes   []time.Duration
	responseTimesMu sync.RWMutex

	// Real-time metrics
	CacheLatency    *metrics.Summary
	ThroughputRPS   *metrics.Summary
	ErrorRate       *metrics.Summary

	startTime       time.Time
}

// PerformancePercentiles holds statistical analysis results
type PerformancePercentiles struct {
	P50   time.Duration `json:"p50"`
	P95   time.Duration `json:"p95"`
	P99   time.Duration `json:"p99"`
	P999  time.Duration `json:"p999"`
	Mean  time.Duration `json:"mean"`
	StdDev time.Duration `json:"stddev"`
	Min   time.Duration `json:"min"`
	Max   time.Duration `json:"max"`
}

// PerformanceReport contains comprehensive test results
type PerformanceReport struct {
	TestName        string                 `json:"test_name"`
	Duration        time.Duration          `json:"duration"`
	TotalRequests   int64                  `json:"total_requests"`
	CacheHitRate    float64               `json:"cache_hit_rate"`
	ErrorRate       float64               `json:"error_rate"`
	ThroughputRPS   float64               `json:"throughput_rps"`
	Percentiles     PerformancePercentiles `json:"percentiles"`
	MemoryUsageMB   float64               `json:"memory_usage_mb"`
	Timestamp       time.Time             `json:"timestamp"`
	Status          string                `json:"status"` // PASS, WARN, FAIL
}

// NewPerformanceMetrics creates a new metrics collector with unique metric names
func NewPerformanceMetrics() *PerformanceMetrics {
	// Use timestamp to create unique metric names across test runs
	suffix := fmt.Sprintf("_%d", time.Now().UnixNano())
	
	return &PerformanceMetrics{
		CacheHits:       metrics.NewCounter("cache_hits_total" + suffix),
		CacheMisses:     metrics.NewCounter("cache_misses_total" + suffix),
		CacheErrors:     metrics.NewCounter("cache_errors_total" + suffix),
		AuthServiceCalls: metrics.NewCounter("auth_service_calls_total" + suffix),
		TotalRequests:   metrics.NewCounter("total_requests" + suffix),
		ResponseTimes:   make([]time.Duration, 0, 10000), // Pre-allocate for performance
		CacheLatency:    metrics.NewSummary("cache_latency_seconds" + suffix),
		ThroughputRPS:   metrics.NewSummary("throughput_requests_per_second" + suffix),
		ErrorRate:       metrics.NewSummary("error_rate_percent" + suffix),
		startTime:       time.Now(),
	}
}

// RecordCacheHit records a successful cache hit
func (pm *PerformanceMetrics) RecordCacheHit(duration time.Duration) {
	pm.CacheHits.Inc()
	pm.TotalRequests.Inc()
	pm.recordLatency(duration)
}

// RecordCacheMiss records a cache miss that resulted in auth service call
func (pm *PerformanceMetrics) RecordCacheMiss(duration time.Duration) {
	pm.CacheMisses.Inc()
	pm.AuthServiceCalls.Inc()
	pm.TotalRequests.Inc()
	pm.recordLatency(duration)
}

// RecordCacheError records a cache operation error
func (pm *PerformanceMetrics) RecordCacheError(duration time.Duration) {
	pm.CacheErrors.Inc()
	pm.TotalRequests.Inc()
	pm.recordLatency(duration)
}

// recordLatency stores response time for statistical analysis
func (pm *PerformanceMetrics) recordLatency(duration time.Duration) {
	pm.responseTimesMu.Lock()
	pm.ResponseTimes = append(pm.ResponseTimes, duration)
	pm.responseTimesMu.Unlock()

	// Update real-time metrics
	pm.CacheLatency.Update(duration.Seconds())
}

// CalculatePercentiles computes statistical percentiles from recorded latencies
func (pm *PerformanceMetrics) CalculatePercentiles() PerformancePercentiles {
	pm.responseTimesMu.RLock()
	defer pm.responseTimesMu.RUnlock()

	if len(pm.ResponseTimes) == 0 {
		return PerformancePercentiles{}
	}

	// Convert to float64 seconds for statistical analysis
	latencies := make([]float64, len(pm.ResponseTimes))
	for i, d := range pm.ResponseTimes {
		latencies[i] = d.Seconds()
	}
	
	// Sort latencies for percentile calculations (required by gonum)
	sort.Float64s(latencies)

	// Calculate statistics using gonum
	return PerformancePercentiles{
		P50:   time.Duration(stat.Quantile(0.50, stat.Empirical, latencies, nil) * float64(time.Second)),
		P95:   time.Duration(stat.Quantile(0.95, stat.Empirical, latencies, nil) * float64(time.Second)),
		P99:   time.Duration(stat.Quantile(0.99, stat.Empirical, latencies, nil) * float64(time.Second)),
		P999:  time.Duration(stat.Quantile(0.999, stat.Empirical, latencies, nil) * float64(time.Second)),
		Mean:  time.Duration(stat.Mean(latencies, nil) * float64(time.Second)),
		StdDev: time.Duration(stat.StdDev(latencies, nil) * float64(time.Second)),
		Min:   pm.getMinLatency(),
		Max:   pm.getMaxLatency(),
	}
}

// GenerateReport creates a comprehensive performance report
func (pm *PerformanceMetrics) GenerateReport(testName string) PerformanceReport {
	duration := time.Since(pm.startTime)
	totalRequests := pm.TotalRequests.Get()
	cacheHits := pm.CacheHits.Get()
	cacheErrors := pm.CacheErrors.Get()

	var cacheHitRate, errorRate, throughputRPS float64

	if totalRequests > 0 {
		cacheHitRate = float64(cacheHits) / float64(totalRequests)
		errorRate = float64(cacheErrors) / float64(totalRequests)
		throughputRPS = float64(totalRequests) / duration.Seconds()
	}

	percentiles := pm.CalculatePercentiles()
	status := pm.calculateStatus(cacheHitRate, errorRate, percentiles.P95)

	return PerformanceReport{
		TestName:      testName,
		Duration:      duration,
		TotalRequests: int64(totalRequests),
		CacheHitRate:  cacheHitRate,
		ErrorRate:     errorRate,
		ThroughputRPS: throughputRPS,
		Percentiles:   percentiles,
		MemoryUsageMB: pm.getMemoryUsage(),
		Timestamp:     time.Now(),
		Status:        status,
	}
}

// Reset clears all metrics for a new test run
func (pm *PerformanceMetrics) Reset() {
	// Reset counters (VictoriaMetrics doesn't support reset, create new instances)
	pm.CacheHits = metrics.NewCounter("cache_hits_total")
	pm.CacheMisses = metrics.NewCounter("cache_misses_total")
	pm.CacheErrors = metrics.NewCounter("cache_errors_total")
	pm.AuthServiceCalls = metrics.NewCounter("auth_service_calls_total")
	pm.TotalRequests = metrics.NewCounter("total_requests")

	// Clear response times
	pm.responseTimesMu.Lock()
	pm.ResponseTimes = pm.ResponseTimes[:0]
	pm.responseTimesMu.Unlock()

	// Reset real-time metrics
	pm.CacheLatency = metrics.NewSummary("cache_latency_seconds")
	pm.ThroughputRPS = metrics.NewSummary("throughput_requests_per_second")
	pm.ErrorRate = metrics.NewSummary("error_rate_percent")

	pm.startTime = time.Now()
}

// getMinLatency finds the minimum response time
func (pm *PerformanceMetrics) getMinLatency() time.Duration {
	if len(pm.ResponseTimes) == 0 {
		return 0
	}

	min := pm.ResponseTimes[0]
	for _, d := range pm.ResponseTimes[1:] {
		if d < min {
			min = d
		}
	}
	return min
}

// getMaxLatency finds the maximum response time
func (pm *PerformanceMetrics) getMaxLatency() time.Duration {
	if len(pm.ResponseTimes) == 0 {
		return 0
	}

	max := pm.ResponseTimes[0]
	for _, d := range pm.ResponseTimes[1:] {
		if d > max {
			max = d
		}
	}
	return max
}

// getMemoryUsage returns current memory usage in MB (simplified implementation)
func (pm *PerformanceMetrics) getMemoryUsage() float64 {
	// In a real implementation, you'd use runtime.MemStats or similar
	// For now, return estimated usage based on stored response times
	estimatedBytes := len(pm.ResponseTimes) * 8 // 8 bytes per time.Duration
	return float64(estimatedBytes) / (1024 * 1024) // Convert to MB
}

// PrintSummary outputs a quick performance summary
func (pm *PerformanceMetrics) PrintSummary(testName string) {
	report := pm.GenerateReport(testName)

	fmt.Printf("\n=== Performance Summary: %s ===\n", testName)
	fmt.Printf("Duration: %v\n", report.Duration)
	fmt.Printf("Total Requests: %d\n", report.TotalRequests)
	fmt.Printf("Throughput: %.2f req/sec\n", report.ThroughputRPS)
	fmt.Printf("Cache Hit Rate: %.2f%%\n", report.CacheHitRate*100)
	fmt.Printf("Error Rate: %.2f%%\n", report.ErrorRate*100)
	fmt.Printf("P50 Latency: %v\n", report.Percentiles.P50)
	fmt.Printf("P95 Latency: %v\n", report.Percentiles.P95)
	fmt.Printf("P99 Latency: %v\n", report.Percentiles.P99)
	fmt.Printf("Memory Usage: %.2f MB\n", report.MemoryUsageMB)
	fmt.Printf("===============================\n\n")
}

// calculateStatus determines test status based on performance thresholds
func (pm *PerformanceMetrics) calculateStatus(cacheHitRate, errorRate float64, p95Latency time.Duration) string {
	// FAIL conditions
	if errorRate > 0.05 { // Error rate > 5%
		return "FAIL"
	}
	if p95Latency > 50*time.Millisecond { // P95 latency > 50ms
		return "FAIL"
	}
	if cacheHitRate < 0.50 { // Cache hit rate < 50%
		return "FAIL"
	}

	// WARN conditions
	if errorRate > 0.02 { // Error rate > 2%
		return "WARN"
	}
	if p95Latency > 20*time.Millisecond { // P95 latency > 20ms
		return "WARN"
	}
	if cacheHitRate < 0.75 { // Cache hit rate < 75%
		return "WARN"
	}

	// All good
	return "PASS"
}
