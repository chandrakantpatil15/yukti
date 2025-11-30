# Test AWS Role Setup

## Quick Test (Same Account)

### 1. Create IAM Role in AWS Console
- Go to IAM → Roles → Create Role
- Select "AWS Account" → "This account"
- Role name: `YuktiTestRole`
- Add policy: `ReadOnlyAccess`

### 2. Add Trust Policy
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
        "sts:ExternalId": "test-external-id-123"
      }
    }
  }]
}
```

### 3. Test in Onboarding UI
- Account ID: YOUR_ACCOUNT_ID (12 digits)
- Role Name: YuktiTestRole
- Backend will auto-generate external ID
- **ISSUE**: External ID won't match trust policy

## Better Test: Mock Mode

Add environment variable to skip real AWS verification during development.
