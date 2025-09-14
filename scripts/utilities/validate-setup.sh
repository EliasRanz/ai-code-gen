#!/bin/bash

# AI Code Generator - Setup Validation Script
# This script validates that the development environment is properly configured

set -e

# Color codes for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

PROJECT_NAME="AI Code Generator"

# Function to print colored output
print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

# Function to check if command exists
check_command() {
    local cmd=$1
    local description=$2

    if command -v "$cmd" >/dev/null 2>&1; then
        print_success "$description found: $cmd"
        return 0
    else
        print_error "$description not found: $cmd"
        return 1
    fi
}

# Function to check Go version
check_go_version() {
    if command -v go >/dev/null 2>&1; then
        local version=$(go version | awk '{print $3}' | sed 's/go//')
        local required="1.21"

        if [[ "$(printf '%s\n' "$required" "$version" | sort -V | head -n1)" == "$required" ]]; then
            print_success "Go version $version (meets requirement ≥$required)"
            return 0
        else
            print_warning "Go version $version (recommended ≥$required)"
            return 0
        fi
    else
        print_error "Go not found"
        return 1
    fi
}

# Function to check Node.js version
check_nodejs_version() {
    if command -v node >/dev/null 2>&1; then
        local version=$(node --version | sed 's/v//')
        local required="18"

        if [[ "$(printf '%s\n' "$required" "$version" | sort -V | head -n1)" == "$required" ]]; then
            print_success "Node.js version $version (meets requirement ≥$required)"
            return 0
        else
            print_warning "Node.js version $version (recommended ≥$required)"
            return 0
        fi
    else
        print_error "Node.js not found"
        return 1
    fi
}

# Function to check Docker
check_docker() {
    if command -v docker >/dev/null 2>&1; then
        if docker info >/dev/null 2>&1; then
            print_success "Docker is running"
            return 0
        else
            print_warning "Docker is installed but not running"
            return 1
        fi
    else
        print_error "Docker not found"
        return 1
    fi
}

# Function to check Docker Compose
check_docker_compose() {
    if command -v docker-compose >/dev/null 2>&1; then
        print_success "Docker Compose found"
        return 0
    elif docker compose version >/dev/null 2>&1; then
        print_success "Docker Compose V2 (docker compose) found"
        return 0
    else
        print_error "Docker Compose not found"
        return 1
    fi
}

# Function to check Kubernetes tools
check_k8s_tools() {
    local issues=0

    if check_command "kubectl" "kubectl"; then
        kubectl version --client --short >/dev/null 2>&1 && print_success "kubectl is properly configured" || print_warning "kubectl may not be properly configured"
    else
        issues=$((issues + 1))
    fi

    if check_command "helm" "Helm"; then
        helm version --short >/dev/null 2>&1 && print_success "Helm is properly configured" || print_warning "Helm may not be properly configured"
    else
        issues=$((issues + 1))
    fi

    check_command "kind" "kind" || issues=$((issues + 1))

    return $issues
}

# Function to check project files
check_project_files() {
    local issues=0

    print_info "Checking project files..."

    [ -f "go.mod" ] && print_success "go.mod found" || { print_error "go.mod not found"; issues=$((issues + 1)); }
    [ -f "Makefile" ] && print_success "Makefile found" || { print_error "Makefile not found"; issues=$((issues + 1)); }
    [ -f ".env" ] && print_success ".env file found" || { print_warning ".env file not found (copy from .env.example)"; }
    [ -d "cmd" ] && print_success "cmd directory found" || { print_error "cmd directory not found"; issues=$((issues + 1)); }
    [ -d "internal" ] && print_success "internal directory found" || { print_error "internal directory not found"; issues=$((issues + 1)); }
    [ -d "web" ] && print_success "web directory found" || { print_error "web directory not found"; issues=$((issues + 1)); }

    return $issues
}

# Function to check Go dependencies
check_go_deps() {
    print_info "Checking Go dependencies..."

    if command -v go >/dev/null 2>&1; then
        if go mod verify >/dev/null 2>&1; then
            print_success "Go dependencies are valid"
            return 0
        else
            print_warning "Go dependencies may need to be downloaded"
            return 1
        fi
    else
        print_error "Go not available to check dependencies"
        return 1
    fi
}

# Function to check Node.js dependencies
check_nodejs_deps() {
    print_info "Checking Node.js dependencies..."

    if [ -d "web" ] && command -v npm >/dev/null 2>&1; then
        cd web
        if [ -d "node_modules" ]; then
            print_success "Node.js dependencies installed"
            cd ..
            return 0
        else
            print_warning "Node.js dependencies not installed (run: cd web && npm install)"
            cd ..
            return 1
        fi
    else
        print_error "Cannot check Node.js dependencies"
        return 1
    fi
}

# Function to check database connectivity
check_databases() {
    print_info "Checking database connectivity..."

    # Check if databases are running via Docker
    if command -v docker >/dev/null 2>&1; then
        if docker ps | grep -q postgres; then
            print_success "PostgreSQL container is running"
        else
            print_warning "PostgreSQL container not found (run: make dev)"
        fi

        if docker ps | grep -q redis; then
            print_success "Redis container is running"
        else
            print_warning "Redis container not found (run: make dev)"
        fi
    else
        print_warning "Docker not available to check database containers"
    fi
}

# Function to check build
check_build() {
    print_info "Checking if services can be built..."

    if command -v go >/dev/null 2>&1; then
        if go build -o /tmp/test-build ./cmd/api-gateway 2>/dev/null; then
            print_success "API Gateway can be built"
            rm -f /tmp/test-build
            return 0
        else
            print_warning "API Gateway build failed (may need dependencies)"
            return 1
        fi
    else
        print_error "Go not available to check build"
        return 1
    fi
}

# Function to show next steps
show_next_steps() {
    echo ""
    print_info "Setup validation complete!"
    echo ""
    print_info "If you have issues, try these commands:"
    echo "  • Install missing tools: ./setup-dev-environment.sh"
    echo "  • Start databases: make dev"
    echo "  • Build services: make build"
    echo "  • Start services: make up"
    echo "  • View logs: make logs"
    echo ""
    print_info "Access your application at:"
    echo "  • Frontend: http://localhost:3000"
    echo "  • API Gateway: http://localhost:8080"
    echo "  • Adminer (DB): http://localhost:8090"
}

# Main validation function
main() {
    echo ""
    echo -e "${BLUE}🔍 ${PROJECT_NAME} - Development Environment Validation${NC}"
    echo "======================================================"
    echo ""

    local total_issues=0
    local section_issues=0

    # 1. Check required tools
    echo -e "${BLUE}📋 Checking Required Tools${NC}"
    echo "---------------------------"

    check_go_version || total_issues=$((total_issues + 1))
    check_nodejs_version || total_issues=$((total_issues + 1))
    check_docker || total_issues=$((total_issues + 1))
    check_docker_compose || total_issues=$((total_issues + 1))

    echo ""

    # 2. Check Kubernetes tools (optional)
    echo -e "${BLUE}☸️  Checking Kubernetes Tools (Optional)${NC}"
    echo "----------------------------------------"

    check_k8s_tools || section_issues=$((section_issues + 1))
    if [ $section_issues -gt 0 ]; then
        print_warning "Some Kubernetes tools are missing (optional for basic development)"
    fi

    echo ""

    # 3. Check project structure
    echo -e "${BLUE}📁 Checking Project Structure${NC}"
    echo "-------------------------------"

    check_project_files || total_issues=$((total_issues + 1))

    echo ""

    # 4. Check dependencies
    echo -e "${BLUE}📦 Checking Dependencies${NC}"
    echo "---------------------------"

    check_go_deps || total_issues=$((total_issues + 1))
    check_nodejs_deps || total_issues=$((total_issues + 1))

    echo ""

    # 5. Check databases
    echo -e "${BLUE}🗄️  Checking Databases${NC}"
    echo "-----------------------"

    check_databases

    echo ""

    # 6. Check build
    echo -e "${BLUE}🔨 Checking Build${NC}"
    echo "-------------------"

    check_build || total_issues=$((total_issues + 1))

    echo ""

    # Summary
    echo -e "${BLUE}📊 Validation Summary${NC}"
    echo "======================"

    if [ $total_issues -eq 0 ]; then
        print_success "All critical components are properly configured!"
        print_success "Your development environment is ready."
    else
        print_warning "Found $total_issues issue(s) that need attention."
        print_info "Run ./setup-dev-environment.sh to fix missing components."
    fi

    show_next_steps
}

# Run main function
main "$@"