# Deployment Options

This directory contains all deployment configurations for the AI Code Generator project. We support **Kubernetes** as the primary deployment method for all environments, with **Docker** available specifically for testing and CI/CD pipelines.

## Directory Structure

```
deployments/
├── k8s/                    # Kubernetes deployments (Primary)
│   ├── helm/              # Helm charts
│   │   └── ai-code-generator/
│   ├── grafana/           # Grafana configuration
│   └── prometheus.yml     # Prometheus configuration
├── docker/                # Docker for testing/CI only
│   ├── docker-compose.yml          # Unified Docker Compose (dev, CI, test)
│   ├── docker-compose.dev.yml      # Development overrides
│   ├── docker-compose.prod.yml     # Production overrides
│   ├── Dockerfile.*                # Service Dockerfiles
│   ├── .dockerignore              # Docker ignore file
│   └── .env.prod.example          # Production env template
└── scripts/               # Deployment scripts
    └── deploy-k8s.sh     # Kubernetes deployment script
```

## Quick Start

### For All Environments (Kubernetes)
```bash
# Deploy to local cluster
./deployments/scripts/deploy-k8s.sh local

# Deploy to staging
./deployments/scripts/deploy-k8s.sh staging

# Deploy to production
./deployments/scripts/deploy-k8s.sh production
```

### For Testing/CI (Docker)
```bash
# Navigate to docker directory
cd deployments/docker

# Run tests in unified environment
docker-compose -f docker-compose.yml up --abort-on-container-exit

# Clean up after testing
docker-compose -f docker-compose.yml down -v
```

## Deployment Methods

### 🐳 Docker (Testing & CI/CD Only)

**Best for:**
- Isolated testing environments
- CI/CD pipelines
- Automated test execution

**Services included:**
- PostgreSQL database
- Redis cache
- All microservices (auth, user, ai, api-gateway)
- Mock AI service for testing
- Monitoring stack (Prometheus, Grafana)

**Usage:**
```bash
# Navigate to docker directory
cd deployments/docker

# Run tests in unified environment
docker-compose -f docker-compose.yml up --abort-on-container-exit

# Run development environment for manual testing
docker-compose -f docker-compose.yml -f docker-compose.dev.yml up

# Clean up containers
docker-compose down -v
```

**Note:** Docker is not intended for production deployment. Use Kubernetes for all production environments.

### ☸️ Kubernetes (Production)

**Best for:**
- Production deployments
- Scalable environments
- High availability
- Enterprise environments

**Features:**
- Helm chart for easy deployment
- Environment-specific configurations
- Service mesh with Istio
- Monitoring with Prometheus/Grafana
- Horizontal Pod Autoscaling

**Usage:**
```bash
# Local cluster
./deployments/scripts/deploy-k8s.sh local

# Staging
./deployments/scripts/deploy-k8s.sh staging

# Production
./deployments/scripts/deploy-k8s.sh production
```

## Environment Configurations

### Docker Environments

| Environment | File | Purpose |
|-------------|------|---------|
| Development | `docker-compose.dev.yml` | Local development with hot reload |
| Production | `docker-compose.prod.yml` | Production-ready with optimizations |
| Testing | `docker-compose.yml` | Unified test environment (same as dev/CI) |

### Kubernetes Environments

| Environment | File | Namespace |
|-------------|------|-----------|
| Local | `values/local.yaml` | `ai-code-generator-local` |
| Staging | `values/staging.yaml` | `ai-code-generator-staging` |
| Production | `values/production.yaml` | `ai-code-generator-prod` |

## Testing Strategy

### Docker-based Testing
- Isolated environment with test database
- No external port conflicts
- Fast startup times
- Coverage reporting
- Clean teardown

```bash
# Navigate to docker directory
cd deployments/docker

# Run tests with coverage
docker-compose -f docker-compose.yml up --abort-on-container-exit

# Run tests and cleanup
docker-compose -f docker-compose.yml down -v --remove-orphans

# View test logs
docker-compose -f docker-compose.yml logs test-runner
```

### CI/CD Integration
The Docker test configuration is designed for CI pipelines:

```yaml
# Example GitHub Actions
- name: Run Tests
  run: |
    cd deployments/docker
    docker-compose -f docker-compose.yml up --abort-on-container-exit
```

## Service Architecture

### Core Services
- **API Gateway**: Entry point for all requests
- **Auth Service**: Authentication and authorization
- **User Service**: User management
- **AI Service**: AI/ML operations
- **AI Generation Service**: Code generation logic

### Infrastructure Services
- **PostgreSQL**: Primary database
- **Redis**: Caching and session storage
- **Prometheus**: Metrics collection
- **Grafana**: Visualization and monitoring

## Configuration Management

### Environment Variables
- Docker: Uses `.env` files and docker-compose environment overrides
- Kubernetes: Uses ConfigMaps and Secrets

### Secrets Management
- Docker: Environment variables (not recommended for production)
- Kubernetes: Kubernetes Secrets with external secret management

## Monitoring & Observability

### Docker Environment
- Redis exporter for metrics
- Basic health checks
- Container logs via docker-compose

### Kubernetes Environment
- Prometheus for metrics collection
- Grafana for visualization
- ServiceMonitors for automatic service discovery
- Distributed tracing support

## Troubleshooting

### Common Docker Issues
```bash
# Check container status
docker-compose ps

# View logs
docker-compose logs -f [service-name]

# Rebuild and restart
docker-compose down && docker-compose up --build
```

### Common Kubernetes Issues
```bash
# Check pod status
kubectl get pods -n [namespace]

# View logs
kubectl logs -f [pod-name] -n [namespace]

# Debug pod
kubectl exec -it [pod-name] -n [namespace] -- /bin/bash
```

## Contributing

When adding new services or modifying deployments:

1. Update both Docker and Kubernetes configurations
2. Test in both environments using docker-compose commands
3. Update this README
4. Update CI/CD pipelines to use docker-compose commands
5. Ensure new services work in isolated test environments

## Migration Guide

### From Docker to Kubernetes
1. Ensure all services are containerized
2. Create Kubernetes manifests
3. Set up ConfigMaps and Secrets
4. Configure Ingress and Services
5. Test in staging environment
6. Deploy to production

### From Kubernetes to Docker
1. Extract environment variables
2. Create docker-compose overrides
3. Configure volume mounts
4. Test service dependencies
5. Validate in development environment