# AI Code Generation Project Makefile
# Simplified for developer efficiency and agent automation

.PHONY: help build test dev generate clean lint security quickstart deps deps-backend deps-frontend ci ci-backend ci-frontend

# Default target
help: ## Show this help message
	@echo 'AI Code Generation Project'
	@echo '========================='
	@echo ''
	@echo 'Quick Start:'
	@echo '  make setup     - Complete development setup'
	@echo '  make dev       - Start development environment (infrastructure only)'
	@echo '  make dev-all   - Start all services (including application services)'
	@echo '  make dev-ci    - Start all services for CI/CD (with health checks)'
	@echo '  make ci         - Run complete CI pipeline (backend + frontend)'
	@echo '  make ci-backend - Run backend CI pipeline only'
	@echo '  make ci-frontend- Run frontend CI pipeline only'
	@echo ''
	@echo 'Script Organization:'
	@echo '  database/      - Database setup and migrations'
	@echo '  development/   - Development workflow scripts'
	@echo '  setup/         - Environment setup scripts'
	@echo '  testing/       - Testing and performance scripts'
	@echo '  utilities/     - General utility scripts'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# ============================================================================
# DEVELOPMENT SETUP
# ============================================================================

setup: ## Complete development environment setup
	@echo "🚀 Setting up complete development environment..."
	./scripts/setup/setup-dev-environment.sh
	./scripts/testing/test.sh install
	./scripts/development/generate.sh all
	@echo "✅ Development environment ready!"

dev: ## Start development environment
	@echo "🔄 Starting development environment..."
	./scripts/development/dev.sh start

dev-all: ## Start all services (including application services)
	@echo "🚀 Starting all services..."
	docker-compose -f deployments/docker/docker-compose.yml up -d
	@echo "⏳ Waiting for services to be healthy..."
	@sleep 10
	@echo "📊 Service status:"
	@docker-compose -f deployments/docker/docker-compose.yml ps

dev-stop: ## Stop development environment
	@echo "🛑 Stopping development environment..."
	./scripts/development/dev.sh stop

dev-stop-all: ## Stop all services
	@echo "🛑 Stopping all services..."
	docker-compose -f deployments/docker/docker-compose.yml down

dev-status: ## Show development environment status
	@echo "📊 Development environment status:"
	./scripts/development/dev.sh status

dev-ci: ## Start all services for CI/CD (waits for health checks)
	@echo "🔄 Starting all services for CI/CD..."
	docker-compose -f deployments/docker/docker-compose.yml up -d
	@echo "⏳ Waiting for all services to be healthy..."
	@timeout=120; \
	counter=0; \
	while [ $$counter -lt $$timeout ]; do \
		healthy=$$(docker-compose -f deployments/docker/docker-compose.yml ps | grep -c "healthy\|Up"); \
		total=$$(docker-compose -f deployments/docker/docker-compose.yml ps | grep -c "Up\|healthy\|unhealthy"); \
		if [ "$$healthy" -eq "$$total" ] && [ "$$total" -gt 0 ]; then \
			echo "✅ All services are healthy!"; \
			break; \
		fi; \
		echo "⏳ Waiting... ($$counter/$$timeout seconds) - $$healthy/$$total services healthy"; \
		sleep 5; \
		counter=$$((counter + 5)); \
	done; \
	if [ $$counter -ge $$timeout ]; then \
		echo "❌ Timeout waiting for services to be healthy"; \
		docker-compose -f deployments/docker/docker-compose.yml ps; \
		exit 1; \
	fi

# ============================================================================
# BUILDING
# ============================================================================

build: ## Build all services (backend + frontend)
	@echo "🔨 Building all services..."
	go build -o bin/api-gateway ./cmd/api-gateway
	go build -o bin/auth-service ./cmd/auth-service
	go build -o bin/user-service ./cmd/user-service
	go build -o bin/ai-service ./cmd/ai-service
	go build -o bin/ai-generation-service ./cmd/ai-generation-service
	@echo "🔨 Building frontend..."
	cd frontend && npm run build
	@echo "✅ All builds completed"

build-backend: ## Build backend services only
	@echo "🔨 Building backend services..."
	go build -o bin/api-gateway ./cmd/api-gateway
	go build -o bin/auth-service ./cmd/auth-service
	go build -o bin/user-service ./cmd/user-service
	go build -o bin/ai-service ./cmd/ai-service
	go build -o bin/ai-generation-service ./cmd/ai-generation-service
	@echo "✅ Backend build completed"

build-frontend: ## Build frontend
	@echo "🔨 Building frontend..."
	cd frontend && npm run build
	@echo "✅ Frontend build completed"

# ============================================================================
# TESTING
# ============================================================================

test: ## Run unit tests only (backend + frontend)
	@echo "🧪 Running backend unit tests..."
	@if ./scripts/testing/test.sh unit; then \
		echo "✅ Backend tests passed"; \
	else \
		echo "❌ Backend tests failed"; \
		exit 1; \
	fi
	@echo "🧪 Running frontend tests..."
	@cd frontend && if npm run test:run; then \
		echo "✅ Frontend tests passed"; \
	else \
		echo "❌ Frontend tests failed"; \
		exit 1; \
	fi
	@echo "✅ All tests completed successfully"

test-all: ## Run unit tests only (backend + frontend)
	@echo "🧪 Running unit tests (backend + frontend)..."
	./scripts/testing/test.sh unit
	@echo "🧪 Running frontend tests..."
	cd frontend && npm run test:run

test-all-coverage: ## Run unit tests with coverage (backend + frontend)
	@echo "📊 Running unit tests with coverage..."
	@if ./scripts/testing/test.sh unit; then \
		echo "✅ Backend coverage generated"; \
	else \
		echo "❌ Backend coverage failed"; \
		exit 1; \
	fi
	@echo "📊 Running frontend tests with coverage..."
	@cd frontend && if npm run test:coverage; then \
		echo "✅ Frontend coverage generated"; \
	else \
		echo "❌ Frontend coverage failed"; \
		exit 1; \
	fi
	@echo "✅ All coverage reports completed successfully"

test-unit: ## Run unit tests only
	@echo "🧪 Running unit tests..."
	./scripts/testing/test.sh unit

test-integration: ## Run integration tests only
	@echo "🔧 Running integration tests..."
	./scripts/testing/test.sh integration

test-coverage: ## Generate coverage reports (backend + frontend)
	@echo "📊 Generating backend coverage reports..."
	@if ./scripts/testing/test.sh unit; then \
		echo "✅ Backend coverage generated"; \
	else \
		echo "❌ Backend coverage failed"; \
		exit 1; \
	fi
	@echo "📊 Generating frontend coverage reports..."
	@cd frontend && if npm run test:coverage; then \
		echo "✅ Frontend coverage generated"; \
	else \
		echo "❌ Frontend coverage failed"; \
		exit 1; \
	fi
	@echo "✅ All coverage reports completed successfully"

test-analyze: ## Analyze test coverage
	@echo "🔍 Analyzing test coverage..."
	./scripts/testing/test.sh analyze

test-frontend: ## Run frontend unit tests
	@echo "🧪 Running frontend tests..."
	cd frontend && npm run test:run

test-frontend-coverage: ## Run frontend tests with coverage
	@echo "📊 Running frontend tests with coverage..."
	cd frontend && npm run test:coverage

# ============================================================================
# CODE GENERATION
# ============================================================================

generate: ## Generate mocks and protobuf files
	@echo "🔧 Generating code..."
	./scripts/development/generate.sh all

generate-mocks: ## Generate mock files only
	@echo "🔧 Generating mocks..."
	./scripts/development/generate.sh mocks

generate-protos: ## Generate protobuf files only
	@echo "🔧 Generating protobufs..."
	./scripts/development/generate.sh protos

# ============================================================================
# QUALITY ASSURANCE
# ============================================================================

lint: ## Run linting
	@echo "🔍 Running linting..."
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run --timeout=5m

format: ## Format code
	@echo "💅 Formatting code..."
	go fmt ./...
	cd frontend && npm run format

security: ## Run security checks
	@echo "🔒 Running security checks..."
	@which gosec > /dev/null || (echo "Installing gosec..." && go install github.com/securego/gosec/v2/cmd/gosec@latest)
	gosec ./...

# ============================================================================
# DATABASE & MIGRATIONS
# ============================================================================

migrate: ## Run database migrations
	@echo "🗃️ Running database migrations..."
	@if command -v pg_isready >/dev/null 2>&1 && pg_isready -h localhost -p 5433 >/dev/null 2>&1; then \
		./scripts/database/init_db.sh; \
	else \
		echo "❌ PostgreSQL is not running on localhost:5433"; \
		echo "💡 To run database migrations, start PostgreSQL:"; \
		echo "   • Using Docker: make dev"; \
		echo "   • Using local PostgreSQL: ensure it's running on port 5433"; \
		echo "   • Using different port: set DB_PORT environment variable"; \
		exit 1; \
	fi

# ============================================================================
# PERFORMANCE TESTING
# ============================================================================

perf-test: ## Run performance tests
	@echo "⚡ Running performance tests..."
	./scripts/testing/performance-test.sh all

perf-benchmark: ## Run benchmark tests only
	@echo "⚡ Running benchmark tests..."
	./scripts/testing/performance-test.sh benchmark

perf-load: ## Run load tests only
	@echo "⚡ Running load tests..."
	./scripts/testing/performance-test.sh load

perf-stress: ## Run stress tests only
	@echo "⚡ Running stress tests..."
	./scripts/testing/performance-test.sh stress

# ============================================================================
# UTILITIES
# ============================================================================

scripts: ## Show scripts directory structure
	@echo "📁 Scripts Directory Structure"
	@echo "=============================="
	@find scripts -name "*.sh" | sort | sed 's|scripts/||' | sed 's|/| ➤ |' | sed 's|\.sh||'
	@echo ""
	@echo "Usage Examples:"
	@echo "  make test                    # Run tests"
	@echo "  make dev                     # Start development"
	@echo "  make generate                # Generate code"
	@echo "  ./scripts/run testing test   # Direct script execution"
	@echo "  ./scripts/run setup setup-dev-db  # Run setup script"

clean: ## Clean build artifacts and temporary files
	@echo "🧹 Cleaning up..."
	rm -rf bin/
	rm -rf frontend/dist/
	rm -rf frontend/node_modules/.cache/
	rm -f coverage*.out coverage*.html tests-report.xml
	go clean
	@echo "✅ Cleanup completed"

deps: ## Update dependencies
	@echo "📦 Updating dependencies..."
	go mod tidy
	go mod download
	cd frontend && npm update

deps-backend: ## Install backend dependencies
	@echo "📦 Installing backend dependencies..."
	go mod download

deps-frontend: ## Install frontend dependencies
	@echo "📦 Installing frontend dependencies..."
	cd frontend && npm ci

# ============================================================================
# CI/CD TARGETS
# ============================================================================

ci-backend: ## Run complete backend CI pipeline (deps + build + test)
	@echo "🔄 Running backend CI pipeline..."
	make deps-backend
	make build-backend
	make test-unit
	@echo "✅ Backend CI pipeline completed"

ci-frontend: ## Run complete frontend CI pipeline (deps + build + test)
	@echo "🔄 Running frontend CI pipeline..."
	make deps-frontend
	make build-frontend
	make test-frontend-coverage
	@echo "✅ Frontend CI pipeline completed"

ci: ## Run complete CI pipeline (backend + frontend)
	@echo "🔄 Running complete CI pipeline..."
	make ci-backend
	make ci-frontend
	@echo "✅ Full CI pipeline completed"

install-tools: ## Install development tools
	@echo "🔧 Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.55.2
	go install github.com/securego/gosec/v2/cmd/gosec@v2.18.2
	go install go.uber.org/mock/mockgen@v0.4.0
	go install gotest.tools/gotestsum@v1.10.0
	@echo "✅ Development tools installed"

# ============================================================================
# QUICK STARTS
# ============================================================================

quickstart: ## Quick start for new developers
	@echo "🚀 Quick start for new developers..."
	make install-tools
	make setup
	make build
	make test
	@echo ""
	@echo "🎉 Quick start completed!"
	@echo ""
	@echo "Next steps:"
	@echo "  1. make dev          - Start development environment"
	@echo "  2. make test         - Run tests"
	@echo "  3. make build        - Build services"

# ============================================================================
# TROUBLESHOOTING
# ============================================================================

troubleshoot-docker: ## Troubleshoot Docker issues
	@echo "🐳 Troubleshooting Docker..."
	@echo "Docker version: $$(docker --version)"
	@echo "Docker Compose version: $$(docker-compose --version)"
	@echo "Docker running: $$(docker info >/dev/null 2>&1 && echo 'Yes' || echo 'No')"
	@echo ""
	@echo "Common solutions:"
	@echo "  - Start Docker Desktop"
	@echo "  - Run: make dev-cleanup && make dev"
	@echo "  - Check logs: docker-compose logs"

troubleshoot-services: ## Troubleshoot service issues
	@echo "🔧 Troubleshooting services..."
	@echo "Service status:"
	@make dev-status
	@echo ""
	@echo "Health checks:"
	@./scripts/utilities/health-check.sh api-gateway 8080 /health >/dev/null 2>&1 && echo "✅ API Gateway: UP" || echo "❌ API Gateway: DOWN"
	@./scripts/utilities/health-check.sh auth-service 8081 /health >/dev/null 2>&1 && echo "✅ Auth Service: UP" || echo "❌ Auth Service: DOWN"
	@./scripts/utilities/health-check.sh user-service 8082 /health >/dev/null 2>&1 && echo "✅ User Service: UP" || echo "❌ User Service: DOWN"
	@./scripts/utilities/health-check.sh ai-service 8083 /health >/dev/null 2>&1 && echo "✅ AI Service: UP" || echo "❌ AI Service: DOWN"

troubleshoot-tests: ## Troubleshoot test issues
	@echo "🧪 Troubleshooting tests..."
	@echo "Go version: $$(go version)"
	@echo "Test directories:"
	@find tests -type d | head -10
	@echo ""
	@echo "Common solutions:"
	@echo "  - Run: make generate-mocks"
	@echo "  - Check: go mod tidy"
	@echo "  - Clean: make clean && make test"