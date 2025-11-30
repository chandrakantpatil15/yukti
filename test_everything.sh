#!/bin/bash

echo "🚀 YUKTI AUTONOMOUS TESTING SCRIPT"
echo "=================================="
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Step 1: Rebuild containers
echo "📦 Step 1: Rebuilding containers..."
docker-compose down
docker-compose up -d --build

echo "⏳ Waiting 30 seconds for services to start..."
sleep 30

# Step 2: Check container status
echo ""
echo "🔍 Step 2: Checking container status..."
docker-compose ps

# Step 3: Test Backend APIs
echo ""
echo "🧪 Step 3: Testing Backend APIs..."
echo ""

# Test health endpoint
echo "Testing /health..."
HEALTH=$(curl -s http://localhost:8080/health)
if [[ $HEALTH == *"healthy"* ]]; then
    echo -e "${GREEN}✅ Health check passed${NC}"
else
    echo -e "${RED}❌ Health check failed${NC}"
fi

# Test admin endpoints WITHOUT auth (should fail)
echo ""
echo "Testing /api/admin/customers without auth (should fail)..."
ADMIN_NO_AUTH=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/api/admin/customers)
if [[ $ADMIN_NO_AUTH == "401" ]]; then
    echo -e "${GREEN}✅ Admin auth working (401 Unauthorized)${NC}"
else
    echo -e "${RED}❌ Admin auth NOT working (got $ADMIN_NO_AUTH, expected 401)${NC}"
fi

# Test admin endpoints WITH auth (should work)
echo ""
echo "Testing /api/admin/customers with auth..."
ADMIN_WITH_AUTH=$(curl -s -H "X-Admin-Key: yukti-admin-2024" http://localhost:8080/api/admin/customers)
if [[ $ADMIN_WITH_AUTH == *"success"* ]]; then
    echo -e "${GREEN}✅ Admin customers API working${NC}"
    echo "Response: $ADMIN_WITH_AUTH" | jq '.' 2>/dev/null || echo "$ADMIN_WITH_AUTH"
else
    echo -e "${RED}❌ Admin customers API failed${NC}"
    echo "Response: $ADMIN_WITH_AUTH"
fi

# Test admin metrics
echo ""
echo "Testing /api/admin/metrics..."
METRICS=$(curl -s -H "X-Admin-Key: yukti-admin-2024" http://localhost:8080/api/admin/metrics)
if [[ $METRICS == *"success"* ]]; then
    echo -e "${GREEN}✅ Admin metrics API working${NC}"
    echo "Response: $METRICS" | jq '.' 2>/dev/null || echo "$METRICS"
else
    echo -e "${RED}❌ Admin metrics API failed${NC}"
    echo "Response: $METRICS"
fi

# Test customer dashboard
echo ""
echo "Testing /api/customers/dashboard?tenant_id=tenant-001..."
DASHBOARD=$(curl -s "http://localhost:8080/api/customers/dashboard?tenant_id=tenant-001")
if [[ $DASHBOARD == *"success"* ]]; then
    echo -e "${GREEN}✅ Customer dashboard API working${NC}"
    echo "Response: $DASHBOARD" | jq '.' 2>/dev/null || echo "$DASHBOARD"
else
    echo -e "${RED}❌ Customer dashboard API failed${NC}"
    echo "Response: $DASHBOARD"
fi

# Test customer findings
echo ""
echo "Testing /api/customers/findings?tenant_id=tenant-001..."
FINDINGS=$(curl -s "http://localhost:8080/api/customers/findings?tenant_id=tenant-001")
if [[ $FINDINGS == *"success"* ]]; then
    echo -e "${GREEN}✅ Customer findings API working${NC}"
    FINDINGS_COUNT=$(echo "$FINDINGS" | jq '.findings | length' 2>/dev/null)
    echo "Found $FINDINGS_COUNT findings"
else
    echo -e "${RED}❌ Customer findings API failed${NC}"
    echo "Response: $FINDINGS"
fi

# Test tenant isolation (invalid tenant)
echo ""
echo "Testing tenant isolation with invalid tenant_id..."
INVALID_TENANT=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:8080/api/customers/dashboard?tenant_id=invalid-tenant")
if [[ $INVALID_TENANT == "403" ]]; then
    echo -e "${GREEN}✅ Tenant isolation working (403 Forbidden)${NC}"
else
    echo -e "${YELLOW}⚠️  Tenant isolation may not be working (got $INVALID_TENANT, expected 403)${NC}"
fi

# Test audit logs
echo ""
echo "Testing /api/admin/audit-logs..."
AUDIT=$(curl -s -H "X-Admin-Key: yukti-admin-2024" "http://localhost:8080/api/admin/audit-logs?limit=10")
if [[ $AUDIT == *"success"* ]]; then
    echo -e "${GREEN}✅ Audit logs API working${NC}"
    AUDIT_COUNT=$(echo "$AUDIT" | jq '.logs | length' 2>/dev/null)
    echo "Found $AUDIT_COUNT audit log entries"
else
    echo -e "${RED}❌ Audit logs API failed${NC}"
    echo "Response: $AUDIT"
fi

# Step 4: Test Frontend
echo ""
echo "🌐 Step 4: Testing Frontend..."
echo ""

FRONTEND=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:3000)
if [[ $FRONTEND == "200" ]]; then
    echo -e "${GREEN}✅ Frontend is accessible${NC}"
else
    echo -e "${RED}❌ Frontend not accessible (HTTP $FRONTEND)${NC}"
fi

# Step 5: Check logs for errors
echo ""
echo "📋 Step 5: Checking for errors in logs..."
echo ""

echo "Backend logs (last 20 lines):"
docker-compose logs --tail=20 backend

echo ""
echo "Frontend logs (last 20 lines):"
docker-compose logs --tail=20 frontend

# Summary
echo ""
echo "=================================="
echo "✅ TESTING COMPLETE"
echo "=================================="
echo ""
echo "Next steps:"
echo "1. Open http://localhost:3000 in browser"
echo "2. Click 'Admin' button (should show customer list)"
echo "3. Click 'Audit Logs' button (should show admin actions)"
echo "4. Click 'View' on a customer (should impersonate)"
echo "5. Verify tenant-specific data loads"
echo ""
echo "To view live logs: docker-compose logs -f"
echo "To stop: docker-compose down"
