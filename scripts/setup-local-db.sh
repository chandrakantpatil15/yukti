#!/bin/bash

echo "🚀 Setting up Yukti Local Database Environment..."

# Detect OS
OS="$(uname -s)"
case "${OS}" in
    Linux*)     MACHINE=Linux;;
    Darwin*)    MACHINE=Mac;;
    CYGWIN*)    MACHINE=Cygwin;;
    MINGW*)     MACHINE=MinGw;;
    *)          MACHINE="UNKNOWN:${OS}"
esac

echo "📱 Detected OS: $MACHINE"

# Install PostgreSQL and Redis based on OS
install_databases() {
    if [ "$MACHINE" = "Mac" ]; then
        echo "🍺 Installing via Homebrew..."
        if ! command -v brew &> /dev/null; then
            echo "❌ Homebrew not found. Please install Homebrew first."
            exit 1
        fi
        brew install postgresql redis
        brew services start postgresql
        brew services start redis
    elif [ "$MACHINE" = "Linux" ]; then
        echo "🐧 Installing via apt..."
        sudo apt update
        sudo apt install -y postgresql postgresql-contrib redis-server
        sudo systemctl start postgresql
        sudo systemctl start redis-server
        sudo systemctl enable postgresql
        sudo systemctl enable redis-server
    else
        echo "❌ Unsupported OS. Please install PostgreSQL and Redis manually."
        exit 1
    fi
}

# Check if PostgreSQL is installed
if ! command -v psql &> /dev/null; then
    echo "📦 PostgreSQL not found. Installing..."
    install_databases
else
    echo "✅ PostgreSQL already installed"
fi

# Check if Redis is installed
if ! command -v redis-cli &> /dev/null; then
    echo "📦 Redis not found. Installing..."
    if [ "$MACHINE" = "Mac" ]; then
        brew install redis
        brew services start redis
    elif [ "$MACHINE" = "Linux" ]; then
        sudo apt install -y redis-server
        sudo systemctl start redis-server
    fi
else
    echo "✅ Redis already installed"
fi

# Create PostgreSQL database and user
echo "🗄️ Setting up PostgreSQL database..."
if [ "$MACHINE" = "Mac" ]; then
    createdb yukti 2>/dev/null || echo "Database yukti already exists"
    psql -d yukti -c "CREATE USER yukti_user WITH PASSWORD 'yukti_pass';" 2>/dev/null || echo "User already exists"
    psql -d yukti -c "GRANT ALL PRIVILEGES ON DATABASE yukti TO yukti_user;" 2>/dev/null
elif [ "$MACHINE" = "Linux" ]; then
    sudo -u postgres createdb yukti 2>/dev/null || echo "Database yukti already exists"
    sudo -u postgres psql -c "CREATE USER yukti_user WITH PASSWORD 'yukti_pass';" 2>/dev/null || echo "User already exists"
    sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE yukti TO yukti_user;" 2>/dev/null
fi

# Test connections
echo "🔍 Testing database connections..."
if psql -h localhost -U yukti_user -d yukti -c "SELECT 1;" &> /dev/null; then
    echo "✅ PostgreSQL connection successful"
else
    echo "❌ PostgreSQL connection failed"
fi

if redis-cli ping | grep -q PONG; then
    echo "✅ Redis connection successful"
else
    echo "❌ Redis connection failed"
fi

# Run schema creation
echo "📋 Creating database schema..."
psql -h localhost -U yukti_user -d yukti -f "$(dirname "$0")/schema.sql"

# Insert sample data
echo "📊 Inserting sample data..."
psql -h localhost -U yukti_user -d yukti -f "$(dirname "$0")/sample-data.sql"

echo "🎉 Local database setup complete!"
echo ""
echo "📝 Connection Details:"
echo "PostgreSQL: jdbc:postgresql://localhost:5432/yukti"
echo "Username: yukti_user"
echo "Password: yukti_pass"
echo "Redis: localhost:6379"
echo ""
echo "🚀 You can now start the Yukti application!"