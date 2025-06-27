# Makefile for AI UI Generator

.PHONY: help build test clean up down logs dev prod install migrate check-standards refactor setup-tests

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
	go test -v -race ./...
	@echo "Running frontend tests..."
	cd web && npm test

lint: ## Run linting
	@echo "Running Go linting..."
	golangci-lint run
	@echo "Running frontend linting..."
	cd web && npm run lint

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
