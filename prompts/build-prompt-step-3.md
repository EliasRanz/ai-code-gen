# Step 3: API Gateway, Auth, and User/Project Service Stubs

## Instructions

Implement the API Gateway, Authentication Service, and User/Project Service as described. Define all routes, middleware, and gRPC interfaces, but do not implement business logic or database access yet.

- API Gateway:
  - Set up all routes and middleware (CORS, logging, tracing, metrics, rate limiting).
  - Reverse proxy stubs for `/api/auth/*`, `/api/users/*`, `/api/projects/*`, `/api/generate/*`.
- Authentication Service:
  - Define endpoints for OAuth login, callback, and refresh.
  - Middleware for JWT validation.
- User/Project Service:
  - Define gRPC interfaces and HTTP stubs for CRUD operations.
- Do not implement business logic or database access yet.
