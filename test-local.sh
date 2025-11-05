#!/bin/bash
set -e

echo "🧪 Starting Yukti Local Testing..."

# Clean up any existing containers
echo "🧹 Cleaning up..."
make local-clean

# Start infrastructure
echo "🏗️ Starting local infrastructure..."
make local-setup

# Wait for services to be ready
echo "⏳ Waiting for services..."
sleep 15

# Test database connection
echo "🔍 Testing database connection..."
docker exec yukti-postgres pg_isready -U yukti_user -d yukti_db

# Start the application in background
echo "🚀 Starting Yukti application..."
SPRING_PROFILES_ACTIVE=local mvn spring-boot:run &
APP_PID=$!

# Wait for app to start
echo "⏳ Waiting for application to start..."
sleep 30

# Test endpoints
echo "🧪 Testing API endpoints..."

# Health check
curl -f http://localhost:8081/actuator/health || {
    echo "❌ Health check failed"
    kill $APP_PID
    exit 1
}

# Test FinOps endpoints
curl -f http://localhost:8081/api/finops/report || {
    echo "❌ FinOps report endpoint failed"
    kill $APP_PID
    exit 1
}

# Test recommendations
curl -f http://localhost:8081/api/finops/recommendations/rightsizing || {
    echo "❌ Rightsizing recommendations failed"
    kill $APP_PID
    exit 1
}

echo "✅ All tests passed!"
echo "🌐 Application running at: http://localhost:8081"
echo "📊 Health: http://localhost:8081/actuator/health"
echo "📈 Metrics: http://localhost:8081/actuator/prometheus"
echo "💰 FinOps Report: http://localhost:8081/api/finops/report"

# Keep app running
wait $APP_PID