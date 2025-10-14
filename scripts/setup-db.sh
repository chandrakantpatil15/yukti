#!/bin/bash

# Setup PostgreSQL database for Yukti
echo "Setting up PostgreSQL database..."

# Check if PostgreSQL is installed
if ! command -v psql &> /dev/null; then
    echo "Installing PostgreSQL..."
    brew install postgresql
    brew services start postgresql
fi

# Create database and user
psql postgres -c "CREATE DATABASE yukti;" 2>/dev/null || echo "Database yukti already exists"
psql postgres -c "CREATE USER postgres WITH PASSWORD 'postgres';" 2>/dev/null || echo "User postgres already exists"
psql postgres -c "GRANT ALL PRIVILEGES ON DATABASE yukti TO postgres;" 2>/dev/null

echo "Database setup complete!"
echo "Connection: postgresql://localhost:5432/yukti"
echo "Username: postgres"
echo "Password: postgres"