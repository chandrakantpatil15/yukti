.PHONY: help build start stop restart clean test backend frontend all

help:
	@echo "Yukti Platform - Local Development Commands:"
	@echo ""
	@echo "  make all       - Build and start backend + frontend"
	@echo "  make build     - Build backend binary"
	@echo "  make backend   - Start backend (port 8081)"
	@echo "  make frontend  - Start frontend (port 3000)"
	@echo "  make stop      - Stop all services"
	@echo "  make restart   - Restart backend"
	@echo "  make clean     - Clean build artifacts"
	@echo "  make test      - Run tests"
	@echo ""

all: build
	@echo "✅ Build complete! Now start services:"
	@echo ""
	@echo "Terminal 1: make backend"
	@echo "Terminal 2: make frontend"
	@echo ""

build:
	@echo "🔨 Building backend..."
	@go build -o bin/api ./cmd/main.go
	@echo "✅ Backend built successfully!"

backend: build
	@echo "🚀 Starting backend on port 8081..."
	@./bin/api

frontend:
	@echo "🚀 Starting frontend on port 3000..."
	@cd frontend && npm start

stop:
	@echo "🛑 Stopping services..."
	@pkill -f "bin/api" 2>/dev/null || true
	@pkill -f "react-scripts" 2>/dev/null || true
	@echo "✅ Services stopped!"

restart: stop backend

clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf bin/api
	@cd frontend && rm -rf build node_modules/.cache 2>/dev/null || true
	@echo "✅ Cleanup complete!"

test:
	@echo "🧪 Running Go tests..."
	@go test ./... -v

# Docker commands (legacy)
docker-start:
	@echo "🐳 Starting Docker services..."
	@docker-compose up -d
	@echo "✅ Docker services started!"

docker-stop:
	@echo "🐳 Stopping Docker services..."
	@docker-compose down
	@echo "✅ Docker services stopped!"

docker-clean:
	@echo "🐳 Cleaning Docker..."
	@docker-compose down -v
	@echo "✅ Docker cleanup complete!"
