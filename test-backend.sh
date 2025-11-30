#!/bin/bash

# Automated Backend Testing Script
# Tests all CRITICAL and HIGH PRIORITY endpoints

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
API_URL="${API_URL:-http://localhost:8080}"
ADMIN_KEY="${ADMIN_KEY:-admin-key-123}"
ADMIN_USER="${ADMIN_USER:-admin@yukti.io}"

# Counters
PASSED=0
FAILED=0

# Helper functions
print_test() {
    echo -e "\n${YELLOW}[TEST]${NC} $1"
}

print_pass() {
    echo -e "${GREEN}✓ PASS${NC} $1"
    ((PASSED++))
}

print_fail() {
    echo -e "${RED}✗ FAIL${NC} $1"
    ((FAILED++))
}

check_response() {
    local response=$1
    local expected=$2
    local test_name=$3
    
    if echo "$response" | grep -q "$expected"; then
        print_pass "$test_name"
        return 0
    else
        print_fail "$test_name"
        echo "Response: $response"
        return 1
    fi
}

# Start tests
echo "=========================================="
echo "Yukti Backend Automated Tests"
echo "=========================================="
echo "API URL: $API_URL"
echo ""

# Test 1: Health Check
print_test "Health Check"
response=$(curl -s "$API_URL/health")
check_response "$response" "healthy" "Health endpoint"

# Test 2: Signup
print_test "User Signup"
TIMESTAMP=$(date +%s)
TEST_EMAIL="test${TIMESTAMP}@example.com"
TEST_PASSWORD="testpass123"
TEST_COMPANY="Test Company ${TIMESTAMP}"

signup_response=$(curl -s -X POST "$API_URL/api/v1/auth/signup" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"$TEST_EMAIL\",
    \"password\": \"$TEST_PASSWORD\",
    \"company_name\": \"$TEST_COMPANY\"
  }")

if echo "$signup_response" | grep -q "success.*true"; then
    print_pass "User signup"
    TENANT_ID=$(echo "$signup_response" | grep -o '"tenant_id":[0-9]*' | grep -o '[0-9]*')
    echo "  Tenant ID: $TENANT_ID"
else
    print_fail "User signup"
    echo "Response: $signup_response"
    exit 1
fi

# Get tenant code from database
TENANT_CODE=$(psql -U yukti -d yukti_finops -t -c "SELECT tenant_code FROM yt_tenants WHERE id = $TENANT_ID;" 2>/dev/null | xargs)
if [ -z "$TENANT_CODE" ]; then
    echo "Warning: Could not get tenant_code from database, using fallback"
    TENANT_CODE="test-company"
fi

# Test 3: Login
print_test "User Login"
login_response=$(curl -s -X POST "$API_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d "{
    \"tenant_code\": \"$TENANT_CODE\",
    \"email\": \"$TEST_EMAIL\",
    \"password\": \"$TEST_PASSWORD\"
  }")

if echo "$login_response" | grep -q "token"; then
    print_pass "User login"
    TOKEN=$(echo "$login_response" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
    echo "  Token: ${TOKEN:0:50}..."
else
    print_fail "User login"
    echo "Response: $login_response"
    exit 1
fi

# Test 4: ML Forecast (should return graceful "no data")
print_test "ML Forecast Endpoint"
forecast_response=$(curl -s -X POST "$API_URL/api/v1/ml/forecast" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{}')
check_response "$forecast_response" "No forecast data available" "ML forecast graceful response"

# Test 5: Resource Details (need to create test resource first)
print_test "Resource Details Endpoint"

# Insert test resource
psql -U yukti -d yukti_finops > /dev/null 2>&1 << EOF
INSERT INTO yt_aws_accounts (tenant_id, account_id, account_name, role_arn, external_id, status)
VALUES ($TENANT_ID, '123456789012', 'Test Account', 'arn:aws:iam::123456789012:role/Test', 'ext-001', 'active')
ON CONFLICT DO NOTHING;

INSERT INTO yt_tenant_resources (tenant_id, aws_account_id, resource_id, resource_type, region, instance_type, state, tags, monthly_cost)
SELECT $TENANT_ID, a.id, 'i-test${TIMESTAMP}', 'ec2', 'us-east-1', 't3.medium', 'running', '{"Environment":"test"}'::jsonb, 45.50
FROM yt_aws_accounts a WHERE a.tenant_id = $TENANT_ID LIMIT 1
ON CONFLICT DO NOTHING;
EOF

resource_response=$(curl -s "$API_URL/api/v1/resources/details?resource_id=i-test${TIMESTAMP}" \
  -H "Authorization: Bearer $TOKEN")
check_response "$resource_response" "i-test${TIMESTAMP}" "Resource details"

# Test 6: Resource Metrics
print_test "Resource Metrics Endpoint"
metrics_response=$(curl -s "$API_URL/api/v1/resources/metrics?resource_id=i-test${TIMESTAMP}" \
  -H "Authorization: Bearer $TOKEN")
check_response "$metrics_response" "No metrics data available" "Resource metrics graceful response"

# Test 7: Resource Cost
print_test "Resource Cost Endpoint"
cost_response=$(curl -s "$API_URL/api/v1/resources/cost?resource_id=i-test${TIMESTAMP}" \
  -H "Authorization: Bearer $TOKEN")
check_response "$cost_response" "No cost history available" "Resource cost graceful response"

# Test 8: Trigger Scan
print_test "Scan Orchestration Endpoint"
scan_response=$(curl -s -X POST "$API_URL/api/v1/scan" \
  -H "Authorization: Bearer $TOKEN")
check_response "$scan_response" "Scan queued" "Scan orchestration"

# Test 9: Admin Sync Pricing
print_test "Admin Sync Pricing Endpoint"
sync_pricing_response=$(curl -s -X POST "$API_URL/api/admin/sync/pricing" \
  -H "X-Admin-Key: $ADMIN_KEY" \
  -H "X-Admin-User: $ADMIN_USER")
check_response "$sync_pricing_response" "Pricing sync queued" "Admin sync pricing"

# Test 10: Admin Sync Inventory
print_test "Admin Sync Inventory Endpoint"
sync_inventory_response=$(curl -s -X POST "$API_URL/api/admin/sync/inventory" \
  -H "X-Admin-Key: $ADMIN_KEY" \
  -H "X-Admin-User: $ADMIN_USER" \
  -H "Content-Type: application/json" \
  -d "{\"tenant_id\": $TENANT_ID}")
check_response "$sync_inventory_response" "Inventory sync queued" "Admin sync inventory"

# Test 11: Pagination on Findings
print_test "Findings Pagination"

# Insert test findings
psql -U yukti -d yukti_finops > /dev/null 2>&1 << EOF
INSERT INTO yt_hidden_cost_findings (id, tenant_id, detector_name, category, severity, title, description, resource_arn, estimated_savings, confidence)
VALUES 
  ('find-test-${TIMESTAMP}-1', 'tenant-$(printf "%03d" $TENANT_ID)', 'test_detector', 'Data Transfer', 'High', 'Test Finding 1', 'Description 1', 'arn:aws:ec2:us-east-1:123456789012:instance/i-1', 100.00, 0.95),
  ('find-test-${TIMESTAMP}-2', 'tenant-$(printf "%03d" $TENANT_ID)', 'test_detector', 'Storage', 'Medium', 'Test Finding 2', 'Description 2', 'arn:aws:s3:::bucket-1', 50.00, 0.85),
  ('find-test-${TIMESTAMP}-3', 'tenant-$(printf "%03d" $TENANT_ID)', 'test_detector', 'Compute', 'Low', 'Test Finding 3', 'Description 3', 'arn:aws:ec2:us-east-1:123456789012:instance/i-2', 25.00, 0.75)
ON CONFLICT DO NOTHING;
EOF

findings_response=$(curl -s "$API_URL/api/customers/findings?tenant_id=tenant-$(printf "%03d" $TENANT_ID)&page=1&per_page=2")
if echo "$findings_response" | grep -q '"total_pages"'; then
    print_pass "Findings pagination"
else
    print_fail "Findings pagination"
    echo "Response: $findings_response"
fi

# Test 12: Admin Customers Pagination
print_test "Admin Customers Pagination"
customers_response=$(curl -s "$API_URL/api/admin/customers?page=1&per_page=10" \
  -H "X-Admin-Key: $ADMIN_KEY" \
  -H "X-Admin-User: $ADMIN_USER")
check_response "$customers_response" "total_pages" "Admin customers pagination"

# Test 13: Dynamic Filters - Resource Types
print_test "Dynamic Filters - Resource Types"
filters_response=$(curl -s "$API_URL/api/v1/filters/resource-types" \
  -H "Authorization: Bearer $TOKEN")
check_response "$filters_response" "success" "Resource types filter"

# Test 14: Dynamic Filters - Services
print_test "Dynamic Filters - Services"
services_response=$(curl -s "$API_URL/api/v1/filters/services" \
  -H "Authorization: Bearer $TOKEN")
check_response "$services_response" "success" "Services filter"

# Test 15: Logout
print_test "User Logout"
logout_response=$(curl -s -X POST "$API_URL/api/v1/auth/logout")
check_response "$logout_response" "success" "User logout"

# Summary
echo ""
echo "=========================================="
echo "Test Summary"
echo "=========================================="
echo -e "${GREEN}Passed: $PASSED${NC}"
echo -e "${RED}Failed: $FAILED${NC}"
echo "Total: $((PASSED + FAILED))"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}✗ Some tests failed${NC}"
    exit 1
fi
