package utils

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"sort"
	"time"
)

// PerformanceReportGenerator handles generation of various report formats
type PerformanceReportGenerator struct {
	reports []PerformanceReport
}

// NewPerformanceReportGenerator creates a new report generator
func NewPerformanceReportGenerator() *PerformanceReportGenerator {
	return &PerformanceReportGenerator{
		reports: make([]PerformanceReport, 0),
	}
}

// AddReport adds a performance report to the generator
func (g *PerformanceReportGenerator) AddReport(report PerformanceReport) {
	g.reports = append(g.reports, report)
}

// GenerateCSVReport creates a CSV report of all performance test results
func (g *PerformanceReportGenerator) GenerateCSVReport() ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write header
	header := []string{
		"Test Name", "Duration (ms)", "Total Requests", "Throughput (req/s)",
		"Cache Hit Rate (%)", "Error Rate (%)", "P50 (ms)", "P95 (ms)", "P99 (ms)",
		"Min (ms)", "Max (ms)", "Memory Usage (MB)",
	}
	if err := writer.Write(header); err != nil {
		return nil, err
	}

	// Write data rows
	for _, report := range g.reports {
		row := []string{
			report.TestName,
			fmt.Sprintf("%.2f", float64(report.Duration.Nanoseconds())/1e6),
			fmt.Sprintf("%d", report.TotalRequests),
			fmt.Sprintf("%.2f", report.ThroughputRPS),
			fmt.Sprintf("%.2f", report.CacheHitRate*100),
			fmt.Sprintf("%.2f", report.ErrorRate*100),
			fmt.Sprintf("%.2f", float64(report.Percentiles.P50.Nanoseconds())/1e6),
			fmt.Sprintf("%.2f", float64(report.Percentiles.P95.Nanoseconds())/1e6),
			fmt.Sprintf("%.2f", float64(report.Percentiles.P99.Nanoseconds())/1e6),
			fmt.Sprintf("%.2f", float64(report.Percentiles.Min.Nanoseconds())/1e6),
			fmt.Sprintf("%.2f", float64(report.Percentiles.Max.Nanoseconds())/1e6),
			fmt.Sprintf("%.2f", report.MemoryUsageMB),
		}
		if err := writer.Write(row); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	return buf.Bytes(), writer.Error()
}

// GenerateJSONReport creates a JSON report of all performance test results
func (g *PerformanceReportGenerator) GenerateJSONReport() ([]byte, error) {
	summary := PerformanceTestSummary{
		GeneratedAt:   time.Now(),
		TotalTests:    len(g.reports),
		TestResults:   g.reports,
		Summary:       g.calculateSummary(),
		Trends:        g.calculateTrends(),
		Recommendations: g.generateRecommendations(),
	}

	return json.MarshalIndent(summary, "", "  ")
}

// GenerateHTMLReport creates an HTML report with charts and analysis
func (g *PerformanceReportGenerator) GenerateHTMLReport() ([]byte, error) {
	tmpl := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Redis Auth Cache Performance Report</title>
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; background-color: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; background: white; padding: 30px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        h1, h2 { color: #333; border-bottom: 2px solid #007acc; padding-bottom: 10px; }
        .summary { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 20px; margin: 20px 0; }
        .metric-card { background: #f8f9fa; padding: 20px; border-radius: 6px; text-align: center; border-left: 4px solid #007acc; }
        .metric-value { font-size: 2em; font-weight: bold; color: #007acc; }
        .metric-label { color: #666; font-size: 0.9em; }
        .chart-container { width: 100%; height: 400px; margin: 20px 0; }
        .recommendations { background: #e8f4fd; padding: 20px; border-radius: 6px; margin: 20px 0; }
        .recommendations h3 { color: #0066cc; margin-top: 0; }
        .recommendations ul { padding-left: 20px; }
        .recommendations li { margin-bottom: 8px; }
        table { width: 100%; border-collapse: collapse; margin: 20px 0; }
        th, td { padding: 12px; text-align: left; border-bottom: 1px solid #ddd; }
        th { background-color: #f2f2f2; font-weight: bold; }
        tr:hover { background-color: #f5f5f5; }
        .status-good { color: #28a745; font-weight: bold; }
        .status-warning { color: #ffc107; font-weight: bold; }
        .status-error { color: #dc3545; font-weight: bold; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Redis Auth Cache Performance Report</h1>
        <p>Generated on {{.GeneratedAt.Format "2006-01-02 15:04:05 MST"}}</p>
        
        <div class="summary">
            <div class="metric-card">
                <div class="metric-value">{{.Summary.TotalTests}}</div>
                <div class="metric-label">Total Tests</div>
            </div>
            <div class="metric-card">
                <div class="metric-value">{{printf "%.1f" .Summary.AvgThroughput}}</div>
                <div class="metric-label">Avg Throughput (req/s)</div>
            </div>
            <div class="metric-card">
                <div class="metric-value">{{printf "%.1f%%" (.Summary.AvgCacheHitRate | mul 100)}}</div>
                <div class="metric-label">Avg Cache Hit Rate</div>
            </div>
            <div class="metric-card">
                <div class="metric-value">{{printf "%.2f%%" (.Summary.AvgErrorRate | mul 100)}}</div>
                <div class="metric-label">Avg Error Rate</div>
            </div>
            <div class="metric-card">
                <div class="metric-value">{{printf "%.1fms" (.Summary.AvgP95Latency | nanosToMs)}}</div>
                <div class="metric-label">Avg P95 Latency</div>
            </div>
        </div>

        <h2>Performance Trends</h2>
        <div class="chart-container">
            <canvas id="throughputChart"></canvas>
        </div>
        
        <div class="chart-container">
            <canvas id="latencyChart"></canvas>
        </div>

        <h2>Detailed Results</h2>
        <table>
            <thead>
                <tr>
                    <th>Test Name</th>
                    <th>Throughput (req/s)</th>
                    <th>Cache Hit Rate</th>
                    <th>Error Rate</th>
                    <th>P95 Latency</th>
                    <th>Status</th>
                </tr>
            </thead>
            <tbody>
                {{range .TestResults}}
                <tr>
                    <td>{{.TestName}}</td>
                    <td>{{printf "%.1f" .ThroughputRPS}}</td>
                    <td>{{printf "%.1f%%" (.CacheHitRate | mul 100)}}</td>
                    <td>{{printf "%.2f%%" (.ErrorRate | mul 100)}}</td>
                    <td>{{printf "%.1fms" (.Percentiles.P95 | nanosToMs)}}</td>
                    <td class="{{.Status | statusClass}}">{{.Status}}</td>
                </tr>
                {{end}}
            </tbody>
        </table>

        <div class="recommendations">
            <h3>Performance Recommendations</h3>
            <ul>
                {{range .Recommendations}}
                <li>{{.}}</li>
                {{end}}
            </ul>
        </div>
    </div>

    <script>
        // Throughput Chart
        const throughputCtx = document.getElementById('throughputChart').getContext('2d');
        new Chart(throughputCtx, {
            type: 'bar',
            data: {
                labels: [{{range .TestResults}}'{{.TestName}}',{{end}}],
                datasets: [{
                    label: 'Throughput (req/s)',
                    data: [{{range .TestResults}}{{.ThroughputRPS}},{{end}}],
                    backgroundColor: 'rgba(0, 122, 204, 0.6)',
                    borderColor: 'rgba(0, 122, 204, 1)',
                    borderWidth: 1
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                scales: {
                    y: { beginAtZero: true }
                }
            }
        });

        // Latency Chart
        const latencyCtx = document.getElementById('latencyChart').getContext('2d');
        new Chart(latencyCtx, {
            type: 'line',
            data: {
                labels: [{{range .TestResults}}'{{.TestName}}',{{end}}],
                datasets: [
                    {
                        label: 'P50 Latency (ms)',
                        data: [{{range .TestResults}}{{.Percentiles.P50 | nanosToMs}},{{end}}],
                        borderColor: 'rgba(40, 167, 69, 1)',
                        backgroundColor: 'rgba(40, 167, 69, 0.1)',
                        fill: false
                    },
                    {
                        label: 'P95 Latency (ms)',
                        data: [{{range .TestResults}}{{.Percentiles.P95 | nanosToMs}},{{end}}],
                        borderColor: 'rgba(255, 193, 7, 1)',
                        backgroundColor: 'rgba(255, 193, 7, 0.1)',
                        fill: false
                    },
                    {
                        label: 'P99 Latency (ms)',
                        data: [{{range .TestResults}}{{.Percentiles.P99 | nanosToMs}},{{end}}],
                        borderColor: 'rgba(220, 53, 69, 1)',
                        backgroundColor: 'rgba(220, 53, 69, 0.1)',
                        fill: false
                    }
                ]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                scales: {
                    y: { beginAtZero: true }
                }
            }
        });
    </script>
</body>
</html>`

	funcMap := template.FuncMap{
		"mul": func(a, b float64) float64 { return a * b },
		"nanosToMs": func(d time.Duration) float64 {
			return float64(d.Nanoseconds()) / 1e6
		},
		"statusClass": func(status string) string {
			switch status {
			case "PASS":
				return "status-good"
			case "WARN":
				return "status-warning"
			case "FAIL":
				return "status-error"
			default:
				return ""
			}
		},
	}

	t, err := template.New("report").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return nil, err
	}

	summary := PerformanceTestSummary{
		GeneratedAt:     time.Now(),
		TotalTests:      len(g.reports),
		TestResults:     g.reports,
		Summary:         g.calculateSummary(),
		Trends:          g.calculateTrends(),
		Recommendations: g.generateRecommendations(),
	}

	var buf bytes.Buffer
	err = t.Execute(&buf, summary)
	return buf.Bytes(), err
}

// WriteTo writes a report to the specified writer in the given format
func (g *PerformanceReportGenerator) WriteTo(w io.Writer, format string) error {
	var data []byte
	var err error

	switch format {
	case "csv":
		data, err = g.GenerateCSVReport()
	case "json":
		data, err = g.GenerateJSONReport()
	case "html":
		data, err = g.GenerateHTMLReport()
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	if err != nil {
		return err
	}

	_, err = w.Write(data)
	return err
}

// calculateSummary computes aggregate statistics across all reports
func (g *PerformanceReportGenerator) calculateSummary() PerformanceSummary {
	if len(g.reports) == 0 {
		return PerformanceSummary{}
	}

	var totalThroughput, totalCacheHitRate, totalErrorRate float64
	var totalP95Latency time.Duration

	for _, report := range g.reports {
		totalThroughput += report.ThroughputRPS
		totalCacheHitRate += report.CacheHitRate
		totalErrorRate += report.ErrorRate
		totalP95Latency += report.Percentiles.P95
	}

	count := float64(len(g.reports))
	return PerformanceSummary{
		TotalTests:       len(g.reports),
		AvgThroughput:    totalThroughput / count,
		AvgCacheHitRate:  totalCacheHitRate / count,
		AvgErrorRate:     totalErrorRate / count,
		AvgP95Latency:    time.Duration(int64(totalP95Latency) / int64(count)),
	}
}

// calculateTrends analyzes performance trends across test runs
func (g *PerformanceReportGenerator) calculateTrends() []TrendAnalysis {
	if len(g.reports) < 2 {
		return []TrendAnalysis{}
	}

	// Sort reports by test name for consistent comparison
	sortedReports := make([]PerformanceReport, len(g.reports))
	copy(sortedReports, g.reports)
	sort.Slice(sortedReports, func(i, j int) bool {
		return sortedReports[i].TestName < sortedReports[j].TestName
	})

	trends := []TrendAnalysis{}

	// Group reports by test name and analyze trends
	testGroups := make(map[string][]PerformanceReport)
	for _, report := range sortedReports {
		testGroups[report.TestName] = append(testGroups[report.TestName], report)
	}

	for testName, reports := range testGroups {
		if len(reports) < 2 {
			continue
		}

		first := reports[0]
		last := reports[len(reports)-1]

		trend := TrendAnalysis{
			TestName:           testName,
			ThroughputTrend:    calculateTrend(first.ThroughputRPS, last.ThroughputRPS),
			LatencyTrend:       calculateLatencyTrend(first.Percentiles.P95, last.Percentiles.P95),
			CacheHitRateTrend:  calculateTrend(first.CacheHitRate, last.CacheHitRate),
			ErrorRateTrend:     calculateTrend(first.ErrorRate, last.ErrorRate),
		}

		trends = append(trends, trend)
	}

	return trends
}

// generateRecommendations provides performance optimization recommendations
func (g *PerformanceReportGenerator) generateRecommendations() []string {
	recommendations := []string{}
	summary := g.calculateSummary()

	// Throughput recommendations
	if summary.AvgThroughput < 500 {
		recommendations = append(recommendations, 
			"Consider optimizing cache hit ratio - current throughput is below target (500 req/s)")
	}

	// Latency recommendations
	if summary.AvgP95Latency > 10*time.Millisecond {
		recommendations = append(recommendations, 
			"P95 latency is high - consider Redis connection pooling optimization")
	}

	// Cache hit rate recommendations
	if summary.AvgCacheHitRate < 0.80 {
		recommendations = append(recommendations, 
			"Cache hit rate is below 80% - review cache TTL settings and data access patterns")
	}

	// Error rate recommendations
	if summary.AvgErrorRate > 0.01 {
		recommendations = append(recommendations, 
			"Error rate is above 1% - investigate Redis connectivity and timeout settings")
	}

	// Memory usage recommendations
	avgMemory := g.calculateAvgMemoryUsage()
	if avgMemory > 100 {
		recommendations = append(recommendations, 
			"High memory usage detected - consider implementing cache eviction policies")
	}

	// General recommendations
	recommendations = append(recommendations, 
		"Monitor cache performance continuously in production")
	recommendations = append(recommendations, 
		"Consider implementing circuit breaker pattern for Redis failures")
	recommendations = append(recommendations, 
		"Set up alerts for cache hit rate drops below 75%")

	return recommendations
}

func (g *PerformanceReportGenerator) calculateAvgMemoryUsage() float64 {
	if len(g.reports) == 0 {
		return 0
	}

	var total float64
	for _, report := range g.reports {
		total += report.MemoryUsageMB
	}
	return total / float64(len(g.reports))
}

func calculateTrend(old, new float64) string {
	if new > old*1.05 {
		return "IMPROVING"
	} else if new < old*0.95 {
		return "DEGRADING"
	}
	return "STABLE"
}

func calculateLatencyTrend(old, new time.Duration) string {
	if new < time.Duration(float64(old)*0.95) {
		return "IMPROVING"
	} else if new > time.Duration(float64(old)*1.05) {
		return "DEGRADING"
	}
	return "STABLE"
}

// Supporting types for report generation
type PerformanceTestSummary struct {
	GeneratedAt     time.Time           `json:"generated_at"`
	TotalTests      int                 `json:"total_tests"`
	TestResults     []PerformanceReport `json:"test_results"`
	Summary         PerformanceSummary  `json:"summary"`
	Trends          []TrendAnalysis     `json:"trends"`
	Recommendations []string            `json:"recommendations"`
}

type PerformanceSummary struct {
	TotalTests      int           `json:"total_tests"`
	AvgThroughput   float64       `json:"avg_throughput"`
	AvgCacheHitRate float64       `json:"avg_cache_hit_rate"`
	AvgErrorRate    float64       `json:"avg_error_rate"`
	AvgP95Latency   time.Duration `json:"avg_p95_latency"`
}

type TrendAnalysis struct {
	TestName          string `json:"test_name"`
	ThroughputTrend   string `json:"throughput_trend"`
	LatencyTrend      string `json:"latency_trend"`
	CacheHitRateTrend string `json:"cache_hit_rate_trend"`
	ErrorRateTrend    string `json:"error_rate_trend"`
}
