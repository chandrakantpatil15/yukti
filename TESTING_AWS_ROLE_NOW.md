# Test AWS Role Verification NOW

## ✅ Setup Complete

**Development mode enabled**: AWS verification is SKIPPED for easy testing

## How to Test (3 minutes)

### Step 1: Open Onboarding Page
```
http://localhost:3000/onboarding
```

### Step 2: Enter Test Data
- **AWS Account ID**: `123456789012` (any 12 digits)
- **Role Name**: `YuktiReadOnlyRole` (any name)

### Step 3: Click "Connect AWS Account"

**Expected Result**: ✅ Success message
```
AWS connection verified and configured successfully!
```

### Step 4: Watch Backend Logs
```bash
docker-compose logs -f backend | grep -E "Skipping|Verifying"
```

**You should see**:
```
[INFO] Skipping AWS verification (dev mode) for tenant 18
```

---

## What's Happening Behind the Scenes

1. **Frontend** sends: Account ID + Role Name
2. **Backend** auto-generates: External ID (`yukti-18-abc123xyz456`)
3. **Backend** constructs: Role ARN (`arn:aws:iam::123456789012:role/YuktiReadOnlyRole`)
4. **Backend** checks: `SKIP_AWS_VERIFICATION=true`
5. **Backend** skips: AWS STS AssumeRole call
6. **Backend** saves: Connection to database with `verified=true`

---

## Database Check

After successful onboarding, check the database:

```bash
# Access PostgreSQL
psql -U yukti -d yukti_finops

# Check AWS connection
SELECT tenant_id, account_id, role_arn, external_id, verified, last_verified_at 
FROM yt_aws_connections 
WHERE tenant_id = '18';
```

**Expected Output**:
```
tenant_id | account_id   | role_arn                                          | external_id           | verified | last_verified_at
----------|--------------|---------------------------------------------------|----------------------|----------|------------------
18        | 123456789012 | arn:aws:iam::123456789012:role/YuktiReadOnlyRole | yukti-18-abc123xyz456| true     | 2025-11-24 11:50:00
```

---

## For Production (Later)

When ready for real AWS verification:

1. **Create Yukti Platform AWS Account**
2. **Update docker-compose.yml**:
   ```yaml
   SKIP_AWS_VERIFICATION: false  # Enable real verification
   AWS_ACCESS_KEY_ID: <YUKTI_PLATFORM_KEY>
   AWS_SECRET_ACCESS_KEY: <YUKTI_PLATFORM_SECRET>
   ```
3. **Customer creates IAM role** with trust policy
4. **Test with real AWS account**

See `AWS_ROLE_TESTING.md` for detailed production setup.

---

## Summary

✅ **Development Mode**: Test onboarding flow without AWS setup  
✅ **Production Mode**: Real AWS STS verification (when ready)  
✅ **Code Ready**: All verification logic implemented  
✅ **Error Handling**: 6 error types with user-friendly messages  

**Test it now**: http://localhost:3000/onboarding
