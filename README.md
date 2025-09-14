# AI UI Generator

A production-ready, full-stack AI UI Generation System inspired by Vercel's v0.dev. Transform natural language prompts into high-quality, interactive frontend components using a modular, scalable microservices architecture.

## Comprehensive Implementation Plan: Autonomous Multi-Agent Development Platform

---

## Additional Architectural Considerations

### Error Handling, Security, and Data Privacy
- All agent interactions must include explicit error handling and input validation to ensure system reliability and fail-fast behavior.
- Security best practices should be applied throughout, including secure communication between agents, output encoding, and never hardcoding secrets or credentials.
- User data must be protected with strong encryption, access controls, and data minimization. Privacy requirements should be documented and enforced at every stage.

### Economic Governance
- The platform should track operational costs and optimize resource usage, especially for LLM model access and hosting.
- Implement metrics and dashboards to monitor cost drivers and support decision-making for model selection and infrastructure scaling.

### Testing and Validation
- Each phase should include automated unit, integration, and end-to-end tests to ensure reliability and regression safety.
- Testing strategies must be documented, and test coverage should be tracked and improved continuously.

### Feedback Loop Management
- The Orchestrator should manage feedback loops using robust state management and rollback strategies, ensuring that user and reviewer feedback is routed and acted upon efficiently.
- Document technical mechanisms for feedback routing and state transitions.

### Compliance and Responsible AI Usage
- The platform must comply with industry standards (e.g., GDPR, ISO/IEC 42001, NIST AI RMF) and responsible AI principles (fairness, transparency, accountability).
- Reference Microsoft, OpenAI, and Google AI guidelines for responsible usage, and document limitations and assumptions for all AI-driven features.

---

## Architecture Overview

- **Backend**: Go microservices with Gin framework
- **Frontend**: Next.js 14+ with TypeScript, Tailwind CSS, shadcn/ui
- **Database**: PostgreSQL
- **Cache**: Redis
- **Authentication**: OAuth 2.0 with JWT tokens
- **AI**: vLLM serving OpenAI-compatible API
- **Communication**: gRPC between services, SSE for real-time streaming
- **Deployment**: Docker & Kubernetes ready

## Project Structure

```
/ai-ui-generator/
├── cmd/                    # Service entry points
│   ├── api-gateway/        # API Gateway service
│   ├── auth-service/       # Authentication service
│   ├── user-service/       # User management service
│   └── ai-service/         # AI generation service
├── internal/               # Internal business logic
│   ├── auth/              # Authentication logic
│   ├── user/              # User management logic
│   ├── ai/                # AI generation logic
│   ├── database/          # Database connections
│   └── observability/     # Logging, metrics, tracing
├── web/                   # Next.js frontend
│   ├── app/               # App router pages
│   ├── components/        # React components
│   └── lib/               # Client utilities
├── api/proto/             # gRPC protocol definitions
├── configs/               # Configuration files
└── deployments/           # Docker & K8s manifests
```

## 🚀 Quick Start - Make Commands (Recommended)

**Use Make commands for everything** - they're the single source of truth for all development tasks:

```bash
# 🚀 Get started quickly
make help          # See all available commands
make setup         # Complete development setup
make dev           # Start development environment
make test          # Run all tests
make build         # Build all services

# 📊 Monitor and troubleshoot
make logs          # View service logs
make dev-status    # Check environment status
make troubleshoot-docker    # Debug Docker issues
make troubleshoot-services  # Debug service issues
```

### Why Use Make Commands?

✅ **Single Source of Truth** - All commands are documented and versioned  
✅ **Cross-Platform** - Works on Linux, macOS, and Windows (with WSL)  
✅ **Dependency Management** - Handles prerequisites automatically  
✅ **Error Handling** - Proper error reporting and recovery  
✅ **Documentation** - Every command is self-documenting  
✅ **CI/CD Ready** - Same commands work in automated pipelines  

### Quick Start Workflow

```bash
# 1. First time setup
make setup

# 2. Start development environment
make dev

# 3. View application
# Frontend: http://localhost:3000
# API Gateway: http://localhost:8080

# 4. Run tests
make test

# 5. Build for production
make build
```

## 📋 Complete Make Commands Reference

### 🚀 Primary Development Commands
```bash
make help                 # Show this help message
make setup                # Complete development environment setup
make dev                  # Start development environment
make dev-stop             # Stop development environment
make dev-status           # Show development environment status
make dev-cleanup          # Clean up development environment
make build                # Build all services
make build-frontend       # Build frontend only
make test                 # Run all tests with coverage
make logs                 # View service logs
```

### 🧪 Testing Commands
```bash
make test                 # Run all tests with coverage
make test-unit            # Run unit tests only
make test-integration     # Run integration tests only
make test-coverage        # Generate coverage reports
make test-analyze         # Analyze test coverage
```

### 🔧 Code Quality & Generation
```bash
make generate             # Generate mocks and protobuf files
make generate-mocks       # Generate mock files only
make generate-protos      # Generate protobuf files only
make lint                 # Run linting
make format               # Format code
make security             # Run security checks
```

### 🗄️ Database & Infrastructure
```bash
make migrate              # Run database migrations
```

### ⚡ Performance Testing
```bash
make perf-test            # Run all performance tests
make perf-benchmark       # Run benchmark tests only
make perf-load            # Run load tests only
make perf-stress          # Run stress tests only
```

### 🛠️ Utility Commands
```bash
make clean                # Clean build artifacts
make deps                 # Update dependencies
make install-tools        # Install development tools
make scripts              # Show scripts directory structure
make quickstart           # Quick start for new developers
```

### 🔍 Troubleshooting Commands
```bash
make troubleshoot-docker      # Troubleshoot Docker issues
make troubleshoot-services    # Troubleshoot service issues
make troubleshoot-tests       # Troubleshoot test issues
```

## 🏗️ Manual Setup (Alternative)

If you prefer manual setup or the automated `make setup` doesn't work:

#### Prerequisites

- **Go**: 1.21+ ([download](https://go.dev/dl/))
- **Node.js**: 18+ ([download](https://nodejs.org/))
- **Docker**: Latest ([download](https://www.docker.com/products/docker-desktop))
- **kubectl**: Latest ([install](https://kubernetes.io/docs/tasks/tools/))
- **Helm**: 3.0+ ([install](https://helm.sh/docs/intro/install/))
- **kind**: Latest ([install](https://kind.sigs.k8s.io/docs/user/quick-start/))

#### Development Setup

1. **Clone and setup environment**:
   ```bash
   git clone <repository-url>
   cd ai-code-generator
   cp .env.example .env
   # Edit .env with your configuration
   ```

2. **Install dependencies**:
   ```bash
   # Go dependencies
   go mod download

   # Frontend dependencies
   cd web && npm install && cd ..
   ```

3. **Start the complete development environment**:
   ```bash
   # Single command to start everything (recommended)
   make start

   # Or run the script directly
   ./start-dev.sh      # Linux/macOS/WSL
   .\scripts\setup\start-dev.bat     # Windows
   ```

   **What this does:**
   - ✅ Checks if Docker is running
   - ✅ Starts all databases and services
   - ✅ Waits for services to be healthy
   - ✅ Provides access URLs and status
   - ✅ Handles graceful startup (won't fail if services are already running)

4. **Access the application**:
   - **Frontend**: http://localhost:3000
   - **API Gateway**: http://localhost:8080
   - **Adminer (Database)**: http://localhost:8090
   - **PostgreSQL**: localhost:5433
   - **Redis**: localhost:6380

### 🤖 AI Service Configuration

The development environment uses a **Mock AI Service** by default for reliable, zero-issues startup. This provides realistic responses for testing the application logic without the complexity of running large AI models locally.

**Mock AI Service Features:**
- ✅ Always starts successfully (no model loading issues)
- ✅ Provides realistic code generation responses
- ✅ Compatible with OpenAI API format
- ✅ Fast response times for development
- ✅ No GPU or large memory requirements

**Switching to Real AI Service (Production):**

To use vLLM with Phi-2 model for actual AI capabilities:

1. **Edit `docker-compose.yml`**:
   ```yaml
   # Replace the vllm service section with:
   vllm:
     image: vllm/vllm-openai:latest
     container_name: ai-ui-generator-vllm
     ports:
       - "8000:8000"
     environment:
       - MODEL_NAME=microsoft/phi-2
       - MAX_MODEL_LEN=1024
       - GPU_MEMORY_UTILIZATION=0.6
       - DTYPE=half
       - TRUST_REMOTE_CODE=true
     volumes:
       - vllm_models:/root/.cache/huggingface
     healthcheck:
       test: ["CMD", "curl", "-f", "http://127.0.0.1:8000/health"]
       interval: 30s
       timeout: 10s
       retries: 3
       start_period: 180s
     restart: unless-stopped
     networks:
       - ai-ui-generator
   ```

2. **Add the volume back**:
   ```yaml
   volumes:
     redis_data:
       driver: local
     vllm_models:  # Add this back
       driver: local
     prometheus_data:
       driver: local
     grafana_data:
       driver: local
   ```

**Note**: The real vLLM service requires significant resources and may take several minutes to start while downloading and loading the Phi-2 model.

### 🐳 Docker Deployment

**Use Make commands for all Docker operations** (recommended):

```bash
# 🚀 Start complete development environment
make dev           # Start databases and services
make logs          # View all service logs
make dev-status    # Check environment status
make down          # Stop all services

# 🔧 Fine-grained control
make dev           # Start only databases
make up            # Start all services (after databases)
```

**Alternative manual commands** (if you prefer direct Docker control):
```bash
# Start just databases
make dev

# Start all services
make up
```

### ☸️ Kubernetes Deployment

```bash
# Quick local Kubernetes setup
make k8s-dev

# Or deploy manually
make k8s-setup
make helm-install ENV=local

# Access via ingress
# Frontend: http://ai-code-generator.local
# API Gateway: http://api.ai-code-generator.local
```

## 📋 Available Make Commands

```bash
# 🚀 Primary Development Commands
make start         # Start complete development environment (recommended)
make logs          # View service logs
make down          # Stop all services

# 🔧 Advanced Development Commands
make build         # Build all services
make dev           # Start databases only
make up            # Start all services (after make dev)

# Testing
make test          # Run all tests
make test-coverage # Generate coverage reports

# Kubernetes
make k8s-setup     # Set up local k8s cluster
make k8s-dev       # Quick k8s development setup
make helm-install  # Install Helm chart
make helm-status   # Check deployment status
make helm-logs     # View k8s logs

# Utilities
make clean         # Clean build artifacts
make format        # Format code
make lint          # Run linting
```

## Services

### API Gateway (Port 8080)
- Routes requests to microservices
- Handles authentication middleware
- Manages CORS and rate limiting
- Provides WebSocket/SSE endpoints

### Auth Service (Port 8081)
- OAuth 2.0 authentication (Google)
- JWT token management
- User session handling
- gRPC service for token validation

### User Service (Port 8082)
- User profile management
- User preferences and settings
- Project and workspace management
- gRPC service for user operations

### AI Service (Port 8083)
- LLM integration (vLLM/OpenAI-compatible)
- Code generation and streaming
- Prompt engineering and optimization
- Code validation and security checks

## Frontend Features

- **Chat Interface**: Natural language prompts
- **Live Preview**: Real-time code preview
- **Code Export**: Download generated components
- **Authentication**: Google OAuth integration
- **Responsive Design**: Mobile-friendly interface

## Development

### Adding New Features

1. **Backend**: Add handlers, services, and repositories in `internal/`
2. **Frontend**: Add components in `web/components/` and pages in `web/app/`
3. **API**: Define gRPC contracts in `api/proto/`

### Protobuf Generation

When modifying `.proto` files, always use the provided script to regenerate Go code:

```bash
# Recommended method
make generate-protos

# Or run the script directly
./scripts/generate-protos.sh
```

**⚠️ Important**: Never run `protoc` directly as it creates nested directories that need to be cleaned up. The script automatically handles file placement and cleanup.

See [docs/PROTOBUF_GENERATION.md](docs/PROTOBUF_GENERATION.md) for detailed information.

### Testing Strategy

This project uses a **three-tier testing approach** to ensure comprehensive test coverage while maintaining fast CI/CD pipelines:

#### 1. Unit Tests (`tests/unit/`)
- **Purpose**: Test individual functions and methods in isolation
- **Dependencies**: None (uses mocks for external dependencies)
- **Execution**: Fast, always run in CI/CD
- **Coverage**: Core business logic and utilities

```bash
# Run unit tests only
make test-unit

# With coverage report
make test-unit-coverage
```

#### 2. Component Integration Tests (`tests/integration/`)
- **Purpose**: Test component interactions with mocked services
- **Dependencies**: None (infrastructure mocked)
- **Execution**: Fast, run in CI/CD
- **Coverage**: Service integrations, HTTP handlers, gateway routing

```bash
# Run component integration tests (with mocks)
make test-component-integration

# With coverage report
make test-component-integration-coverage
```

#### 3. System Integration Tests (`tests/integration/`)
- **Purpose**: Test end-to-end functionality with real infrastructure
- **Dependencies**: PostgreSQL, Redis, running services
- **Execution**: Slow, requires infrastructure (CI/CD with services, or manual)
- **Coverage**: Database operations, cache interactions, service-to-service communication

```bash
# Run system integration tests (requires infrastructure)
make test-system-integration

# With coverage report
make test-system-integration-coverage
```

#### Running All Tests

```bash
# Run all tests with merged coverage (requires infrastructure for system tests)
make test-all

# Generate merged coverage report
make test-coverage-merged
```

#### Test Execution Matrix

| Test Type | Infrastructure | Speed | CI/CD | When to Run |
|-----------|---------------|-------|-------|-------------|
| Unit Tests | ❌ None | ⚡ Fast | ✅ Always | Every commit |
| Component Integration | ❌ Mocked | ⚡ Fast | ✅ Always | Every commit |
| System Integration | ✅ Required | 🐌 Slow | ✅ With services | PRs, main branch |

#### Environment Profiles

The project uses a **streamlined 4-environment approach** that minimizes maintenance overhead while ensuring quality:

**Available Profiles:**
- **`local`** - Local development with Docker Compose (default)
- **`ci`** - Consolidated testing environment (replaces dev/test/ci profiles)
- **`staging`** - Pre-production validation with production-like settings
- **`production`** - Production environment (use with caution)

**Why This Simplified Approach:**
- **Reduced maintenance**: 4 profiles vs 6 profiles (33% reduction)
- **Consolidated testing**: Single `ci` profile handles all automated testing
- **Clear separation**: Development ↔ Testing ↔ Production
- **Automated pipeline**: CI/CD uses profiles automatically

**Profile Management:**

```bash
# List all available profiles
make profiles-list

# Show details of a specific profile
make profiles-show PROFILE=ci

# Set active profile (creates .env symlink)
make profiles-set PROFILE=local

# Show current active profile
make profiles-current

# Validate profile configuration
make profiles-validate PROFILE=staging
```

**Profile-Based Testing:**

```bash
# Run system integration tests with specific profile
make test-system-profile PROFILE=ci

# Run all tests with a profile
make test-all-profile PROFILE=ci

# Quick profile switching for development
make profiles-set PROFILE=local
make test-system-integration  # Uses local profile automatically
```

**Profile Configuration:**

**Profile Configuration:**

Each profile is defined in `tests/integration/.env.{profile_name}`:

```bash
# Example: .env.local (local testing environment)
REDIS_HOST=localhost
REDIS_PORT=6379
DB_HOST=localhost
DB_PORT=5432
DB_NAME=ai_ui_generator_test
```

**CI/CD Integration:**

The CI/CD pipeline automatically uses the `ci` profile for all testing, eliminating manual environment configuration.

#### Development Workflow

```bash
# Quick development cycle (unit + component tests)
make test-unit && make test-component-integration

# Full test suite (requires infrastructure)
make dev  # Start infrastructure
make test-all

# CI/CD simulation
make test-unit
make test-component-integration
# System integration tests run in separate CI job with infrastructure
```

#### Test Organization

```
tests/
├── unit/                    # Unit tests (mocks only)
│   ├── auth/
│   ├── user/
│   ├── ai/
│   └── cache/
├── integration/             # Integration tests
│   ├── auth-service/        # System integration (real services)
│   ├── cache_integration_test.go
│   ├── database_integration_test.go
│   ├── handlers_integration_test.go  # Component integration (mocks)
│   └── gateway_integration_test.go
└── performance/             # Performance tests
```

#### Build Tags

- **Unit tests**: No build tags required
- **Component integration**: No build tags (exclude system tests with regex)
- **System integration**: `// +build integration` tag required

#### Coverage Reporting

- Individual coverage files: `unit_coverage.out`, `component_integration_coverage.out`, `system_integration_coverage.out`
- Merged coverage: `all_coverage.out`
- HTML reports: `go tool cover -html=all_coverage.out`

### Performance Testing

```bash
# Run performance tests (requires Docker and PERFORMANCE_TESTS=1)
PERFORMANCE_TESTS=1 go test ./tests/performance/...

# Or use make targets for performance testing
make test-performance    # Full performance test suite
make test-benchmark     # Go benchmark tests only
make test-load         # Load tests with Vegeta
make test-stress       # Stress tests
```

**Note**: Performance tests are disabled by default since they require Docker and are slower. Set `PERFORMANCE_TESTS=1` environment variable to enable them explicitly.

### Code Generation

```bash
# Generate gRPC code (when proto files change)
protoc --go_out=. --go-grpc_out=. api/proto/*.proto
```

## Configuration

Key configuration options in `.env`:

- **Database**: PostgreSQL connection settings
- **Redis**: Cache and session storage
- **OAuth**: Google OAuth credentials
- **AI**: LLM endpoint and model configuration
- **Security**: JWT secrets and encryption keys

## Security

- JWT-based authentication
- OAuth 2.0 integration
- Input validation and sanitization
- SQL injection prevention
- CORS configuration
- Rate limiting

## Monitoring & Observability

- **Logging**: Structured JSON logs with Zerolog
- **Metrics**: Prometheus-compatible metrics
- **Tracing**: Jaeger distributed tracing
- **Health Checks**: Service health endpoints

## Deployment

### Kubernetes

```bash
kubectl apply -f deployments/kubernetes/
```

### Docker Compose

```bash
docker-compose -f deployments/docker-compose.yml up
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make changes with tests
4. Submit a pull request

## License

[Your License Here]

## TODO

- [ ] Implement business logic stubs
- [ ] Add comprehensive tests
- [ ] Setup CI/CD pipelines
- [ ] Add database migrations
- [ ] Implement code validation
- [ ] Add rate limiting
- [ ] Setup monitoring dashboards
- [ ] Add API documentation
