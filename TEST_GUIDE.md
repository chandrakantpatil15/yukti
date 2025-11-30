# Yukti FinOps - Complete Test Guide

**Status**: Ready for Testing  
**Last Updated**: 2025-11-12  
**Services**: All Running ✅

---

## 🚀 Quick Start

### 1. Check Services
```bash
docker-compose ps
```

**Expected**: All 5 services running
- ✅ yukti-backend (port 8081)
- ✅ yukti-frontend (port 3000)
- ✅ yukti-ml (port 8000)
- ✅ yukti-prometheus (port 9090)
- ✅ yukti-grafana (port 3001)

### 2. Access Application
```
Frontend: http://localhost:3000
Backend API: http://localhost:8081
ML Service: http://localhost:8000
Prometheus: http://localhost:9090
Grafana: http://localhost:3001
```

---

## 🧪 Test Scenarios

### Scenario 1: New User Signup & Login

#### Step 1: Signup
1. Go to http://localhost:3000/signup
2. Enter:
   - Email: `test@example.com`
   - Password: `Test123!@#`
   - Company Name: `Test Company`
3. Click "Sign Up"
4. **Expected**: Verification code screen

#### Step 2: Verify Email
1. Check backend logs for OTP code:
   ```bash
   docker-compose logs backend | grep "OTP code"
   ```
2. Enter the 6-digit code
3. Click "Verify"
4. **Expected**: Redirect to onboarding

#### Step 3: Onboarding
1. **Welcome Screen**: Click "Get Started"
2. **AWS Connection**: 
   - See auto-generated External ID
   - See CloudFormation template
   - Click "Skip AWS Connection (Mock)" (yellow button)
3. **Expected**: Redirect to dashboard

---

### Scenario 2: Existing User Login

#### Test Account
```
Email: yourname123@example.com
Password: Chandra!@#$143
Tenant ID: 18
```

#### Step 1: Login
1. Go to http://localhost:3000/login
2. Enter credentials above
3. Click "Sign In"
4. **Expected**: Redirect to dashboard

#### Step 2: View Dashboard
**URL**: http://localhost:3000/dashboard

**Expected Data**:
- Total Savings: **$425.60/month**
- Findings Count: **7**
- Budget Usage: **0%** (no budget set)
- RI Savings: **$0** (no RI recommendations)

**Visual Check**:
- ✅ 4 metric cards displayed
- ✅ Green color for savings
- ✅ Blue color for findings
- ✅ Orange color for budget
- ✅ Purple color for RI savings

#### Step 3: View Hidden Costs
**URL**: http://localhost:3000/hidden-costs

**Expected Data**:
- **7 findings** displayed
- **Total savings**: $425.60/month
- **Categories**: Storage, Compute, Networking, Database
- **Severities**: High (3), Medium (2), Low (2)

**Test Filters**:
1. Filter by Severity: "High"
   - **Expected**: 3 findings (Unused EBS, Idle NAT Gateway, Oversized RDS)
2. Filter by Category: "Storage"
   - **Expected**: 3 findings (Unused EBS, Unoptimized S3, Old Snapshot)
3. Clear filters
   - **Expected**: Back to 7 findings

**Test Detail Panel**:
1. Click on any finding
2. **Expected**: Modal opens with:
   - Full description
   - Estimated savings
   - Resource ARN
   - Confidence meter
   - "Generate IaC" button
   - "Whitelist" button
3. Click X to close
4. **Expected**: Modal closes

---

### Scenario 3: API Testing

#### Test 1: Health Check
```bash
curl http://localhost:8081/health
```
**Expected**:
```json
{"status":"healthy"}
```

#### Test 2: Login API
```bash
curl -X POST http://localhost:8081/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"yourname123@example.com","password":"Chandra!@#$143"}'
```
**Expected**:
```json
{
  "success": true,
  "token": "eyJhbGc...",
  "refresh_token": "...",
  "user": {
    "id": "...",
    "email": "yourname123@example.com",
    "tenant_id": 18,
    "role": "viewer"
  }
}
```

#### Test 3: Dashboard API
```bash
# Replace TOKEN with actual token from login
curl http://localhost:8081/api/customers/dashboard?tenant_id=18 \
  -H "Authorization: Bearer TOKEN"
```
**Expected**:
```json
{
  "success": true,
  "data": {
    "total_savings": 425.6,
    "findings_count": 7,
    "budget_amount": 0,
    "current_spend": 0,
    "ri_savings": 0
  }
}
```

#### Test 4: Findings API
```bash
curl "http://localhost:8081/api/customers/findings?tenant_id=18" \
  -H "Authorization: Bearer TOKEN"
```
**Expected**:
```json
{
  "success": true,
  "findings": [
    {
      "id": "finding-001",
      "title": "Unused EBS Volume",
      "severity": "High",
      "estimated_savings": 50.00,
      ...
    },
    ...
  ],
  "meta": {
    "page": 1,
    "per_page": 50,
    "total": 7,
    "total_pages": 1
  }
}
```

#### Test 5: ML Service Health
```bash
curl http://localhost:8000/health
```
**Expected**:
```json
{
  "status": "healthy",
  "service": "ml-enhanced-detection",
  "version": "2.0",
  "models_trained": {
    "transfer_predictor": false,
    "anomaly_detector": false,
    "confidence_estimator": false,
    "workload_classifier": false
  }
}
```

---

### Scenario 4: Filter Testing

#### Test Severity Filters
1. Go to Hidden Costs page
2. Select "High" from severity dropdown
3. **Expected**: 3 findings
   - Unused EBS Volume ($50)
   - Idle NAT Gateway ($45)
   - Oversized RDS Instance ($200)
4. Select "Medium" from severity dropdown
5. **Expected**: 2 findings
   - Underutilized EC2 Instance ($100)
   - Unattached Elastic IP ($3.60)
6. Select "Low" from severity dropdown
7. **Expected**: 2 findings
   - Unoptimized S3 Storage Class ($15)
   - Old EBS Snapshot ($12)

#### Test Category Filters
1. Select "Storage" from category dropdown
2. **Expected**: 3 findings
3. Select "Compute" from category dropdown
4. **Expected**: 1 finding
5. Select "Networking" from category dropdown
6. **Expected**: 2 findings
7. Select "Database" from category dropdown
8. **Expected**: 1 finding

#### Test Combined Filters
1. Select "High" severity + "Storage" category
2. **Expected**: 1 finding (Unused EBS Volume)
3. Click "Clear Filters"
4. **Expected**: Back to 7 findings

---

### Scenario 5: Navigation Testing

#### Test All Pages
1. **Dashboard** (`/dashboard`)
   - ✅ Shows metrics
   - ✅ No navigation menu (if coming from onboarding)
   - ✅ Shows navigation menu (if logged in directly)

2. **Hidden Costs** (`/hidden-costs`)
   - ✅ Shows 7 findings
   - ✅ Filters work
   - ✅ Detail panel works

3. **Resources** (`/resources`)
   - ⚠️ Shows empty state (no resources in DB)
   - ✅ Page loads without errors

4. **Whitelists** (`/whitelists`)
   - ⚠️ Shows empty state (no whitelists in DB)
   - ✅ Page loads without errors

5. **Admin** (`/admin`)
   - ❌ 403 Forbidden (user is not admin)
   - ✅ Correct behavior

---

### Scenario 6: Error Handling

#### Test 1: Invalid Login
1. Go to login page
2. Enter wrong password
3. **Expected**: Error message displayed

#### Test 2: Unauthenticated Access
1. Clear localStorage (F12 → Application → Clear)
2. Try to access `/dashboard`
3. **Expected**: Redirect to `/login`

#### Test 3: Invalid Tenant ID
```bash
curl "http://localhost:8081/api/customers/dashboard?tenant_id=999"
```
**Expected**: Empty data or error

---

## 🐛 Troubleshooting

### Issue: Dashboard shows $0
**Solution**:
1. Check if logged in as tenant 18
2. Verify findings exist:
   ```bash
   psql -h localhost -U yukti -d yukti_finops -c "SELECT COUNT(*) FROM yt_hidden_cost_findings WHERE tenant_id = '18';"
   ```
3. Should return 7

### Issue: Hidden Costs shows empty
**Solution**:
1. Check browser console for errors
2. Verify API call:
   ```bash
   curl "http://localhost:8081/api/customers/findings?tenant_id=18"
   ```
3. Check backend logs:
   ```bash
   docker-compose logs backend | tail -50
   ```

### Issue: Login fails
**Solution**:
1. Check backend logs:
   ```bash
   docker-compose logs backend | grep ERROR
   ```
2. Verify database connection:
   ```bash
   psql -h localhost -U yukti -d yukti_finops -c "SELECT email FROM yt_users WHERE email = 'yourname123@example.com';"
   ```

### Issue: Frontend not loading
**Solution**:
1. Check frontend logs:
   ```bash
   docker-compose logs frontend
   ```
2. Rebuild frontend:
   ```bash
   docker-compose stop frontend && docker-compose rm -f frontend && docker-compose build frontend && docker-compose up -d frontend
   ```

---

## ✅ Success Criteria

### Must Pass
- ✅ Login works
- ✅ Dashboard shows $425.60 savings
- ✅ Hidden Costs shows 7 findings
- ✅ Filters work on Hidden Costs
- ✅ Detail panel opens/closes
- ✅ Navigation between pages works

### Should Pass
- ✅ Signup creates new user
- ✅ Email verification works
- ✅ Onboarding flow completes
- ✅ API endpoints return correct data
- ✅ Error handling works

### Nice to Have
- ⚠️ Resources page shows data (needs AWS integration)
- ⚠️ Budget cards show data (needs budget seeding)
- ⚠️ RI savings show data (needs RI recommendations)
- ⚠️ Generate IaC button works (needs implementation)
- ⚠️ Whitelist button works (needs implementation)

---

## 📊 Test Results Template

```
Date: ___________
Tester: ___________

Scenario 1: New User Signup
- Signup: [ ] Pass [ ] Fail
- Verification: [ ] Pass [ ] Fail
- Onboarding: [ ] Pass [ ] Fail

Scenario 2: Existing User Login
- Login: [ ] Pass [ ] Fail
- Dashboard: [ ] Pass [ ] Fail
- Hidden Costs: [ ] Pass [ ] Fail

Scenario 3: API Testing
- Health Check: [ ] Pass [ ] Fail
- Login API: [ ] Pass [ ] Fail
- Dashboard API: [ ] Pass [ ] Fail
- Findings API: [ ] Pass [ ] Fail

Scenario 4: Filter Testing
- Severity Filters: [ ] Pass [ ] Fail
- Category Filters: [ ] Pass [ ] Fail
- Combined Filters: [ ] Pass [ ] Fail

Scenario 5: Navigation
- All Pages Load: [ ] Pass [ ] Fail
- Navigation Works: [ ] Pass [ ] Fail

Scenario 6: Error Handling
- Invalid Login: [ ] Pass [ ] Fail
- Unauthenticated Access: [ ] Pass [ ] Fail

Overall Status: [ ] Pass [ ] Fail
Notes: ___________________________________________
```

---

**Ready to test!** Start with Scenario 2 (existing user login) for quickest results.
