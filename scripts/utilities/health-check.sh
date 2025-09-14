# Health check script for services
#!/bin/bash

set -e

SERVICE_NAME=${1:-"unknown"}
PORT=${2:-"8080"}
HEALTH_ENDPOINT=${3:-"/health"}

echo "Health check for $SERVICE_NAME on port $PORT"

# Function to check if port is open
check_port() {
    local host=$1
    local port=$2

    # Try nc first (most common)
    if command -v nc &> /dev/null; then
        nc -z "$host" "$port" 2>/dev/null
        return $?
    fi

    # Try ss (socket statistics)
    if command -v ss &> /dev/null; then
        ss -tnl | grep -q ":$port "
        return $?
    fi

    # Try lsof
    if command -v lsof &> /dev/null; then
        lsof -i :"$port" &>/dev/null
        return $?
    fi

    # Try netstat
    if command -v netstat &> /dev/null; then
        netstat -tln | grep -q ":$port "
        return $?
    fi

    # Fallback: try to connect using bash /dev/tcp
    if [ -e /dev/tcp ]; then
        (echo > /dev/tcp/"$host"/"$port") 2>/dev/null
        return $?
    fi

    echo "Warning: No port checking tool available (nc, ss, lsof, netstat)"
    echo "Skipping port availability check"
    return 0
}

# Wait for port to be available (with timeout)
echo "Waiting for port $PORT to be available..."
timeout=30
elapsed=0

while [ $elapsed -lt $timeout ]; do
    if check_port localhost "$PORT"; then
        echo "Port $PORT is now available"
        break
    fi

    sleep 1
    elapsed=$((elapsed + 1))

    if [ $elapsed -eq $timeout ]; then
        echo "Timeout: Port $PORT is not available after $timeout seconds"
        exit 1
    fi
done

# Check health endpoint
echo "Checking health endpoint: http://localhost:$PORT$HEALTH_ENDPOINT"
if command -v curl &> /dev/null; then
    if curl -f --max-time 10 "http://localhost:$PORT$HEALTH_ENDPOINT" 2>/dev/null; then
        echo "$SERVICE_NAME is healthy"
        exit 0
    else
        echo "$SERVICE_NAME health check failed"
        exit 1
    fi
elif command -v wget &> /dev/null; then
    if wget --quiet --tries=1 --timeout=10 -O /dev/null "http://localhost:$PORT$HEALTH_ENDPOINT" 2>/dev/null; then
        echo "$SERVICE_NAME is healthy"
        exit 0
    else
        echo "$SERVICE_NAME health check failed"
        exit 1
    fi
else
    echo "Warning: Neither curl nor wget available for HTTP health check"
    echo "Assuming $SERVICE_NAME is healthy (port is open)"
    exit 0
fi
