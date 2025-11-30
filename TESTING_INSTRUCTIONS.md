# Testing AWS Resource Scanning - Step by Step

## Current Status ✅
- **Backend**: Enhanced logging deployed (port 8081)
- **Frontend**: Scan troubleshooting deployed (port 3000)
- **User Account**: chandrakantpatil1594@gmail.com (tenant_id: 25)
- **AWS Account**: 424851482219 with $100 credit

## Step 1: Access Dashboard
1. Go to: http://localhost:3000
2. Login with: chandrakantpatil1594@gmail.com
3. Navigate to Dashboard

## Step 2: Check AWS Connection Status
Look for the "AWS Connection" section at the top of dashboard:
- **🟢 Connected**: Ready to scan
- **🔴 Disconnected**: Need to fix connection

## Step 3: Trigger Resource Scan
1. Click **"Scan Resources"** button
2. Watch for success/error messages
3. Check console logs (F12 → Console tab)

## Step 4: Monitor Scan Progress
### Real-time Backend Logs
```bash
# Terminal 1: Watch all scan activity
docker-compose logs -f backend | grep -E "(Scanner|ScanAPI)"

# Terminal 2: Watch your specific tenant (25)
docker-compose logs -f backend | grep "tenant.*25"
```

### Expected Log Messages
```
[ScanAPI] ========== SCAN REQUEST RECEIVED ===========
[ScanAPI] Tenant ID: 25
[Scanner] ========== STARTING AWS SCAN ===========
[Scanner] AWS Connection Details:
[Scanner]   Account ID: 424851482219
[Scanner]   Verified: true
[Scanner] ✓ Successfully assumed IAM role
[Scanner] Fetching AWS resources...
[Scanner] Found EC2 instance: i-xxxxx (t3.large) - running
[Scanner] ✓ Found X EC2 instances
[Scanner] ✓ Found X RDS instances  
[Scanner] ✓ Found X S3 buckets
[Scanner] ✓ Successfully fetched X resources
[Scanner] ✓ Detectors completed successfully
[Scanner] Found X cost optimization opportunities
[Scanner] ========== SCAN COMPLETED ===========
```

## Step 5: Debug Issues (If Scan Fails)

### Use Debug Button
1. Click **"Debug"** button next to "Scan Resources"
2. Check scan status and troubleshooting info
3. Look for specific error codes:
   - `NO_AWS_CONNECTION`: Complete onboarding
   - `AWS_NOT_VERIFIED`: Fix IAM role setup
   - `DATABASE_ERROR`: Check database connection

### Common Issues & Solutions

#### Issue: "No AWS connection configured"
**Fix**: Go to Settings → AWS Connection and complete setup

#### Issue: "AWS connection not verified"  
**Fix**: Check IAM role trust policy:
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

#### Issue: "Failed to assume role"
**Check**:
1. IAM role exists: `YuktiFinOpsRole`
2. Trust policy includes Yukti account: `144403604430`
3. External ID matches pattern: `yukti-25-xxxxxxxxxxxx`

#### Issue: "No resources found"
**Possible Causes**:
1. Resources in different region (scan checks us-east-1 by default)
2. No EC2/RDS/S3 resources in account
3. Missing IAM permissions

## Step 6: Verify Results
After successful scan:
1. **Dashboard**: Should show updated resource counts
2. **Hidden Costs**: Should show new findings
3. **Database**: Check findings table
```bash
docker exec -it yukti-postgres psql -U yukti -d yukti_finops -c "SELECT COUNT(*) FROM yt_hidden_cost_findings WHERE tenant_id = '25';"
```

## Step 7: Test with Your EC2 Instance

### Verify EC2 Instance Exists
```bash
# Check if your EC2 instance is running
aws ec2 describe-instances --region us-east-1 --query 'Reservations[].Instances[].{ID:InstanceId,Type:InstanceType,State:State.Name}'
```

### Expected Scan Results
If you have a t3.large EC2 instance, scan should find:
- **Right-sizing opportunity**: t3.large → t3.medium ($60/month savings)
- **Spot instance opportunity**: 70% savings for dev workloads
- **Detailed monitoring**: Disable if not needed

## Troubleshooting Commands

### Check Service Status
```bash
docker-compose ps
```

### View All Logs
```bash
docker-compose logs -f
```

### Scan-Specific Logs
```bash
# All scan activity
docker-compose logs backend | grep -E "(Scanner|ScanAPI)"

# Your tenant only
docker-compose logs backend | grep "tenant.*25"

# Recent logs only
docker-compose logs --tail=50 backend
```

### Test API Directly
```bash
# Check scan status
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" http://localhost:8081/api/v1/scan/status

# Trigger scan
curl -X POST -H "Authorization: Bearer YOUR_JWT_TOKEN" http://localhost:8081/api/v1/scan
```

## Expected Timeline
- **Scan trigger**: Immediate response
- **Resource discovery**: 10-30 seconds
- **Detector analysis**: 5-15 seconds
- **Results in dashboard**: 30-60 seconds total

## Success Indicators
✅ **Scan button works** (no errors)
✅ **Backend logs show scan progress**
✅ **Resources discovered** (EC2/RDS/S3 counts)
✅ **Findings generated** (cost optimization opportunities)
✅ **Dashboard updates** (new resource counts/savings)

## If Issues Persist
Collect this information:
1. **Error messages** from dashboard alerts
2. **Console logs** (F12 → Console)
3. **Backend logs** (docker-compose logs backend)
4. **Scan status** (Debug button output)
5. **AWS resources** (what you expect to see)

Then we can debug together! 🚀