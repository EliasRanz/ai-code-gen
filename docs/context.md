# AI Code Generator – Project Context

## Overview
AI Code Generator is a production-ready, full-stack system inspired by Vercel's v0.dev. It transforms natural language prompts into complete applications including frontend components, backend services, and infrastructure code using a modular, scalable microservices architecture. The system is designed for extensibility, observability, and agent-driven development.

## End-User Workflow
The system follows a structured approach to application generation:
1. **Requirements Analysis** - Understands what the user wants to build and determines optimal languages/frameworks for the use case
2. **Frontend Wireframing** - Creates initial wireframes and UI structure
3. **Functionality Implementation** - Adds business logic, API integrations, and core features
4. **Production Enhancement** - Refines the GUI for production readiness with proper styling, error handling, and UX
5. **Deployment** - Generates infrastructure code and deployment configurations for the complete application

## Architecture Summary
- **Backend:** Go microservices (Gin framework)
- **Frontend:** Next.js 14+ (TypeScript, Tailwind CSS, shadcn/ui)
- **Database:** PostgreSQL
- **Cache/Message Broker:** Redis
- **Authentication:** OAuth 2.0 (Google) + JWT
- **AI Inference:** vLLM (OpenAI-compatible API)
- **Inter-Service Communication:** gRPC, Protobuf
- **Streaming:** Server-Sent Events (SSE)
- **Deployment:** Docker, Docker Compose, Kubernetes

## Microservice-Focused Architecture
Following ADR-013 and ADR-017, the system uses a fully consolidated service-based architecture where each service owns its complete implementation stack:

### Core Services (Complete Ownership Model)
- **API Gateway:** Unified entry point, routing, CORS, rate limiting, JWT validation
- **Auth Service:** OAuth 2.0 flow, JWT issuance/validation, session management, HTTP handlers, middleware
- **User Service:** User profiles, projects, chat sessions, CRUD APIs, database adapters, HTTP/gRPC handlers
- **AI Service:** LLM integration, code generation, streaming, database adapters, HTTP handlers, LLM clients

### Package Structure (Current Architecture)
- **Service Packages:** (`internal/user/`, `internal/ai/`, `internal/auth/`, `internal/gateway/`) - Complete service implementation including:
  - Domain entities and business logic
  - Database adapters and repositories  
  - HTTP handlers and gRPC servers
  - Service-specific clients and adapters
- **Shared Packages:** 
  - `internal/utilities/` - Shared types (UserID, ProjectID, errors, pagination)
  - `internal/config/` - Configuration management
  - `internal/observability/` - Logging, metrics, tracing, validation (was infrastructure/)
  - `internal/database/` - Shared database utilities
  - `internal/cache/` - Redis cache utilities

### ✅ Infrastructure Migration Complete (ADR-026)
**Completed**: Successfully eliminated `internal/infrastructure/` over-abstractions (January 2025)
- **ObservabilityProvider Pattern**: Implemented comprehensive observability with Prometheus/OpenTelemetry providers
- **Legacy Infrastructure Removed**: Deleted entire `internal/infrastructure/` directory with zero breaking changes
- **Test Coverage**: Achieved 37.4% coverage in new observability package vs 0% in legacy infrastructure
- **Backward Compatibility**: All services compile and run with new observability system
- **Migration Metrics**: 16+ files migrated from infrastructure/observability, 2 files from infrastructure/validation

## Authentication Architecture
The authentication system is centralized in the `internal/auth` package with complete business logic implementation:

### Core Auth Package (`internal/auth/`)
- **Authentication Use Cases:** Login, logout, refresh token, token validation
- **Authorization Use Cases:** Role checking, user context resolution, session management
- **JWT Management:** Token generation, validation, and expiration handling
- **Password Security:** BCrypt hashing with salt and validation
- **HTTP Middleware:** Complete set of auth middleware functions
- **Domain Types:** UserID/SessionID type aliases, session status, error handling

## Frontend
- **Framework:** Next.js (App Router), React, TypeScript
- **UI:** Tailwind CSS, shadcn/ui
- **Features:**
  - Conversational chat interface for application requirements
  - Multi-stage code generation workflow (wireframe → functionality → production → deployment)
  - Live, sandboxed preview pane (iframe) for frontend components
  - Code editor with syntax highlighting for backend and infrastructure code
  - Authentication (Google OAuth)
  - Project and settings management
  - Responsive, modern design

## Database Schema (Key Entities)
- Users, Projects, Chat Sessions, Chat Messages, UI Generations, User Settings, API Keys

## Testing & Quality
- **Unit, integration, and E2E tests** for all layers
- **Test Coverage:** 43.8% total coverage (Jan 2025), with 74.0% in observability package
- **Coverage Methodology Investigation**: Resolved measurement methodology issues - proper coverage aggregation via enhanced testing script shows actual progress toward 75% target
- **Comprehensive Testing Infrastructure**: Enhanced testing framework with third-party tools:
  - **GoTestSum**: Enhanced test output with progress indicators and color formatting
  - **Go-Cover-Treemap**: Visual coverage analysis with SVG treemap generation
  - **Go-Test-Coverage**: Coverage threshold enforcement (configurable via .testcoverage.yml)
  - **Multiple Coverage Formats**: HTML, JSON, and visual treemap reports
  - **Fallback Testing**: Reliable basic testing using only built-in Go tools when enhanced tools fail
- **Gateway Package Testing**: Improved from 0% to 14.5% coverage with 32 comprehensive test cases
- **Infrastructure Testing**: Eliminated 28+ untested functions from legacy infrastructure
- **Mock Integration Strategy**: GoMock-based generated mocks replacing manual mocks (ADR-024)
- **Interface Segregation Testing**: Comprehensive mock coverage for segregated interfaces (ADR-023)
- **CI/CD:** Enhanced GitHub Actions workflow with coverage artifacts and threshold validation
- **Test Organization:** Domain-based test structure following Go conventions (*_test.go pattern)
- **Generated Mock Framework**: Automated mock generation with `scripts/generate-mocks.sh`
- **Testing Scripts**: 
  - `scripts/enhanced-testing.sh`: Comprehensive testing with third-party tools and error handling
  - `scripts/basic-testing.sh`: Reliable fallback using only Go standard tools
- **Make Targets**: Multiple testing commands (test-enhanced, test-basic, test-visual, test-coverage)

## Security & Observability
- JWT secret management, password hashing (bcrypt)
- Input validation, error handling, rate limiting
- **Comprehensive Observability**: ObservabilityProvider pattern with Prometheus, OpenTelemetry, and multi-provider support
- **Structured logging** (zerolog), metrics, tracing with MonitoringDecorator pattern
- **Health Checks**: Provider-based health monitoring with lifecycle management

## Current Architecture Changes (2025-01-13)
- **✅ Infrastructure Migration Complete**: ADR-026 - Successfully eliminated `internal/infrastructure` directory, implemented ObservabilityProvider pattern with 37.4% test coverage
- **✅ Domain Cleanup Completed**: Successfully moved from `internal/common` to `internal/utilities`, eliminated `internal/application` layer
- **✅ Interface Segregation Completed**: ADR-023 implementation with segregated cache, LLM, config, and auth interfaces
- **✅ Mock Integration Strategy Completed**: ADR-024 implementation with GoMock-based generated mocks across all services
- **✅ HTTP Handlers Consolidation**: ADR-025 Phase C.1 - Moved handlers to service packages for complete ownership
- **Auth Centralization**: Centralized auth server implementation (see ADR-012)
- **Implementation Status**: See `CENTRALIZED_AUTH_PLAN.md` for auth progress, ADR-017 for infrastructure migration
- **Service Ownership**: Moving toward complete service ownership model where each service contains domain + adapters + handlers

## Agent Guidance
- **All code follows Clean Architecture and SOLID principles within service boundaries**
- **ADRs**: Architectural decisions are documented in `ADRs/` - See ADR-017 for current infrastructure migration
- **Service Ownership**: Each service should own its complete implementation (domain + adapters + handlers)
- **Migration Priority**: Work toward ADR-017 target architecture - eliminate `internal/infrastructure` and `internal/interfaces` over-abstractions
- **All business logic is isolated and testable within service packages**
- **All changes must be incremental, reviewed, and tested**
- **Auth Changes**: Follow centralized auth server pattern per ADR-012
- **Test Standards**: Follow Go test conventions with *_test.go naming pattern
- **Type Safety**: Use proper type aliases (UserID, SessionID) and ensure type compatibility
- **Package Structure**: Prefer service-specific implementations over shared abstractions
- **See `README.md` for service details and development instructions**

## Entry Points
- **Backend:** `cmd/` (main.go for each service)
- **Frontend:** `web/` (Next.js app)
- **Protobuf:** `api/proto/`
- **Migrations:** `migrations/`
- **Implementation Plans:** `docs/implementation/` (organized phase-specific documentation)

## For AI Agents
- Use this file as the main entry point for project context
- Review ADRs for architectural decisions
- Follow Clean Architecture and coding standards
- All major changes should be documented with an ADR
