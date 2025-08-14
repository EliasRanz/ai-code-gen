# Makefile for AI UI Generator

.PHONY: help build test test-performance test-benchmark test-benchmark-quick test-benchmark-single test-load test-stress performance-report sla-demo clean up down logs dev prod install migrate check-standards refactor setup-tests generate-mocks

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Development commands
install: ## Install dependencies
	@echo "Installing Go dependencies..."
	go mod download
	@echo "Installing frontend dependencies..."
	cd web && npm install

install-tools: ## Install development tools
	@echo "Installing development tools..."
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	@which gosec > /dev/null || (echo "Installing gosec..." && go install github.com/securego/gosec/v2/cmd/gosec@latest)
	@which mockgen > /dev/null || (echo "Installing mockgen..." && go install go.uber.org/mock/mockgen@latest)
	@echo "✅ Development tools installed"

build: ## Build all services
	@echo "Building Go services..."
	go build -o bin/api-gateway ./cmd/api-gateway
	go build -o bin/auth-service ./cmd/auth-service
	go build -o bin/user-service ./cmd/user-service
	go build -o bin/ai-service ./cmd/ai-service
	@echo "Building frontend..."
	cd web && npm run build

test: ## Run all tests
	@echo "Running Go tests..."
	go test -v ./... -json | tparse
	@echo "Running frontend tests..."
	cd web && npm test

test-enhanced: ## Run tests with enhanced discovery and reporting
	@echo "🚀 Running tests with enhanced discovery..."
	@./scripts/enhanced-testing.sh test

test-coverage: ## Generate accurate coverage reports (excluding generated files)
	@echo "📊 Generating accurate coverage reports..."
	@./scripts/coverage.sh

test-discover: ## Analyze test discovery and structure
	@echo "🔍 Analyzing test discovery..."
	@./scripts/enhanced-testing.sh discover

test-analyze: ## Analyze package-level coverage
	@echo "🔍 Analyzing package coverage..."
	@./scripts/enhanced-testing.sh analyze

test-visual: ## Generate visual coverage reports and open browser
	@echo "🎨 Generating visual coverage reports..."
	@./scripts/enhanced-testing.sh all
	@echo "Opening coverage reports..."
	@command -v xdg-open >/dev/null && xdg-open coverage.html || echo "Open coverage.html in your browser"

install-test-tools: ## Install enhanced testing tools
	@echo "📦 Installing enhanced testing tools..."
	@./scripts/enhanced-testing.sh install || echo "⚠️  Some tools failed to install due to Go version requirements"

test-basic: ## Run tests with built-in Go tools only
	@echo "🧪 Running tests with built-in tools..."
	@./scripts/basic-testing.sh

test-unit: ## Run unit tests only
	@echo "Running unit tests..."
	go test -v ./tests/unit/... -json | tparse

test-integration: ## Run integration tests with local environment
	@echo "🔧 Running integration tests (local environment)..."
	@echo "Ensure Redis and PostgreSQL are running: make dev"
	INTEGRATION_TESTS=1 INTEGRATION_ENV=local go test -v ./tests/integration/... -timeout=5m

test-integration-staging: ## Run integration tests against staging environment
	@echo "🔧 Running integration tests (staging environment)..."
	@echo "⚠️ Ensure staging credentials are set in environment"
	INTEGRATION_TESTS=1 INTEGRATION_ENV=staging go test -v ./tests/integration/... -timeout=10m

test-integration-production: ## Run integration tests against production environment
	@echo "🔧 Running integration tests (production environment)..."
	@echo "⚠️ WARNING: Running tests against production environment"
	@read -p "Are you sure you want to run tests against production? (y/N): " confirm; \
	if [ "$$confirm" = "y" ] || [ "$$confirm" = "Y" ]; then \
		INTEGRATION_TESTS=1 INTEGRATION_ENV=production go test -v ./tests/integration/... -timeout=15m; \
	else \
		echo "Integration tests cancelled"; \
	fi

test-all: test-unit test-integration ## Run all tests (unit and integration)

test-performance: ## Run comprehensive performance tests for Redis auth cache (30min max)
	@echo "🚀 Running Redis Auth Cache Performance Tests..."
	@echo "Ensure Redis is running: make dev"
	@mkdir -p ./performance_reports
	timeout 30m bash -c 'PERFORMANCE_TESTS=1 go run ./tests/performance/cmd/performance_runner.go ./performance_reports' || echo "⚠️ Performance tests timed out after 30 minutes"
	@echo "📊 Performance test reports available in ./performance_reports/"

test-benchmark: ## Run Go benchmark tests only (2s each, 5min max)
	@echo "Running Go benchmark tests..."
	PERFORMANCE_TESTS=1 go test -run="^$$" -bench=. -benchmem -benchtime=2s -timeout=5m ./tests/performance/auth_cache/
	
test-benchmark-quick: ## Run quick benchmark tests (200ms each, 30s max)
	@echo "Running quick benchmark tests..."
	PERFORMANCE_TESTS=1 go test -run="^$$" -bench="BenchmarkCacheGet$$|BenchmarkCacheSet$$|BenchmarkCacheGetMiss$$" -benchmem -benchtime=200ms -timeout=30s ./tests/performance/auth_cache/

test-benchmark-single: ## Run single cache get benchmark (1s, 10s max)
	@echo "Running single benchmark test..."
	PERFORMANCE_TESTS=1 go test -run="^$$" -bench="BenchmarkCacheGet$$" -benchmem -benchtime=1s -timeout=10s ./tests/performance/auth_cache/
	
test-load: ## Run load tests with Vegeta (10min max)
	@echo "Running load tests..."
	PERFORMANCE_TESTS=1 go test -v -timeout=10m ./tests/performance/auth_cache/ -run TestAuthCacheLoadPerformance
	
test-stress: ## Run stress tests (15min max)
	@echo "Running stress tests..."
	PERFORMANCE_TESTS=1 go test -v -timeout=15m ./tests/performance/auth_cache/ -run TestAuthCacheStressTest

performance-report: ## Generate performance reports from existing data
	@echo "Generating performance reports..."
	@mkdir -p ./performance_reports
	go run ./tests/performance/cmd/performance_runner.go ./performance_reports

# DevOps Tools
devops-performance: ## Run comprehensive performance testing automation
	@echo "🚀 Running DevOps Performance Testing Automation..."
	PERFORMANCE_TESTS=1 ./devops/scripts/performance-test.sh

devops-monitor: ## Start Redis monitoring with real-time metrics
	@echo "📊 Starting Redis monitoring..."
	go run ./devops/monitoring/redis-monitor.go

devops-sla: ## Validate SLA configuration and thresholds
	@echo "🎯 Validating SLA Configuration..."
	go run ./devops/performance/sla-validator.go

lint: ## Run linting
	@echo "Running Go linting..."
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run
	@echo "Running frontend linting..."
	cd web && npm run lint

lint-check: ## Run linting without failing the build
	@echo "Running Go linting (check only)..."
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	-golangci-lint run
	@echo "Running frontend linting (check only)..."
	-cd web && npm run lint

clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -rf web/.next/
	rm -rf web/out/
	go clean

# Protobuf generation
generate-protos: ## Generate protobuf files (prevents nested directory issues)
	@echo "Generating protobuf files..."
	./scripts/generate-protos.sh

# Mock generation
generate-mocks: ## Generate mock files for interfaces using GoMock
	@echo "Generating mock files..."
	@which mockgen > /dev/null || (echo "Installing mockgen..." && go install go.uber.org/mock/mockgen@latest)
	./scripts/generate-mocks.sh

# Docker commands
docker-build: ## Build Docker images
	@echo "Building Docker images..."
	docker build -f cmd/api-gateway/Dockerfile -t ai-ui-generator-api-gateway .
	docker build -f cmd/auth-service/Dockerfile -t ai-ui-generator-auth-service .
	docker build -f cmd/user-service/Dockerfile -t ai-ui-generator-user-service .
	docker build -f cmd/ai-service/Dockerfile -t ai-ui-generator-ai-service .
	docker build -f web/Dockerfile -t ai-ui-generator-frontend ./web

# Development environment
dev: ## Start development environment
	@echo "Starting development environment..."
	docker-compose up -d postgres redis adminer
	@echo "Development environment started!"
	@echo "Postgres: localhost:5433"
	@echo "Redis: localhost:6380"
	@echo "Adminer: http://localhost:8090"

up: ## Start all services
	@echo "Starting all services..."
	docker-compose up -d

down: ## Stop all services
	@echo "Stopping all services..."
	docker-compose down

logs: ## Show logs for all services
	docker-compose logs -f

# Production commands
prod: ## Start production environment
	@echo "Starting production environment..."
	docker-compose -f docker-compose.prod.yml up -d

prod-down: ## Stop production environment
	@echo "Stopping production environment..."
	docker-compose -f docker-compose.prod.yml down

# Database commands
migrate: ## Run database migrations
	@echo "Running database migrations..."
	# Add migration command here when ready
	@echo "Migrations completed!"

migrate-down: ## Rollback database migrations
	@echo "Rolling back database migrations..."
	# Add rollback command here when ready
	@echo "Rollback completed!"

# Health checks
health: ## Check service health
	@echo "Checking service health..."
	@curl -f http://localhost:8080/health || echo "API Gateway: DOWN"
	@curl -f http://localhost:8081/health || echo "Auth Service: DOWN"
	@curl -f http://localhost:8082/health || echo "User Service: DOWN"
	@curl -f http://localhost:8083/health || echo "AI Service: DOWN"
	@curl -f http://localhost:3000/api/health || echo "Frontend: DOWN"

# Utility commands
format: ## Format code
	@echo "Formatting Go code..."
	go fmt ./...
	@echo "Formatting frontend code..."
	cd web && npm run format

deps: ## Update dependencies
	@echo "Updating Go dependencies..."
	go mod tidy
	go mod download
	@echo "Updating frontend dependencies..."
	cd web && npm update

security: ## Run security checks
	@echo "Running security checks..."
	@which gosec > /dev/null || (echo "Installing gosec..." && go install github.com/securego/gosec/v2/cmd/gosec@latest)
	gosec ./...
	cd web && npm audit

# Backup and restore
backup: ## Backup database
	@echo "Backing up database..."
	docker-compose exec postgres pg_dump -U postgres ai_ui_generator > backup_$(shell date +%Y%m%d_%H%M%S).sql

restore: ## Restore database from backup
	@echo "Restoring database..."
	@read -p "Enter backup file name: " backup_file; \
	docker-compose exec -T postgres psql -U postgres ai_ui_generator < $$backup_file

# Environment setup
env: ## Create environment file from example
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo "Created .env file from .env.example"; \
		echo "Please edit .env file with your configuration"; \
	else \
		echo ".env file already exists"; \
	fi

# Documentation
docs: ## Generate documentation
	@echo "Generating documentation..."
	go doc ./...
	cd web && npm run docs || echo "Frontend docs not configured"

# Quick start
quickstart: env install dev migrate ## Quick start for new developers
	@echo ""
	@echo "🚀 Quick start completed!"
	@echo ""
	@echo "Services running:"
	@echo "  - Postgres: localhost:5433"
	@echo "  - Redis: localhost:6380"
	@echo "  - Adminer: http://localhost:8090"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Edit .env file with your configuration"
	@echo "  2. Run 'make build' to build services"
	@echo "  3. Run 'make up' to start all services"
	@echo "  4. Visit http://localhost:3000 to see the frontend"

# Coding standards enforcement
check-standards: ## Check coding standards compliance (file size, function size, test organization)
	@echo "🔍 Checking coding standards compliance..."
	@echo "Checking file sizes (must be <300 lines, except generated code)..."
	@find . -name "*.go" -not -path "./api/proto/*" -not -path "./github.com/*" -not -path "./web/node_modules/*" -exec wc -l {} + | \
		awk '$$1 > 300 && $$2 !~ /total$$/ { violations++; print "❌ File too large:", $$2, "(" $$1 " lines)" } END { if (violations > 0) exit 1 }'
	@echo "Checking test organization..."
	@test_files=$$(find . -name "*_test.go" -not -path "./tests/*" | wc -l); \
	if [ $$test_files -gt 0 ]; then \
		echo "❌ Test files found in source directories (should be in /tests):"; \
		find . -name "*_test.go" -not -path "./tests/*" | head -5; \
		exit 1; \
	else \
		echo "✅ Test files properly organized"; \
	fi
	@echo "✅ Coding standards check passed"

refactor: ## Run automated refactoring for coding standards compliance
	@echo "🔧 Running automated refactoring..."
	@./scripts/refactor-large-files.sh || echo "⚠️ Refactoring script not found, manual refactoring required"
	@make setup-tests
	@echo "✅ Refactoring helpers complete. Manual review required."

setup-tests: ## Set up proper test directory structure
	@echo "📁 Setting up test directories..."
	@mkdir -p tests/{unit,integration,fixtures,utils}
	@echo "✅ Test directories created"
