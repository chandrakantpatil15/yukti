#!/bin/bash

# Setup Script - Prepares environment for testing

set -e

echo "=========================================="
echo "Yukti Testing Environment Setup"
echo "=========================================="

# Check if PostgreSQL is running
echo "Checking PostgreSQL..."
if pg_isready -U yukti > /dev/null 2>&1; then
    echo "✓ PostgreSQL is running"
else
    echo "✗ PostgreSQL is not running"
    echo "  Start it with: brew services start postgresql"
    exit 1
fi

# Check database exists
echo "Checking database..."
if psql -U yukti -lqt | cut -d \| -f 1 | grep -qw yukti_finops; then
    echo "✓ Database yukti_finops exists"
else
    echo "✗ Database yukti_finops does not exist"
    echo "  Create it with: createdb -U yukti yukti_finops"
    exit 1
fi

# Apply migrations
echo ""
echo "Applying database migrations..."

if [ -f "scripts/014_create_hidden_cost_findings.sql" ]; then
    echo "  Applying 014_create_hidden_cost_findings.sql..."
    psql -U yukti -d yukti_finops -f scripts/014_create_hidden_cost_findings.sql > /dev/null 2>&1 || true
    echo "  ✓ Applied"
fi

if [ -f "scripts/015_add_performance_indexes.sql" ]; then
    echo "  Applying 015_add_performance_indexes.sql..."
    psql -U yukti -d yukti_finops -f scripts/015_add_performance_indexes.sql > /dev/null 2>&1 || true
    echo "  ✓ Applied"
fi

# Check required tables
echo ""
echo "Verifying database schema..."
REQUIRED_TABLES=(
    "yt_tenants"
    "yt_users"
    "yt_aws_accounts"
    "yt_tenant_resources"
    "yt_hidden_cost_findings"
    "yt_customers"
    "yt_cost_data"
)

for table in "${REQUIRED_TABLES[@]}"; do
    if psql -U yukti -d yukti_finops -t -c "SELECT to_regclass('$table');" | grep -q "$table"; then
        echo "  ✓ $table exists"
    else
        echo "  ✗ $table missing"
        echo "    Run: psql -U yukti -d yukti_finops -f scripts/schema.sql"
        exit 1
    fi
done

# Set environment variables
echo ""
echo "Setting environment variables..."
export DATABASE_URL="postgres://yukti:yukti123@localhost:5432/yukti_finops?sslmode=disable"
export JWT_SECRET="test-secret-key-for-local-development-only"
export ENVIRONMENT="development"
export CORS_ALLOWED_ORIGINS="http://localhost:3000"
export ADMIN_KEY="admin-key-123"
export ADMIN_USER="admin@yukti.io"

echo "  ✓ Environment variables set"

# Check if backend is built
echo ""
echo "Checking backend binary..."
if [ ! -f "yukti-api" ]; then
    echo "  Building backend..."
    go build -o yukti-api cmd/main.go
    echo "  ✓ Backend built"
else
    echo "  ✓ Backend binary exists"
fi

# Summary
echo ""
echo "=========================================="
echo "Setup Complete!"
echo "=========================================="
echo ""
echo "Environment Variables:"
echo "  DATABASE_URL: $DATABASE_URL"
echo "  JWT_SECRET: ${JWT_SECRET:0:20}..."
echo "  ENVIRONMENT: $ENVIRONMENT"
echo "  CORS_ALLOWED_ORIGINS: $CORS_ALLOWED_ORIGINS"
echo ""
echo "Next Steps:"
echo "  1. Start backend: ./yukti-api"
echo "  2. Run tests: ./test-backend.sh"
echo ""
