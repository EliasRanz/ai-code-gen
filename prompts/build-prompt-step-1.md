# Step 1: Project Structure, Tech Stack, and High-Level Goal

## High-Level Goal

Your task is to generate the directory and file structure for a full-stack, production-ready AI UI Generator application, similar in core functionality to v0.dev. The system must be built as a Go-based microservices monorepo. It will take natural language prompts from a user and stream back generated React/Next.js component code. The architecture must be modular, scalable, and include robust authentication, testing, and observability.

## Core Technology Stack

- **Backend:** Go 1.22+
- **Backend Framework:** Gin (`github.com/gin-gonic/gin`)
- **Frontend:** Next.js 14+ with App Router, TypeScript, Tailwind CSS, shadcn/ui
- **Database:** PostgreSQL
- **Cache/Message Broker:** Redis
- **Inter-Service Communication:** gRPC
- **Real-time Streaming:** Server-Sent Events (SSE)
- **Authentication:** OAuth 2.0 (with Google as a provider) and JWTs
- **AI Inference (Self-Hosted):** vLLM serving an OpenAI-compatible API
- **Logging:** Zerolog (`github.com/rs/zerolog`) for structured JSON logging.
- **Testing:** Go's native testing package, `testify/assert`, `testify/mock`, and `testcontainers-go`.
- **Deployment:** Docker, Docker Compose for local development, Kubernetes manifests.

## Project Monorepo Structure

Generate the following directory and file structure. Provide placeholder content for all files.

```
/ai-ui-generator/
├──.github/
│   └── workflows/
│       └── ci.yml
├── api/
│   └── proto/
│       ├── auth.proto
│       └── user.proto
├── cmd/
│   ├── api-gateway/
│   │   └── main.go
│   ├── auth-service/
│   │   └── main.go
│   ├── user-service/
│   │   └── main.go
│   └── ai-service/
│       └── main.go
├── configs/
│   └── config.yaml
├── deployments/
│   ├── docker-compose.yml
│   └── kubernetes/
│       ├── api-gateway-deployment.yaml
│       ├── auth-service-deployment.yaml
│       #... other k8s manifests
├── internal/
│   ├── auth/
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── token.go
│   │   └── middleware.go
│   ├── user/
│   │   ├── handler.go
│   │   ├── service.go
│   │   └── repository.go
│   ├── project/
│   │   #... similar structure for project management
│   ├── ai/
│   │   ├── handler.go
│   │   ├── service.go
│   │   └── llm_abstractor.go
│   ├── database/
│   │   └── postgres.go
│   └── observability/
│       ├── logging.go
│       ├── metrics.go
│       └── tracing.go
├── web/
│   ├── app/
│   │   ├── (auth)/
│   │   │   └── page.tsx
│   │   ├── (dashboard)/
│   │   │   ├── layout.tsx
│   │   │   └── page.tsx
│   │   └── api/
│   │       └── auth/
│   │           └── [...nextauth]/
│   │               └── route.ts
│   ├── components/
│   │   ├── ui/ # shadcn components
│   │   ├── ChatInterface.tsx
│   │   └── PreviewPane.tsx
│   ├── lib/
│   │   └── sse.ts
│   ├── next.config.mjs
│   ├── tailwind.config.ts
│   └── tsconfig.json
├──.env.example
├── Dockerfile.gateway
├── Dockerfile.auth
├── Dockerfile.user
├── Dockerfile.ai
├── go.mod
└── README.md
```
