# Business Logic Gaps - UI vs Backend Disconnect

**Critical Issue**: Frontend pages exist but don't connect to real backend data

---

## 🔴 Dashboard Page Issues

### Current State
- **File**: `frontend/src/pages/Dashboard.tsx`
- **API Call**: `/api/customers/dashboard?tenant_id=${tenantId}`
- **Backend Route**: ✅ EXISTS in `routes.go`
- **Handler**: `customerHandler.GetDashboard`

### Problems
1. **Wrong tenant_id source**: Uses `localStorage.getItem('tenant_id')` instead of user object
2. **No real data**: Database has no aggregated dashboard data
3. **Handler returns mock data**: Need to calculate real metrics

### Fix Required
```typescript
// Frontend: Get tenant_id from user object
const user = JSON.parse(localStorage.getItem('yukti_user') || '{}');
const data = await api.getDashboard(); // Use centralized API
```

```go
// Backend: Calculate real metrics from database
func (h *CustomerHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
    tenantID := r.Header.Get("X-Tenant-ID")
    
    // Query real data
    totalSavings := calculateTotalSavings(tenantID)
    findingsCount := countFindings(tenantID)
    budgetData := getBudgetStatus(tenantID)
    riSavings := calculateRISavings(tenantID)
    
    // Return real data
}
```

---

## 🔴 Hidden Costs Page Issues

### Current State
- **File**: `frontend/src/pages/HiddenCosts.tsx`
- **API Call**: `/api/customers/findings?tenant_id=${tenantId}`
- **Backend Route**: ✅ EXISTS
- **Handler**: `customerHandler.GetFindings`

### Problems
1. **No findings in database**: `yt_hidden_cost_findings` table is empty
2. **77 detectors not triggered**: Detectors exist but never run
3. **Filters don't work**: Backend doesn't filter by severity/category

### Fix Required
1. **Seed sample findings**:
```sql
INSERT INTO yt_hidden_cost_findings (tenant_id, detector_name, resource_arn, severity, category, estimated_monthly_cost, estimated_savings)
VALUES 
('18', 'unused-ebs-volumes', 'arn:aws:ec2:us-east-1:123456789012:volume/vol-abc123', 'high', 'storage', 50.00, 50.00),
('18', 'underutilized-ec2', 'arn:aws:ec2:us-east-1:123456789012:instance/i-xyz789', 'medium', 'compute', 200.00, 100.00);
```

2. **Trigger detectors after onboarding**:
```go
// After AWS connection successful
go triggerHiddenCostScan(tenantID, awsAccountID)
```

3. **Implement filters in backend**:
```go
func (h *CustomerHandler) GetFindings(w http.ResponseWriter, r *http.Request) {
    severity := r.URL.Query().Get("severity")
    category := r.URL.Query().Get("category")
    
    query := `SELECT * FROM yt_hidden_cost_findings WHERE tenant_id = $1`
    if severity != "" {
        query += ` AND severity = $2`
    }
    // ... apply filters
}
```

---

## 🔴 Resources Page Issues

### Current State
- **File**: `frontend/src/pages/Resources.tsx`
- **API Call**: `/api/v1/resources`
- **Backend Route**: ✅ EXISTS
- **Handler**: `resourceHandler.ListResources`

### Problems
1. **No resources in database**: `yt_tenant_resources` table is empty
2. **No AWS data fetching**: Backend doesn't call AWS APIs
3. **Filters return empty**: No data to filter

### Fix Required
1. **Implement AWS resource discovery**:
```go
// internal/aws/resource_discovery.go
func DiscoverResources(accountID, roleARN, externalID string) ([]Resource, error) {
    // Assume role
    creds := assumeRole(roleARN, externalID)
    
    // List EC2 instances
    ec2Client := ec2.New(creds)
    instances := ec2Client.DescribeInstances()
    
    // List S3 buckets
    s3Client := s3.New(creds)
    buckets := s3Client.ListBuckets()
    
    // Store in database
    saveResources(accountID, instances, buckets)
}
```

2. **Trigger discovery after onboarding**:
```go
// After AWS connection
go DiscoverResources(accountID, roleARN, externalID)
```

---

## 🔴 Whitelists Page Issues

### Current State
- **File**: `frontend/src/pages/Whitelists.tsx`
- **API Call**: `/api/whitelists`
- **Backend Route**: ✅ EXISTS
- **Handler**: `whitelistHandler.ListWhitelists`

### Problems
1. **No whitelists in database**: Table exists but empty
2. **Create/Delete not tested**: Endpoints exist but no UI testing

### Fix Required
- Seed sample whitelists for testing
- Test create/delete flows

---

## 🔴 Admin Dashboard Issues

### Current State
- **File**: `frontend/src/pages/AdminDashboard.tsx`
- **API Call**: `/api/admin/customers`, `/api/admin/metrics`
- **Backend Route**: ✅ EXISTS
- **Handler**: `adminHandler.GetCustomers`, `adminHandler.GetMetrics`

### Problems
1. **Returns all customers**: No pagination
2. **Metrics are mock**: Need real aggregation
3. **No admin user**: Can't test admin features

### Fix Required
1. **Create admin user**:
```sql
UPDATE yt_users SET role = 'admin' WHERE email = 'admin@yukti.io';
```

2. **Implement real metrics**:
```go
func (h *AdminHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
    totalCustomers := countCustomers()
    totalSavings := sumAllSavings()
    activeScans := countActiveScans()
    // Return real data
}
```

---

## 🔴 Onboarding Flow Issues

### Current State
- **File**: `frontend/src/pages/Onboarding.tsx`
- **Steps**: Welcome → AWS Connection → Metrics → Complete
- **Backend Routes**: ✅ EXIST

### Problems
1. **Step 3 (Metrics) not implemented**: UI exists but no backend logic
2. **Step 4 (Complete) doesn't trigger scan**: Should start resource discovery
3. **No progress tracking**: Can't resume onboarding

### Fix Required
1. **Implement metrics integration**:
```go
func (h *OnboardingHandler) ConfigureMetrics(w http.ResponseWriter, r *http.Request) {
    // Validate Prometheus/CloudWatch endpoint
    // Test connection
    // Save configuration
    // Trigger initial data sync
}
```

2. **Complete onboarding triggers**:
```go
func (h *OnboardingHandler) CompleteOnboarding(w http.ResponseWriter, r *http.Request) {
    // Mark onboarding complete
    h.service.CompleteOnboarding(tenantID)
    
    // Trigger background jobs
    go DiscoverResources(tenantID)
    go RunHiddenCostDetectors(tenantID)
    go SyncCostData(tenantID)
}
```

---

## 📊 Missing Backend Business Logic

### 1. AWS STS AssumeRole
**Status**: ❌ NOT IMPLEMENTED  
**Required**: Assume customer IAM roles to fetch AWS data

```go
// internal/aws/sts_client.go
func AssumeRole(roleARN, externalID string) (*credentials.Credentials, error) {
    sess := session.Must(session.NewSession())
    stsClient := sts.New(sess)
    
    result, err := stsClient.AssumeRole(&sts.AssumeRoleInput{
        RoleArn:         aws.String(roleARN),
        RoleSessionName: aws.String("yukti-session"),
        ExternalId:      aws.String(externalID),
        DurationSeconds: aws.Int64(3600),
    })
    
    return credentials.NewStaticCredentials(
        *result.Credentials.AccessKeyId,
        *result.Credentials.SecretAccessKey,
        *result.Credentials.SessionToken,
    ), nil
}
```

### 2. Resource Discovery
**Status**: ❌ NOT IMPLEMENTED  
**Required**: Fetch EC2, S3, RDS, Lambda resources from AWS

### 3. Cost Data Sync
**Status**: ❌ NOT IMPLEMENTED  
**Required**: Fetch cost data from AWS Cost Explorer

```go
// internal/aws/cost_explorer.go
func FetchCostData(accountID string, startDate, endDate time.Time) ([]CostRecord, error) {
    creds := getCredentials(accountID)
    ceClient := costexplorer.New(creds)
    
    result, err := ceClient.GetCostAndUsage(&costexplorer.GetCostAndUsageInput{
        TimePeriod: &costexplorer.DateInterval{
            Start: aws.String(startDate.Format("2006-01-02")),
            End:   aws.String(endDate.Format("2006-01-02")),
        },
        Granularity: aws.String("DAILY"),
        Metrics:     []*string{aws.String("UnblendedCost")},
    })
    
    // Parse and store in yt_cost_data table
}
```

### 4. Hidden Cost Detectors Trigger
**Status**: ❌ NOT IMPLEMENTED  
**Required**: Run 77 detectors and store findings

```go
// internal/hiddencosts/scanner.go
func RunScan(tenantID, accountID string) error {
    resources := getResources(accountID)
    
    // Run all 77 detectors
    findings := []Finding{}
    findings = append(findings, detectUnusedEBS(resources)...)
    findings = append(findings, detectUnderutilizedEC2(resources)...)
    // ... run all detectors
    
    // Store findings
    storeFindings(tenantID, findings)
}
```

### 5. RI/SP Recommendations
**Status**: ❌ NOT IMPLEMENTED  
**Required**: Calculate Reserved Instance savings

### 6. Budget Tracking
**Status**: ❌ NOT IMPLEMENTED  
**Required**: Track spend vs budget, send alerts

---

## 🎯 Priority Fix Order

### Phase 1: Make Dashboard Work (1-2 hours)
1. Fix tenant_id retrieval in all pages
2. Seed sample data in database
3. Implement real dashboard metrics calculation
4. Test dashboard displays data

### Phase 2: Make Hidden Costs Work (2-3 hours)
1. Seed sample findings
2. Implement filter logic in backend
3. Test findings display and filters
4. Add pagination

### Phase 3: Implement AWS Integration (4-6 hours)
1. Implement STS AssumeRole
2. Implement resource discovery (EC2, S3, RDS)
3. Store resources in database
4. Test resources page displays data

### Phase 4: Connect Detectors (3-4 hours)
1. Create scan trigger endpoint
2. Run detectors on discovered resources
3. Store findings in database
4. Test end-to-end flow

### Phase 5: Complete Onboarding (2-3 hours)
1. Implement metrics integration
2. Trigger scans after onboarding complete
3. Test full onboarding → dashboard flow

---

## 📝 Quick Win: Seed Sample Data

Run this SQL to make UI show data immediately:

```sql
-- Sample findings for tenant 18
INSERT INTO yt_hidden_cost_findings (tenant_id, detector_name, resource_arn, resource_type, severity, category, estimated_monthly_cost, estimated_savings, description, recommendation, status, created_at)
VALUES 
('18', 'unused-ebs-volumes', 'arn:aws:ec2:us-east-1:999888777666:volume/vol-abc123', 'ebs-volume', 'high', 'storage', 50.00, 50.00, 'EBS volume not attached to any instance', 'Delete unused volume or create snapshot', 'open', NOW()),
('18', 'underutilized-ec2', 'arn:aws:ec2:us-east-1:999888777666:instance/i-xyz789', 'ec2-instance', 'medium', 'compute', 200.00, 100.00, 'EC2 instance CPU < 10% for 7 days', 'Downsize to t3.small', 'open', NOW()),
('18', 'unoptimized-s3-storage', 'arn:aws:s3:::my-bucket', 's3-bucket', 'low', 'storage', 30.00, 15.00, 'S3 bucket using Standard storage for infrequent access', 'Move to S3 Intelligent-Tiering', 'open', NOW());

-- Sample resources
INSERT INTO yt_tenant_resources (tenant_id, resource_id, resource_type, resource_name, region, tags, monthly_cost, created_at)
VALUES
('18', 'i-xyz789', 'ec2-instance', 'web-server-1', 'us-east-1', '{"env":"production","team":"backend"}', 200.00, NOW()),
('18', 'vol-abc123', 'ebs-volume', 'unused-volume', 'us-east-1', '{}', 50.00, NOW()),
('18', 'my-bucket', 's3-bucket', 'my-bucket', 'us-east-1', '{"project":"analytics"}', 30.00, NOW());

-- Update dashboard metrics
UPDATE yt_customers 
SET onboarding_status = 'completed', 
    completed_at = NOW() 
WHERE tenant_id = '18';
```

---

## 🚨 Root Cause

**The core issue**: Frontend was built first with mock data expectations, but backend handlers were never fully implemented with real AWS integration and data aggregation logic.

**Solution**: Either:
1. **Quick**: Seed mock data to make UI functional for demo
2. **Proper**: Implement full AWS integration (4-6 hours of work)
3. **Hybrid**: Seed data now, implement AWS integration incrementally

---

**Recommendation**: Start with Phase 1 (seed data + fix tenant_id) to make the UI functional immediately, then implement AWS integration in parallel.
