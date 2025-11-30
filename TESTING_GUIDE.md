# Yukti Platform Testing Guide

## 🚀 Quick Start

### 1. Start All Services
```bash
cd /Users/chandrakantpatil/workspace/yukti
docker-compose up -d
```

**Services Started:**
- PostgreSQL: `localhost:5432`
- Backend API: `localhost:8080`
- ML Service: `localhost:8000`
- Frontend: `localhost:3000`
- Prometheus: `localhost:9090`
- Grafana: `localhost:3001`

### 2. Verify Services
```bash
# Check all containers are running
docker-compose ps

# Check logs
docker-compose logs -f backend
docker-compose logs -f frontend
```

### 3. Access Applications
- **Marketing Site**: http://localhost:3000
- **Customer Dashboard**: http://localhost:3000/dashboard
- **Admin Dashboard**: http://localhost:3000/admin
- **API**: http://localhost:8080
- **Grafana**: http://localhost:3001 (admin/admin)

---

## 👥 Test Accounts

### Customer Accounts (Pre-seeded)

#### 1. Acme Corp
- **Email**: admin@acme.com
- **Tenant ID**: tenant-001
- **Status**: Completed onboarding
- **Findings**: 3 (Total savings: $486.20/month)
- **Budget**: $15,000/month (83% used)

#### 2. TechStart Inc
- **Email**: cto@techstart.io
- **Tenant ID**: tenant-002
- **Status**: In progress (initial scan)
- **Findings**: 2 (Total savings: $428/month)
- **Budget**: $5,000/month (64% used)

#### 3. CloudScale LLC
- **Email**: ops@cloudscale.com
- **Tenant ID**: tenant-003
- **Status**: Completed onboarding
- **Findings**: 2 (Total savings: $3,520/month)
- **Budget**: $50,000/month (84% used)

### Admin Account
- **URL**: http://localhost:3000/admin
- **Access**: Direct access (no login required in dev)

---

## 🧪 Testing Scenarios

### Scenario 1: New Customer Onboarding

**Steps:**
1. Go to http://localhost:3000
2. Click "Start Free Trial"
3. Fill in company info:
   - Company: Test Company
   - Email: test@example.com
4. Configure AWS:
   - Account ID: 999888777666
   - Role ARN: arn:aws:iam::999888777666:role/YuktiRole
   - External ID: test-external-id
   - Regions: us-east-1
5. Configure Metrics (optional):
   - Source: Prometheus
   - Endpoint: http://prometheus:9090
   - Skip or add token
6. Wait for initial scan
7. View findings

**Expected Result:**
- Customer created with new tenant_id
- Onboarding status progresses through steps
- Initial scan completes
- Findings displayed

---

### Scenario 2: View Hidden Costs

**Steps:**
1. Go to http://localhost:3000/dashboard
2. Impersonate "Acme Corp" (tenant-001)
3. Click "Hidden Costs" in navigation
4. View findings list
5. Filter by category (Data Transfer, Storage, etc.)
6. Filter by severity (Critical, High, Medium, Low)
7. Click on a finding to see details
8. View slide-out panel with:
   - Full description
   - Estimated savings
   - Recommendation
   - IaC code (if available)

**Expected Result:**
- 3 findings displayed for Acme Corp
- Total savings: $486.20/month
- Filters work correctly
- Detail panel shows complete information

---

### Scenario 3: Generate IaC Code

**Steps:**
1. Go to Hidden Costs page
2. Select finding: "EBS gp2 to gp3 migration"
3. Click "Generate IaC"
4. Select format: Terraform
5. View generated code
6. Download .tf file
7. Repeat with CloudFormation format

**Expected Result:**
- Terraform code generated with:
  - Resource definitions
  - Data sources
  - Outputs
  - Comments with savings
- CloudFormation YAML generated
- Download works correctly

---

### Scenario 4: Budget Tracking

**Steps:**
1. Go to Dashboard
2. View budget widget
3. See current spend vs budget
4. Check if alert triggered (>80%)
5. View budget details
6. See breakdown by service

**Expected Result:**
- Acme Corp: $12,500 / $15,000 (83% - Alert!)
- Visual progress bar
- Alert notification
- Service breakdown (EC2, RDS, S3)

---

### Scenario 5: RI/SP Recommendations

**Steps:**
1. Go to Dashboard
2. View RI recommendations widget
3. See recommended Reserved Instances:
   - Instance type
   - Region
   - Term
   - Estimated savings
4. View SP recommendations
5. See coverage analysis

**Expected Result:**
- Acme Corp: 1 RI recommendation ($450/month savings)
- CloudScale: 1 RI recommendation ($2,000/month savings)
- Coverage percentages displayed
- Break-even calculations shown

---

### Scenario 6: Admin Dashboard

**Steps:**
1. Go to http://localhost:3000/admin
2. View platform metrics:
   - Total customers: 3
   - Total savings: $85,700/month
   - Active trials: 1
   - MRR: $1,497
3. View customer list
4. Search for "Acme"
5. Click "View" to impersonate Acme Corp
6. View their dashboard
7. Return to admin dashboard

**Expected Result:**
- All 3 customers listed
- Metrics calculated correctly
- Search works
- Impersonation works (backdoor access)
- Can view customer data

---

### Scenario 7: Whitelisting

**Steps:**
1. Go to Whitelists page
2. Click "Create Whitelist"
3. Fill in:
   - Type: Resource
   - Resource ARN: arn:aws:ec2:us-east-1:123:instance/i-test
   - Reason: Production critical instance
   - Expiry: 30 days
4. Submit
5. View whitelist in list
6. Revoke whitelist

**Expected Result:**
- Whitelist created
- Appears in list with status "active"
- Can be revoked
- Audit trail recorded

---

### Scenario 8: Cost Anomaly Detection

**Steps:**
1. Go to Dashboard
2. View "Cost Anomalies" widget
3. See detected anomalies:
   - Service
   - Impact ($)
   - Date range
4. Click on anomaly for details
5. View recommendation

**Expected Result:**
- Anomalies detected (if any)
- Impact quantified
- Root cause identified
- Recommendation provided

---

## 🔧 Admin Backdoor Functions

### Impersonate Customer
```javascript
// In browser console
localStorage.setItem('admin_impersonate', 'tenant-001');
window.location.href = '/dashboard';
```

### View Customer Data
```bash
# Connect to database
docker exec -it yukti-postgres psql -U yukti -d yukti

# View customers
SELECT * FROM yt_customers;

# View findings for customer
SELECT * FROM yt_hidden_cost_findings WHERE tenant_id = 'tenant-001';

# View budgets
SELECT * FROM yt_budgets WHERE tenant_id = 'tenant-001';
```

### Trigger Manual Scan
```bash
# API call to trigger scan
curl -X POST http://localhost:8080/api/scan \
  -H "Content-Type: application/json" \
  -d '{"tenant_id": "tenant-001"}'
```

### Update Customer Data
```sql
-- Update onboarding status
UPDATE yt_customers 
SET onboarding_status = 'completed', completed_at = NOW() 
WHERE tenant_id = 'tenant-002';

-- Add more findings
INSERT INTO yt_hidden_cost_findings (...) VALUES (...);

-- Update budget spend
UPDATE yt_budgets 
SET current_spend = 14000.00 
WHERE tenant_id = 'tenant-001';
```

---

## 📊 Monitoring & Debugging

### View Logs
```bash
# Backend logs
docker-compose logs -f backend

# Frontend logs
docker-compose logs -f frontend

# ML service logs
docker-compose logs -f ml-service

# Database logs
docker-compose logs -f postgres
```

### Check Database
```bash
# Connect to PostgreSQL
docker exec -it yukti-postgres psql -U yukti -d yukti

# List tables
\dt

# View table schema
\d yt_customers

# Query data
SELECT * FROM yt_customers;
SELECT * FROM yt_hidden_cost_findings;
```

### API Health Check
```bash
# Backend health
curl http://localhost:8080/health

# ML service health
curl http://localhost:8000/health
```

### Prometheus Metrics
- Go to http://localhost:9090
- Query: `yukti_findings_total`
- Query: `yukti_customers_total`
- Query: `yukti_savings_total`

### Grafana Dashboards
- Go to http://localhost:3001
- Login: admin/admin
- View pre-configured dashboards:
  - Customer metrics
  - Savings metrics
  - System metrics

---

## 🐛 Common Issues

### Issue: Containers won't start
```bash
# Stop all containers
docker-compose down

# Remove volumes
docker-compose down -v

# Rebuild and start
docker-compose up --build -d
```

### Issue: Database connection error
```bash
# Check PostgreSQL is running
docker-compose ps postgres

# Check logs
docker-compose logs postgres

# Restart PostgreSQL
docker-compose restart postgres
```

### Issue: Frontend not loading
```bash
# Check frontend logs
docker-compose logs frontend

# Rebuild frontend
docker-compose up --build frontend
```

### Issue: Seed data not loaded
```bash
# Manually run seed script
docker exec -it yukti-postgres psql -U yukti -d yukti -f /docker-entrypoint-initdb.d/seed_data.sql
```

---

## ✅ Testing Checklist

### Customer Experience
- [ ] Marketing site loads
- [ ] Sign-up flow works
- [ ] Onboarding completes
- [ ] Dashboard displays data
- [ ] Hidden costs page works
- [ ] Filters work correctly
- [ ] Detail panel opens
- [ ] IaC generation works
- [ ] Download works
- [ ] Whitelisting works
- [ ] Budget tracking works

### Admin Experience
- [ ] Admin dashboard loads
- [ ] Customer list displays
- [ ] Search works
- [ ] Impersonation works
- [ ] Metrics are correct
- [ ] Can view customer data

### Data Integrity
- [ ] Customers created correctly
- [ ] Findings stored properly
- [ ] Budgets calculated correctly
- [ ] RI/SP recommendations accurate
- [ ] Cost data aggregated correctly

### Performance
- [ ] Pages load <2 seconds
- [ ] API responses <500ms
- [ ] Database queries optimized
- [ ] No memory leaks

---

## 🚀 Ready for Production?

### Pre-Launch Checklist
- [ ] All tests passing
- [ ] No critical bugs
- [ ] Performance acceptable
- [ ] Security reviewed
- [ ] Documentation complete
- [ ] Monitoring configured
- [ ] Backup strategy in place
- [ ] Rollback plan ready

### Go-Live Steps
1. Deploy to production environment
2. Run smoke tests
3. Monitor for 24 hours
4. Enable marketing campaigns
5. Onboard first customers
6. Collect feedback
7. Iterate and improve

---

**Status**: Ready for testing! 🎉
**Next**: Run through all scenarios, fix bugs, then deploy to production.
