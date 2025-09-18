# AI UI Generator - Workspace Instructions

This file provides general guidance for AI agents working on the AI UI Generator project. These instructions apply to all AI interactions within this workspace, regardless of the specific chat mode being used.

## Project Overview

The AI UI Generator is a production-ready, full-stack AI UI Generation System inspired by Vercel's v0.dev. It transforms natural language prompts into high-quality, interactive frontend components using a modular, scalable microservices architecture.

### Technology Stack
- **Backend**: Go microservices with Gin framework
- **Frontend**: Next.js 14+ with TypeScript, Tailwind CSS, shadcn/ui
- **Database**: PostgreSQL with Redis caching
- **Authentication**: OAuth 2.0 with JWT tokens
- **AI Integration**: vLLM serving OpenAI-compatible API
- **Communication**: gRPC between services, SSE for real-time streaming
- **Deployment**: Docker & Kubernetes ready

## Core Principles

### 1. Microservices-First Architecture
The system is designed around focused, independent microservices:
- Respect existing service boundaries
- Use gRPC for inter-service communication
- Maintain service independence and avoid tight coupling

### 2. Developer Productivity
- Always use existing Makefile commands for consistency
- Leverage established tooling and frameworks
- Focus on business logic rather than infrastructure concerns

### 3. Quality & Security
- Security considerations are paramount
- Comprehensive testing strategy with proper mocking
- Performance monitoring and optimization
- Clean, maintainable code following established patterns

## Essential Commands

### Development Workflow
```bash
# Complete environment setup
make setup

# Start development environment
make dev

# Run complete validation pipeline
make ci

# Quick unit tests
make test

# Integration tests with testcontainers
make test-integration

# Performance testing
make perf-test

# Security scanning
make security

# Code generation (mocks, protobuf)
make generate
```

### Service Architecture
```
internal/
├── ai/          # AI generation and LLM orchestration
├── auth/        # Authentication and authorization
├── user/        # User management and profiles
├── gateway/     # API gateway and routing
├── cache/       # Caching strategies and Redis integration
├── observability/ # Monitoring, logging, and metrics
└── utilities/   # Shared utilities and interfaces
```

## Guidelines for AI Agents

### Before Starting Work
1. **Understand the Context**: Review existing code and documentation
2. **Check Current State**: Use `make ci` to ensure everything is working
3. **Plan Your Changes**: Consider impact on microservice boundaries
4. **Validate Assumptions**: Ask for clarification when uncertain

### During Development
1. **Follow Existing Patterns**: Maintain consistency with established code patterns
2. **Use Project Tooling**: Leverage Makefile commands and existing infrastructure
3. **Test Thoroughly**: Write appropriate unit and integration tests
4. **Document Changes**: Update relevant documentation and add ADRs for significant decisions

### Code Quality Standards
1. **Security First**: Never log sensitive data, validate all inputs
2. **Proper Testing**: Mock external dependencies, use testcontainers for integration tests
3. **Performance Aware**: Consider performance impact of changes
4. **Clean Code**: Follow established coding standards and naming conventions

### Communication Patterns
1. **gRPC for Internal**: Use gRPC for service-to-service communication
2. **REST for External**: Use REST APIs for frontend and external integrations
3. **SSE for Streaming**: Use Server-Sent Events for real-time features

## Common Pitfalls to Avoid

1. **Don't Break Service Boundaries**: Avoid direct coupling between microservices
2. **Don't Skip Testing**: All code changes should include appropriate tests
3. **Don't Hardcode Secrets**: Use environment variables and proper configuration
4. **Don't Ignore Performance**: Consider performance impact of all changes
5. **Don't Reinvent Solutions**: Use existing libraries and frameworks when appropriate

## Getting Help

### Documentation Locations
- **Architecture Decisions**: `ADRs/` directory
- **Setup Instructions**: `README.md` and `docs/` directory
- **API Documentation**: Service-specific documentation in each package
- **Testing Guides**: `docs/ENHANCED_TESTING.md`

### Troubleshooting
- **Build Issues**: Run `make ci` to diagnose problems systematically
- **Service Issues**: Use `make dev-status` to check service health
- **Performance Issues**: Use `make perf-*` commands for analysis
- **Test Failures**: Check both unit and integration test results

### When to Ask Questions
- Unclear requirements or specifications
- Uncertainty about architectural decisions
- Security concerns or potential vulnerabilities
- Performance implications of proposed changes
- Questions about existing code patterns or conventions

## Project-Specific Context

### Current Development Focus
The project emphasizes:
- Clean architecture with proper service boundaries
- Comprehensive testing strategy with high coverage
- Performance optimization and monitoring
- Security best practices throughout
- Developer productivity and tooling

### Key Success Metrics
- Code quality and maintainability
- Test coverage and reliability
- Performance and scalability
- Security and compliance
- Developer experience and productivity

This workspace is designed to support rapid, confident development while maintaining high standards for quality, security, and performance.