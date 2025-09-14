#!/bin/bash
# Unified development setup script with security improvements
# Combines functionality from setup-dev-db.sh and other dev setup tasks

set -euo pipefail

# Security: Use PROJECT_ROOT variable instead of absolute paths
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
COMPOSE_FILE="${COMPOSE_FILE:-$PROJECT_ROOT/../deployments/docker/docker-compose.yml}"
SERVICES_TIMEOUT="${SERVICES_TIMEOUT:-60}"

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

# Security: Validate Docker environment
check_docker() {
    log_info "Checking Docker environment..."

    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed"
        exit 1
    fi

    if ! docker info &> /dev/null; then
        log_error "Docker is not running or you don't have permissions"
        exit 1
    fi

    if ! command -v docker-compose &> /dev/null; then
        log_error "docker-compose is not installed"
        exit 1
    fi

    log_success "Docker environment is ready"
}

# Start development services
start_services() {
    log_info "Starting development services..."

    docker-compose -f "$COMPOSE_FILE" up -d postgres redis

    # Wait for PostgreSQL
    wait_for_postgres

    # Wait for Redis
    wait_for_redis

    log_success "Development services started"
}

# Stop development services
stop_services() {
    log_info "Stopping development services..."
    docker-compose -f "$COMPOSE_FILE" down
    log_success "Development services stopped"
}

# Wait for PostgreSQL to be ready
wait_for_postgres() {
    log_info "Waiting for PostgreSQL to be ready..."

    local counter=0
    while ! docker-compose -f "$COMPOSE_FILE" exec -T postgres pg_isready -U postgres -d ai_ui_generator &> /dev/null; do
        if [ $counter -ge "$SERVICES_TIMEOUT" ]; then
            log_error "PostgreSQL failed to start within $SERVICES_TIMEOUT seconds"
            docker-compose -f "$COMPOSE_FILE" logs postgres
            exit 1
        fi
        sleep 2
        counter=$((counter + 2))
        log_info "Waiting... (${counter}s/${SERVICES_TIMEOUT}s)"
    done

    log_success "PostgreSQL is ready"
}

# Wait for Redis to be ready
wait_for_redis() {
    log_info "Waiting for Redis to be ready..."

    local counter=0
    while ! docker-compose -f "$COMPOSE_FILE" exec -T redis redis-cli ping &> /dev/null; do
        if [ $counter -ge "$SERVICES_TIMEOUT" ]; then
            log_error "Redis failed to start within $SERVICES_TIMEOUT seconds"
            docker-compose -f "$COMPOSE_FILE" logs redis
            exit 1
        fi
        sleep 2
        counter=$((counter + 2))
        log_info "Waiting... (${counter}s/${SERVICES_TIMEOUT}s)"
    done

    log_success "Redis is ready"
}

# Run database migrations
run_migrations() {
    log_info "Running database migrations..."

    # Check if services are running
    if ! docker-compose -f "$COMPOSE_FILE" ps postgres | grep -q "Up"; then
        log_error "PostgreSQL is not running. Start services first with: $0 start"
        exit 1
    fi

    # For now, just check connection - actual migrations would go here
    if docker-compose -f "$COMPOSE_FILE" exec -T postgres psql -U postgres -d ai_ui_generator -c "SELECT 1;" &> /dev/null; then
        log_success "Database connection successful"
        log_info "Migration system ready (implement actual migrations as needed)"
    else
        log_error "Database connection failed"
        exit 1
    fi
}

# Show service status
show_status() {
    log_info "Service Status:"
    echo ""
    docker-compose -f "$COMPOSE_FILE" ps
    echo ""

    # Test connections
    log_info "Testing connections..."

    if docker-compose -f "$COMPOSE_FILE" exec -T postgres pg_isready -U postgres -d ai_ui_generator &> /dev/null; then
        echo -e "PostgreSQL: ${GREEN}Connected${NC}"
    else
        echo -e "PostgreSQL: ${RED}Not Connected${NC}"
    fi

    if docker-compose -f "$COMPOSE_FILE" exec -T redis redis-cli ping &> /dev/null; then
        echo -e "Redis: ${GREEN}Connected${NC}"
    else
        echo -e "Redis: ${RED}Not Connected${NC}"
    fi
}

# Clean up development environment
cleanup() {
    log_info "Cleaning up development environment..."

    # Stop services
    docker-compose -f "$COMPOSE_FILE" down -v 2>/dev/null || true

    # Remove generated files
    rm -f coverage*.out coverage*.html tests-report.xml

    log_success "Cleanup completed"
}

# Main execution
main() {
    local command="${1:-status}"

    case "$command" in
        "start")
            check_docker
            start_services
            ;;
        "stop")
            stop_services
            ;;
        "restart")
            stop_services
            sleep 2
            check_docker
            start_services
            ;;
        "migrate")
            run_migrations
            ;;
        "status")
            show_status
            ;;
        "cleanup")
            cleanup
            ;;
        "setup")
            check_docker
            start_services
            run_migrations
            show_status
            ;;
        *)
            echo "Usage: $0 [start|stop|restart|migrate|status|cleanup|setup]"
            echo ""
            echo "Commands:"
            echo "  start   - Start development services"
            echo "  stop    - Stop development services"
            echo "  restart - Restart development services"
            echo "  migrate - Run database migrations"
            echo "  status  - Show service status"
            echo "  cleanup - Clean up development environment"
            echo "  setup   - Complete development setup (start + migrate + status)"
            exit 1
            ;;
    esac
}

main "$@"