# Health check script for services
#!/bin/bash

set -e

SERVICE_NAME=${1:-"unknown"}
PORT=${2:-"8080"}
HEALTH_ENDPOINT=${3:-"/health"}

echo "Health check for $SERVICE_NAME on port $PORT"

# Wait for port to be available
timeout 30 bash -c "until nc -z localhost $PORT; do sleep 1; done"

# Check health endpoint
curl -f "http://localhost:$PORT$HEALTH_ENDPOINT" || exit 1

echo "$SERVICE_NAME is healthy"
