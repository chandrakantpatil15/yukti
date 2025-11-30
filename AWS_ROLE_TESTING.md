# AWS Role Verification Testing Guide

## Current Setup

**Development Mode**: `SKIP_AWS_VERIFICATION=true` (in docker-compose.yml)
- AWS verification is SKIPPED
- Any Account ID + Role Name will be accepted
- Allows testing without real AWS setup

**Production Mode**: `SKIP_AWS_VERIFICATION=false` or unset
- Real AWS STS AssumeRole verification
- Requires proper IAM role setup

---

## Testing Options

### Option 1: Development Mode (Current - EASIEST)

**Status**: ✅ Already configured

**How it works**:
- Backend skips AWS verification
- Accepts any Account ID (12 digits) + Role Name
- Saves connection immediately
- Perfect for UI/UX testing

**Test now**:
1. Go to http://localhost:3000/onboarding
2. Enter any 12-digit Account ID: `123456789012`
3. Enter any Role Name: `YuktiReadOnlyRole`
4. Click "Connect AWS Account"
5. ✅ Should succeed immediately

**Logs to watch**:
```bash
docker-compose logs -f backend | grep "Skipping AWS verification"
```

---

### Option 2: Production Mode (Real AWS Verification)

**Requirements**:
1. Yukti Platform AWS Account (separate from personal)
2. Customer creates IAM role in their account
3. Role trusts Yukti's AWS account

**Setup Steps**:

#### Step 1: Get Yukti Platform AWS Account
- Create new AWS account for Yukti platform
- Get Account ID (e.g., `999888777666`)
- Create IAM user with STS permissions
- Get Access Key + Secret Key

#### Step 2: Customer Creates IAM Role

**In Customer's AWS Console**:
1. IAM → Roles → Create Role
2. Select "AWS Account" → "Another AWS account"
3. Enter Yukti's Account ID: `999888777666`
4. Check "Require external ID"
5. Enter external ID: `yukti-18-abc123xyz456` (from backend)
6. Role name: `YuktiReadOnlyRole`
7. Attach policy: `ReadOnlyAccess`

**Trust Policy**:
```json
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {
      "AWS": "arn:aws:iam::999888777666:root"
    },
    "Action": "sts:AssumeRole",
    "Condition": {
      "StringEquals": {
        "sts:ExternalId": "yukti-18-abc123xyz456"
      }
    }
  }]
}
```

#### Step 3: Update Backend Configuration

**In docker-compose.yml**:
```yaml
environment:
  SKIP_AWS_VERIFICATION: false  # Enable real verification
  AWS_ACCESS_KEY_ID: <YUKTI_PLATFORM_KEY>
  AWS_SECRET_ACCESS_KEY: <YUKTI_PLATFORM_SECRET>
  AWS_REGION: us-east-1
```

#### Step 4: Test Onboarding
1. Rebuild backend: `docker-compose up -d --build backend`
2. Go to onboarding page
3. Enter customer's Account ID: `123456789012`
4. Enter role name: `YuktiReadOnlyRole`
5. Backend will:
   - Generate external ID: `yukti-18-abc123xyz456`
   - Call STS AssumeRole
   - Verify credentials
   - Save connection if successful

**Success Response**:
```json
{
  "verified": true,
  "message": "AWS connection verified and configured successfully!"
}
```

**Error Response** (if trust policy wrong):
```json
{
  "success": false,
  "verified": false,
  "error": "Access denied. Please check the trust policy on your IAM role.",
  "error_code": "ACCESS_DENIED",
  "error_details": "The IAM role trust policy must allow Yukti's AWS account..."
}
```

---

### Option 3: Same-Account Testing (Quick Test)

**For quick testing without separate accounts**:

1. Create role in YOUR AWS account that trusts itself
2. Use your AWS credentials in backend
3. Test AssumeRole within same account

**IAM Role Setup**:
```json
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {
      "AWS": "arn:aws:iam::YOUR_ACCOUNT_ID:root"
    },
    "Action": "sts:AssumeRole",
    "Condition": {
      "StringEquals": {
        "sts:ExternalId": "test-123"
      }
    }
  }]
}
```

**Limitation**: External ID won't match (backend auto-generates it)

---

## Recommended Approach

### For Development (Now):
✅ **Use Option 1** - Development mode with `SKIP_AWS_VERIFICATION=true`
- Test UI/UX flow
- Test database operations
- Test error handling
- No AWS setup needed

### For Production (Later):
✅ **Use Option 2** - Real AWS verification
- Create Yukti Platform AWS account
- Update backend credentials
- Test with real customer roles
- Full security validation

---

## Current Status

- ✅ Development mode enabled
- ✅ AWS verification code implemented
- ✅ Error handling with 6 error types
- ✅ External ID auto-generation
- ⏸️ Real AWS testing (waiting for Yukti Platform account)

---

## Testing Commands

```bash
# Rebuild backend with changes
docker-compose up -d --build backend

# Watch verification logs
docker-compose logs -f backend | grep -E "Verifying|Skipping|verification"

# Test onboarding API directly
curl -X POST http://localhost:8081/api/onboarding/aws-connection \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "18",
    "account_id": "123456789012",
    "role_arn": "arn:aws:iam::123456789012:role/YuktiReadOnlyRole",
    "regions": ["us-east-1"]
  }'
```

---

## Next Steps

1. ✅ Test onboarding in dev mode (SKIP_AWS_VERIFICATION=true)
2. ⏸️ Create Yukti Platform AWS account
3. ⏸️ Update backend with Yukti credentials
4. ⏸️ Test with real customer role
5. ⏸️ Document customer setup instructions
