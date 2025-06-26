# Step 2: Backend Service Scaffolding and Shared Libraries

## Instructions

For the monorepo structure above, implement the entrypoint files for each Go service (`main.go`), the shared observability package, and configuration loading. Ensure each service can start, log, and shut down gracefully, but do not implement any business logic or API endpoints yet.

- Each service should:
  - Load configuration from environment variables (use `godotenv` for local dev).
  - Initialize logging (Zerolog, structured JSON).
  - Initialize observability (Prometheus, OpenTelemetry) via a shared package (`internal/observability`).
  - Implement graceful shutdown.
- Do not implement any business logic, API endpoints, or database access yet.
