# 🧪 End-to-End Testing Guide

## ✅ Database Cleaned - Ready for Fresh Test

All user data has been cleared:
- ✅ 0 Users
- ✅ 0 Customers (Tenants)
- ✅ 0 AWS Connections
- ✅ 0 Resources
- ✅ 0 Findings
- ✅ 0 Budgets
- ✅ 0 AWS Accounts
- ✅ 0 Refresh Tokens

---

## 🚀 Complete E2E Test Flow

### Step 1: User Signup (5 minutes)
1. **Open Platform**: http://localhost:3000
2. **Click**: "Sign Up" button
3. **Enter Details**:
   - Email: `test@yukti.com` (or your email)
   - Password: `Test@123456` (min 8 chars, 1 uppercase, 1 special)
   - Confirm Password: `Test@123456`
4. **Click**: "Sign Up"
5. **Check Console**: Backend logs will show OTP code
   ```bash
   docker-compose logs backend | grep "OTP"
   ```
6. **Enter OTP**: Copy OTP from logs and paste in verification screen
7. **Expected Result**: ✅ Email verified, redirected to login

### Step 2: Login (2 minutes)
1. **Enter Credentials**:
   - Email: `test@yukti.com`
   - Password: `Test@123456`
2. **Click**: "Login"
3. **Expected Result**: ✅ Redirected to Dashboard

### Step 3: Dashboard - Initial State (2 minutes)
**What You'll See**:
- Total Savings: $0
- Findings: 0
- Budget Usage: 0%
- RI Savings: $0
- ⚠️ No AWS Connection banner

**Expected**: Dashboard loads successfully with empty state

### Step 4: AWS Onboarding (10 minutes)
1. **Click**: "Settings" or "Configure AWS" button
2. **Enter AWS Details**:
   - AWS Account ID: `424851482219` (your AWS account)
   - IAM Role Name: `YuktiFinOpsRole`
3. **Backend Auto-Generates**:
   - External ID: `yukti-1-xxxxxxxxxxxx` (shown in UI)
4. **Copy Trust Policy**: Click copy button
5. **Go to AWS Console**:
   - Navigate to IAM → Roles
   - Create role: `YuktiFinOpsRole`
   - Trusted entity: Another AWS account
   - Account ID: `144403604430` (Yukti platform account)
   - Require external ID: Paste the external ID from UI
   - Attach policy: `ReadOnlyAccess`
6. **Back to Yukti**: Click "Verify Connection"
7. **Expected Result**: ✅ Connection verified, green status

### Step 5: First Scan (5 minutes)
1. **Dashboard**: Click "Scan Resources" button
2. **Alert**: "Scan started! Scanning all AWS regions..."
3. **Wait**: 30-60 seconds for scan to complete
4. **Auto-Refresh**: Dashboard polls every 5 seconds (12 times)
5. **Check Backend Logs**:
   ```bash
   docker-compose logs -f backend | grep Scanner
   ```
6. **Expected Output**:
   ```
   [Scanner] ========== STARTING AWS SCAN ===========
   [Scanner] Scanning 16 AWS regions...
   [Scanner] Region ap-southeast-2: Found 1 EC2 instance
   [Scanner] ✓ Total resources found: 1
   [Scanner] ✓ Successfully stored 1 resources
   [Scanner] Running cost optimization detectors...
   [Scanner] ✓ Detectors completed successfully
   [Scanner] ========== SCAN COMPLETED ===========
   ```

### Step 6: Verify Resources (5 minutes)
1. **Navigate**: Resources page
2. **Expected**:
   - EC2 count: 1 (if you have EC2 in ap-southeast-2)
   - RDS count: 0
   - S3 count: 0
3. **Click**: EC2 instance row
4. **Verify Details**:
   - Instance ID: `i-0a046ebb489ff3cd7`
   - Instance Type: `t3.micro`
   - State: `running`
   - Region: `ap-southeast-2`
   - Tags: All AWS tags displayed
   - Metadata: Complete AWS metadata

### Step 7: Check Findings (5 minutes)
1. **Navigate**: Hidden Costs page
2. **Expected Findings** (if resources exist):
   - Right-sizing opportunities
   - Spot instance recommendations
   - Storage optimization
   - Reserved instance suggestions
3. **If 0 Findings**:
   - ⚠️ Normal for first scan (need 24h CloudWatch metrics)
   - ⚠️ Detectors need historical data
   - ✅ Resources are discovered and stored

### Step 8: Wait for CloudWatch Metrics (24 hours)
**Why Wait?**
- CloudWatch needs 24 hours of metrics history
- Detectors analyze: CPU utilization, network traffic, disk I/O
- Without metrics: Detectors can't identify optimization opportunities

**What to Do**:
1. Let resources run for 24 hours
2. Come back tomorrow
3. Trigger second scan
4. Findings will appear with metrics data

### Step 9: Second Scan (Next Day)
1. **Dashboard**: Click "Scan Resources" again
2. **This Time**:
   - Scanner fetches CloudWatch metrics
   - Stores: `avg_cpuutilization`, `avg_networkin`, etc.
   - Detectors analyze metrics
   - Generates cost optimization findings
3. **Expected**: 5-10 findings with savings recommendations

### Step 10: Review Findings (10 minutes)
1. **Hidden Costs Page**:
   - Filter by severity: Critical, High, Medium, Low
   - Filter by category: Compute, Storage, Network, etc.
2. **Click Finding**: View details
   - Description
   - Estimated savings
   - Confidence score
   - Resource ARN
3. **Actions**:
   - Generate IaC (Terraform/CloudFormation)
   - Whitelist (exclude from recommendations)

### Step 11: Generate IaC (5 minutes)
1. **Copy Finding ID**: From Hidden Costs page
2. **Navigate**: IaC Generator page
3. **Enter**: Finding ID
4. **Select**: Terraform or CloudFormation
5. **Click**: Generate Code
6. **Expected**: Infrastructure code for optimization
7. **Copy**: Use in your AWS account

### Step 12: Test Whitelisting (5 minutes)
1. **Navigate**: Whitelists page
2. **Click**: "Add Whitelist"
3. **Enter**:
   - Resource ARN: Copy from finding
   - Reason: "Production critical workload"
   - Expires: 90 days
4. **Click**: Create
5. **Expected**: Resource excluded from future scans

### Step 13: Profile & Settings (5 minutes)
1. **Navigate**: Profile page
2. **Verify**:
   - Email address
   - Tenant ID
   - Role (user/admin)
   - AWS Account ID
   - IAM Role ARN
   - Regions configured
3. **Copy**: Any field with copy button

---

## 🎯 Success Criteria

### ✅ Signup & Login
- [ ] User can sign up with email
- [ ] OTP verification works
- [ ] Login successful with JWT token
- [ ] Auto-logout on token expiration

### ✅ AWS Onboarding
- [ ] External ID auto-generated
- [ ] Trust policy displayed
- [ ] IAM role verification works
- [ ] Connection status shows "Connected"

### ✅ Resource Discovery
- [ ] Scan completes without errors
- [ ] Resources discovered (EC2/RDS/S3)
- [ ] Metadata stored correctly
- [ ] Tags displayed in UI

### ✅ Cost Optimization
- [ ] Detectors run successfully
- [ ] Findings generated (after 24h)
- [ ] Savings calculated
- [ ] Recommendations actionable

### ✅ User Experience
- [ ] All pages load without errors
- [ ] Navigation works smoothly
- [ ] Real-time updates (auto-refresh)
- [ ] Error messages clear and helpful

---

## 🐛 Troubleshooting

### Issue: OTP Not Received
**Solution**: Check backend logs
```bash
docker-compose logs backend | grep "OTP"
```
OTP is printed in console (dev mode)

### Issue: AWS Connection Failed
**Possible Causes**:
1. Wrong Account ID
2. IAM role not created
3. Trust policy incorrect
4. External ID mismatch

**Solution**: Check backend logs
```bash
docker-compose logs backend | grep "ERROR"
```

### Issue: No Resources Found
**Possible Causes**:
1. No resources in AWS account
2. IAM role lacks permissions
3. Resources in different regions

**Solution**: 
- Deploy Terraform resources
- Check IAM role has `ReadOnlyAccess`
- Verify regions configured

### Issue: 0 Findings Generated
**Expected Behavior**:
- First scan: 0 findings (no metrics yet)
- After 24h: Findings appear with metrics

**Solution**: Wait 24 hours, trigger second scan

### Issue: Backend Crash
**Solution**: Check logs for panic
```bash
docker-compose logs backend --tail=100
```
Restart backend:
```bash
docker-compose restart backend
```

---

## 📊 Expected Timeline

| Step | Duration | Cumulative |
|------|----------|------------|
| Signup + Login | 7 min | 7 min |
| Dashboard Check | 2 min | 9 min |
| AWS Onboarding | 10 min | 19 min |
| First Scan | 5 min | 24 min |
| Verify Resources | 5 min | 29 min |
| Check Findings | 5 min | 34 min |
| **Wait 24 hours** | - | - |
| Second Scan | 5 min | 39 min |
| Review Findings | 10 min | 49 min |
| Test IaC | 5 min | 54 min |
| Test Whitelist | 5 min | 59 min |
| Profile Check | 5 min | 64 min |

**Total Active Time**: ~1 hour (+ 24h wait for metrics)

---

## 🎉 Test Complete!

After completing all steps, you've validated:
- ✅ Complete user onboarding flow
- ✅ AWS integration working
- ✅ Resource discovery functional
- ✅ Cost optimization detectors operational
- ✅ All frontend pages dynamic
- ✅ End-to-end platform functionality

**Platform Status**: Production Ready! 🚀

---

## 🔄 Reset for Another Test

To run E2E test again:
```bash
psql -U yukti -d yukti_finops -f scripts/cleanup-for-e2e-test.sql
```

This clears all data and resets the platform for fresh testing.
