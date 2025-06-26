# AI UI Generator

A production-ready, full-stack AI UI Generation System inspired by Vercel's v0.dev. Transform natural language prompts into high-quality, interactive frontend components using a modular, scalable microservices architecture.

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

## Quick Start

### Prerequisites

- Go 1.22+
- Node.js 18+
- Docker & Docker Compose
- PostgreSQL
- Redis

### Development Setup

1. **Clone and setup environment**:
   ```bash
   git clone <repository-url>
   cd ai-ui-generator
   cp .env.example .env
   # Edit .env with your configuration
   ```

2. **Start infrastructure**:
   ```bash
   docker-compose up postgres redis -d
   ```

3. **Run backend services**:
   ```bash
   # Install Go dependencies
   go mod download
   
   # Run services (in separate terminals)
   go run cmd/api-gateway/main.go
   go run cmd/auth-service/main.go
   go run cmd/user-service/main.go
   go run cmd/ai-service/main.go
   ```

4. **Run frontend**:
   ```bash
   cd web
   npm install
   npm run dev
   ```

5. **Access the application**:
   - Frontend: http://localhost:3000
   - API Gateway: http://localhost:8080
   - Services: Ports 8081-8083

### Docker Deployment

```bash
# Build and run all services
docker-compose up --build

# Or run in detached mode
docker-compose up -d --build
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

### Testing

```bash
# Run Go tests
go test ./...

# Run frontend tests
cd web && npm test

# Run integration tests
go test -tags=integration ./...
```

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
