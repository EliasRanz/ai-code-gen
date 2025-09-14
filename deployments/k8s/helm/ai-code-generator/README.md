# AI Code Generator Helm Chart

A comprehensive Helm chart for deploying the AI Code Generator microservices application on Kubernetes with support for multiple environments (local, staging, production).

## Features

- **Microservices Architecture**: Deploys 6 services (API Gateway, Auth Service, User Service, AI Service, AI Generation Service, Frontend)
- **Database Support**: PostgreSQL and Redis with environment-specific configurations
- **Monitoring Stack**: Prometheus and Grafana for observability
- **Ingress Support**: NGINX ingress controller for external access
- **Environment Configurations**: Separate values files for local, staging, and production
- **Security**: TLS support, secrets management, and RBAC
- **Scalability**: Horizontal Pod Autoscaling support
- **High Availability**: Multi-replica deployments for production

## Prerequisites

- Kubernetes 1.19+
- Helm 3.0+
- kubectl configured to access your cluster

## Quick Start

### Local Development

1. **Set up local Kubernetes cluster**:
   ```bash
   make k8s-setup
   ```

2. **Deploy the application**:
   ```bash
   make k8s-dev
   ```

3. **Access the application**:
   - Frontend: http://ai-code-generator.local
   - API Gateway: http://api.ai-code-generator.local

### Production Deployment

1. **Install the chart**:
   ```bash
   helm install ai-code-generator ./deployments/helm/ai-code-generator \
     -f ./deployments/helm/ai-code-generator/values/production.yaml \
     --namespace ai-code-generator --create-namespace
   ```

2. **Verify deployment**:
   ```bash
   kubectl get pods --namespace ai-code-generator
   ```

## Configuration

### Environment Values Files

- `values.yaml`: Default configuration
- `values/local.yaml`: Local development (minimal resources, no persistence)
- `values/staging.yaml`: Staging environment (moderate resources, external databases)
- `values/production.yaml`: Production environment (full resources, high availability)

### Key Configuration Options

| Parameter | Description | Default |
|-----------|-------------|---------|
| `apiGateway.replicaCount` | Number of API Gateway replicas | `1` |
| `postgresql.enabled` | Enable PostgreSQL deployment | `true` |
| `redis.enabled` | Enable Redis deployment | `true` |
| `monitoring.enabled` | Enable Prometheus and Grafana | `true` |
| `ingress.enabled` | Enable ingress controller | `true` |
| `global.imageRegistry` | Global image registry | `""` |

### Database Configuration

#### Local Development
- Uses embedded PostgreSQL and Redis pods
- Data persistence disabled (data lost on pod restart)
- Minimal resource allocation

#### Staging/Production
- Uses external database services
- Set `postgresql.enabled: false` and `redis.enabled: false`
- Configure connection strings via environment variables

### Monitoring Configuration

The chart includes a complete monitoring stack:

#### Prometheus
- Service discovery via ServiceMonitor
- Custom alerting rules
- 30-day data retention (configurable)

#### Grafana
- Pre-configured dashboards
- Admin user: `admin` / `admin` (change in production!)
- Persistent storage for dashboards and data

#### Alerting Rules
- High error rate detection
- Service availability monitoring
- Pod restart alerts
- Latency monitoring

## Services

### API Gateway
- **Port**: 80
- **Health Check**: `/health`
- **Metrics**: `/metrics`
- Routes requests to appropriate microservices

### Auth Service
- **Port**: 8081
- **Purpose**: User authentication and authorization
- **Database**: PostgreSQL for user sessions

### User Service
- **Port**: 8082
- **Purpose**: User management and profiles
- **Database**: PostgreSQL for user data

### AI Service
- **Port**: 8083
- **Purpose**: AI model interactions
- **Dependencies**: External AI APIs

### AI Generation Service
- **Port**: 8084
- **Purpose**: Code generation using AI
- **Resources**: Higher CPU/memory allocation

### Frontend
- **Port**: 3000
- **Purpose**: React-based user interface
- **Build**: Static file serving

## Usage Examples

### Install with Custom Values

```bash
helm install ai-code-generator ./deployments/helm/ai-code-generator \
  --namespace ai-code-generator \
  --create-namespace \
  -f values/production.yaml \
  --set apiGateway.replicaCount=5 \
  --set global.imageTag=v1.2.3
```

### Upgrade Deployment

```bash
helm upgrade ai-code-generator ./deployments/helm/ai-code-generator \
  --namespace ai-code-generator \
  -f values/production.yaml
```

### Rollback Deployment

```bash
helm rollback ai-code-generator --namespace ai-code-generator
```

### Uninstall

```bash
helm uninstall ai-code-generator --namespace ai-code-generator
```

## Monitoring and Observability

### Accessing Grafana

```bash
# Port forward Grafana service
kubectl port-forward --namespace ai-code-generator svc/ai-code-generator-grafana 3000:80

# Access at: http://localhost:3000
# Default credentials: admin / admin
```

### Accessing Prometheus

```bash
# Port forward Prometheus service
kubectl port-forward --namespace ai-code-generator svc/ai-code-generator-prometheus-server 9090:80

# Access at: http://localhost:9090
```

### Viewing Logs

```bash
# All service logs
kubectl logs -f --namespace ai-code-generator -l app.kubernetes.io/instance=ai-code-generator

# Specific service logs
kubectl logs -f --namespace ai-code-generator deployment/ai-service
```

## Troubleshooting

### Common Issues

1. **Pods not starting**:
   ```bash
   kubectl describe pod [pod-name] --namespace ai-code-generator
   kubectl logs [pod-name] --namespace ai-code-generator
   ```

2. **Service unavailable**:
   ```bash
   kubectl get services --namespace ai-code-generator
   kubectl get ingress --namespace ai-code-generator
   ```

3. **Database connection issues**:
   ```bash
   kubectl exec -it [db-pod] --namespace ai-code-generator -- psql -U postgres
   ```

4. **Resource constraints**:
   ```bash
   kubectl top pods --namespace ai-code-generator
   kubectl describe node
   ```

### Debug Commands

```bash
# Check all resources
kubectl get all --namespace ai-code-generator

# Check events
kubectl get events --namespace ai-code-generator --sort-by=.metadata.creationTimestamp

# Check persistent volumes
kubectl get pv,pvc --namespace ai-code-generator

# Debug specific pod
kubectl exec -it [pod-name] --namespace ai-code-generator -- /bin/sh
```

## Security Considerations

### Production Deployment

1. **Change default passwords**:
   ```yaml
   monitoring:
     grafana:
       adminPassword: "your-secure-password"
   ```

2. **Enable TLS**:
   ```yaml
   ingress:
     tls:
     - secretName: ai-code-generator-tls
       hosts:
       - ai-code-generator.local
   ```

3. **Use secrets for sensitive data**:
   ```yaml
   apiGateway:
     config:
       databaseUrl: "postgresql://user:password@host:5432/db"
   ```

4. **RBAC configuration**:
   - Use service accounts with minimal permissions
   - Implement network policies for pod-to-pod communication

### Image Security

- Scan images for vulnerabilities before deployment
- Use specific image tags (avoid `latest`)
- Implement image pull secrets for private registries

## Scaling

### Horizontal Pod Autoscaling

Enable HPA for services:

```yaml
apiGateway:
  autoscaling:
    enabled: true
    minReplicas: 2
    maxReplicas: 10
    targetCPUUtilizationPercentage: 70
    targetMemoryUtilizationPercentage: 80
```

### Manual Scaling

```bash
kubectl scale deployment ai-service --replicas=5 --namespace ai-code-generator
```

## Backup and Recovery

### Database Backup

```bash
# PostgreSQL backup
kubectl exec -it [postgres-pod] --namespace ai-code-generator -- pg_dump -U postgres ai_ui_generator > backup.sql

# Redis backup
kubectl exec -it [redis-pod] --namespace ai-code-generator -- redis-cli save
```

### Chart Backup

```bash
# Save current values
helm get values ai-code-generator --namespace ai-code-generator > current-values.yaml
```

## Contributing

1. Update chart version in `Chart.yaml`
2. Test changes in local environment
3. Update documentation
4. Create pull request

## Support

For issues and questions:
- Check the troubleshooting section
- Review Kubernetes and Helm documentation
- Create an issue in the project repository

## License

This Helm chart is part of the AI Code Generator project.