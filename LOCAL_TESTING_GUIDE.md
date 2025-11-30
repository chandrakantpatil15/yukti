# Local Testing Guide - Complete Walkthrough

**Goal**: Test all CRITICAL and HIGH PRIORITY changes locally before AWS deployment.

---

## Prerequisites

1. **Database Running**: PostgreSQL on localhost:5432
2. **Environment Variables Set**:
   ```bash
   export DATABASE_URL="postgres://yukti:yukti123@localhost:5432/yukti_finops?sslmode=disable"
   export JWT_SECRET="your-secret-key-min-32-chars"
   export ENVIRONMENT="development"
   export CORS_ALLOWED_ORIGINS="http://localhost:3000"
   ```

---

## Step 1: Apply Database Migrations

```bash
cd /Users/chandrakantpatil/workspace/yukti

# Apply new migrations
psql -U yukti -d yukti_finops -f scripts/014_create_hidden_cost_findings.sql
psql -U yukti -d yukti_finops -f scripts/015_add_performance_indexes.sql

# Verify tables exist
psql -U yukti -d yukti_finops -c "\dt yt_*"
```

**Expected Output**: Should see `yt_hidden_cost_findings` table listed.

---

## Step 2: Start Backend Server

```bash
cd /Users/chandrakantpatil/workspace/yukti

# Build and run
go build -o yukti-api cmd/main.go
./yukti-api
```

**Expected Output**:
```
[INFO] ========================================
[INFO] Yukti FinOps Platform Starting...
[INFO] ========================================
[INFO] Loading configuration...
[INFO] Configuration loaded successfully
[INFO] Connecting to database...
[INFO] Database connection established
[INFO] Initializing API server...
[INFO] Setting up API routes...
[INFO] Configuring CORS for origins: [http://localhost:3000]
[INFO] Starting HTTP server on :8080
```

**If you see errors about JWT_SECRET or CORS**, the config validation is working! ✅

---

## Step 3: Test Health Check

```bash
curl http://localhost:8080/health
```

**Expected**:
```json
{"status":"healthy"}
```

---

## Step 4: Test Authentication Flow

### 4.1 Signup (Create Test User)

```bash
curl -X POST http://localhost:8080/api/v1/auth/signup \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "testpass123",
    "company_name": "Test Company"
  }'
```

**Expected**:
```json
{
  "success": true,
  "user_id": "uuid-here",
  "tenant_id": 1,
  "message": "User created successfully"
}
```

**Save the tenant_id for later tests!**

### 4.2 Login (Get JWT Token)

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_code": "test-company-xxxxx",
    "email": "test@example.com",
    "password": "testpass123"
  }'
```

**Note**: Use the actual tenant_code from signup response or check database:
```bash
psql -U yukti -d yukti_finops -c "SELECT id, tenant_code FROM yt_tenants ORDER BY id DESC LIMIT 1;"
```

**Expected**:
```json
{
  "success": true,
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2024-...",
  "user": {
    "id": "uuid",
    "email": "test@example.com",
    "role": "admin",
    "tenant_id": 1
  }
}
```

**Save the token as an environment variable**:
```bash
export TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

---

## Step 5: Test NEW HIGH PRIORITY Endpoints

### 5.1 ML Forecast (Should return "no data" gracefully)

```bash
curl -X POST http://localhost:8080/api/v1/ml/forecast \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{}'
```

**Expected**:
```json
{
  "success": true,
  "message": "No forecast data available yet. Data will be available once ML service is fully integrated.",
  "data": []
}
```

✅ **PASS**: No 501 error, graceful response

---

### 5.2 Resource Details

First, insert a test resource:
```bash
psql -U yukti -d yukti_finops << EOF
INSERT INTO yt_aws_accounts (tenant_id, account_id, account_name, role_arn, external_id, status)
VALUES (1, '123456789012', 'Test Account', 'arn:aws:iam::123456789012:role/Test', 'ext-001', 'active')
ON CONFLICT DO NOTHING;

INSERT INTO yt_tenant_resources (tenant_id, aws_account_id, resource_id, resource_type, region, instance_type, state, tags, monthly_cost)
VALUES (1, 1, 'i-test123', 'ec2', 'us-east-1', 't3.medium', 'running', '{"Environment":"test"}', 45.50)
ON CONFLICT DO NOTHING;
EOF
```

Now test the endpoint:
```bash
curl "http://localhost:8080/api/v1/resources/details?resource_id=i-test123" \
  -H "Authorization: Bearer $TOKEN"
```

**Expected**:
```json
{
  "success": true,
  "data": {
    "id": 1,
    "resource_id": "i-test123",
    "resource_type": "ec2",
    "region": "us-east-1",
    "instance_type": "t3.medium",
    "state": "running",
    "tags": {"Environment": "test"},
    "monthly_cost": 45.5,
    "account_id": "123456789012",
    "metadata": {}
  }
}
```

✅ **PASS**: Resource details returned

---

### 5.3 Resource Metrics (Should return "no data" gracefully)

```bash
curl "http://localhost:8080/api/v1/resources/metrics?resource_id=i-test123" \
  -H "Authorization: Bearer $TOKEN"
```

**Expected**:
```json
{
  "success": true,
  "message": "No metrics data available yet. Metrics collection will be available once monitoring is fully integrated.",
  "data": []
}
```

✅ **PASS**: Graceful "no data" response

---

### 5.4 Resource Cost (Should return "no data" gracefully)

```bash
curl "http://localhost:8080/api/v1/resources/cost?resource_id=i-test123" \
  -H "Authorization: Bearer $TOKEN"
```

**Expected**:
```json
{
  "success": true,
  "message": "No cost history available yet. Cost tracking will be available once resource-level billing is implemented.",
  "data": []
}
```

✅ **PASS**: Graceful "no data" response

---

### 5.5 Trigger Scan

```bash
curl -X POST http://localhost:8080/api/v1/scan \
  -H "Authorization: Bearer $TOKEN"
```

**Expected**:
```json
{
  "success": true,
  "message": "Scan queued for processing. Results will be available shortly.",
  "tenant_id": 1
}
```

✅ **PASS**: Scan accepted

---

### 5.6 Admin Sync Endpoints

Set admin credentials:
```bash
export ADMIN_KEY="admin-key-123"
export ADMIN_USER="admin@yukti.io"
```

**Sync Pricing**:
```bash
curl -X POST http://localhost:8080/api/admin/sync/pricing \
  -H "X-Admin-Key: $ADMIN_KEY" \
  -H "X-Admin-User: $ADMIN_USER"
```

**Expected**:
```json
{
  "success": true,
  "message": "Pricing sync queued for processing"
}
```

**Sync Inventory**:
```bash
curl -X POST http://localhost:8080/api/admin/sync/inventory \
  -H "X-Admin-Key: $ADMIN_KEY" \
  -H "X-Admin-User: $ADMIN_USER" \
  -H "Content-Type: application/json" \
  -d '{"tenant_id": 1}'
```

**Expected**:
```json
{
  "success": true,
  "message": "Inventory sync queued for processing",
  "tenant_id": 1
}
```

✅ **PASS**: Admin endpoints working

---

## Step 6: Test CRITICAL Fixes

### 6.1 Test Pagination on Findings

Insert test findings:
```bash
psql -U yukti -d yukti_finops << EOF
INSERT INTO yt_hidden_cost_findings (id, tenant_id, detector_name, category, severity, title, description, resource_arn, estimated_savings, confidence)
VALUES 
  ('find-test-1', 'tenant-001', 'test_detector', 'Data Transfer', 'High', 'Test Finding 1', 'Description 1', 'arn:aws:ec2:us-east-1:123456789012:instance/i-1', 100.00, 0.95),
  ('find-test-2', 'tenant-001', 'test_detector', 'Storage', 'Medium', 'Test Finding 2', 'Description 2', 'arn:aws:s3:::bucket-1', 50.00, 0.85),
  ('find-test-3', 'tenant-001', 'test_detector', 'Compute', 'Low', 'Test Finding 3', 'Description 3', 'arn:aws:ec2:us-east-1:123456789012:instance/i-2', 25.00, 0.75)
ON CONFLICT DO NOTHING;
EOF
```

Test pagination:
```bash
# Page 1, 2 items per page
curl "http://localhost:8080/api/customers/findings?tenant_id=tenant-001&page=1&per_page=2"
```

**Expected**:
```json
{
  "success": true,
  "findings": [...2 items...],
  "meta": {
    "page": 1,
    "per_page": 2,
    "total": 3,
    "total_pages": 2
  }
}
```

✅ **PASS**: Pagination working

---

### 6.2 Test CORS Configuration

```bash
# Should work from localhost:3000
curl -H "Origin: http://localhost:3000" \
  -H "Access-Control-Request-Method: GET" \
  -H "Access-Control-Request-Headers: Authorization" \
  -X OPTIONS \
  http://localhost:8080/health -v
```

**Expected**: Should see `Access-Control-Allow-Origin: http://localhost:3000` in response headers.

✅ **PASS**: CORS configured correctly

---

### 6.3 Test JWT Secret Validation

Stop the server and try to start without JWT_SECRET in production:
```bash
unset JWT_SECRET
export ENVIRONMENT="production"
./yukti-api
```

**Expected**: Server should FAIL to start with error:
```
[FATAL] Configuration validation failed: JWT_SECRET must be set in production environment
```

✅ **PASS**: Production safety working

Reset for development:
```bash
export ENVIRONMENT="development"
export JWT_SECRET="test-secret-key"
```

---

## Step 7: Test Frontend

### 7.1 Start Frontend

```bash
cd frontend
npm install  # if not already done
npm start
```

**Expected**: Opens http://localhost:3000

---

### 7.2 Test Login Flow

1. Navigate to http://localhost:3000/login
2. Enter credentials:
   - Tenant Code: (from database query above)
   - Email: test@example.com
   - Password: testpass123
3. Click "Sign in"

**Expected**: Redirects to /dashboard

✅ **PASS**: Login working

---

### 7.3 Test Onboarding (Consolidated)

1. Navigate to http://localhost:3000/onboarding
2. Should see the Onboarding.tsx component (NOT SimpleOnboarding)
3. Verify no console errors

✅ **PASS**: Single onboarding flow

---

### 7.4 Test Dynamic Filters

1. Navigate to http://localhost:3000/hidden-costs
2. Filters should load from API
3. Check browser console for API calls to:
   - `/api/v1/filters/resource-types`
   - `/api/v1/filters/tags`
   - `/api/v1/filters/services`
   - `/api/v1/filters/regions`

**Expected**: All filter endpoints return data or empty arrays

✅ **PASS**: Filters working

---

## Step 8: Check Logs

Review server logs for any errors:
```bash
# Check for warnings
grep WARN server.log

# Check for errors
grep ERROR server.log

# Check for critical issues
grep FATAL server.log
```

**Expected**: Only development warnings (JWT_SECRET, CORS), no errors.

---

## Step 9: Database Verification

```bash
# Check all tables exist
psql -U yukti -d yukti_finops -c "\dt yt_*"

# Check indexes
psql -U yukti -d yukti_finops -c "\di yt_*"

# Check data
psql -U yukti -d yukti_finops << EOF
SELECT COUNT(*) as users FROM yt_users;
SELECT COUNT(*) as tenants FROM yt_tenants;
SELECT COUNT(*) as findings FROM yt_hidden_cost_findings;
SELECT COUNT(*) as resources FROM yt_tenant_resources;
EOF
```

**Expected**: All tables exist, indexes created, data present.

---

## ✅ Testing Checklist

Mark each as you test:

**Backend - CRITICAL**:
- [ ] JWT secret validation (production fails without it)
- [ ] CORS environment configuration
- [ ] Pagination on findings
- [ ] Pagination on admin customers
- [ ] Database migrations applied
- [ ] Performance indexes created

**Backend - HIGH PRIORITY**:
- [ ] ML forecast endpoint (graceful "no data")
- [ ] Resource details endpoint
- [ ] Resource metrics endpoint (graceful "no data")
- [ ] Resource cost endpoint (graceful "no data")
- [ ] Scan orchestration endpoint
- [ ] Admin sync pricing endpoint
- [ ] Admin sync inventory endpoint

**Frontend**:
- [ ] Login flow works
- [ ] Signup flow works
- [ ] Dashboard loads
- [ ] Onboarding page (single version)
- [ ] Dynamic filters load
- [ ] No console errors

**Integration**:
- [ ] JWT authentication end-to-end
- [ ] Tenant isolation working
- [ ] CORS allows frontend requests
- [ ] All API responses follow standard format

---

## 🐛 Troubleshooting

**Issue**: Server won't start
- Check DATABASE_URL is correct
- Check PostgreSQL is running: `psql -U yukti -d yukti_finops -c "SELECT 1;"`
- Check JWT_SECRET is set (development mode)

**Issue**: 401 Unauthorized on API calls
- Check JWT token is valid: Decode at jwt.io
- Check token hasn't expired (24 hours)
- Re-login to get fresh token

**Issue**: CORS errors in browser
- Check CORS_ALLOWED_ORIGINS includes http://localhost:3000
- Check browser console for exact error
- Restart backend after changing CORS config

**Issue**: Database connection failed
- Check PostgreSQL is running: `pg_isready`
- Check credentials: `psql -U yukti -d yukti_finops`
- Check DATABASE_URL format

**Issue**: Frontend can't reach backend
- Check backend is running on port 8080
- Check REACT_APP_API_URL in frontend/.env
- Check no firewall blocking localhost

---

## 📊 Success Criteria

All tests pass when:
- ✅ Backend starts without errors
- ✅ All 7 new endpoints return expected responses
- ✅ JWT authentication works end-to-end
- ✅ Frontend loads and connects to backend
- ✅ No console errors in browser
- ✅ Database has all tables and indexes
- ✅ Pagination works on findings and customers
- ✅ CORS allows frontend requests

---

## 🚀 Next Steps After Local Testing

Once all tests pass:
1. Document any issues found
2. Fix any bugs discovered
3. Prepare for AWS deployment
4. Set up production environment variables
5. Configure RDS, EKS, S3

---

**Ready to start testing? Run through each step and let me know if you hit any issues!**
