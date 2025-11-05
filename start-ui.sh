#!/bin/bash

echo "🚀 Starting Yukti FinOps UI"
echo "=========================="

# Check if Node.js is installed
if ! command -v node &> /dev/null; then
    echo "❌ Node.js is not installed. Please install Node.js first."
    exit 1
fi

# Check if npm is installed
if ! command -v npm &> /dev/null; then
    echo "❌ npm is not installed. Please install npm first."
    exit 1
fi

cd frontend

# Install dependencies if node_modules doesn't exist
if [ ! -d "node_modules" ]; then
    echo "📦 Installing dependencies..."
    npm install
fi

# Start the development server
echo "🌐 Starting React development server..."
echo "📍 UI will be available at: http://localhost:3000"
echo "🔗 API should be running at: http://localhost:8085"
echo ""
echo "Press Ctrl+C to stop the server"

npm start