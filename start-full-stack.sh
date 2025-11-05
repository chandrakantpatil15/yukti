#!/bin/bash

echo "🚀 YUKTI FINOPS - FULL STACK STARTUP"
echo "===================================="

# Check prerequisites
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go first."
    exit 1
fi

if ! command -v node &> /dev/null; then
    echo "❌ Node.js is not installed. Please install Node.js first."
    exit 1
fi

if ! command -v psql &> /dev/null; then
    echo "❌ PostgreSQL is not installed. Please install PostgreSQL first."
    exit 1
fi

# Set environment variables
export DATABASE_URL="postgres://chandrakantpatil@localhost:5432/yukti?sslmode=disable"
export PORT=8085
export HEALTH_PORT=8081

echo "🗄️  Database: $DATABASE_URL"
echo "🌐 API Server: http://localhost:$PORT"
echo "🏥 Health Monitor: http://localhost:$HEALTH_PORT"
echo "📱 React UI: http://localhost:3000"
echo ""

# Function to cleanup processes on exit
cleanup() {
    echo ""
    echo "🧹 Cleaning up processes..."
    pkill -f "api-server" 2>/dev/null || true
    pkill -f "health-monitor" 2>/dev/null || true
    pkill -f "react-scripts" 2>/dev/null || true
    pkill -f "node.*start" 2>/dev/null || true
    exit 0
}

# Set trap to cleanup on script exit
trap cleanup SIGINT SIGTERM EXIT

# Test database connection
echo "🔍 Testing database connection..."
if ! psql $DATABASE_URL -c "SELECT 1;" > /dev/null 2>&1; then
    echo "❌ Database connection failed. Please check PostgreSQL is running."
    exit 1
fi
echo "✅ Database connection successful"

# Sync latest AWS data
echo ""
echo "📊 Syncing AWS data..."
go run cmd/sync-all-aws-data.go
if [ $? -ne 0 ]; then
    echo "⚠️  AWS data sync had issues, but continuing..."
fi

# Start API server
echo ""
echo "🚀 Starting API Server..."
go run cmd/api-server.go &
API_PID=$!
sleep 3

# Test API server
if curl -s http://localhost:$PORT/health > /dev/null; then
    echo "✅ API Server started successfully"
else
    echo "❌ API Server failed to start"
    exit 1
fi

# Start Health Monitor
echo ""
echo "🏥 Starting Health Monitor..."
go run cmd/health-monitor.go &
HEALTH_PID=$!
sleep 2

# Test Health Monitor
if curl -s http://localhost:$HEALTH_PORT/health > /dev/null; then
    echo "✅ Health Monitor started successfully"
else
    echo "⚠️  Health Monitor failed to start, but continuing..."
fi

# Install frontend dependencies if needed
echo ""
echo "📦 Checking frontend dependencies..."
cd frontend
if [ ! -d "node_modules" ]; then
    echo "Installing React dependencies..."
    npm install
fi

# Start React frontend
echo ""
echo "🌐 Starting React Frontend..."
echo "📍 Opening browser at http://localhost:3000"
echo ""
echo "🎯 FULL STACK READY!"
echo "   - API: http://localhost:8085"
echo "   - Health: http://localhost:8081" 
echo "   - UI: http://localhost:3000"
echo ""
echo "Press Ctrl+C to stop all services"
echo ""

# Start React (this will block)
npm start

# This line should never be reached due to npm start blocking
wait