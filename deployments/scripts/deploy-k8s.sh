#!/bin/bash

# Kubernetes Deployment Script
# Usage: ./deploy-k8s.sh [environment]
# Environments: local, staging, production

set -e

ENVIRONMENT=${1:-local}
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$(dirname "$SCRIPT_DIR")")"

echo "🚀 Deploying to Kubernetes - Environment: $ENVIRONMENT"

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo "❌ kubectl not found. Please install kubectl."
    exit 1
fi

# Check if helm is available
if ! command -v helm &> /dev/null; then
    echo "❌ helm not found. Please install helm."
    exit 1
fi

# Navigate to helm directory
cd "$PROJECT_ROOT/deployments/k8s/helm/ai-code-generator"

# Update dependencies
echo "📦 Updating Helm dependencies..."
helm dependency update

# Deploy based on environment
case $ENVIRONMENT in
    local)
        echo "🏠 Deploying to local Kubernetes cluster..."
        helm upgrade --install ai-code-generator . \
            --namespace ai-code-generator-local \
            --create-namespace \
            --values values/local.yaml \
            --wait
        ;;
    staging)
        echo "🧪 Deploying to staging environment..."
        helm upgrade --install ai-code-generator . \
            --namespace ai-code-generator-staging \
            --create-namespace \
            --values values/staging.yaml \
            --wait
        ;;
    production)
        echo "🏭 Deploying to production environment..."
        echo "⚠️  This will deploy to production. Are you sure? (y/N)"
        read -r confirm
        if [[ $confirm =~ ^[Yy]$ ]]; then
            helm upgrade --install ai-code-generator . \
                --namespace ai-code-generator-prod \
                --create-namespace \
                --values values/production.yaml \
                --wait
        else
            echo "❌ Deployment cancelled."
            exit 1
        fi
        ;;
    *)
        echo "❌ Invalid environment: $ENVIRONMENT"
        echo "📖 Valid environments: local, staging, production"
        exit 1
        ;;
esac

echo "✅ Deployment completed successfully!"
echo "🔍 Check status with: kubectl get pods -n ai-code-generator-$ENVIRONMENT"