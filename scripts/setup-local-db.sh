#!/bin/bash
set -e

echo "🔧 Setting up Yukti database on local PostgreSQL..."

# Database configuration
DB_NAME="yukti_db"
DB_USER="yukti_user"
DB_PASSWORD="yukti_password"
POSTGRES_USER="chandrakantpatil"

echo "📊 Creating database and user..."

# Create database and user
psql -h localhost -p 5432 -U $POSTGRES_USER -d postgres << EOF
-- Create user if not exists
DO \$\$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = '$DB_USER') THEN
        CREATE USER $DB_USER WITH PASSWORD '$DB_PASSWORD';
    END IF;
END
\$\$;

-- Create database if not exists
SELECT 'CREATE DATABASE $DB_NAME OWNER $DB_USER'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = '$DB_NAME')\gexec

-- Grant privileges
GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;
EOF

echo "📋 Creating schema and loading data..."

# Run schema creation (local version without partitioning)
psql -h localhost -p 5432 -U $DB_USER -d $DB_NAME -f src/main/resources/schema-local.sql

# Load realistic FinOps data
psql -h localhost -p 5432 -U $DB_USER -d $DB_NAME -f scripts/realistic-finops-data.sql

echo "✅ Database setup complete!"
echo "🔗 Connection details:"
echo "   Host: localhost"
echo "   Port: 5432"
echo "   Database: $DB_NAME"
echo "   User: $DB_USER"
echo "   Password: $DB_PASSWORD"

# Test connection
echo "🧪 Testing connection..."
psql -h localhost -p 5432 -U $DB_USER -d $DB_NAME -c "SELECT COUNT(*) as tables FROM information_schema.tables WHERE table_schema = 'public';"

echo "🎯 Ready to run: make local"