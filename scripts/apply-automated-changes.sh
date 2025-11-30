#!/bin/bash

echo "🔄 Starting Automated Changes"
echo "==========================="

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

# Track changes
CHANGES_LOG="automated_changes_$(date +%Y%m%d_%H%M%S).log"
echo "Changes made on $(date)" > $CHANGES_LOG

# Function to log changes
log_change() {
    echo "$(date +"%Y-%m-%d %H:%M:%S"): $1" >> $CHANGES_LOG
    echo -e "${GREEN}✓${NC} $1"
}

# 1. Update port configurations
echo "🔧 Updating port configurations..."

# Update docker-compose.yml
sed -i'.bak' 's/8090:8090/8080:8080/g' docker-compose.yml
sed -i'.bak' 's/8091:8091/8000:8000/g' docker-compose.yml
log_change "Updated ports in docker-compose.yml"

# Update prometheus.yml
sed -i'.bak' 's/:8090/:8080/g' prometheus.yml
log_change "Updated ports in prometheus.yml"

# 2. Apply infrastructure changes
echo "🏗️  Applying infrastructure changes..."

# Rebuild and restart services
make clean
log_change "Cleaned existing containers"

make start
log_change "Started services with new configuration"

# 3. Run tests
echo "🧪 Running tests..."
./test_everything.sh
if [ $? -eq 0 ]; then
    log_change "All tests passed successfully"
else
    echo -e "${RED}❌ Tests failed${NC}"
    echo "Tests failed - manual intervention required" >> $CHANGES_LOG
    exit 1
fi

echo -e "\n${GREEN}✅ All changes applied successfully!${NC}"
echo "Changes log saved to: $CHANGES_LOG"