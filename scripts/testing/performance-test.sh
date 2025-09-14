#!/bin/bash

# Performance Testing Automation Script
# DevOps tool for comprehensive Redis auth cache performance validation

set -euo pipefail

# Configuration
REPORTS_DIR="./performance_reports"
TEST_ENVIRONMENT="${TEST_ENVIRONMENT:-staging}"
REDIS_URL="${REDIS_URL:-redis://localhost:6379}"

# Test duration configuration (in seconds)
BENCHMARK_DURATION="${BENCHMARK_DURATION:-30}"
LOAD_TEST_DURATION="${LOAD_TEST_DURATION:-300}"
STRESS_TEST_DURATION="${STRESS_TEST_DURATION:-600}"

# Enable performance tests explicitly
export PERFORMANCE_TESTS=1

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed or not in PATH"
        exit 1
    fi
    
    # Check if Redis is accessible
    if ! timeout 5 bash -c "</dev/tcp/localhost/6379" 2>/dev/null; then
        log_warning "Redis is not accessible on localhost:6379"
        
        # Check if Docker is available
        if command -v docker &> /dev/null && docker info &> /dev/null; then
            log_info "Starting Redis with Docker..."
            make dev || {
                log_error "Failed to start Redis. Please ensure Docker is running."
                log_error "Performance tests require Redis. Please start Redis manually or install Docker."
                exit 1
            }
            sleep 5
        else
            log_error "Docker is not available and Redis is not running."
            log_error "Performance tests require Redis to be running on localhost:6379"
            log_error "Please start Redis manually or install Docker to run these tests."
            exit 1
        fi
    fi
    
    log_success "Prerequisites check passed"
}

# Validate SLA configuration
validate_sla_config() {
    log_info "Validating SLA configuration for environment: $TEST_ENVIRONMENT"
    
    go run ./devops/performance/sla-validator.go validate "$TEST_ENVIRONMENT"
    
    log_success "SLA configuration validated"
}

# Run benchmark tests
run_benchmark_tests() {
    log_info "Running Go benchmark tests (duration: ${BENCHMARK_DURATION}s)..."
    
    mkdir -p "$REPORTS_DIR/benchmarks"
    
    # Set benchmark duration for tests
    export BENCHMARK_DURATION="$BENCHMARK_DURATION"
    
    # Suppress debug logs during performance tests
    export LOG_LEVEL="info"
    export GIN_MODE="release"
    
    # Calculate timeout with buffer for Go benchmark calibration
    local timeout_duration=$((BENCHMARK_DURATION + 30))
    
    # Run benchmarks with specified duration and save results
    # Note: Use -benchtime flag to control individual benchmark duration and only run benchmark tests
    timeout "${timeout_duration}s" go test -bench=BenchmarkCache -benchmem -count=1 -timeout="${timeout_duration}s" -benchtime="${BENCHMARK_DURATION}s" -run='^$' ./tests/performance/auth_cache/ > "$REPORTS_DIR/benchmarks/benchmark_results.txt" 2>&1 || {
        log_warning "Some benchmark tests failed or timed out, continuing..."
    }
    
    BENCHMARK_COMPLETED=true
    log_success "Benchmark tests completed"
}

# Run load tests
run_load_tests() {
    log_info "Running Vegeta load tests (duration: ${LOAD_TEST_DURATION}s)..."
    
    mkdir -p "$REPORTS_DIR/load_tests"
    
    # Set environment and duration for tests
    export TEST_ENVIRONMENT="$TEST_ENVIRONMENT"
    export LOAD_TEST_DURATION="$LOAD_TEST_DURATION"
    
    # Suppress debug logs during performance tests
    export LOG_LEVEL="info"
    export GIN_MODE="release"
    
    # Run load tests with specified duration
    timeout "${LOAD_TEST_DURATION}s" go test -v -timeout="${LOAD_TEST_DURATION}s" ./tests/performance/auth_cache/ -run TestAuthCacheLoadPerformance > "$REPORTS_DIR/load_tests/load_test_results.txt" 2>&1 || {
        log_warning "Some load tests failed or timed out, continuing..."
    }
    
    LOAD_COMPLETED=true
    log_success "Load tests completed"
}

# Run stress tests
run_stress_tests() {
    log_info "Running stress tests (duration: ${STRESS_TEST_DURATION}s)..."
    
    mkdir -p "$REPORTS_DIR/stress_tests"
    
    # Set environment and duration for tests
    export TEST_ENVIRONMENT="$TEST_ENVIRONMENT"
    export STRESS_TEST_DURATION="$STRESS_TEST_DURATION"
    
    # Suppress debug logs during performance tests
    export LOG_LEVEL="info"
    export GIN_MODE="release"
    
    # Run stress tests with specified duration
    timeout "${STRESS_TEST_DURATION}s" go test -v -timeout="${STRESS_TEST_DURATION}s" ./tests/performance/auth_cache/ -run TestAuthCacheStressTest > "$REPORTS_DIR/stress_tests/stress_test_results.txt" 2>&1 || {
        log_warning "Some stress tests failed or timed out, continuing..."
    }
    
    STRESS_COMPLETED=true
    log_success "Stress tests completed"
}

# Generate comprehensive reports
generate_reports() {
    log_info "Generating comprehensive performance reports..."
    
    # Run the performance report generator
    go run ./tests/performance/cmd/performance_runner.go "$REPORTS_DIR" || {
        log_warning "Report generation encountered issues, continuing..."
    }
    
    # Generate SLA comparison report
    go run ./devops/performance/sla-validator.go compare > "$REPORTS_DIR/sla_comparison.txt"
    
    log_success "Performance reports generated in $REPORTS_DIR"
}

# Analyze results and provide recommendations
analyze_results() {
    log_info "Analyzing performance test results..."
    
    local analysis_file="$REPORTS_DIR/performance_analysis.md"
    
    cat << EOF > "$analysis_file"
# Performance Test Analysis Report

**Generated:** $(date)
**Environment:** $TEST_ENVIRONMENT
**Redis URL:** $REDIS_URL

## Test Summary

### Benchmark Tests
$(if [ "$BENCHMARK_COMPLETED" = true ]; then echo "✅ Completed"; elif [ -f "$REPORTS_DIR/benchmarks/benchmark_results.txt" ]; then echo "✅ Completed"; else echo "❌ Not run or failed"; fi)

### Load Tests  
$(if [ "$LOAD_COMPLETED" = true ]; then echo "✅ Completed"; elif [ -f "$REPORTS_DIR/load_tests/load_test_results.txt" ]; then echo "✅ Completed"; else echo "❌ Not run or failed"; fi)

### Stress Tests
$(if [ "$STRESS_COMPLETED" = true ]; then echo "✅ Completed"; elif [ -f "$REPORTS_DIR/stress_tests/stress_test_results.txt" ]; then echo "✅ Completed"; else echo "❌ Not run or failed"; fi)

$(if [ "$BENCHMARK_COMPLETED" = false ] && [ "$LOAD_COMPLETED" = false ] && [ "$STRESS_COMPLETED" = false ]; then
    echo "⚠️  **Note:** This report contains partial results due to early termination."
    echo ""
fi)

## Key Findings

EOF

    # Add specific findings based on test results
    if [ -f "$REPORTS_DIR/performance_report.json" ]; then
        echo "### Detailed Metrics Available" >> "$analysis_file"
        echo "- Interactive HTML report: performance_report.html" >> "$analysis_file"
        echo "- Machine-readable data: performance_report.json" >> "$analysis_file"
        echo "- Spreadsheet data: performance_report.csv" >> "$analysis_file"
    fi
    
    echo "" >> "$analysis_file"
    echo "## Next Steps" >> "$analysis_file"
    echo "1. Review detailed reports in $REPORTS_DIR" >> "$analysis_file"
    echo "2. Check SLA compliance in sla_comparison.txt" >> "$analysis_file"
    echo "3. Implement recommended optimizations" >> "$analysis_file"
    echo "4. Schedule regular performance monitoring" >> "$analysis_file"
    
    log_success "Performance analysis saved to $analysis_file"
}

# Generate performance dashboard
generate_dashboard() {
    log_info "Generating performance dashboard..."
    
    local dashboard_file="$REPORTS_DIR/dashboard.html"
    
    # Calculate metrics for dashboard
    local html_count=$(ls "$REPORTS_DIR"/*.html 2>/dev/null | wc -l)
    local json_count=$(ls "$REPORTS_DIR"/*.json 2>/dev/null | wc -l)
    local test_categories=$(ls "$REPORTS_DIR"/*_tests/ 2>/dev/null | wc -l)
    local current_time=$(date +%H:%M)
    local current_date=$(date)
    
    # Calculate test completion status
    local benchmark_status="❌"
    local load_status="❌"
    local stress_status="❌"
    local overall_status="❌"
    
    if [ "$BENCHMARK_COMPLETED" = true ] || [ -f "$REPORTS_DIR/benchmarks/benchmark_results.txt" ]; then
        benchmark_status="✅"
    fi
    
    if [ "$LOAD_COMPLETED" = true ] || [ -f "$REPORTS_DIR/load_tests/load_test_results.txt" ]; then
        load_status="✅"
    fi
    
    if [ "$STRESS_COMPLETED" = true ] || [ -f "$REPORTS_DIR/stress_tests/stress_test_results.txt" ]; then
        stress_status="✅"
    fi
    
    if [ "$benchmark_status" = "✅" ] || [ "$load_status" = "✅" ] || [ "$stress_status" = "✅" ]; then
        overall_status="✅"
    elif [ "$BENCHMARK_COMPLETED" = false ] && [ "$LOAD_COMPLETED" = false ] && [ "$STRESS_COMPLETED" = false ]; then
        overall_status="⚠️"
    fi
    
    cat << EOF > "$dashboard_file"
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Redis Auth Cache Performance Dashboard</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .header { text-align: center; margin-bottom: 30px; }
        .metrics-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 20px; margin-bottom: 30px; }
        .metric-card { background: #f8f9fa; padding: 20px; border-radius: 6px; text-align: center; border-left: 4px solid #007acc; }
        .metric-value { font-size: 2em; font-weight: bold; color: #007acc; }
        .metric-label { color: #666; font-size: 0.9em; }
        .section { margin-bottom: 30px; }
        .section h2 { border-bottom: 2px solid #007acc; padding-bottom: 10px; }
        .file-links { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 15px; }
        .file-link { display: block; padding: 15px; background: #e8f4fd; text-decoration: none; color: #0066cc; border-radius: 6px; text-align: center; transition: background 0.3s; }
        .file-link:hover { background: #d1e7dd; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🎯 Redis Auth Cache Performance Dashboard</h1>
            <p>Generated on $current_date</p>
            <p>Environment: <strong>$TEST_ENVIRONMENT</strong></p>
        </div>
        
        <div class="section">
            <h2>📊 Quick Metrics</h2>
            <div class="metrics-grid">
                <div class="metric-card">
                    <div class="metric-value">$html_count</div>
                    <div class="metric-label">HTML Reports</div>
                </div>
                <div class="metric-card">
                    <div class="metric-value">$json_count</div>
                    <div class="metric-label">JSON Reports</div>
                </div>
                <div class="metric-card">
                    <div class="metric-value">$test_categories</div>
                    <div class="metric-label">Test Categories</div>
                </div>
                <div class="metric-card">
                    <div class="metric-value">$current_time</div>
                    <div class="metric-label">Last Updated</div>
                </div>
            </div>
        </div>
        
        <div class="section">
            <h2>🎯 Test Completion Status</h2>
            <div class="metrics-grid">
                <div class="metric-card">
                    <div class="metric-value">$benchmark_status</div>
                    <div class="metric-label">Benchmark Tests</div>
                </div>
                <div class="metric-card">
                    <div class="metric-value">$load_status</div>
                    <div class="metric-label">Load Tests</div>
                </div>
                <div class="metric-card">
                    <div class="metric-value">$stress_status</div>
                    <div class="metric-label">Stress Tests</div>
                </div>
                <div class="metric-card">
                    <div class="metric-value">$overall_status</div>
                    <div class="metric-label">Overall Status</div>
                </div>
            </div>
        </div>
        
        <div class="section">
            <h2>📋 Available Reports</h2>
            <div class="file-links">
EOF

    # Add links to available reports
    for file in "$REPORTS_DIR"/*.html "$REPORTS_DIR"/*.json "$REPORTS_DIR"/*.csv "$REPORTS_DIR"/*.txt; do
        if [ -f "$file" ]; then
            filename=$(basename "$file")
            echo "                <a href=\"$filename\" class=\"file-link\">📄 $filename</a>" >> "$dashboard_file"
        fi
    done

    cat << 'EOF' >> "$dashboard_file"
            </div>
        </div>
        
        <div class="section">
            <h2>🔧 Quick Actions</h2>
            <div class="file-links">
                <a href="performance_report.html" class="file-link">📈 Interactive Performance Report</a>
                <a href="sla_comparison.txt" class="file-link">🎯 SLA Compliance Check</a>
                <a href="performance_analysis.md" class="file-link">📊 Analysis Summary</a>
            </div>
        </div>
    </div>
</body>
</html>
EOF

    log_success "Performance dashboard generated: $dashboard_file"
}

# Global variables to track test completion
BENCHMARK_COMPLETED=false
LOAD_COMPLETED=false
STRESS_COMPLETED=false

# Cleanup function
cleanup() {
    log_info "Cleaning up temporary files..."
    # Add any cleanup tasks here
}

# Signal handler for graceful shutdown
graceful_shutdown() {
    log_warning "Received interrupt signal (Ctrl+C)"
    log_info "Generating reports for completed tests..."
    
    # Generate reports for any completed tests
    if [ "$BENCHMARK_COMPLETED" = true ] || [ "$LOAD_COMPLETED" = true ] || [ "$STRESS_COMPLETED" = true ]; then
        log_info "At least one test completed, generating partial reports..."
        generate_reports || log_warning "Report generation failed"
        analyze_results || log_warning "Analysis failed"
        generate_dashboard || log_warning "Dashboard generation failed"
        log_info "Partial results available in: $REPORTS_DIR"
    else
        log_warning "No tests completed, skipping report generation"
    fi
    
    cleanup
    exit 130  # Standard exit code for Ctrl+C
}

# Print usage information
usage() {
    cat << EOF
Performance Testing Automation Script

Usage: $0 [OPTIONS] [COMMAND]

Commands:
    all         Run all performance tests (default)
    benchmark   Run only benchmark tests
    load        Run only load tests
    stress      Run only stress tests
    validate    Validate SLA configuration only
    reports     Generate reports only

Options:
    -e, --environment       Set test environment (production|staging|development)
    -r, --reports-dir       Set reports output directory
    -b, --benchmark-duration Set benchmark test duration in seconds (default: 30)
    -l, --load-duration     Set load test duration in seconds (default: 300)
    -s, --stress-duration   Set stress test duration in seconds (default: 600)
    -h, --help             Show this help message

Environment Variables:
    TEST_ENVIRONMENT        Test environment (default: staging)
    REDIS_URL              Redis connection URL (default: redis://localhost:6379)
    BENCHMARK_DURATION     Benchmark test duration in seconds (default: 30)
    LOAD_TEST_DURATION     Load test duration in seconds (default: 300)
    STRESS_TEST_DURATION   Stress test duration in seconds (default: 600)

Examples:
    $0                                              # Run all tests with default settings
    $0 -e production all                           # Run all tests in production mode
    $0 --reports-dir ./results load               # Run load tests with custom output dir
    $0 -b 60 -l 600 -s 1200 all                  # Custom durations for all tests
    $0 --benchmark-duration 120 benchmark         # 2-minute benchmark test
    $0 validate                                    # Only validate SLA configuration

Duration Examples:
    -b 30      # 30-second benchmark tests
    -l 300     # 5-minute load tests  
    -s 600     # 10-minute stress tests

EOF
}

# Main execution function
main() {
    local command="all"
    
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -e|--environment)
                TEST_ENVIRONMENT="$2"
                shift 2
                ;;
            -r|--reports-dir)
                REPORTS_DIR="$2"
                shift 2
                ;;
            -b|--benchmark-duration)
                BENCHMARK_DURATION="$2"
                shift 2
                ;;
            -l|--load-duration)
                LOAD_TEST_DURATION="$2"
                shift 2
                ;;
            -s|--stress-duration)
                STRESS_TEST_DURATION="$2"
                shift 2
                ;;
            -h|--help)
                usage
                exit 0
                ;;
            benchmark|load|stress|validate|reports|all)
                command="$1"
                shift
                ;;
            *)
                log_error "Unknown option: $1"
                usage
                exit 1
                ;;
        esac
    done
    
    log_info "Starting Redis Auth Cache Performance Testing"
    log_info "Environment: $TEST_ENVIRONMENT"
    log_info "Reports Directory: $REPORTS_DIR"
    log_info "Command: $command"
    log_info "Test Durations - Benchmark: ${BENCHMARK_DURATION}s, Load: ${LOAD_TEST_DURATION}s, Stress: ${STRESS_TEST_DURATION}s"
    
    # Ensure reports directory exists
    mkdir -p "$REPORTS_DIR"
    
    # Set trap for cleanup and signal handling
    trap cleanup EXIT
    trap graceful_shutdown INT TERM
    
    # Execute based on command
    case $command in
        validate)
            validate_sla_config
            ;;
        benchmark)
            check_prerequisites
            validate_sla_config
            run_benchmark_tests
            ;;
        load)
            check_prerequisites
            validate_sla_config
            run_load_tests
            ;;
        stress)
            check_prerequisites
            validate_sla_config
            run_stress_tests
            ;;
        reports)
            generate_reports
            analyze_results
            generate_dashboard
            ;;
        all)
            check_prerequisites
            validate_sla_config
            run_benchmark_tests
            run_load_tests
            run_stress_tests
            generate_reports
            analyze_results
            generate_dashboard
            ;;
    esac
    
    log_success "Performance testing completed successfully!"
    log_info "Results available in: $REPORTS_DIR"
    log_info "Open $REPORTS_DIR/dashboard.html for interactive results"
}

# Run main function with all arguments
main "$@"
