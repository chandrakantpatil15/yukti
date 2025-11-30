# AWS Resource Scanning Troubleshooting Guide

## Quick Diagnosis

If your EC2 instances or other AWS resources are not showing in the Yukti dashboard, follow this step-by-step guide.

## Step 1: Check AWS Connection Status

### Dashboard Check
1. Go to your Yukti Dashboard
2. Look for the "AWS Connection" section at the top
3. Check the connection status indicator:
   - 🟢 **Connected** = Good to go
   - 🔴 **Disconnected** = Needs fixing

### If Disconnected
**Action Required**: Complete or fix your AWS onboarding
- Go to **Settings > AWS Connection**
- Follow the setup instructions
- Verify your IAM role configuration

## Step 2: Verify IAM Role Setup

### Required IAM Permissions
Your IAM role must have these permissions:
```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "ec2:DescribeInstances",
                "ec2:DescribeVolumes",
                "rds:DescribeDBInstances",
                "s3:ListAllMyBuckets",
                "s3:GetBucketLocation"
            ],
            "Resource": "*"
        }
    ]
}
```

### Trust Policy Check
Your IAM role trust policy should look like this:
```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Principal": {
                "AWS": "arn:aws:iam::144403604430:user/yukti-platform-user"
            },
            "Action": "sts:AssumeRole",
            "Condition": {
                "StringLike": {
                    "sts:ExternalId": "yukti-*"
                }
            }
        }
    ]
}
```

## Step 3: Trigger Manual Scan

### From Dashboard
1. Go to your Dashboard
2. Look for the "AWS Connection" section
3. Click **"Scan Resources"** button
4. Wait 30-60 seconds for results

### Check Scan Logs
If scan fails, check the backend logs:
```bash
# View real-time logs
docker-compose logs -f backend

# Look for scan-related messages
docker-compose logs backend | grep -i scanner
```

## Step 4: Common Issues & Solutions

### Issue: "No AWS connection configured"
**Cause**: Onboarding not completed
**Solution**: 
1. Go to **Settings > AWS Connection**
2. Enter your AWS Account ID and Role Name
3. Complete verification process

### Issue: "AWS connection not verified"
**Cause**: IAM role setup incorrect
**Solution**:
1. Check IAM role trust policy (see above)
2. Verify external ID matches
3. Re-verify connection in Yukti

### Issue: "Failed to assume role"
**Cause**: Permission or trust policy issues
**Solutions**:
- Verify Yukti account ID: `144403604430`
- Check external ID format: `yukti-{tenant_id}-{random}`
- Ensure role ARN is correct
- Test role assumption manually:
```bash
aws sts assume-role \
  --role-arn "arn:aws:iam::YOUR_ACCOUNT:role/YuktiFinOpsRole" \
  --role-session-name "test-session" \
  --external-id "YOUR_EXTERNAL_ID"
```

### Issue: "No resources found"
**Possible Causes**:
1. **Wrong Region**: Resources in different region than scanned
2. **No Resources**: Account actually has no EC2/RDS/S3 resources
3. **Permissions**: Missing describe permissions

**Solutions**:
1. Check which region your resources are in
2. Verify resources exist in AWS Console
3. Add missing IAM permissions

### Issue: Resources exist but not detected
**Debugging Steps**:
1. Check resource state (stopped instances won't be optimized)
2. Verify resource tags and metadata
3. Check if resources are in supported regions

## Step 5: Enable Debug Logging

### View Detailed Scan Logs
```bash
# Real-time scanning logs
docker-compose logs -f backend | grep -E "(Scanner|ScanAPI)"

# Filter for your tenant ID (replace 25 with your tenant ID)
docker-compose logs backend | grep "tenant.*25"
```

### Log Messages to Look For
- `[Scanner] Starting scan for tenant: X`
- `[Scanner] Found EC2 instance: i-xxxxx`
- `[Scanner] Found X resources for tenant Y`
- `[Scanner] Found X cost optimization opportunities`

## Step 6: Manual Verification

### Test AWS CLI Access
Use the same credentials to test access:
```bash
# Test EC2 access
aws ec2 describe-instances --region us-east-1

# Test RDS access  
aws rds describe-db-instances --region us-east-1

# Test S3 access
aws s3 ls
```

### Check Resource Regions
Ensure your resources are in the region being scanned:
```bash
# List instances by region
aws ec2 describe-instances --region us-east-1 --query 'Reservations[].Instances[].{ID:InstanceId,Type:InstanceType,State:State.Name}'
aws ec2 describe-instances --region us-west-2 --query 'Reservations[].Instances[].{ID:InstanceId,Type:InstanceType,State:State.Name}'
```

## Step 7: Contact Support

If issues persist, provide these details:

### Information to Collect
1. **Tenant ID**: Found in dashboard URL or user profile
2. **AWS Account ID**: Your 12-digit AWS account number
3. **IAM Role ARN**: Full ARN of your Yukti role
4. **Error Messages**: From dashboard or logs
5. **Resource Details**: What resources you expect to see

### Log Collection
```bash
# Collect recent logs
docker-compose logs --tail=100 backend > yukti-logs.txt

# Collect scan-specific logs
docker-compose logs backend | grep -E "(Scanner|ScanAPI)" > scan-logs.txt
```

### Support Channels
- **Email**: support@yukti.io
- **Documentation**: Check latest docs for updates
- **Status Page**: Check for known issues

## Expected Scan Results

### Typical Scan Duration
- **Small accounts** (< 50 resources): 10-30 seconds
- **Medium accounts** (50-200 resources): 30-60 seconds  
- **Large accounts** (200+ resources): 1-3 minutes

### What Gets Scanned
- **EC2 Instances**: All states (running, stopped, terminated)
- **RDS Instances**: All database engines
- **S3 Buckets**: All buckets (global service)
- **EBS Volumes**: Attached and unattached
- **Load Balancers**: ALB, NLB, CLB

### Expected Findings
After successful scan, you should see:
- Cost optimization opportunities
- Right-sizing recommendations  
- Unused resource alerts
- Storage optimization suggestions
- Reserved Instance recommendations

---

## Quick Reference

### Scan Button Location
Dashboard → AWS Connection section → "Scan Resources" button

### Log Commands
```bash
# View all logs
docker-compose logs -f backend

# Scan-specific logs
docker-compose logs backend | grep Scanner

# Your tenant logs (replace 25)
docker-compose logs backend | grep "tenant.*25"
```

### Key Files
- IAM Role: `YuktiFinOpsRole` (or your custom name)
- Trust Policy: Must include Yukti account `144403604430`
- Permissions: EC2, RDS, S3 describe/list actions

### Support Info
- Yukti Account ID: `144403604430`
- Platform User: `yukti-platform-user`
- External ID Pattern: `yukti-{tenant_id}-{random}`