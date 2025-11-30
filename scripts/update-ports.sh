#!/bin/bash

# Yukti Platform - Port Update Script
# This script updates ports across the entire codebase

set -e

echo "🔧 Yukti Port Configuration Updater"
echo "===================================="
echo ""

# Load current ports from .env.ports
if [ -f .env.ports ]; then
    source .env.ports
    echo "✅ Loaded ports from .env.ports"
else
    echo "❌ .env.ports file not found!"
    exit 1
fi

echo ""
echo "Current Port Configuration:"
echo "  Backend:    $BACKEND_PORT"
echo "  Frontend:   $FRONTEND_PORT"
echo "  PostgreSQL: $POSTGRES_PORT"
echo "  ML Service: $ML_SERVICE_PORT"
echo "  Prometheus: $PROMETHEUS_PORT"
echo "  Grafana:    $GRAFANA_PORT"
echo ""

# Update frontend API URL in api.ts
echo "📝 Updating frontend/src/services/api.ts..."
sed -i.bak "s|http://localhost:[0-9]*|http://localhost:$BACKEND_PORT|g" frontend/src/services/api.ts
rm -f frontend/src/services/api.ts.bak

# Update backend default port in config.go
echo "📝 Updating internal/config/config.go..."
sed -i.bak "s/getEnv(\"PORT\", \"[0-9]*\")/getEnv(\"PORT\", \"$BACKEND_PORT\")/g" internal/config/config.go
rm -f internal/config/config.go.bak

# Update backend default port in main.go
echo "📝 Updating cmd/main.go..."
sed -i.bak "s/port = \"[0-9]*\"/port = \"$BACKEND_PORT\"/g" cmd/main.go
rm -f cmd/main.go.bak

# Update README.md
echo "📝 Updating README.md..."
sed -i.bak "s|http://localhost:[0-9]* (Docker container)|http://localhost:$BACKEND_PORT (yukti-backend container)|g" README.md
sed -i.bak "s|http://localhost:[0-9]* (yukti-frontend container)|http://localhost:$FRONTEND_PORT (yukti-frontend container)|g" README.md
sed -i.bak "s|localhost:[0-9]* (yukti-postgres container)|localhost:$POSTGRES_PORT (yukti-postgres container)|g" README.md
sed -i.bak "s|http://localhost:[0-9]* (yukti-ml container)|http://localhost:$ML_SERVICE_PORT (yukti-ml container)|g" README.md
sed -i.bak "s|http://localhost:[0-9]* (yukti-prometheus container)|http://localhost:$PROMETHEUS_PORT (yukti-prometheus container)|g" README.md
sed -i.bak "s|http://localhost:[0-9]* (yukti-grafana container)|http://localhost:$GRAFANA_PORT (yukti-grafana container)|g" README.md
rm -f README.md.bak

echo ""
echo "✅ All ports updated successfully!"
echo ""
echo "Next steps:"
echo "  1. Review changes: git diff"
echo "  2. Rebuild containers: docker-compose up -d --build"
echo "  3. Test services at new ports"
echo ""
