# Kubernetes Deployment Guide

This guide explains how to deploy the AI Code Generator application using Kubernetes and Helm for consistent environments across local development, staging, and production.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Local Development Setup](#local-development-setup)
3. [Environment Configuration](#environment-configuration)
4. [Deployment Commands](#deployment-commands)
5. [Monitoring and Debugging](#monitoring-and-debugging)
6. [CI/CD Integration](#cicd-integration)
7. [Troubleshooting](#troubleshooting)

## Prerequisites

### Required Tools

- **kind** (Kubernetes in Docker): `brew install kind` (macOS) or [download from GitHub](https://kind.sigs.k8s.io/)
- **kubectl**: `brew install kubectl` (macOS) or [install via official docs](https://kubernetes.io/docs/tasks/tools/)
- **Helm**: `brew install helm` (macOS) or [install via official docs](https://helm.sh/docs/intro/install/)
- **Docker**: For building container images

### Verify Installation

```bash
# Check all tools are installed
kind version
kubectl version --client
helm version
docker --version
```

## Local Development Setup

### 1. Set Up Local Kubernetes Cluster

```bash
# Create local cluster with ingress support
make k8s-setup
```

This command will:
- Create a kind cluster named `ai-code-generator`
- Install NGINX Ingress Controller
- Set up port forwarding (80/443 → localhost)
- Add required Helm repositories

### 2. Build and Deploy

```bash
# Quick start for development
make k8s-dev
```

This will:
- Set up the local cluster (if not already done)
- Build Docker images for all services
- Deploy using Helm with local configuration
- Wait for all services to be ready

### 3. Access Your Application

Once deployed, access your application at:
- **Frontend**: http://ai-code-generator.local
- **API Gateway**: http://api.ai-code-generator.local
- **Individual Services**: http://[service-name].ai-code-generator.local

## Environment Configuration

The Helm chart supports three environments with different configurations:

### Local Environment (`values/local.yaml`)
- **Purpose**: Development and testing
- **Resources**: Minimal (256Mi RAM, 0.2 CPU per service)
- **Replicas**: 1 per service
- **Persistence**: Disabled (data doesn't persist between restarts)
- **Databases**: PostgreSQL and Redis as separate pods

### Staging Environment (`values/staging.yaml`)
- **Purpose**: Pre-production testing
- **Resources**: Moderate (512Mi RAM, 0.5 CPU per service)
- **Replicas**: 2 per service (high availability)
- **Persistence**: Enabled with PVCs
- **Databases**: External PostgreSQL and Redis services

### Production Environment (`values/production.yaml`)
- **Purpose**: Live production deployment
- **Resources**: Full (1Gi RAM, 1 CPU per service)
- **Replicas**: 3 per service (high availability)
- **Persistence**: Enabled with robust PVCs
- **Databases**: External managed services
- **Security**: TLS enabled, secrets management

## Deployment Commands

### Basic Commands

```bash
# Install/upgrade deployment
make helm-install ENV=local          # Deploy to local
make helm-install ENV=staging        # Deploy to staging
make helm-install ENV=production     # Deploy to production

# Check status
make helm-status                    # Show release status
make helm-pods                      # List all pods
make helm-services                  # List all services
make helm-ingress                   # Show ingress config

# View logs
make helm-logs                      # Follow all pod logs
kubectl logs -f [pod-name]          # Follow specific pod logs

# Debug issues
make helm-debug                     # Comprehensive debug info
```

### Advanced Commands

```bash
# Validate chart before deployment
make helm-validate

# Show rendered templates (dry run)
make helm-template ENV=local

# Update dependencies
make helm-deps

# Uninstall deployment
make helm-uninstall
```

### Environment-Specific Deployments

```bash
# Local development
make k8s-local                     # Full local setup
make k8s-local-cleanup             # Clean up local environment

# Staging/Production (requires cluster access)
make k8s-staging                   # Deploy to staging
make k8s-production                # Deploy to production
```

## Monitoring and Debugging

### Health Checks

```bash
# Check pod health
kubectl get pods --namespace ai-code-generator

# Check service endpoints
kubectl get services --namespace ai-code-generator

# Check ingress
kubectl get ingress --namespace ai-code-generator

# View detailed pod information
kubectl describe pod [pod-name] --namespace ai-code-generator
```

### Logs and Troubleshooting

```bash
# View all logs
make helm-logs

# View specific service logs
kubectl logs -f deployment/ai-service --namespace ai-code-generator

# Debug pod issues
kubectl exec -it [pod-name] --namespace ai-code-generator -- /bin/sh

# Check events
kubectl get events --namespace ai-code-generator --sort-by=.metadata.creationTimestamp
```

### Common Issues

1. **Pods not starting**: Check resource limits and image pull errors
2. **Services not accessible**: Verify ingress configuration and DNS
3. **Database connection issues**: Check database service and credentials
4. **Image pull errors**: Ensure images are built and pushed to registry

## CI/CD Integration

The project includes a GitHub Actions workflow (`.github/workflows/deploy-k8s.yml`) that:

### Automated Pipeline

1. **Test Stage**: Runs all tests and coverage analysis
2. **Build Stage**: Builds and pushes Docker images to GitHub Container Registry
3. **Validate Stage**: Validates Helm chart syntax and templates
4. **Deploy Stage**: Deploys to staging (develop branch) or production (main branch)

### Required Secrets

For the CI/CD pipeline to work, set these GitHub secrets:

```bash
# For container registry access
GITHUB_TOKEN                 # Auto-provided by GitHub

# For staging deployment
STAGING_KUBECONFIG          # Base64-encoded kubeconfig for staging

# For production deployment
PRODUCTION_KUBECONFIG       # Base64-encoded kubeconfig for production

# For notifications (optional)
SLACK_WEBHOOK_URL           # Slack webhook for deployment notifications
```

### Manual Deployment

You can also deploy manually using the provided scripts:

```bash
# Deploy to staging
ENVIRONMENT=staging make helm-upgrade

# Deploy to production
ENVIRONMENT=production make helm-upgrade
```

## Troubleshooting

### Cluster Issues

```bash
# Reset local cluster
make k8s-cleanup
make k8s-setup

# Check cluster status
kubectl cluster-info
kubectl get nodes
```

### Deployment Issues

```bash
# Check Helm release status
helm status ai-code-generator --namespace ai-code-generator

# Rollback to previous version
helm rollback ai-code-generator --namespace ai-code-generator

# Force redeploy
helm upgrade --install ai-code-generator ./deployments/helm/ai-code-generator \
  --namespace ai-code-generator \
  --force
```

### Network Issues

```bash
# Check ingress controller
kubectl get pods --namespace ingress-nginx

# Debug DNS resolution
nslookup ai-code-generator.local

# Check port forwarding
kubectl port-forward --namespace ingress-nginx service/ingress-nginx-controller 8080:80
```

### Resource Issues

```bash
# Check resource usage
kubectl top pods --namespace ai-code-generator
kubectl top nodes

# Scale deployment
kubectl scale deployment ai-service --replicas=2 --namespace ai-code-generator
```

## Architecture Overview

The Helm chart deploys:

- **6 Microservices**: API Gateway, Auth Service, User Service, AI Service, AI Generation Service, Frontend
- **Databases**: PostgreSQL and Redis (local) or external connections (staging/production)
- **Ingress**: NGINX ingress controller for external access
- **ConfigMaps/Secrets**: Environment-specific configuration
- **Services**: Internal service discovery
- **Persistent Volumes**: Data persistence for databases

### Service Dependencies

```
Frontend → API Gateway → Auth Service
                    → User Service
                    → AI Service
                    → AI Generation Service

All Services → PostgreSQL
All Services → Redis
```

This architecture ensures:
- **Scalability**: Each service can be scaled independently
- **Resilience**: Services are loosely coupled
- **Consistency**: Same deployment process across all environments
- **Observability**: Centralized logging and monitoring

## Next Steps

1. **Customize Configuration**: Update `values/local.yaml` for your development needs
2. **Add Monitoring**: Integrate Prometheus and Grafana for observability
3. **Security**: Configure TLS certificates and secrets management
4. **Backup**: Set up database backup procedures
5. **Scaling**: Configure horizontal pod autoscaling

For more advanced configurations, refer to the Helm chart documentation in `deployments/helm/ai-code-generator/README.md`.