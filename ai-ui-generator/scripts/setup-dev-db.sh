#!/bin/bash

# Development database setup script
# This script sets up PostgreSQL and Redis for local development

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 Setting up AI UI Generator Development Environment${NC}"
echo "================================================================"

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo -e "${RED}❌ Docker is not running. Please start Docker and try again.${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Docker is running${NC}"

# Check if docker-compose is available
if ! command -v docker-compose &> /dev/null; then
    echo -e "${RED}❌ docker-compose is not installed. Please install docker-compose and try again.${NC}"
    exit 1
fi

echo -e "${GREEN}✅ docker-compose is available${NC}"

# Start the services
echo -e "${YELLOW}🔄 Starting development services...${NC}"
docker-compose -f docker-compose.dev.yml up -d

# Wait for PostgreSQL to be ready
echo -e "${YELLOW}⏳ Waiting for PostgreSQL to be ready...${NC}"
timeout=60
counter=0

while ! docker-compose -f docker-compose.dev.yml exec -T postgres pg_isready -U postgres -d ai_ui_generator > /dev/null 2>&1; do
    if [ $counter -ge $timeout ]; then
        echo -e "${RED}❌ PostgreSQL failed to start within $timeout seconds${NC}"
        docker-compose -f docker-compose.dev.yml logs postgres
        exit 1
    fi
    sleep 2
    counter=$((counter + 2))
    echo -n "."
done

echo ""
echo -e "${GREEN}✅ PostgreSQL is ready${NC}"

# Wait for Redis to be ready
echo -e "${YELLOW}⏳ Waiting for Redis to be ready...${NC}"
timeout=30
counter=0

while ! docker-compose -f docker-compose.dev.yml exec -T redis redis-cli ping > /dev/null 2>&1; do
    if [ $counter -ge $timeout ]; then
        echo -e "${RED}❌ Redis failed to start within $timeout seconds${NC}"
        docker-compose -f docker-compose.dev.yml logs redis
        exit 1
    fi
    sleep 2
    counter=$((counter + 2))
    echo -n "."
done

echo ""
echo -e "${GREEN}✅ Redis is ready${NC}"

# Run database migrations
echo -e "${YELLOW}📊 Running database migrations...${NC}"
docker-compose -f docker-compose.dev.yml exec -T postgres bash -c "
    cd /docker-entrypoint-initdb.d
    for migration in \$(ls *.sql | sort); do
        echo \"Running migration: \$migration\"
        # Extract the 'Up' part of the migration
        awk '/-- \\+migrate Up/,/-- \\+migrate Down/{if(/-- \\+migrate Down/) exit; if(!/-- \\+migrate Up/) print}' \"\$migration\" | \\
        psql -U postgres -d ai_ui_generator > /dev/null
        if [ \$? -eq 0 ]; then
            echo \"✅ \$migration completed\"
        else
            echo \"❌ \$migration failed\"
            exit 1
        fi
    done
"

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Database migrations completed${NC}"
else
    echo -e "${RED}❌ Database migrations failed${NC}"
    exit 1
fi

# Show service status
echo ""
echo -e "${BLUE}📋 Development Environment Status:${NC}"
echo "=================================="
echo -e "${GREEN}PostgreSQL:${NC} localhost:5432 (ai_ui_generator)"
echo -e "${GREEN}Redis:${NC}      localhost:6379"
echo -e "${GREEN}Adminer:${NC}    http://localhost:8090"
echo ""
echo -e "${YELLOW}Database Connection Details:${NC}"
echo "Host:     localhost"
echo "Port:     5432"
echo "Database: ai_ui_generator"
echo "Username: postgres"
echo "Password: password"
echo ""
echo -e "${YELLOW}Redis Connection Details:${NC}"
echo "Host:     localhost"
echo "Port:     6379"
echo "Password: (none)"
echo ""
echo -e "${GREEN}🎉 Development environment is ready!${NC}"
echo ""
echo -e "${BLUE}Useful commands:${NC}"
echo "  Stop services:    docker-compose -f docker-compose.dev.yml down"
echo "  View logs:        docker-compose -f docker-compose.dev.yml logs"
echo "  Database shell:   docker-compose -f docker-compose.dev.yml exec postgres psql -U postgres -d ai_ui_generator"
echo "  Redis shell:      docker-compose -f docker-compose.dev.yml exec redis redis-cli"
echo ""
