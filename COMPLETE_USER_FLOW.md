# Complete User Flow - Yukti FinOps Platform

## Overview
This document describes the complete user journey from signup to dashboard, including authentication, AWS onboarding, resource scanning, and cost optimization features.

---

## FLOW 1: User Signup & Email Verification

### Step 1: Signup Page
```
┌─────────────────────────────────────────────────────────────┐
│                    YUKTI FINOPS                             │
│              Cloud Cost Optimization Platform               │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│                   Create Your Account                       │
│                                                             │
│  Email Address                                              │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ user@company.com                                    │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  Password (min 8 characters)                                │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ ••••••••••••                                        │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  Company Name (optional)                                    │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ Acme Corporation                                    │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │              Create Account                         │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  Already have an account? Sign in                          │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

**Backend Process**:
1. Validate email format and password strength
2. Check if email already exists
3. Create tenant record in `yt_tenants` (status: active, tier: FREE)
4. Hash password with bcrypt
5. Create user record in `yt_users`
6. Create tenant-user mapping in `yt_tenant_users` (role: owner)
7. Generate 6-digit OTP code
8. Send OTP email via AWS SES
9. Return success response

**Database Changes**:
```sql
-- yt_tenants
INSERT INTO yt_tenants (tenant_code, company_name, subscription_tier, status)
VALUES ('acme-corp-abc123', 'Acme Corporation', 'FREE', 'active');

-- yt_users
INSERT INTO yt_users (email, password_hash, is_active, email_verified)
VALUES ('user@company.com', '$2a$10$...', true, false);

-- yt_tenant_users
INSERT INTO yt_tenant_users (user_id, tenant_id, role)
VALUES ('uuid-123', 18, 'owner');
```

---

### Step 2: OTP Verification
```
┌─────────────────────────────────────────────────────────────┐
│                    YUKTI FINOPS                             │
│              Verify Your Email Address                      │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  We've sent a 6-digit code to:                              │
│  user@company.com                                           │
│                                                             │
│  Enter Verification Code                                    │
│  ┌───┐ ┌───┐ ┌───┐ ┌───┐ ┌───┐ ┌───┐                      │
│  │ 4 │ │ 5 │ │ 7 │ │ 2 │ │ 9 │ │ 1 │                      │
│  └───┘ └───┘ └───┘ └───┘ └───┘ └───┘                      │
│                                                             │
│  Code expires in 10 minutes                                 │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │              Verify Email                           │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  Didn't receive code? Resend                                │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

**Backend Process**:
1. Validate OTP code against stored hash
2. Check expiration (10 minutes)
3. Update `yt_users.email_verified = true`
4. Generate JWT token (24-hour expiry)
5. Return token + user info
6. Redirect to onboarding

**Email Template** (AWS SES):
```
Subject: Yukti Verification Code

Your verification code: 457291

This code expires in 10 minutes.

If you didn't request this, ignore this email.
```

---

## FLOW 2: Login & Authentication

### Step 1: Login Page
```
┌─────────────────────────────────────────────────────────────┐
│                    YUKTI FINOPS                             │
│              "We know the pain"                             │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│                   Sign In to Your Account                   │
│                                                             │
│  Email Address                                              │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ user@company.com                                    │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  Password                                                   │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ ••••••••••••                                        │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  🔒 Your financial data is protected with                   │
│     enterprise-grade security                               │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │              Sign In                                │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  Don't have an account? Sign up                             │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

**Backend Process**:
1. Validate email and password
2. Query `yt_users` by email
3. Verify password hash with bcrypt
4. Check `email_verified = true`
5. Generate JWT token with claims:
   - user_id
   - tenant_id
   - email
   - role (owner/admin/editor/viewer)
6. Return token + user info
7. Check AWS connection status
8. Redirect to onboarding OR dashboard

**JWT Token Structure**:
```json
{
  "user_id": "uuid-123",
  "tenant_id": 18,
  "email": "user@company.com",
  "role": "owner",
  "scopes": ["read", "write"],
  "exp": 1738368000
}
```

---

## FLOW 3: AWS Onboarding (IAM Role Setup)

### Step 1: IAM Role Instructions
```
┌─────────────────────────────────────────────────────────────┐
│  YUKTI FINOPS - ONBOARDING                  [Skip] [Logout] │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Welcome to Yukti!                                          │
│  Let's connect your AWS account to start optimizing costs  │
│                                                             │
│  Step 1: Create IAM Role in Your AWS Account               │
│                                                             │
│  📋 Quick Setup Instructions                                │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ 1. Go to AWS IAM Console → Roles                    │   │
│  │ 2. Click "Create role" → "AWS account"              │   │
│  │ 3. Choose "Another AWS account"                     │   │
│  │ 4. Enter Yukti Account ID: 144403604430             │   │
│  │ 5. Check "Require external ID" → Enter: yukti-*    │   │
│  │ 6. Attach policy: ReadOnlyAccess                    │   │
│  │ 7. Name: YuktiReadOnlyRole                          │   │
│  │ 8. Click "Create role"                              │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  🔐 Trust Policy (Copy & Paste)              [Copy]        │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ {                                                   │   │
│  │   "Version": "2012-10-17",                          │   │
│  │   "Statement": [{                                   │   │
│  │     "Effect": "Allow",                              │   │
│  │     "Principal": {                                  │   │
│  │       "AWS": "arn:aws:iam::144403604430:user/..."  │   │
│  │     },                                              │   │
│  │     "Action": "sts:AssumeRole",                     │   │
│  │     "Condition": {                                  │   │
│  │       "StringLike": {                               │   │
│  │         "sts:ExternalId": "yukti-*"                 │   │
│  │       }                                             │   │
│  │     }                                               │   │
│  │   }]                                                │   │
│  │ }                                                   │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  ⚠️ Important Details                                       │
│  • Yukti Account ID: 144403604430                           │
│  • External ID Pattern: yukti-* (auto-generated)            │
│  • Required Permission: ReadOnlyAccess                      │
│  • Security: We can only READ, never modify                 │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  ✅ I've Created the IAM Role → Continue           │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

### Step 2: AWS Connection Form
```
┌─────────────────────────────────────────────────────────────┐
│  YUKTI FINOPS - ONBOARDING                          [Logout]│
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Connect AWS Account                                        │
│  Enter your AWS account details                             │
│                                                             │
│  AWS Account ID                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ 424851482219                                        │   │
│  └─────────────────────────────────────────────────────┘   │
│  12-digit AWS account number                                │
│                                                             │
│  IAM Role ARN                                               │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ arn:aws:iam::424851482219:role/YuktiReadOnlyRole   │   │
│  └─────────────────────────────────────────────────────┘   │
│  Full ARN of the IAM role you created                       │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │              Connect AWS Account                    │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

**Backend Process**:
1. Validate Account ID (12 digits)
2. Validate Role ARN format
3. Auto-generate external ID: `yukti-{tenant_id}-{random_12_chars}`
4. Test AWS STS AssumeRole with external ID
5. Verify credentials with GetCallerIdentity
6. Store connection in `yt_metrics_integrations`:
   - account_id
   - role_arn
   - external_id (encrypted)
   - verified = true
   - last_verified_at = NOW()
7. Return success response

**Database Changes**:
```sql
INSERT INTO yt_metrics_integrations (
  tenant_id, account_id, role_arn, external_id, 
  verified, last_verified_at, regions
)
VALUES (
  18, '424851482219', 
  'arn:aws:iam::424851482219:role/YuktiReadOnlyRole',
  'yukti-18-abc123xyz456',
  true, NOW(),
  ARRAY['us-east-1', 'us-west-2', 'eu-west-1']
);
```

---

### Step 3: Connection Success
```
┌─────────────────────────────────────────────────────────────┐
│  YUKTI FINOPS - ONBOARDING                          [Logout]│
├─────────────────────────────────────────────────────────────┤
│                                                             │
│                    ✅ Setup Complete!                        │
│                                                             │
│  Your AWS account is connected.                             │
│  We're analyzing your resources and costs.                  │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  🔄 Scanning AWS Resources...                       │   │
│  │                                                     │   │
│  │  ✅ EC2 Instances: 45 found                         │   │
│  │  ✅ RDS Databases: 12 found                         │   │
│  │  ✅ S3 Buckets: 23 found                            │   │
│  │  🔄 Analyzing cost optimization opportunities...    │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  Redirecting to dashboard in 3 seconds...                   │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │              Go to Dashboard Now                    │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## FLOW 4: Resource Scanning & Data Storage

### Backend Scanning Process

**Trigger**: Automatic after AWS connection OR manual "Scan Resources" button

**Step 1: Assume IAM Role**
```go
// internal/scanner/aws_scanner.go
func (s *AWSScanner) assumeRole(accountID, roleARN, externalID string) (*aws.Credentials, error) {
    stsClient := sts.NewFromConfig(s.cfg)
    
    result, err := stsClient.AssumeRole(ctx, &sts.AssumeRoleInput{
        RoleArn:         aws.String(roleARN),
        RoleSessionName: aws.String("yukti-scanner"),
        ExternalId:      aws.String(externalID),
        DurationSeconds: aws.Int32(3600),
    })
    
    return result.Credentials, nil
}
```

**Step 2: Scan All 16 AWS Regions**
```
Regions: us-east-1, us-west-2, eu-west-1, ap-southeast-1, etc.

For each region:
  1. Fetch EC2 instances
  2. Fetch RDS databases
  3. Fetch S3 buckets
  4. Collect CloudWatch metrics
  5. Extract tags
  6. Store in PostgreSQL
```

**Step 3: Store Resources in PostgreSQL**
```sql
-- yt_tenant_resources
INSERT INTO yt_tenant_resources (
  tenant_id, resource_id, resource_type, region,
  metadata, tags, monthly_cost, discovered_at
)
VALUES (
  18, 'i-0a046ebb489ff3cd7', 'ec2', 'us-east-1',
  '{"instance_type": "t3.large", "state": "running", "cpu_avg": 12.5}',
  '{"Environment": "production", "Team": "backend"}',
  145.60, NOW()
);
```

**Step 4: Run 77 Cost Detectors**
```
For each resource:
  - Idle EC2 Detector (CPU < 5%)
  - Unattached EBS Detector
  - Old Snapshot Detector
  - Unencrypted RDS Detector
  - S3 Lifecycle Detector
  ... (77 total detectors)
```

**Step 5: Store Findings in PostgreSQL**
```sql
-- yt_hidden_cost_findings
INSERT INTO yt_hidden_cost_findings (
  tenant_id, resource_id, finding_type, severity,
  monthly_savings, description, status
)
VALUES (
  18, 'i-0a046ebb489ff3cd7', 'idle_ec2', 'high',
  120.00, 'EC2 instance with CPU < 5% for 7 days', 'open'
);
```

---

## FLOW 5: Dashboard (Post-Onboarding)

### Main Dashboard
```
┌─────────────────────────────────────────────────────────────┐
│  YUKTI FINOPS - DASHBOARD        [Scan] [Profile] [Logout]  │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Welcome back, user@company.com!                            │
│  Last updated: 2:45 PM                          [Refresh]   │
│                                                             │
│  ☁️ AWS Connection: 424851482219 • YuktiReadOnlyRole        │
│  🟢 Connected • Last verified: 5 minutes ago   [Scan Now]   │
│                                                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │ Total Cost  │  │   Savings   │  │  Resources  │         │
│  │  $12,450    │  │   $425.60   │  │     847     │         │
│  │  ↑ 5% MoM   │  │  Potential  │  │   Scanned   │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
│                                                             │
│  Cost Trend (Last 90 Days)      [30D] [90D] [1Y] [Custom]  │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  $600 ┤                                    ╱╲       │   │
│  │  $500 ┤                    ╱╲  ╱  ╲    ╱    ╲     │   │
│  │  $400 ┤        ╱╲  ╱  ╲╱            ╲╱        ╲   │   │
│  │  $300 ┤╱  ╲╱                                     ╲ │   │
│  │       └────────────────────────────────────────────│   │
│  │       Oct 1      Nov 1      Dec 1      Jan 1      │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  Cost Breakdown by Service                                  │
│  EC2 ████████████████████ 45% ($5,602)                     │
│  RDS ████████████ 25% ($3,112)                              │
│  S3  ████████ 15% ($1,867)                                  │
│  ELB ████ 8% ($996)                                         │
│                                                             │
│  Top Findings (7)                              [View All]   │
│  🔴 Idle EC2 instances (3) - Save $120/month   [Fix Now]   │
│  🟡 Unattached EBS volumes (2) - Save $80/mo   [Fix Now]   │
│  🟢 Old snapshots (5) - Save $50/month         [Fix Now]   │
│                                                             │
│  Budget Status: 78% used ($9,750 / $12,500)                │
│  ████████████████████░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░        │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

**Data Sources**:
- **Total Cost**: Sum from `yt_tenant_resources.monthly_cost`
- **Savings**: Sum from `yt_hidden_cost_findings.monthly_savings`
- **Resources**: Count from `yt_tenant_resources`
- **Cost Trend**: Aggregated from `yt_cost_data` (daily)
- **Findings**: Query `yt_hidden_cost_findings` WHERE status='open'
- **Budget**: From `yt_budgets` table

---

## FLOW 6: Hidden Costs Page

```
┌─────────────────────────────────────────────────────────────┐
│  YUKTI FINOPS - HIDDEN COSTS            [Dashboard] [Logout]│
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Cost Optimization Opportunities                            │
│  Total Potential Savings: $425.60/month                     │
│                                                             │
│  Filters:                                                   │
│  Severity: [All ▼] Category: [All ▼] Status: [Open ▼]      │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ 🔴 HIGH • Idle EC2 Instance                         │   │
│  │ Resource: i-0a046ebb489ff3cd7 (us-east-1)          │   │
│  │ Savings: $120/month                                 │   │
│  │ Details: CPU utilization < 5% for 7 days           │   │
│  │ [View Details] [Generate IaC] [Whitelist]          │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ 🟡 MEDIUM • Unattached EBS Volume                   │   │
│  │ Resource: vol-abc123 (us-west-2)                    │   │
│  │ Savings: $40/month                                  │   │
│  │ Details: Volume not attached for 30+ days          │   │
│  │ [View Details] [Generate IaC] [Whitelist]          │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  Showing 7 findings • Page 1 of 1                           │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## FLOW 7: Resources Inventory

```
┌─────────────────────────────────────────────────────────────┐
│  YUKTI FINOPS - RESOURCES               [Dashboard] [Logout]│
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  AWS Resources Inventory                                    │
│  Total: 847 resources across 3 regions                      │
│                                                             │
│  ┌───────┐ ┌───────┐ ┌───────┐ ┌───────┐                  │
│  │  EC2  │ │  RDS  │ │  S3   │ │  All  │                  │
│  │  45   │ │  12   │ │  23   │ │  847  │                  │
│  └───────┘ └───────┘ └───────┘ └───────┘                  │
│                                                             │
│  Search: [i-0a046ebb...]  Region: [All ▼]  Type: [All ▼]   │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ Resource ID         │ Type │ Region    │ Cost/mo   │   │
│  ├─────────────────────────────────────────────────────┤   │
│  │ i-0a046ebb489ff3cd7 │ EC2  │ us-east-1 │ $145.60   │   │
│  │ db-prod-mysql       │ RDS  │ us-west-2 │ $280.00   │   │
│  │ customer-data-prod  │ S3   │ us-east-1 │ $45.20    │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  [Export CSV] [Export PDF]                                  │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## Data Flow Summary

```
User Signup
    ↓
Email Verification (OTP)
    ↓
Login (JWT Token)
    ↓
AWS Onboarding (IAM Role)
    ↓
Resource Scanning (16 Regions)
    ↓
PostgreSQL Storage
    ├── yt_tenant_resources (EC2, RDS, S3)
    ├── yt_hidden_cost_findings (77 detectors)
    └── yt_cost_data (daily aggregates)
    ↓
Dashboard Display
    ├── Total Cost
    ├── Savings Opportunities
    ├── Resource Inventory
    └── Budget Tracking
```

---

## Redis Caching Strategy

**Cache Keys**:
- `tenant:{tenant_id}:dashboard` (TTL: 5 minutes)
- `tenant:{tenant_id}:resources` (TTL: 10 minutes)
- `tenant:{tenant_id}:findings` (TTL: 5 minutes)
- `tenant:{tenant_id}:aws_connection` (TTL: 1 hour)

**Cache Invalidation**:
- After resource scan completes
- After finding status change (open → resolved)
- After AWS connection update

---

## Performance Metrics

**Target Response Times**:
- Login: < 500ms
- Dashboard load: < 1 second
- Resource scan: 30-60 seconds (background)
- Finding generation: 10-20 seconds (background)

**Database Queries**:
- Dashboard: 3 queries (cached)
- Resources: 1 query with pagination
- Findings: 1 query with filters

**Scalability**:
- PostgreSQL: Up to 10,000 tenants
- ClickHouse: 50,000+ tenants (future migration)
- Redis: Session + cache storage
