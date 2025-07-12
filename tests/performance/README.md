# Redis Auth Cache Performance Testing Framework

A comprehensive performance testing framework for Redis-based authentication caching, designed to validate cache performance under various load conditions and generate detailed analytics reports.

## 🚀 Quick Start

### Prerequisites

```bash
# Install required dependencies
go mod download

# Start Redis for testing
make dev  # Starts Docker containers including Redis
```

### Run Complete Performance Test Suite

```bash
# Run all performance tests and generate reports
make test-performance

# Reports will be generated in ./performance_reports/
```

## 📊 Test Coverage

### 1. Benchmark Tests (`benchmark_test.go`)
- **Purpose**: Micro-benchmarks for individual cache operations
- **Coverage**: Cache hits, misses, concurrent operations, memory pressure
- **Metrics**: Operations/second, memory allocations, latency percentiles
- **Run Command**: `make test-benchmark`

```bash
# Example output:
BenchmarkCacheGet-8              1000000    1.2 μs/op    64 B/op    2 allocs/op
BenchmarkCacheSet-8               500000    2.1 μs/op   128 B/op    3 allocs/op
BenchmarkCacheHitRatio-8          200000    5.3 μs/op   256 B/op    5 allocs/op
```

### 2. Load Tests (`load_test.go`)
- **Purpose**: Realistic traffic simulation using Vegeta
- **Coverage**: Various traffic patterns, cache warmup, sustained load
- **Scenarios**:
  - Baseline Load (300 req/s, 80% hit rate)
  - Peak Traffic (1000 req/s, 75% hit rate) 
  - Burst Load (2000 req/s, 60% hit rate)
  - Sustained High Load (800 req/s, 70% hit rate)
- **Run Command**: `make test-load`

### 3. Stress Tests
- **Purpose**: System limits and failure point identification
- **Coverage**: Extreme load, memory pressure, connection limits
- **Scenarios**:
  - Extreme Load (2000+ req/s)
  - Burst Spike (5000 req/s)
- **Run Command**: `make test-stress`

### 4. Cache Warmup Tests
- **Purpose**: Performance during cache initialization
- **Coverage**: Cold start performance, hit rate improvement over time
- **Metrics**: Latency degradation, throughput during warmup

## 🛠️ Framework Architecture

### Core Components

```
tests/performance/
├── auth_cache/                 # Test implementations
│   ├── benchmark_test.go       # Go benchmark tests
│   └── load_test.go           # Vegeta load tests
├── utils/                     # Testing utilities
│   ├── metrics_collector.go   # Performance metrics collection
│   ├── redis_setup.go        # Test environment setup
│   ├── test_data_generator.go # Realistic test data
│   └── report_generator.go   # Report generation
└── cmd/
    └── performance_runner.go  # Test suite orchestrator
```

### Technology Stack

- **Load Testing**: [Vegeta](https://github.com/tsenart/vegeta) v12.12.0 - Professional HTTP load testing
- **Container Management**: [Testcontainers](https://github.com/testcontainers/testcontainers-go) v0.37.0 - Real Redis integration
- **Metrics Collection**: [VictoriaMetrics](https://github.com/VictoriaMetrics/metrics) v1.38.0 - High-performance metrics
- **Statistical Analysis**: [Gonum](https://github.com/gonum/gonum) v0.16.0 - Percentile calculations
- **Redis Client**: [go-redis](https://github.com/redis/go-redis) v9.11.0 - Redis connectivity

## 📈 Reports and Analytics

### Generated Report Formats

1. **HTML Report** (`performance_report.html`)
   - Interactive charts and visualizations
   - Detailed metrics breakdown
   - Performance recommendations
   - Trend analysis

2. **JSON Report** (`performance_report.json`)
   - Machine-readable format
   - Complete test data
   - Integration with monitoring systems

3. **CSV Report** (`performance_report.csv`)
   - Tabular data format
   - Easy Excel/spreadsheet import
   - Historical analysis support

### Key Metrics Tracked

- **Throughput**: Requests per second
- **Latency Percentiles**: P50, P95, P99, P999
- **Cache Performance**: Hit rate, miss rate, error rate
- **Resource Usage**: Memory consumption, connection count
- **System Health**: Error rates, timeout rates

### Performance SLAs

| Scenario | P95 Latency | Error Rate | Cache Hit Rate | Throughput |
|----------|-------------|------------|----------------|------------|
| Baseline Load | < 5ms | < 1% | > 75% | > 250 req/s |
| Peak Traffic | < 10ms | < 2% | > 70% | > 700 req/s |
| Burst Load | < 20ms | < 5% | > 50% | > 1500 req/s |
| Sustained Load | < 15ms | < 3% | > 65% | > 600 req/s |

## 🔧 Configuration

### Environment Variables

```bash
export REDIS_URL="redis://localhost:6379"    # Redis connection
export TEST_DURATION="30s"                   # Load test duration
export CONCURRENT_USERS="50"                 # Concurrent connections
export CACHE_TTL="300s"                      # Cache time-to-live
```

### Test Data Configuration

```go
// Customize test scenarios in utils/test_data_generator.go
scenarios := []LoadTestScenario{
    {
        Name:          "Custom Load",
        RequestRate:   1500,           // req/s
        Duration:      60 * time.Second,
        CacheHitRatio: 0.80,           // 80% hit rate
        UserPattern:   "zipf",         // Distribution pattern
        Concurrency:   100,            // Concurrent connections
    },
}
```

## 📊 Usage Examples

### Run Specific Test Scenarios

```bash
# Run only benchmark tests
go test -bench=. ./tests/performance/auth_cache/

# Run specific load test scenario
go test -v ./tests/performance/auth_cache/ -run TestAuthCacheLoadPerformance

# Run with custom parameters
go test -bench=BenchmarkCacheGet -benchtime=30s ./tests/performance/auth_cache/
```

### Generate Reports Only

```bash
# Generate reports from existing test data
make performance-report

# Generate specific format
go run ./tests/performance/cmd/performance_runner.go ./reports
```

### Integration with CI/CD

```yaml
# Example GitHub Actions workflow
- name: Run Performance Tests
  run: |
    make dev
    make test-performance
    
- name: Upload Performance Reports
  uses: actions/upload-artifact@v3
  with:
    name: performance-reports
    path: ./performance_reports/
```

## 🐛 Troubleshooting

### Common Issues

1. **Redis Connection Failed**
   ```bash
   # Ensure Redis is running
   make dev
   
   # Check Redis connectivity
   redis-cli ping
   ```

2. **Port Already in Use**
   ```bash
   # Check running containers
   docker ps
   
   # Stop existing containers
   make down
   ```

3. **High Memory Usage**
   ```bash
   # Monitor test execution
   go test -memprofile=mem.prof ./tests/performance/auth_cache/
   go tool pprof mem.prof
   ```

### Performance Issues

1. **Low Throughput**
   - Check Redis connection pooling
   - Verify network latency
   - Review cache hit rates

2. **High Latency**
   - Monitor Redis performance
   - Check concurrent connection limits
   - Review system resources

3. **Test Failures**
   - Check error logs in reports
   - Verify SLA thresholds
   - Review test data patterns

## 🔮 Advanced Features

### Custom Metrics Collection

```go
// Add custom metrics to PerformanceMetrics
metrics := utils.NewPerformanceMetrics()
metrics.RecordCustomMetric("auth_validation_time", duration)
```

### Load Pattern Customization

```go
// Define custom load patterns
generator := utils.NewTestDataGenerator(userCount, tokenVariety)
tokens := generator.GenerateWithPattern("pareto") // 80/20 distribution
```

### Real-time Monitoring

```go
// Stream metrics to external systems
metrics.SetOutputTarget("http://prometheus:9090/metrics")
```

## 📝 Contributing

### Adding New Test Scenarios

1. Create test function in appropriate `*_test.go` file
2. Add scenario to `LoadTestScenario` configurations
3. Update SLA thresholds in validation functions
4. Add documentation and examples

### Extending Metrics Collection

1. Add new fields to `PerformanceMetrics` struct
2. Implement collection methods
3. Update report generation templates
4. Add visualization to HTML reports

### Performance Optimizations

1. Use connection pooling for Redis
2. Implement metric batching for high throughput
3. Add memory optimization for large datasets
4. Use background processing for report generation

## 📚 References

- [Redis Performance Best Practices](https://redis.io/docs/management/optimization/)
- [Go Benchmarking Guidelines](https://golang.org/pkg/testing/#hdr-Benchmarks)
- [Vegeta Load Testing](https://github.com/tsenart/vegeta)
- [VictoriaMetrics Documentation](https://docs.victoriametrics.com/)

---

**🎯 Goal**: Ensure Redis auth cache maintains sub-10ms P95 latency with >75% hit rate under production load conditions.
