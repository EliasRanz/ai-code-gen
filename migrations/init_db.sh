#!/bin/bash

# Database initialization script for AI UI Generator
# This script creates the database and runs all migrations

set -e

# Database configuration
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5433}
DB_NAME=${DB_NAME:-ai_ui_generator}
DB_USER=${DB_USER:-postgres}
DB_PASSWORD=${DB_PASSWORD:-}

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}🚀 AI UI Generator Database Setup${NC}"
echo "=================================="

# Check if PostgreSQL is running
if ! pg_isready -h $DB_HOST -p $DB_PORT > /dev/null 2>&1; then
    echo -e "${RED}❌ PostgreSQL is not running on $DB_HOST:$DB_PORT${NC}"
    exit 1
fi

echo -e "${GREEN}✅ PostgreSQL is running${NC}"

# Create database if it doesn't exist
echo -e "${YELLOW}📦 Creating database '$DB_NAME' if it doesn't exist...${NC}"
createdb -h $DB_HOST -p $DB_PORT -U $DB_USER $DB_NAME 2>/dev/null || true

# Check if database exists and is accessible
if ! psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "SELECT 1;" > /dev/null 2>&1; then
    echo -e "${RED}❌ Cannot connect to database '$DB_NAME'${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Database '$DB_NAME' is accessible${NC}"

# Enable required extensions
echo -e "${YELLOW}🔧 Enabling required PostgreSQL extensions...${NC}"
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";
CREATE EXTENSION IF NOT EXISTS \"pgcrypto\";
" > /dev/null

echo -e "${GREEN}✅ Extensions enabled${NC}"

# Run migrations
echo -e "${YELLOW}📊 Running database migrations...${NC}"

MIGRATION_DIR="$(dirname "$0")"
MIGRATIONS=$(ls $MIGRATION_DIR/*.sql | sort)

for migration in $MIGRATIONS; do
    migration_name=$(basename "$migration")
    echo -e "  🔄 Running migration: $migration_name"
    
    # Extract the "Up" part of the migration
    awk '/-- \+migrate Up/,/-- \+migrate Down/{if(/-- \+migrate Down/) exit; if(!/-- \+migrate Up/) print}' "$migration" | \
    psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME > /dev/null
    
    if [ $? -eq 0 ]; then
        echo -e "  ${GREEN}✅ $migration_name completed${NC}"
    else
        echo -e "  ${RED}❌ $migration_name failed${NC}"
        exit 1
    fi
done

# Verify schema
echo -e "${YELLOW}🔍 Verifying database schema...${NC}"
TABLES=$(psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "
SELECT table_name 
FROM information_schema.tables 
WHERE table_schema = 'public' 
AND table_type = 'BASE TABLE'
ORDER BY table_name;
" | tr -d ' ')

echo -e "${GREEN}📋 Created tables:${NC}"
for table in $TABLES; do
    if [ ! -z "$table" ]; then
        echo -e "  📄 $table"
    fi
done

# Show database statistics
echo -e "${YELLOW}📈 Database statistics:${NC}"
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
SELECT 
    schemaname,
    tablename,
    attname AS column_name,
    typname AS data_type
FROM pg_attribute 
JOIN pg_class ON pg_attribute.attrelid = pg_class.oid
JOIN pg_namespace ON pg_class.relnamespace = pg_namespace.oid
JOIN pg_type ON pg_attribute.atttypid = pg_type.oid
WHERE 
    pg_namespace.nspname = 'public'
    AND pg_class.relkind = 'r'
    AND pg_attribute.attnum > 0
    AND NOT pg_attribute.attisdropped
ORDER BY tablename, attname;
" > /dev/null

echo ""
echo -e "${GREEN}🎉 Database setup completed successfully!${NC}"
echo -e "${GREEN}🔗 Connection string: postgresql://$DB_USER@$DB_HOST:$DB_PORT/$DB_NAME${NC}"
echo ""
