# Yukti FinOps - Infrastructure Makefile
# Usage: make <target>

.PHONY: help setup db-setup db-migrate sync-data start-api start-ui start-health start-all stop-all clean test status

# Configuration
DB_URL := postgres://chandrakantpatil@localhost:5432/yukti?sslmode=disable
API_PORT := 8085
HEALTH_PORT := 8081
UI_PORT := 3000

# Default target
help:
	@echo "🚀 Yukti FinOps - Infrastructure Commands"
	@echo "========================================"
	@echo ""
	@echo "📋 Setup Commands:"
	@echo "  make setup      - Complete setup (DB + dependencies)"
	@echo "  make db-setup   - Setup PostgreSQL database"
	@echo "  make db-migrate - Run database migrations"
	@echo ""
	@echo "📊 Data Commands:"
	@echo "  make sync-data  - Sync AWS data (pricing + resources)"
	@echo "  make test-data  - Test data integrity"
	@echo ""
	@echo "🚀 Service Commands:"
	@echo "  make start-all  - Start full stack (API + UI + Health)"
	@echo "  make start-api  - Start API server only"
	@echo "  make start-ui   - Start React UI only"
	@echo "  make start-health - Start health monitor only"
	@echo ""
	@echo "🛑 Control Commands:"
	@echo "  make stop-all   - Stop all services"
	@echo "  make status     - Check service status"
	@echo "  make clean      - Clean build artifacts"
	@echo ""
	@echo "🧪 Testing Commands:"
	@echo "  make test       - Run integration tests"
	@echo "  make test-api   - Test API endpoints"

# Complete setup
setup: db-setup db-migrate
	@echo "✅ Complete setup finished"

# Database setup
db-setup:
	@echo "🗄️  Setting up PostgreSQL database..."
	@createdb yukti 2>/dev/null || echo "Database already exists"
	@echo "✅ Database setup complete"

# Database migrations
db-migrate:
	@echo "📊 Running database migrations..."
	@for script in scripts/*.sql; do \
		echo "Running $$script..."; \
		psql $(DB_URL) -f $$script; \
	done
	@echo "✅ Database migrations complete"

# Sync AWS data
sync-data:
	@echo "📊 Syncing AWS data..."
	@DATABASE_URL=$(DB_URL) go run cmd/sync-all-aws-data.go
	@echo "✅ AWS data sync complete"

# Test data integrity
test-data:
	@echo "🧪 Testing data integrity..."
	@DATABASE_URL=$(DB_URL) go test -v tests/integration_test.go

# Start API server
start-api:
	@echo "🚀 Starting API server on port $(API_PORT)..."
	@DATABASE_URL=$(DB_URL) PORT=$(API_PORT) go run cmd/api-server.go &
	@echo $$! > .api.pid
	@sleep 2
	@curl -s http://localhost:$(API_PORT)/health > /dev/null && echo "✅ API server started" || echo "❌ API server failed"

# Start React UI
start-ui:
	@echo "🌐 Starting React UI on port $(UI_PORT)..."
	@cd frontend && npm install --silent 2>/dev/null || true
	@cd frontend && npm start &
	@echo $$! > .ui.pid
	@echo "✅ React UI starting (will open browser)"

# Start health monitor
start-health:
	@echo "🏥 Starting health monitor on port $(HEALTH_PORT)..."
	@DATABASE_URL=$(DB_URL) HEALTH_PORT=$(HEALTH_PORT) go run cmd/health-monitor.go &
	@echo $$! > .health.pid
	@sleep 2
	@curl -s http://localhost:$(HEALTH_PORT)/health > /dev/null && echo "✅ Health monitor started" || echo "❌ Health monitor failed"

# Start all services
start-all: sync-data start-api start-health start-ui
	@echo ""
	@echo "🎯 YUKTI FINOPS - FULL STACK READY!"
	@echo "=================================="
	@echo "📱 React UI:      http://localhost:$(UI_PORT)"
	@echo "🔧 API Server:    http://localhost:$(API_PORT)"
	@echo "🏥 Health Monitor: http://localhost:$(HEALTH_PORT)"
	@echo ""
	@echo "Press Ctrl+C or run 'make stop-all' to stop services"

# Stop all services
stop-all:
	@echo "🛑 Stopping all services..."
	@-pkill -f "api-server" 2>/dev/null || true
	@-pkill -f "health-monitor" 2>/dev/null || true
	@-pkill -f "react-scripts" 2>/dev/null || true
	@-pkill -f "node.*start" 2>/dev/null || true
	@-rm -f .api.pid .ui.pid .health.pid 2>/dev/null || true
	@echo "✅ All services stopped"

# Check service status
status:
	@echo "📊 Service Status:"
	@echo "=================="
	@curl -s http://localhost:$(API_PORT)/health > /dev/null && echo "✅ API Server: Running" || echo "❌ API Server: Stopped"
	@curl -s http://localhost:$(HEALTH_PORT)/health > /dev/null && echo "✅ Health Monitor: Running" || echo "❌ Health Monitor: Stopped"
	@curl -s http://localhost:$(UI_PORT) > /dev/null && echo "✅ React UI: Running" || echo "❌ React UI: Stopped"
	@psql $(DB_URL) -c "SELECT 1;" > /dev/null 2>&1 && echo "✅ Database: Connected" || echo "❌ Database: Disconnected"

# Test API endpoints
test-api:
	@echo "🧪 Testing API endpoints..."
	@curl -s http://localhost:$(API_PORT)/health | jq . || echo "❌ Health endpoint failed"
	@curl -s http://localhost:$(API_PORT)/api/v1/resources | jq '.data | length' || echo "❌ Resources endpoint failed"

# Run integration tests
test: test-data test-api
	@echo "✅ All tests completed"

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf frontend/node_modules/.cache 2>/dev/null || true
	@rm -rf frontend/build 2>/dev/null || true
	@rm -f .api.pid .ui.pid .health.pid 2>/dev/null || true
	@go clean -cache
	@echo "✅ Clean complete"

# Development shortcuts
dev: start-all
local: start-all
full-stack: start-all