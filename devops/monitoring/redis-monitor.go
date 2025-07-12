package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisMetrics represents Redis performance metrics
type RedisMetrics struct {
	Timestamp         time.Time `json:"timestamp"`
	ConnectedClients  int64     `json:"connected_clients"`
	UsedMemory        int64     `json:"used_memory"`
	UsedMemoryPeak    int64     `json:"used_memory_peak"`
	KeyspaceHits      int64     `json:"keyspace_hits"`
	KeyspaceMisses    int64     `json:"keyspace_misses"`
	HitRate           float64   `json:"hit_rate"`
	TotalCommands     int64     `json:"total_commands_processed"`
	CommandsPerSecond float64   `json:"commands_per_second"`
	AvgTTL            float64   `json:"avg_ttl"`
	TotalKeys         int64     `json:"total_keys"`
	ExpiredKeys       int64     `json:"expired_keys"`
}

// Monitor handles Redis monitoring operations
type Monitor struct {
	client   *redis.Client
	interval time.Duration
	output   string
	verbose  bool
}

// NewMonitor creates a new Redis monitor
func NewMonitor(redisURL string, interval time.Duration, output string, verbose bool) (*Monitor, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	client := redis.NewClient(opts)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Monitor{
		client:   client,
		interval: interval,
		output:   output,
		verbose:  verbose,
	}, nil
}

// CollectMetrics gathers current Redis metrics
func (m *Monitor) CollectMetrics(ctx context.Context) (*RedisMetrics, error) {
	info, err := m.client.Info(ctx, "all").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get Redis info: %w", err)
	}

	metrics := &RedisMetrics{
		Timestamp: time.Now(),
	}

	// Parse info response
	lines := strings.Split(info, "\r\n")
	infoMap := make(map[string]string)

	for _, line := range lines {
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				infoMap[parts[0]] = parts[1]
			}
		}
	}

	// Extract metrics
	if val, ok := infoMap["connected_clients"]; ok {
		metrics.ConnectedClients, _ = strconv.ParseInt(val, 10, 64)
	}

	if val, ok := infoMap["used_memory"]; ok {
		metrics.UsedMemory, _ = strconv.ParseInt(val, 10, 64)
	}

	if val, ok := infoMap["used_memory_peak"]; ok {
		metrics.UsedMemoryPeak, _ = strconv.ParseInt(val, 10, 64)
	}

	if val, ok := infoMap["keyspace_hits"]; ok {
		metrics.KeyspaceHits, _ = strconv.ParseInt(val, 10, 64)
	}

	if val, ok := infoMap["keyspace_misses"]; ok {
		metrics.KeyspaceMisses, _ = strconv.ParseInt(val, 10, 64)
	}

	if val, ok := infoMap["total_commands_processed"]; ok {
		metrics.TotalCommands, _ = strconv.ParseInt(val, 10, 64)
	}

	if val, ok := infoMap["expired_keys"]; ok {
		metrics.ExpiredKeys, _ = strconv.ParseInt(val, 10, 64)
	}

	// Calculate hit rate
	totalRequests := metrics.KeyspaceHits + metrics.KeyspaceMisses
	if totalRequests > 0 {
		metrics.HitRate = float64(metrics.KeyspaceHits) / float64(totalRequests) * 100
	}

	// Get database info for key count
	dbInfo, err := m.client.Info(ctx, "keyspace").Result()
	if err == nil {
		for _, line := range strings.Split(dbInfo, "\r\n") {
			if strings.HasPrefix(line, "db0:") {
				parts := strings.Split(line, ",")
				for _, part := range parts {
					if strings.HasPrefix(part, "keys=") {
						keyCountStr := strings.TrimPrefix(part, "keys=")
						metrics.TotalKeys, _ = strconv.ParseInt(keyCountStr, 10, 64)
						break
					}
				}
				break
			}
		}
	}

	return metrics, nil
}

// StartMonitoring begins continuous monitoring
func (m *Monitor) StartMonitoring(ctx context.Context) error {
	var previousMetrics *RedisMetrics
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	fmt.Printf("🔍 Starting Redis monitoring (interval: %v)\n", m.interval)
	fmt.Printf("📊 Output format: %s\n", m.output)
	fmt.Printf("🔗 Redis connection established\n\n")

	if m.output == "table" {
		m.printTableHeader()
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			metrics, err := m.CollectMetrics(ctx)
			if err != nil {
				log.Printf("Error collecting metrics: %v", err)
				continue
			}

			// Calculate commands per second
			if previousMetrics != nil {
				timeDiff := metrics.Timestamp.Sub(previousMetrics.Timestamp).Seconds()
				commandDiff := metrics.TotalCommands - previousMetrics.TotalCommands
				if timeDiff > 0 {
					metrics.CommandsPerSecond = float64(commandDiff) / timeDiff
				}
			}

			// Output metrics
			switch m.output {
			case "json":
				m.outputJSON(metrics)
			case "table":
				m.outputTable(metrics)
			case "summary":
				m.outputSummary(metrics)
			}

			previousMetrics = metrics
		}
	}
}

// printTableHeader prints the table header for table output
func (m *Monitor) printTableHeader() {
	fmt.Printf("%-20s %-8s %-12s %-10s %-8s %-8s %-6s %-6s\n",
		"Time", "Clients", "Memory(MB)", "Hit Rate%", "Cmd/Sec", "Keys", "Hits", "Misses")
	fmt.Println(strings.Repeat("-", 85))
}

// outputTable outputs metrics in table format
func (m *Monitor) outputTable(metrics *RedisMetrics) {
	memoryMB := float64(metrics.UsedMemory) / 1024 / 1024
	fmt.Printf("%-20s %-8d %-12.1f %-10.2f %-8.1f %-8d %-6d %-6d\n",
		metrics.Timestamp.Format("15:04:05"),
		metrics.ConnectedClients,
		memoryMB,
		metrics.HitRate,
		metrics.CommandsPerSecond,
		metrics.TotalKeys,
		metrics.KeyspaceHits,
		metrics.KeyspaceMisses)
}

// outputJSON outputs metrics in JSON format
func (m *Monitor) outputJSON(metrics *RedisMetrics) {
	jsonData, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		return
	}
	fmt.Println(string(jsonData))
}

// outputSummary outputs metrics in summary format
func (m *Monitor) outputSummary(metrics *RedisMetrics) {
	fmt.Printf("\n🕐 %s\n", metrics.Timestamp.Format("15:04:05 MST"))
	fmt.Printf("👥 Connected Clients: %d\n", metrics.ConnectedClients)
	fmt.Printf("💾 Memory Used: %.1f MB (Peak: %.1f MB)\n",
		float64(metrics.UsedMemory)/1024/1024,
		float64(metrics.UsedMemoryPeak)/1024/1024)
	fmt.Printf("🎯 Hit Rate: %.2f%% (%d hits, %d misses)\n",
		metrics.HitRate, metrics.KeyspaceHits, metrics.KeyspaceMisses)
	fmt.Printf("⚡ Commands/sec: %.1f (Total: %d)\n",
		metrics.CommandsPerSecond, metrics.TotalCommands)
	fmt.Printf("🔑 Total Keys: %d (Expired: %d)\n",
		metrics.TotalKeys, metrics.ExpiredKeys)
	fmt.Println(strings.Repeat("-", 50))
}

// GenerateReport creates a comprehensive monitoring report
func (m *Monitor) GenerateReport(ctx context.Context, duration time.Duration) error {
	fmt.Printf("📋 Generating %v monitoring report...\n\n", duration)

	var metrics []*RedisMetrics
	startTime := time.Now()
	ticker := time.NewTicker(time.Second * 5) // Sample every 5 seconds
	defer ticker.Stop()

	timeout := time.After(duration)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			return m.saveReport(metrics, duration)
		case <-ticker.C:
			metric, err := m.CollectMetrics(ctx)
			if err != nil {
				log.Printf("Error collecting metrics: %v", err)
				continue
			}
			metrics = append(metrics, metric)
			
			if m.verbose {
				elapsed := time.Since(startTime)
				fmt.Printf("Collected sample %d (%.1fs elapsed)\n", len(metrics), elapsed.Seconds())
			}
		}
	}
}

// saveReport saves the collected metrics to a report file
func (m *Monitor) saveReport(metrics []*RedisMetrics, duration time.Duration) error {
	if len(metrics) == 0 {
		return fmt.Errorf("no metrics collected")
	}

	filename := fmt.Sprintf("redis_monitor_report_%s.json", time.Now().Format("20060102_150405"))
	
	report := struct {
		GeneratedAt time.Time       `json:"generated_at"`
		Duration    string          `json:"duration"`
		SampleCount int             `json:"sample_count"`
		Metrics     []*RedisMetrics `json:"metrics"`
		Summary     interface{}     `json:"summary"`
	}{
		GeneratedAt: time.Now(),
		Duration:    duration.String(),
		SampleCount: len(metrics),
		Metrics:     metrics,
		Summary:     m.calculateSummary(metrics),
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create report file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(report); err != nil {
		return fmt.Errorf("failed to write report: %w", err)
	}

	fmt.Printf("✅ Report saved to: %s\n", filename)
	fmt.Printf("📊 Collected %d samples over %v\n", len(metrics), duration)
	
	return nil
}

// calculateSummary calculates summary statistics from metrics
func (m *Monitor) calculateSummary(metrics []*RedisMetrics) interface{} {
	if len(metrics) == 0 {
		return nil
	}

	var totalHitRate, totalCmdPerSec, totalMemory float64
	var maxClients, maxKeys, maxMemory int64

	for _, metric := range metrics {
		totalHitRate += metric.HitRate
		totalCmdPerSec += metric.CommandsPerSecond
		totalMemory += float64(metric.UsedMemory)

		if metric.ConnectedClients > maxClients {
			maxClients = metric.ConnectedClients
		}
		if metric.TotalKeys > maxKeys {
			maxKeys = metric.TotalKeys
		}
		if metric.UsedMemory > maxMemory {
			maxMemory = metric.UsedMemory
		}
	}

	count := float64(len(metrics))
	
	return map[string]interface{}{
		"avg_hit_rate":        totalHitRate / count,
		"avg_commands_per_sec": totalCmdPerSec / count,
		"avg_memory_mb":       (totalMemory / count) / 1024 / 1024,
		"max_clients":         maxClients,
		"max_keys":           maxKeys,
		"max_memory_mb":      float64(maxMemory) / 1024 / 1024,
		"first_sample":       metrics[0].Timestamp,
		"last_sample":        metrics[len(metrics)-1].Timestamp,
	}
}

func main() {
	var (
		redisURL = flag.String("redis-url", "redis://localhost:6379", "Redis connection URL")
		interval = flag.Duration("interval", time.Second*2, "Monitoring interval")
		output   = flag.String("output", "table", "Output format: json, table, summary")
		report   = flag.Duration("report", 0, "Generate report for specified duration (e.g. 5m, 1h)")
		verbose  = flag.Bool("verbose", false, "Enable verbose output")
	)
	flag.Parse()

	monitor, err := NewMonitor(*redisURL, *interval, *output, *verbose)
	if err != nil {
		log.Fatalf("Failed to create monitor: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		fmt.Println("\n🛑 Shutting down monitor...")
		cancel()
	}()

	// Run in report mode or continuous monitoring mode
	if *report > 0 {
		if err := monitor.GenerateReport(ctx, *report); err != nil {
			log.Fatalf("Report generation failed: %v", err)
		}
	} else {
		if err := monitor.StartMonitoring(ctx); err != nil && err != context.Canceled {
			log.Fatalf("Monitoring failed: %v", err)
		}
	}

	fmt.Println("👋 Monitor stopped")
}
