#!/bin/bash

# Local Kubernetes Setup Script for AI Code Generator
# This script sets up a local Kubernetes cluster using kind for development

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 Setting up Local Kubernetes Cluster for AI Code Generator${NC}"
echo "================================================================="

# Check if kind is installed
if ! command -v kind &> /dev/null; then
    echo -e "${RED}❌ kind is not installed. Please install kind first:${NC}"
    echo "  - macOS: brew install kind"
    echo "  - Linux: curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.20.0/kind-linux-amd64 && chmod +x ./kind && sudo mv ./kind /usr/local/bin/"
    echo "  - Windows: choco install kind"
    exit 1
fi

# Check if kubectl is installed
if ! command -v kubectl &> /dev/null; then
    echo -e "${RED}❌ kubectl is not installed. Please install kubectl first.${NC}"
    exit 1
fi

# Check if helm is installed
if ! command -v helm &> /dev/null; then
    echo -e "${RED}❌ helm is not installed. Please install helm first.${NC}"
    exit 1
fi

echo -e "${GREEN}✅ All prerequisites are installed${NC}"

# Create kind cluster configuration
cat > kind-config.yaml << EOF
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
  extraPortMappings:
  - containerPort: 80
    hostPort: 80
    protocol: TCP
  - containerPort: 443
    hostPort: 443
    protocol: TCP
EOF

# Create the cluster
echo -e "${BLUE}📦 Creating kind cluster...${NC}"
kind create cluster --name ai-code-generator --config kind-config.yaml

# Install NGINX Ingress Controller
echo -e "${BLUE}🌐 Installing NGINX Ingress Controller...${NC}"
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.8.1/deploy/static/provider/cloud/deploy.yaml

# Wait for ingress controller to be ready
echo -e "${BLUE}⏳ Waiting for ingress controller to be ready...${NC}"
kubectl wait --namespace ingress-nginx \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/component=controller \
  --timeout=300s

# Add Helm repositories
echo -e "${BLUE}📚 Adding Helm repositories...${NC}"
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update

# Create namespace for the application
echo -e "${BLUE}📁 Creating namespace...${NC}"
kubectl create namespace ai-code-generator --dry-run=client -o yaml | kubectl apply -f -

echo -e "${GREEN}✅ Local Kubernetes cluster is ready!${NC}"
echo ""
echo -e "${BLUE}Next steps:${NC}"
echo "1. Build and push your Docker images to a registry"
echo "2. Update the image references in values/local.yaml"
echo "3. Deploy using: helm install ai-code-generator ./deployments/helm/ai-code-generator -f ./deployments/helm/ai-code-generator/values/local.yaml"
echo "4. Access your application at: http://ai-code-generator.local"
echo ""
echo -e "${YELLOW}To clean up:${NC}"
echo "  kind delete cluster --name ai-code-generator"
echo "  rm kind-config.yaml"

# Clean up the config file
rm kind-config.yaml