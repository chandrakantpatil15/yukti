#!/bin/bash

echo "🔄 Hard restarting frontend and backend..."

# Stop containers
docker-compose stop backend frontend

# Remove containers
docker-compose rm -f backend frontend

# Rebuild and start
docker-compose up -d --build backend frontend

# Wait for services to start
sleep 5

# Check status
docker-compose ps backend frontend

echo "✅ Restart complete!"
echo "🌐 Frontend: http://localhost:3000"
echo "🔧 Backend: http://localhost:8081"
