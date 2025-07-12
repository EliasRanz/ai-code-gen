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

## Core Services
- **API Gateway:** Unified entry point, routing, CORS, rate limiting, JWT validation
- **Auth Service:** OAuth 2.0 flow, JWT issuance/validation, session management
- **User Service:** User profiles, projects, chat sessions, CRUD APIs
- **AI Service:** LLM integration, multi-language code generation (frontend, backend, infrastructure), prompt engineering, validation

## Clean Architecture Layers
- **Domain:** Pure business logic, entities, interfaces (no external deps)
- **Application:** Use cases, orchestrates domain logic
- **Infrastructure:** Adapters for DB, LLM, auth, observability
- **Interfaces:** HTTP/gRPC handlers, middleware, routing

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
- **Test coverage:** All business logic, adapters, and handlers
- **CI/CD:** Automated testing and deployment pipeline (planned)
- **Test Organization:** Domain-based test structure following Go conventions (*_test.go pattern)

## Security & Observability
- JWT secret management, password hashing (bcrypt)
- Input validation, error handling, rate limiting
- Structured logging (zerolog), metrics, tracing (OpenTelemetry)

## Current Architecture Changes (2025-01-11)
- **Auth Centralization**: Centralized auth server implementation (see ADR-012)
- **Implementation Status**: See `CENTRALIZED_AUTH_PLAN.md` for current progress
- **Auth Service**: Single source of truth for all authentication/authorization
- **Service Cleanup**: Removing distributed auth logic from other services

## Agent Guidance
- **All code follows Clean Architecture and SOLID principles**
- **ADRs**: Architectural decisions are documented in `ADRs/`
- **All business logic is isolated and testable**
- **All changes must be incremental, reviewed, and tested**
- **Auth Changes**: Follow centralized auth server pattern per ADR-012
- **Test Standards**: Follow Go test conventions with *_test.go naming pattern
- **Type Safety**: Use proper type aliases (UserID, SessionID) and ensure type compatibility
- **See `README.md` for service details and development instructions**

## Entry Points
- **Backend:** `cmd/` (main.go for each service)
- **Frontend:** `web/` (Next.js app)
- **Protobuf:** `api/proto/`
- **Migrations:** `migrations/`

## For AI Agents
- Use this file as the main entry point for project context
- Review ADRs for architectural decisions
- Follow Clean Architecture and coding standards
- All major changes should be documented with an ADR
