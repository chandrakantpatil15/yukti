# Quick Test Guide - Multi-Tenant Flow

## ✅ All Services Running

```bash
docker-compose ps
# All 6 services should be "Up"
```

## 🧪 Test Scenarios

### 1. Admin Dashboard (Empty → Populated)

**Access**: http://localhost:3000/admin

**Expected**:
- ✅ Shows 3 customers (Acme Corp, TechStart Inc, CloudScale LLC)
- ✅ Metrics: 3 customers, $4,434 total savings, 1 active trial, $198 MRR
- ✅ Search box works
- ✅ "View" buttons are clickable

**Test**:
1. Open http://localhost:3000/admin
2. See customer list populated
3. Search for "Acme" - filters to 1 result
4. Click "View" on Acme Corp

---

### 2. Customer Dashboard (Dynamic Data)

**After clicking "View" on Acme Corp**:

**Expected**:
- ✅ Shows Acme Corp data (tenant-001)
- ✅ Total Savings: $486
- ✅ Findings: 3
- ✅ Budget: 83% used ($12,500 / $15,000)
- ✅ RI Savings: $450
- ✅ Budget bar shows orange/red (>80%)

**Test**:
1. Verify all numbers match Acme Corp
2. Click "View Hidden Costs" button
3. Should navigate to /hidden-costs

---

### 3. Hidden Costs Page (Filtering & Sorting)

**Access**: http://localhost:3000/hidden-costs

**Expected**:
- ✅ Shows 3 findings for Acme Corp
- ✅ Total savings: $486.20/month
- ✅ Sorted by savings (highest first)
- ✅ Filters work

**Test**:
1. See 3 findings listed
2. Select "Data Transfer Costs" category - shows 1 finding
3. Clear filters - shows all 3 again
4. Click on any finding card
5. Detail panel opens with full info
6. "Generate IaC" and "Whitelist" buttons visible

---

### 4. Switch Tenants

**Test different tenant data**:

```javascript
// In browser console
localStorage.setItem('tenant_id', 'tenant-002');
window.location.href = '/dashboard';
```

**Expected for TechStart Inc (tenant-002)**:
- ✅ Total Savings: $428
- ✅ Findings: 2
- ✅ Budget: 64% used ($3,200 / $5,000)
- ✅ Different findings in hidden costs

**Test for CloudScale LLC (tenant-003)**:
```javascript
localStorage.setItem('tenant_id', 'tenant-003');
window.location.href = '/dashboard';
```

**Expected**:
- ✅ Total Savings: $3,520
- ✅ Findings: 2
- ✅ Budget: 84% used ($42,000 / $50,000)
- ✅ RI Savings: $2,000

---

### 5. Onboarding Flow

**Access**: http://localhost:3000/onboarding

**Test**:
1. Fill in:
   - Company: "Test Company"
   - Email: "test@example.com"
2. Click "Continue"
3. See tenant ID generated (e.g., "tenant-abc12345")
4. Fill in AWS details:
   - Account ID: "999888777666"
   - Role ARN: "arn:aws:iam::999888777666:role/YuktiRole"
5. Click "Continue"
6. See success screen
7. Click "Go to Dashboard"
8. Should redirect to dashboard with new tenant

**Verify**:
```bash
# Check new customer created
curl http://localhost:8080/api/admin/customers | jq '.customers | length'
# Should show 4 (was 3, now 4)
```

---

### 6. API Testing

**Admin APIs**:
```bash
# Get all customers
curl http://localhost:8080/api/admin/customers | jq '.customers[0]'

# Get metrics
curl http://localhost:8080/api/admin/metrics | jq '.metrics'
```

**Customer APIs**:
```bash
# Dashboard data for tenant-001
curl "http://localhost:8080/api/customers/dashboard?tenant_id=tenant-001" | jq '.'

# Findings for tenant-001
curl "http://localhost:8080/api/customers/findings?tenant_id=tenant-001" | jq '.findings | length'

# Filtered findings
curl "http://localhost:8080/api/customers/findings?tenant_id=tenant-001&category=Data%20Transfer%20Costs" | jq '.findings | length'
```

---

## 🎯 Verification Checklist

### Admin Dashboard
- [ ] Page loads without errors
- [ ] Shows 3 customers
- [ ] Metrics display correctly
- [ ] Search filters customers
- [ ] "View" button works
- [ ] Redirects to customer dashboard

### Customer Dashboard
- [ ] Shows correct tenant data
- [ ] All 4 metric cards populated
- [ ] Budget progress bar displays
- [ ] Alert shows if >80%
- [ ] Quick action buttons work

### Hidden Costs
- [ ] Findings load for current tenant
- [ ] Category filter works
- [ ] Severity filter works
- [ ] Clear filters works
- [ ] Findings sorted by savings
- [ ] Click opens detail panel
- [ ] Detail panel shows all info

### Multi-Tenant Isolation
- [ ] tenant-001 shows 3 findings
- [ ] tenant-002 shows 2 findings
- [ ] tenant-003 shows 2 findings
- [ ] No data leakage between tenants

### Onboarding
- [ ] Step 1 creates customer
- [ ] Tenant ID generated
- [ ] Step 2 accepts AWS details
- [ ] Step 3 shows success
- [ ] Redirects to dashboard
- [ ] New customer appears in admin

---

## 🐛 Troubleshooting

### Admin dashboard empty
```bash
# Check backend logs
docker-compose logs backend | tail -20

# Test API directly
curl http://localhost:8080/api/admin/customers

# Verify data in database
docker exec yukti-postgres psql -U yukti -d yukti -c "SELECT COUNT(*) FROM yt_customers;"
```

### Dashboard shows zeros
```bash
# Check tenant_id in localStorage
# Open browser console:
localStorage.getItem('tenant_id')

# Should return: "tenant-001" or similar

# Test API with tenant_id
curl "http://localhost:8080/api/customers/dashboard?tenant_id=tenant-001"
```

### Findings not loading
```bash
# Check findings in database
docker exec yukti-postgres psql -U yukti -d yukti -c "SELECT tenant_id, COUNT(*) FROM yt_hidden_cost_findings GROUP BY tenant_id;"

# Test API
curl "http://localhost:8080/api/customers/findings?tenant_id=tenant-001"
```

### CORS errors
```bash
# Backend should allow localhost:3000
# Check backend logs for CORS errors
docker-compose logs backend | grep -i cors
```

---

## 📊 Expected Data Summary

| Tenant | Company | Findings | Savings | Budget Used |
|--------|---------|----------|---------|-------------|
| tenant-001 | Acme Corp | 3 | $486 | 83% |
| tenant-002 | TechStart Inc | 2 | $428 | 64% |
| tenant-003 | CloudScale LLC | 2 | $3,520 | 84% |

**Total Platform**:
- Customers: 3
- Total Savings: $4,434/month
- Active Trials: 1 (TechStart Inc)
- MRR: $198 (2 completed × $99)

---

## ✨ Success Criteria

All tests pass when:
1. ✅ Admin dashboard shows real data
2. ✅ Customer dashboard shows tenant-specific data
3. ✅ Hidden costs page shows filtered findings
4. ✅ Switching tenants shows different data
5. ✅ Onboarding creates new customer
6. ✅ All buttons are clickable and functional
7. ✅ Sorting and filtering work
8. ✅ No console errors
9. ✅ APIs return correct data
10. ✅ Multi-tenant isolation enforced
