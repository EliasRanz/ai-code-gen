#!/bin/bash

# Integration Test Runner
# Runs integration tests against specific deployed runtime environments
# Usage: ./run_integration_tests.sh [environment] [test_pattern]
# Environments: local, staging, production
# Examples:
#   ./run_integration_tests.sh local
#   ./run_integration_tests.sh staging
#   ./run_integration_tests.sh production TestCacheIntegration

set -e

# Default values
ENVIRONMENT="${1:-local}"
TEST_PATTERN="${2:-.*}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

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

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

# Validate environment parameter
validate_environment() {
    local env="$1"
    case "$env" in
        local|staging|production)
            return 0
            ;;
        *)
            log_error "Invalid environment: $env"
            log_error "Valid environments: local, staging, production"
            exit 1
            ;;
    esac
}

# Check if environment file exists
check_environment_file() {
    local env="$1"
    local env_file="$SCRIPT_DIR/.env.$env"

    if [[ ! -f "$env_file" ]]; then
        log_error "Environment file not found: $env_file"
        exit 1
    fi

    log_info "Using environment file: $env_file"
}

# Validate environment connectivity
validate_environment_connectivity() {
    local env="$1"

    log_info "Validating environment connectivity for: $env"

    # Load environment variables
    set -a
    source "$SCRIPT_DIR/.env.$env"
    set +a

    # Test Redis connectivity
    if command -v redis-cli &> /dev/null; then
        log_info "Testing Redis connectivity..."
        if redis-cli -h "${REDIS_HOST:-localhost}" -p "${REDIS_PORT:-6379}" ping &> /dev/null; then
            log_success "Redis connection successful"
        else
            log_warn "Redis connection failed - tests may fail"
        fi
    else
        log_warn "redis-cli not available - skipping Redis connectivity test"
    fi

    # Test database connectivity (if psql is available)
    if command -v psql &> /dev/null; then
        log_info "Testing database connectivity..."
        local conn_string="host=${DB_HOST:-localhost} port=${DB_PORT:-5432} user=${DB_USER:-postgres} password=${DB_PASSWORD:-password} dbname=${DB_NAME:-ai_code_gen_test} sslmode=${DB_SSLMODE:-disable}"
        if psql "$conn_string" -c "SELECT 1;" &> /dev/null; then
            log_success "Database connection successful"
        else
            log_warn "Database connection failed - tests may fail"
        fi
    else
        log_warn "psql not available - skipping database connectivity test"
    fi
}

# Setup environment for Docker services (local environments)
setup_docker_services() {
    local env="$1"

    if [[ "$env" == "local" ]]; then
        log_info "Setting up Docker services for $env environment..."

        # Check if docker-compose is available
        if ! command -v docker-compose &> /dev/null && ! command -v docker &> /dev/null; then
            log_error "Docker/Docker Compose not available. Required for $env environment."
            exit 1
        fi

        # Start services
        cd "$PROJECT_ROOT/deployments/docker"
        if command -v docker-compose &> /dev/null; then
            docker-compose -f docker-compose.yml up -d
        else
            docker compose -f docker-compose.yml up -d
        fi

        # Wait for services to be ready
        log_info "Waiting for services to be ready..."
        sleep 10

        # Health checks
        log_info "Performing health checks..."
        for i in {1..30}; do
            if redis-cli -h localhost -p 6379 ping &> /dev/null && \
               psql "host=localhost port=5432 user=postgres password=password dbname=ai_ui_generator_test sslmode=disable" -c "SELECT 1;" &> /dev/null; then
                log_success "All services are ready"
                break
            fi
            if [[ $i -eq 30 ]]; then
                log_error "Services failed to start within timeout"
                exit 1
            fi
            log_info "Waiting for services... ($i/30)"
            sleep 2
        done
    fi
}

# Cleanup Docker services
cleanup_docker_services() {
    local env="$1"

    if [[ "$env" == "local" ]]; then
        log_info "Cleaning up Docker services..."

        cd "$PROJECT_ROOT/deployments/docker"
        if command -v docker-compose &> /dev/null; then
            docker-compose -f docker-compose.yml down -v
        else
            docker compose -f docker-compose.yml down -v
        fi
    fi
}

# Run integration tests
run_integration_tests() {
    local env="$1"
    local pattern="$2"

    log_info "Running integration tests for environment: $env"
    log_info "Test pattern: $pattern"

    cd "$PROJECT_ROOT"

    # Set environment variable for config loading
    export INTEGRATION_ENV="$env"

    # Run tests with integration build tag
    go test -tags=integration ./tests/integration/... \
        -v \
        -run "$pattern" \
        -timeout 5m \
        --count=1

    local test_exit_code=$?

    if [[ $test_exit_code -eq 0 ]]; then
        log_success "Integration tests passed for $env environment"
    else
        log_error "Integration tests failed for $env environment"
    fi

    return $test_exit_code
}

# Show usage information
show_usage() {
    cat << EOF
Integration Test Runner

Runs integration tests against specific deployed runtime environments.

USAGE:
    $0 [environment] [test_pattern]

ENVIRONMENTS:
    local      - Local development environment (uses Docker services)
    staging    - Staging/pre-production environment
    production - Production environment (use with extreme caution)

ARGUMENTS:
    environment    - Target environment (default: local)
    test_pattern   - Test pattern to run (default: .*, runs all tests)

EXAMPLES:
    $0 local                    # Run all tests against local environment
    $0 staging                  # Run all tests against staging
    $0 production TestCache     # Run cache tests against production

ENVIRONMENT FILES:
    .env.local      - Local environment configuration
    .env.staging    - Staging environment configuration
    .env.production - Production environment configuration

REQUIREMENTS:
    - Go 1.19+
    - For local: Docker and Docker Compose
    - For staging/production: Network access to target environment
    - Environment-specific credentials configured

EOF
}

# Main execution
main() {
    # Show usage if requested
    if [[ "$1" == "--help" || "$1" == "-h" ]]; then
        show_usage
        exit 0
    fi

    # Validate environment
    validate_environment "$ENVIRONMENT"

    # Check environment file
    check_environment_file "$ENVIRONMENT"

    # Validate connectivity (skip for production to avoid accidental connections)
    if [[ "$ENVIRONMENT" != "production" ]]; then
        validate_environment_connectivity "$ENVIRONMENT"
    else
        log_warn "Skipping connectivity validation for production environment"
        log_warn "Ensure you have proper credentials and network access"
    fi

    # Setup Docker services if needed
    setup_docker_services "$ENVIRONMENT"

    # Run tests
    local exit_code=0
    if run_integration_tests "$ENVIRONMENT" "$TEST_PATTERN"; then
        log_success "All integration tests completed successfully"
    else
        exit_code=$?
        log_error "Integration tests failed"
    fi

    # Cleanup
    cleanup_docker_services "$ENVIRONMENT"

    exit $exit_code
}

# Run main function
main "$@"