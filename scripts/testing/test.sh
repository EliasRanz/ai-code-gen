#!/bin/bash
# Unified testing script with clean CLI output and automatic cleanup
# Supports local development environment
#
# Environment Guidelines:
# - local: Uses developer's running local environment. Tests fail immediately if local services are not running.

set -euo pipefail

# Security: Use PROJECT_ROOT variable instead of absolute paths
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$PROJECT_ROOT"

# Environment configuration
ENVIRONMENT="${ENVIRONMENT:-local}"
SUPPORTED_ENVIRONMENTS=("local")

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

# Validate environment parameter
validate_environment() {
    local env="$1"
    for supported_env in "${SUPPORTED_ENVIRONMENTS[@]}"; do
        if [ "$env" = "$supported_env" ]; then
            return 0
        fi
    done
    return 1
}

# Load environment-specific configuration
load_environment_config() {
    local env="$1"
    local env_file="$PROJECT_ROOT/tests/integration/.env.$env"

    if [ -f "$env_file" ]; then
        log_info "Loading environment configuration: $env"
        # Use set +u temporarily to allow unbound variables in env files
        set +u
        source "$env_file"
        set -u
    else
        log_info "No environment file found for $env, using defaults"
    fi

    # Set environment-specific variables
    export ENVIRONMENT="$env"

    # Set compose file based on environment
    case "$env" in
        "local")
            export COMPOSE_FILE="$PROJECT_ROOT/deployments/docker/docker-compose.yml"
            ;;
        *)
            export COMPOSE_FILE="$PROJECT_ROOT/deployments/docker/docker-compose.yml"
            ;;
    esac

    # Use main compose file for all environments
    log_info "Using unified compose file: $COMPOSE_FILE"
}

# Display current environment configuration
show_environment_info() {
    echo ""
    echo "🌍 Environment Configuration"
    echo "==========================="
    echo "Environment: $ENVIRONMENT"
    echo "Compose File: $COMPOSE_FILE"
    echo "Project Root: $PROJECT_ROOT"

    if [ -n "${DATABASE_URL:-}" ]; then
        echo "Database URL: [CONFIGURED]"
    fi
    if [ -n "${REDIS_URL:-}" ]; then
        echo "Redis URL: [CONFIGURED]"
    fi
    echo ""
}

# Security: Validate Go version before proceeding
check_go_version() {
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed or not in PATH"
        exit 1
    fi

    local go_version
    go_version=$(go version | awk '{print $3}' | sed 's/go//')
    log_info "Go version: $go_version"
}

# Security: Safe tool installation with version pinning
install_tools() {
    log_info "Installing testing tools with version pinning..."

    # Install gotestsum (pinned version for security)
    if ! command -v gotestsum &> /dev/null; then
        log_info "Installing gotestsum..."
        go install gotest.tools/gotestsum@v1.10.0
    fi

    # Install gocovmerge (pinned version)
    if ! command -v gocovmerge &> /dev/null; then
        log_info "Installing gocovmerge..."
        go install github.com/wadey/gocovmerge@latest
    fi

    log_success "Tool installation completed"
}

# Start Docker services required for testing
start_docker_services() {
    log_info "Starting Docker services for testing (Environment: $ENVIRONMENT)..."

    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed or not in PATH"
        exit 1
    fi

    if ! command -v docker-compose &> /dev/null; then
        log_error "docker-compose is not installed or not in PATH"
        exit 1
    fi

    if [ ! -f "$COMPOSE_FILE" ]; then
        log_error "Docker Compose file not found: $COMPOSE_FILE"
        exit 1
    fi

    # Start services in detached mode
    log_info "Starting services defined in $COMPOSE_FILE..."
    if ! docker-compose -f "$COMPOSE_FILE" up -d; then
        log_error "Failed to start Docker services"
        exit 1
    fi

    # Wait for services to be healthy
    log_info "Waiting for services to be ready..."
    wait_for_services
}

# Wait for Docker services to be healthy
wait_for_services() {
    local max_attempts=20
    local attempt=1

    while [ $attempt -le $max_attempts ]; do
        log_info "Checking service health (attempt $attempt/$max_attempts)..."

        # Check Redis health
        if docker-compose -f "$COMPOSE_FILE" exec -T redis redis-cli ping 2>/dev/null | grep -q "PONG"; then
            redis_ready=true
            log_info "Redis is ready"
        else
            redis_ready=false
            log_info "Redis is not ready yet"
        fi

        # Check PostgreSQL health
        if docker-compose -f "$COMPOSE_FILE" exec -T postgres pg_isready -U postgres >/dev/null 2>&1; then
            postgres_ready=true
            log_info "PostgreSQL is ready"
        else
            postgres_ready=false
            log_info "PostgreSQL is not ready yet"
        fi

        if [ "$redis_ready" = true ] && [ "$postgres_ready" = true ]; then
            log_success "All services are ready!"
            return 0
        fi

        log_info "Services not ready yet. Waiting 3 seconds..."
        sleep 3
        ((attempt++))
    done

    log_error "Services failed to start within $max_attempts attempts"
    stop_docker_services
    exit 1
}

# Check if required services are available for the current environment
check_service_availability() {
    log_info "Checking service availability for $ENVIRONMENT environment..."

    case "$ENVIRONMENT" in
        "local")
            # In local environment, check if developer's full application stack is running
            # This uses the main docker-compose.yml which includes all services
            log_info "Checking local development environment status..."

            # Use make dev-status to check if services are running
            if ! make dev-status >/dev/null 2>&1; then
                log_error "Local development services are not running"
                log_error "Please start your local services: make dev"
                return 1
            fi

            # Check if key application services are responding
            if ! check_service_health "localhost" "8080"; then
                log_error "API Gateway service is not available on localhost:8080"
                log_error "Please ensure all services are running: make dev"
                return 1
            fi

            log_success "All local services are available"
            ;;

        *)
            log_error "Unknown environment: $ENVIRONMENT"
            return 1
            ;;
    esac
}

# Check Redis connection
check_redis_connection() {
    # Try to connect to Redis (use environment variables if available, otherwise defaults)
    local redis_host="${REDIS_HOST:-localhost}"
    local redis_port="${REDIS_PORT:-6379}"

    if command -v redis-cli &> /dev/null; then
        if timeout 5 redis-cli -h "$redis_host" -p "$redis_port" ping 2>/dev/null | grep -q "PONG"; then
            return 0
        fi
    fi

    # Fallback: try netcat/telnet if redis-cli is not available
    if command -v nc &> /dev/null; then
        if echo "PING" | timeout 5 nc "$redis_host" "$redis_port" 2>/dev/null | grep -q "PONG"; then
            return 0
        fi
    fi

    return 1
}

# Check PostgreSQL connection
check_postgres_connection() {
    # Try to connect to PostgreSQL (use environment variables if available, otherwise defaults)
    local db_host="${DB_HOST:-localhost}"
    local db_port="${DB_PORT:-5432}"
    local db_user="${DB_USER:-postgres}"
    local db_name="${DB_NAME:-postgres}"

    if command -v pg_isready &> /dev/null; then
        if pg_isready -h "$db_host" -p "$db_port" -U "$db_user" -d "$db_name" >/dev/null 2>&1; then
            return 0
        fi
    fi

    # Fallback: try psql if available
    if command -v psql &> /dev/null; then
        if PGPASSWORD="${DB_PASSWORD:-}" psql -h "$db_host" -p "$db_port" -U "$db_user" -d "$db_name" -c "SELECT 1;" >/dev/null 2>&1; then
            return 0
        fi
    fi

    return 1
}

# Check if a service is healthy by making an HTTP request
check_service_health() {
    local host="$1"
    local port="$2"
    local url="http://$host:$port/health"

    # Try wget first
    if command -v wget &> /dev/null; then
        if wget --no-verbose --tries=1 --spider "$url" >/dev/null 2>&1; then
            return 0
        fi
    fi

    # Fallback: try curl
    if command -v curl &> /dev/null; then
        if curl --silent --fail --max-time 5 "$url" >/dev/null 2>&1; then
            return 0
        fi
    fi

    return 1
}

# Run tests with coverage
run_tests() {
    local test_type="${1:-all}"

    log_info "Running $test_type tests (Environment: $ENVIRONMENT)..."
    show_environment_info

    # Check service availability for integration tests
    if [ "$test_type" = "integration" ] || [ "$test_type" = "all" ]; then
        if ! check_service_availability; then
            log_error "Required services are not available. Cannot run integration tests."
            exit 1
        fi
    fi

    # Set up cleanup trap for files
    trap cleanup_files EXIT

    local test_cmd="go test"
    local cover_profile="/tmp/coverage_$$.out"

    case "$test_type" in
        "unit")
            test_cmd="$test_cmd ./tests/unit/..."
            ;;
        "integration")
            test_cmd="$test_cmd -tags=integration ./tests/integration/..."
            ;;
        "all")
            # Run unit tests first (no service dependencies)
            local unit_cover_profile="/tmp/unit_coverage_$$.out"
            log_info "Running unit tests..."
            if ! go test ./tests/unit/... -coverpkg=./internal/... -coverprofile=$unit_cover_profile -timeout=30s >/tmp/unit_test_output_$$.txt 2>&1; then
                echo "Unit tests failed. Output:"
                cat /tmp/unit_test_output_$$.txt
                rm -f /tmp/unit_test_output_$$.txt
                exit 1
            fi
            rm -f /tmp/unit_test_output_$$.txt

            # Check service availability before running integration tests
            if ! check_service_availability; then
                log_error "Required services are not available. Cannot run integration tests."
                exit 1
            fi

            # Then run integration tests (with build tags)
            local integration_cover_profile="/tmp/integration_coverage_$$.out"
            log_info "Running integration tests..."
            if ! go test -tags=integration ./tests/integration/... -coverpkg=./internal/... -coverprofile=$integration_cover_profile -timeout=30s >/tmp/integration_test_output_$$.txt 2>&1; then
                echo "Integration tests failed. Output:"
                cat /tmp/integration_test_output_$$.txt
                rm -f /tmp/integration_test_output_$$.txt
                exit 1
            fi
            rm -f /tmp/integration_test_output_$$.txt

            # Merge coverage profiles if both exist
            if [ -f "$unit_cover_profile" ] && [ -f "$integration_cover_profile" ]; then
                if command -v gocovmerge &> /dev/null; then
                    log_info "Merging coverage profiles..."
                    gocovmerge "$unit_cover_profile" "$integration_cover_profile" > "$cover_profile"
                else
                    # Fallback: use the integration coverage (more comprehensive)
                    cp "$integration_cover_profile" "$cover_profile"
                fi
            elif [ -f "$unit_cover_profile" ]; then
                cp "$unit_cover_profile" "$cover_profile"
            elif [ -f "$integration_cover_profile" ]; then
                cp "$integration_cover_profile" "$cover_profile"
            fi

            # Clean up temporary coverage files
            rm -f "$unit_cover_profile" "$integration_cover_profile"

            log_success "All tests completed successfully"
            display_coverage "$cover_profile"
            return 0
            ;;
        *)
            log_error "Invalid test type: $test_type"
            echo "Valid options: unit, integration, all"
            return 1
            ;;
    esac

    # For unit and integration cases, continue with the standard test execution
    # Add coverage flags (remove -v for cleaner output)
    test_cmd="$test_cmd -coverpkg=./internal/... -coverprofile=$cover_profile -timeout=30s"

    # Use gotestsum if available for better output (without XML file generation)
    if false && command -v gotestsum &> /dev/null; then
        log_info "Using gotestsum for enhanced output..."
        echo ""
        echo "🧪 Running tests..."
        echo ""

        # Run tests with gotestsum (no XML output, simpler format)
        if gotestsum --format=testname -- $test_cmd 2>/dev/null; then
            log_success "Tests completed successfully"
            display_coverage "$cover_profile"
            return 0
        else
            log_warning "gotestsum failed, falling back to go test..."
            run_tests_fallback "$test_cmd" "$cover_profile"
            return $?
        fi
    else
        log_info "Using standard go test..."
        run_tests_fallback "$test_cmd" "$cover_profile"
        return $?
    fi

    # Clean up temporary coverage file
    rm -f "$cover_profile"
}

# Fallback test runner with clean output
run_tests_fallback() {
    local test_cmd="$1"
    local cover_profile="$2"

    echo ""
    echo "🧪 Running tests..."
    echo ""

    # Run tests and capture results
    local output_file="/tmp/test_output_$$.txt"
    local exit_code=0

    # Run tests quietly but capture output
    if eval "$test_cmd" > "$output_file" 2>&1; then
        log_success "Tests completed successfully"
        display_coverage "$cover_profile"
    else
        exit_code=$?
        log_error "Some tests failed"
        display_coverage "$cover_profile"
    fi

    # Parse and display results in a clean format
    display_test_results "$output_file" "$exit_code"

    # Clean up
    rm -f "$output_file"

    return $exit_code
}

# Display coverage information directly to CLI
display_coverage() {
    local cover_profile="$1"

    if [ ! -f "$cover_profile" ]; then
        return
    fi

    echo ""
    echo "📊 Coverage Summary"
    echo "=================="

    local coverage_summary
    if ! coverage_summary=$(go tool cover -func="$cover_profile" 2>/dev/null | tail -1 2>/dev/null); then
        echo "Unable to generate coverage summary"
        return
    fi

    if [ -n "$coverage_summary" ]; then
        local coverage_percent
        coverage_percent=$(echo "$coverage_summary" | awk '{print $3}' 2>/dev/null || echo "")

        if [ -n "$coverage_percent" ] && [[ "$coverage_percent" =~ ^[0-9.]+$ ]]; then
            local percent_num
            percent_num=$(echo "$coverage_percent" | sed 's/%//' 2>/dev/null || echo "0")

            if [ -n "$percent_num" ] && [[ "$percent_num" =~ ^[0-9.]+$ ]]; then
                if (( $(echo "$percent_num >= 80" | bc -l 2>/dev/null || echo "0") )); then
                    echo -e "🎯 Overall Coverage: ${GREEN}$coverage_percent${NC}"
                elif (( $(echo "$percent_num >= 60" | bc -l 2>/dev/null || echo "0") )); then
                    echo -e "⚠️  Overall Coverage: ${YELLOW}$coverage_percent${NC}"
                else
                    echo -e "❌ Overall Coverage: ${RED}$coverage_percent${NC}"
                fi
            else
                echo "📈 Overall Coverage: $coverage_percent"
            fi
        else
            echo "📈 Overall Coverage: $coverage_summary"
        fi
    fi

    echo ""
}

# Display test results in a clean, readable format
display_test_results() {
    local output_file="$1"
    local exit_code="$2"

    echo ""
    echo "📊 Test Results Summary"
    echo "======================"

    # Count total tests, passes, fails
    local total_tests passed_tests failed_tests skipped_tests
    total_tests=$(grep -c "^=== RUN" "$output_file" 2>/dev/null | tr -d '[:space:]' || echo "0")
    total_tests=$(echo "$total_tests" | sed 's/[^0-9]*//g' || echo "0")
    if [ -z "$total_tests" ] || [ "$total_tests" = "0" ]; then
        total_tests=0
    fi

    passed_tests=$(grep -c "^--- PASS:" "$output_file" 2>/dev/null | tr -d '[:space:]' || echo "0")
    passed_tests=$(echo "$passed_tests" | sed 's/[^0-9]*//g' || echo "0")
    if [ -z "$passed_tests" ] || [ "$passed_tests" = "0" ]; then
        passed_tests=0
    fi

    failed_tests=$(grep -c "^--- FAIL:" "$output_file" 2>/dev/null | tr -d '[:space:]' || echo "0")
    failed_tests=$(echo "$failed_tests" | sed 's/[^0-9]*//g' || echo "0")
    if [ -z "$failed_tests" ] || [ "$failed_tests" = "0" ]; then
        failed_tests=0
    fi

    skipped_tests=$(grep -c "^--- SKIP:" "$output_file" 2>/dev/null | tr -d '[:space:]' || echo "0")
    skipped_tests=$(echo "$skipped_tests" | sed 's/[^0-9]*//g' || echo "0")
    if [ -z "$skipped_tests" ] || [ "$skipped_tests" = "0" ]; then
        skipped_tests=0
    fi

    # Display summary
    if [ "$total_tests" -gt 0 ]; then
        echo "Total Tests:  $total_tests"
        echo "✅ Passed:    $passed_tests"
        echo "❌ Failed:    $failed_tests"
        echo "⏭️  Skipped:   $skipped_tests"
        echo ""

        if [ "$failed_tests" -gt 0 ]; then
            echo ""
            echo "❌ Failed Tests:"
            echo "---------------"
            grep "^--- FAIL:" "$output_file" | sed 's/--- FAIL: //' | head -5
            if [ "$failed_tests" -gt 5 ]; then
                local remaining=$((failed_tests - 5))
                echo "... and $remaining more"
            fi
        fi

        # Show test duration
        local duration=$(grep "^PASS$" "$output_file" | tail -1 | awk '{print $2}' || echo "")
        if [ -n "$duration" ]; then
            echo ""
            echo "⏱️  Duration:  $duration"
        fi
    else
        # If no individual test results found, try to parse package-level results
        local package_count=$(grep -c "^ok " "$output_file" 2>/dev/null | tr -d '[:space:]' || echo "0")
        package_count=$(echo "$package_count" | sed 's/[^0-9]*//g' || echo "0")
        if [ -z "$package_count" ] || [ "$package_count" = "0" ]; then
            package_count=0
        fi

        local failed_packages=$(grep -c "^FAIL" "$output_file" 2>/dev/null | tr -d '[:space:]' || echo "0")
        failed_packages=$(echo "$failed_packages" | sed 's/[^0-9]*//g' || echo "0")
        if [ -z "$failed_packages" ] || [ "$failed_packages" = "0" ]; then
            failed_packages=0
        fi

        if [ "$package_count" -gt 0 ]; then
            local passed_packages=$((package_count - failed_packages))
            echo "Package Tests: $package_count"
            echo "✅ Passed:     $passed_packages"
            echo "❌ Failed:     $failed_packages"
            echo ""

            if [ "$failed_packages" -gt 0 ]; then
                echo "❌ Failed Packages:"
                echo "------------------"
                grep "^FAIL" "$output_file" | head -5
            fi
        else
            echo "No tests found or executed"
        fi
    fi

    echo ""
}

# Generate coverage reports (CLI output only)
generate_coverage() {
    log_info "Generating coverage reports..."

    # Look for any existing coverage files
    local cover_file=""
    if [ -f "coverage.out" ]; then
        cover_file="coverage.out"
    elif [ -f "unit_coverage.out" ]; then
        cover_file="unit_coverage.out"
    elif [ -f "integration_coverage.out" ]; then
        cover_file="integration_coverage.out"
    else
        log_error "No coverage files found. Run tests first."
        return 1
    fi

    # Display detailed coverage information
    display_detailed_coverage "$cover_file"

    log_success "Coverage analysis completed"
}

# Display detailed coverage information
display_detailed_coverage() {
    local cover_file="$1"

    echo ""
    echo "📊 Detailed Coverage Analysis"
    echo "============================"

    # Show overall coverage
    local coverage_summary
    coverage_summary=$(go tool cover -func="$cover_file" | tail -1)

    if [ -n "$coverage_summary" ]; then
        local coverage_percent
        coverage_percent=$(echo "$coverage_summary" | awk '{print $3}')

        # Color code coverage percentage
        if [[ "$coverage_percent" =~ ^[0-9.]+$ ]]; then
            local percent_num
            percent_num=$(echo "$coverage_percent" | sed 's/%//')

            if (( $(echo "$percent_num >= 80" | bc -l) )); then
                echo -e "🎯 Overall Coverage: ${GREEN}$coverage_percent${NC}"
            elif (( $(echo "$percent_num >= 60" | bc -l) )); then
                echo -e "⚠️  Overall Coverage: ${YELLOW}$coverage_percent${NC}"
            else
                echo -e "❌ Overall Coverage: ${RED}$coverage_percent${NC}"
            fi
        else
            echo "📈 Overall Coverage: $coverage_percent"
        fi
    fi

    echo ""
    echo "📋 Coverage by Function:"
    echo "-----------------------"
    go tool cover -func="$cover_file" | head -20

    echo ""
    echo "💡 Tip: Run 'make test-analyze' for package-level analysis"
}

# Analyze coverage by package
analyze_coverage() {
    log_info "Analyzing coverage by package..."

    # Look for any existing coverage files
    local cover_file=""
    if [ -f "coverage.out" ]; then
        cover_file="coverage.out"
    elif [ -f "unit_coverage.out" ]; then
        cover_file="unit_coverage.out"
    elif [ -f "integration_coverage.out" ]; then
        cover_file="integration_coverage.out"
    else
        log_error "No coverage file found. Run tests first."
        return 1
    fi

    echo ""
    echo "📊 Package Coverage Analysis"
    echo "==========================="

    # Show coverage by service in a nice table format
    echo ""
    printf "%-12s %-10s %-10s %-10s\n" "Service" "Coverage" "Files" "Status"
    printf "%-12s %-10s %-10s %-10s\n" "--------" "--------" "-----" "------"

    local total_files=0
    local service_count=0

    for service in ai auth user cache config gateway observability utilities; do
        local coverage
        local file_count
        coverage=$(go tool cover -func="$cover_file" | grep "internal/${service}" | awk '{sum += $3; count++} END {if (count > 0) printf "%.1f%%", sum/count}' || echo "")
        file_count=$(go tool cover -func="$cover_file" | grep -c "internal/${service}" || echo "0")

        if [ -n "$coverage" ] && [ "$file_count" -gt 0 ]; then
            # Determine status based on coverage
            local status="❌ Low"
            if [[ "$coverage" =~ ^[0-9.]+% ]]; then
                local percent_num
                percent_num=$(echo "$coverage" | sed 's/%//')
                if (( $(echo "$percent_num >= 80" | bc -l) )); then
                    status="✅ Good"
                elif (( $(echo "$percent_num >= 60" | bc -l) )); then
                    status="⚠️  Medium"
                fi
            fi

            printf "%-12s %-10s %-10s %-10s\n" "$service" "$coverage" "$file_count" "$status"

            total_files=$((total_files + file_count))
            service_count=$((service_count + 1))
        fi
    done

    echo ""
    echo "📈 Low Coverage Files (< 50%)"
    echo "=============================="

    local low_coverage_files
    low_coverage_files=$(go tool cover -func="$cover_file" | awk '$3 != "total:" && $3 != "(statements)" && substr($3, 1, length($3)-1) < 50.0 { print $1 " - " $3 }' | head -5)

    if [ -n "$low_coverage_files" ]; then
        echo "$low_coverage_files"
    else
        echo "🎉 No files with coverage below 50%!"
    fi

    echo ""
    log_success "Coverage analysis completed"
}

# Display coverage summary
display_coverage() {
    local cover_file="$1"

    if [ ! -f "$cover_file" ]; then
        log_warning "Coverage file not found: $cover_file"
        return
    fi

    echo ""
    echo "📊 Coverage Summary"
    echo "=================="

    # Show overall coverage
    local coverage_summary
    coverage_summary=$(go tool cover -func="$cover_file" | tail -1)

    if [ -n "$coverage_summary" ]; then
        local coverage_percent
        coverage_percent=$(echo "$coverage_summary" | awk '{print $3}')

        if [ -n "$coverage_percent" ]; then
            echo "🎯 Overall Coverage: $coverage_percent"
        fi
    fi

    echo ""
}
# Cleanup function to remove generated files
cleanup_files() {
    log_info "Cleaning up generated files..."

    # Remove coverage files (ignore if they don't exist)
    rm -f coverage*.out coverage*.html 2>/dev/null || true

    # Remove test report files (ignore if they don't exist)
    rm -f tests-report.xml 2>/dev/null || true

    # Remove any temporary files created during testing (ignore if they don't exist)
    rm -f /tmp/test_output_*.txt 2>/dev/null || true
    rm -f /tmp/coverage_*.out 2>/dev/null || true

    log_success "Cleanup completed"
}

# Check Go version
check_go_version() {
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed or not in PATH"
        exit 1
    fi

    local go_version
    go_version=$(go version | awk '{print $3}' | sed 's/go//')
    log_info "Using Go version: $go_version"
}

# Main execution
main() {
    local command="${1:-test}"

    check_go_version
    load_environment_config "$ENVIRONMENT"
    show_environment_info

    # Set up cleanup trap
    trap cleanup_files EXIT

    case "$command" in
        "install")
            install_tools
            ;;
        "test")
            run_tests "${2:-all}"
            exit $?
            ;;
        "unit")
            run_tests "unit"
            exit $?
            ;;
        "integration")
            run_tests "integration"
            exit $?
            ;;
        "coverage")
            generate_coverage
            ;;
        "analyze")
            analyze_coverage
            ;;
        "all")
            install_tools
            run_tests "all"
            ;;
        *)
            echo "Usage: $0 [install|test|unit|integration|coverage|analyze|all]"
            echo ""
            echo "Commands:"
            echo "  install     - Install testing tools"
            echo "  test        - Run all tests with coverage (CLI output only)"
            echo "  unit        - Run unit tests only (CLI output only)"
            echo "  integration - Run integration tests only (CLI output only)"
            echo "  coverage    - Show coverage reports (CLI output only)"
            echo "  analyze     - Analyze coverage by package (CLI output only)"
            echo "  all         - Install tools and run everything (CLI output only)"
            exit 1
            ;;
    esac
}

# Call the main function
main "$@"
