#!/bin/bash

echo "=== Week 7: API Gateway Testing ==="
echo ""

# Start API Gateway
export DATABASE_URL="postgres://chandrakantpatil@localhost:5432/yukti?sslmode=disable"
go run cmd/week7-api-gateway.go > /tmp/week7-api.log 2>&1 &
API_PID=$!
echo "🚀 Starting API Gateway (PID: $API_PID)..."
sleep 4

# Test endpoints
echo ""
echo "📋 Test 1: Health Check"
echo "─────────────────────────"
curl -s http://localhost:8080/health | jq .

echo ""
echo "📋 Test 2: List Resources (with API key)"
echo "────────────────────────────────────────"
curl -s -H "X-API-Key: democorp-test_demo-api-key-12345" \
  http://localhost:8080/api/v1/resources | jq '.success, .meta'

echo ""
echo "📋 Test 3: Resource Stats"
echo "─────────────────────────"
curl -s -H "X-API-Key: democorp-test_demo-api-key-12345" \
  http://localhost:8080/api/v1/resources/stats | jq '.data'

echo ""
echo "📋 Test 4: Unauthorized Access"
echo "──────────────────────────────"
curl -s http://localhost:8080/api/v1/resources | jq .

echo ""
echo "✅ Week 7 API Gateway Tests Complete!"

# Cleanup
kill $API_PID 2>/dev/null
rm -f /tmp/week7-api.log
