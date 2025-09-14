#!/bin/bash

# AI Code Generator - Local Development Environment Script
# This script provides a single command to start the entire local development environment

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Project name
PROJECT_NAME="AI Code Generator"
COMPOSE_FILE="docker-compose.yml"

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if Docker is running
check_docker() {
    if ! docker info >/dev/null 2>&1; then
        print_error "Docker is not running. Please start Docker and try again."
        exit 1
    fi
    print_success "Docker is running"
}

# Function to check if docker-compose file exists
check_compose_file() {
    if [ ! -f "$COMPOSE_FILE" ]; then
        print_error "docker-compose.yml not found in current directory"
        exit 1
    fi
    print_success "Found docker-compose.yml"
}

# Function to start services
start_services() {
    print_status "Starting all services..."

    # Start services in detached mode
    if docker-compose up -d; then
        print_success "Services started successfully"
    else
        print_error "Failed to start services"
        exit 1
    fi
}

# Function to wait for services to be healthy
wait_for_services() {
    print_status "Waiting for services to be ready..."

    # List of services to check
    services=("postgres" "redis" "vllm" "auth-service" "user-service" "ai-service" "api-gateway" "frontend")

    # Wait for each service
    for service in "${services[@]}"; do
        print_status "Waiting for $service..."

        # Try to check service health for up to 60 seconds
        timeout=60
        counter=0

        while [ $counter -lt $timeout ]; do
            if docker-compose ps "$service" | grep -q "Up"; then
                print_success "$service is ready"
                break
            fi

            sleep 2
            counter=$((counter + 2))

            if [ $counter -ge $timeout ]; then
                print_warning "$service is taking longer than expected to start"
                break
            fi
        done
    done
}

# Function to show service status
show_status() {
    echo
    print_success "=== $PROJECT_NAME Development Environment ==="
    echo
    echo "Services Status:"
    docker-compose ps
    echo
    print_success "Access URLs:"
    echo "  🌐 Frontend:        http://localhost:3000"
    echo "  🚪 API Gateway:     http://localhost:8080"
    echo "  🔐 Auth Service:    http://localhost:8081"
    echo "  👤 User Service:    http://localhost:8082"
    echo "  🤖 AI Service:      http://localhost:8083"
    echo "  ⚡ AI Generation:   http://localhost:8084"
    echo "  🗄️  Database Admin:  http://localhost:8090"
    echo "  🤖 vLLM AI Server:  http://localhost:8000"
    echo
    print_success "Database Connections:"
    echo "  PostgreSQL: localhost:5433 (user: postgres, db: ai_ui_generator)"
    echo "  Redis:      localhost:6380"
    echo
    print_status "Useful Commands:"
    echo "  📋 View logs:       docker-compose logs -f"
    echo "  🛑 Stop services:   docker-compose down"
    echo "  🔄 Restart service: docker-compose restart <service-name>"
    echo "  📊 Service status:  docker-compose ps"
}

# Function to check if services are already running
check_running_services() {
    if docker-compose ps | grep -q "Up"; then
        print_warning "Some services are already running"
        echo "Current status:"
        docker-compose ps
        echo
        read -p "Do you want to restart all services? (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            print_status "Restarting services..."
            docker-compose down >/dev/null 2>&1
        else
            print_status "Using existing services..."
            show_status
            exit 0
        fi
    fi
}

# Main function
main() {
    echo
    print_success "🚀 Starting $PROJECT_NAME Local Development Environment"
    echo "=================================================="
    echo

    # Pre-flight checks
    check_docker
    check_compose_file

    # Check if services are already running
    check_running_services

    # Start services
    start_services

    # Wait for services to be ready
    wait_for_services

    # Show final status
    show_status

    echo
    print_success "🎉 Development environment is ready!"
    print_status "Happy coding! 🚀"
    echo
}

# Handle script interruption
trap 'echo -e "\n${YELLOW}Script interrupted by user${NC}"; exit 1' INT TERM

# Run main function
main "$@"