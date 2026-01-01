# Yukti Platform - Simple User Journey & Implementation Plan

## 🎯 PROJECT GOAL
**Simple AWS Cost Optimization Platform** - Users connect their AWS account, we scan it, show savings opportunities.

---

## 👤 USER JOURNEY (SIMPLE VERSION)

### **Step 1: Signup** (30 seconds)
```
User visits: http://localhost:3000/signup

Enters:
- Email: user@company.com
- Password: SecurePass123!

Clicks "Sign Up"

Backend:
1. Creates company account (tenant)
2. Creates user account
3. Sends success message

Result: "Account created! Please login"
```

### **Step 2: Login** (15 seconds)
```
User visits: http://localhost:3000/login

Enters:
- Email: user@company.com
- Password: SecurePass123!

Clicks "Login"

Backend:
1. Finds user by email
2. Checks password
3. Creates JWT token with tenant info

Result: Redirected to Dashboard
```

### **Step 3: Dashboard (First Time)** (5 seconds)
```
User sees: "Welcome! Connect your AWS account to start"

Shows:
- Big "Connect AWS" button
- Why we need AWS access
- What we'll scan

Clicks "Connect AWS"

Result: Redirected to Onboarding
```

### **Step 4: AWS Onboarding** (2 minutes)
```
User sees simple form:

Field 1: AWS Account ID
Input: 123456789012

Field 2: IAM Role ARN
Input: arn:aws:iam::123456789012:role/YuktiFinOpsRole

Shows:
- Copy-paste instructions
- Trust policy JSON
- Step-by-step AWS Console guide

Clicks "Connect"

Backend:
1. Tests AWS connection (STS AssumeRole)
2. If success: Saves connection
3. If fail: Shows clear error

Result: "AWS Connected! Starting scan..."
```

### **Step 5: First Scan** (30 seconds)
```
Backend automatically:
1. Scans 16 AWS regions
2. Finds EC2, RDS, S3 resources
3. Runs 77 cost detectors
4. Stores findings

User sees:
- Progress bar
- "Scanning us-east-1..."
- "Found 45 resources"

Result: Redirected to Dashboard with data
```

### **Step 6: Dashboard (With Data)** (Ongoing)
```
User sees:

Top Cards:
- Total Monthly Cost: $12,450
- Potential Savings: $3,200 (25%)
- Resources Scanned: 847
- Findings: 23

Findings List:
1. 🔴 Idle EC2 instance (i-abc123) - Save $156/month
2. 🟡 Unattached EBS volume - Save $45/month
3. 🟢 Old snapshot (90+ days) - Save $12/month

Each finding shows:
- What it is
- Why it costs money
- How to fix it
- One-click fix (future)

Actions:
- View all findings
- Filter by severity
- Export report
- Trigger new scan
```

---

## 🗄️ DATABASE SCHEMA (SIMPLE)

### **yt_tenants** (Companies)
```sql
id: 1, 2, 3...
tenant_code: "acme-corp-a1b2c3d4"
company_name: "Acme Corp"
status: "active"
created_at: timestamp
```

### **yt_users** (User Accounts)
```sql
id: uuid
tenant_id: 1 (FK to yt_tenants)
email: "user@company.com" (UNIQUE)
password_hash: bcrypt hash
role: "admin" or "viewer"
email_verified: true/false
created_at: timestamp
```

### **yt_aws_connections** (AWS Accounts)
```sql
id: 1, 2, 3...
tenant_id: 1 (FK to yt_tenants)
account_id: "123456789012"
role_arn: "arn:aws:iam::..."
external_id: "yukti-1-abc123"
verified: true/false
last_scan_at: timestamp
```

### **yt_tenant_resources** (Discovered Resources)
```sql
id: 1, 2, 3...
tenant_id: 1
resource_id: "i-abc123"
resource_type: "ec2"
region: "us-east-1"
metadata: JSON (tags, size, etc)
monthly_cost: 156.00
created_at: timestamp
```

### **yt_hidden_cost_findings** (Savings Opportunities)
```sql
id: 1, 2, 3...
tenant_id: 1
resource_id: "i-abc123"
finding_type: "idle_ec2"
severity: "high"
monthly_savings: 156.00
description: "EC2 instance with <5% CPU"
recommendation: "Stop or downsize"
status: "open"
created_at: timestamp
```

---

## 🔐 AUTHENTICATION FLOW (SIMPLE)

### **Signup**
```
POST /api/v1/auth/signup
Body: { "email": "user@company.com", "password": "Pass123!" }

Backend:
1. Check if email exists → Error if yes
2. Create tenant (company_name = email domain)
3. Create user (tenant_id = new tenant, role = "admin")
4. Return success

Response: { "success": true, "message": "Account created" }
```

### **Login**
```
POST /api/v1/auth/login
Body: { "email": "user@company.com", "password": "Pass123!" }

Backend:
1. Find user by email (email is UNIQUE across all tenants)
2. Check password
3. Get tenant_id from user record
4. Generate JWT with: user_id, tenant_id, email, role
5. Return token

Response: { 
  "success": true, 
  "token": "eyJhbGc...",
  "user": { "email": "...", "tenant_id": 1 }
}
```

### **Protected Routes**
```
GET /api/v1/dashboard
Headers: { "Authorization": "Bearer eyJhbGc..." }

Backend:
1. Extract JWT from header
2. Validate JWT signature
3. Extract tenant_id from JWT
4. Fetch data for that tenant_id ONLY
5. Return data

Response: { "metrics": {...}, "findings": [...] }
```

---

## 🚀 DEPLOYMENT (SIMPLE)

### **Development (Docker)**
```bash
# Start everything
docker-compose up -d

# Services:
- Backend: http://localhost:8081
- Frontend: http://localhost:3000
- PostgreSQL: localhost:5432
```

### **Production (Customer Hosted)**
```bash
# Customer runs on their cloud:
docker-compose -f docker-compose.prod.yml up -d

# All data stays in customer's infrastructure
# We provide:
- Docker images
- Database migrations
- Configuration guide
```

### **Production (SaaS)**
```bash
# We run on our cloud:
- Kubernetes cluster
- Managed PostgreSQL
- Load balancer
- Auto-scaling

# Customers just visit: https://app.yukti.io
```

---

## 📋 IMPLEMENTATION CHECKLIST

### **Phase 1: Basic Auth (1 hour)**
- [ ] Fix auth.go to NOT require tenant_code
- [ ] User login with email + password only
- [ ] JWT contains tenant_id
- [ ] Test: Signup → Login → Get JWT

### **Phase 2: Dashboard (30 minutes)**
- [ ] Dashboard reads tenant_id from JWT
- [ ] Shows placeholder data
- [ ] Test: Login → See dashboard

### **Phase 3: AWS Onboarding (1 hour)**
- [ ] Form: AWS Account ID + Role ARN
- [ ] Backend: Test STS AssumeRole
- [ ] Save connection if successful
- [ ] Test: Connect AWS → See success

### **Phase 4: AWS Scanner (2 hours)**
- [ ] Scan button triggers scan
- [ ] Fetch EC2, RDS, S3 from AWS
- [ ] Store in yt_tenant_resources
- [ ] Test: Scan → See resources

### **Phase 5: Cost Detectors (1 hour)**
- [ ] Run 77 detectors on resources
- [ ] Store findings in yt_hidden_cost_findings
- [ ] Test: Scan → See findings

### **Phase 6: Dashboard with Data (30 minutes)**
- [ ] Show real metrics from database
- [ ] Show findings list
- [ ] Test: Full flow end-to-end

---

## 🎯 SUCCESS CRITERIA

**User can:**
1. ✅ Signup with email + password (no tenant code)
2. ✅ Login with email + password (no tenant code)
3. ✅ See dashboard (empty at first)
4. ✅ Connect AWS account (2 fields only)
5. ✅ Trigger scan (one button)
6. ✅ See findings (list of savings)

**Total Time**: User onboarded in **3 minutes**

---

## 🔧 TECHNICAL DECISIONS (SIMPLE)

### **Authentication**
- ✅ Email is UNIQUE across all tenants
- ✅ No tenant_code needed for login
- ✅ JWT contains tenant_id
- ✅ Backend uses JWT tenant_id for all queries

### **Database**
- ✅ Use yt_tenants (integer ID) as main tenant table
- ✅ Ignore yt_customers (legacy/onboarding tracking)
- ✅ yt_users.tenant_id references yt_tenants.id

### **AWS Integration**
- ✅ User provides: Account ID + Role ARN
- ✅ Backend auto-generates external_id
- ✅ Backend tests connection before saving
- ✅ Scan runs in background (non-blocking)

### **Deployment**
- ✅ Docker Compose for development
- ✅ Same Docker Compose for customer-hosted
- ✅ Kubernetes for SaaS (future)

---

## 📝 NEXT STEPS

1. **Review this document** - Is the user journey clear?
2. **Confirm decisions** - Email-only login? No tenant code?
3. **Start coding** - Fix auth.go first
4. **Test each phase** - Don't move forward until working
5. **Deploy** - Docker Compose up and running

---

**Status**: Ready for implementation
**Estimated Time**: 6 hours total
**Complexity**: Simple (no over-engineering)
