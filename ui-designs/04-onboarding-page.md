# Feature: Onboarding Page

## Priority: HIGH (IMPLEMENTED ✅)

## What It Does
Guides new users through AWS account connection setup with IAM role configuration and real-time verification.

## Visual Reference
```
┌─────────────────────────────────────────────────────────────┐
│  YUKTI FINOPS - AWS ONBOARDING                               │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  Connect Your AWS Account                                    │
│  Follow these steps to grant Yukti read-only access          │
│                                                               │
│  ┌─────────────────────────────────────────────────────┐    │
│  │  STEP 1: Create IAM Role in AWS Console            │    │
│  │                                                      │    │
│  │  1. Log in to AWS Console                           │    │
│  │  2. Go to IAM → Roles → Create Role                 │    │
│  │  3. Select "Another AWS Account"                    │    │
│  │  4. Enter Yukti Account ID: 144403604430            │    │
│  │  5. Check "Require external ID"                     │    │
│  │  6. Copy this External ID:                          │    │
│  │     yukti-27-a8f3d9e2b1c4  [Copy]                  │    │
│  │  7. Attach policy: ReadOnlyAccess                   │    │
│  │  8. Name role: YuktiFinOpsRole                      │    │
│  │  9. Copy Role ARN                                   │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                               │
│  ┌─────────────────────────────────────────────────────┐    │
│  │  STEP 2: Enter AWS Connection Details              │    │
│  │                                                      │    │
│  │  AWS Account ID *                                   │    │
│  │  ┌───────────────────────────────────────────────┐ │    │
│  │  │ 424851482219                                  │ │    │
│  │  └───────────────────────────────────────────────┘ │    │
│  │  (12-digit AWS account number)                     │    │
│  │                                                      │    │
│  │  IAM Role ARN *                                     │    │
│  │  ┌───────────────────────────────────────────────┐ │    │
│  │  │ arn:aws:iam::424851482219:role/YuktiFinOps   │ │    │
│  │  └───────────────────────────────────────────────┘ │    │
│  │  (Full ARN from AWS Console)                       │    │
│  │                                                      │    │
│  │  ┌───────────────────────────────────────────────┐ │    │
│  │  │         VERIFY & CONNECT                      │ │    │
│  │  └───────────────────────────────────────────────┘ │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                               │
│  ┌─────────────────────────────────────────────────────┐    │
│  │  ✅ Connection Verified!                            │    │
│  │  Successfully connected to AWS Account              │    │
│  │  424851482219                                       │    │
│  │                                                      │    │
│  │  Redirecting to dashboard in 3 seconds...          │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                               │
└─────────────────────────────────────────────────────────────┘
```

## User Flow
1. User completes signup/login
2. System checks if AWS connection exists
3. If no connection, redirect to /onboarding
4. User sees 2-step onboarding wizard
5. Step 1: Shows Yukti Account ID + External ID (auto-generated)
6. User creates IAM role in AWS Console
7. Step 2: User enters AWS Account ID + Role ARN
8. User clicks "Verify & Connect"
9. Backend validates Account ID format (12 digits)
10. Backend validates Role ARN format
11. Backend attempts STS AssumeRole with external ID
12. Backend verifies credentials with GetCallerIdentity
13. If successful, saves connection with verified=true
14. Shows success message
15. Auto-redirects to /dashboard after 3 seconds

## Data Requirements

### Input
- `account_id` (string, required, 12 digits)
- `role_arn` (string, required, ARN format)

### Output (Success)
```json
{
  "message": "AWS connection verified and saved successfully",
  "connection": {
    "account_id": "424851482219",
    "role_arn": "arn:aws:iam::424851482219:role/YuktiFinOpsRole",
    "external_id": "yukti-27-a8f3d9e2b1c4",
    "verified": true,
    "last_verified_at": "2025-01-31T10:30:00Z"
  }
}
```

### Output (Error)
```json
{
  "error": "ACCESS_DENIED",
  "message": "Failed to assume IAM role. Please check trust policy.",
  "details": "Ensure trust policy includes Yukti Account ID (144403604430) and correct external ID."
}
```

## API Endpoints

### GET /api/v1/onboarding/external-id
**Headers**:
```
Authorization: Bearer <JWT_TOKEN>
```

**Response (200)**:
```json
{
  "external_id": "yukti-27-a8f3d9e2b1c4",
  "yukti_account_id": "144403604430"
}
```

### POST /api/v1/onboarding/aws-connection
**Headers**:
```
Authorization: Bearer <JWT_TOKEN>
```

**Request**:
```json
{
  "account_id": "424851482219",
  "role_arn": "arn:aws:iam::424851482219:role/YuktiFinOpsRole"
}
```

**Response (200)**:
```json
{
  "message": "AWS connection verified and saved successfully"
}
```

**Response (400)**:
```json
{
  "error": "INVALID_ACCOUNT_ID",
  "message": "AWS Account ID must be exactly 12 digits"
}
```

**Response (403)**:
```json
{
  "error": "ACCESS_DENIED",
  "message": "Failed to assume IAM role. Please check trust policy.",
  "details": "Ensure trust policy includes Yukti Account ID (144403604430) and correct external ID."
}
```

### GET /api/v1/onboarding/aws-connection
**Headers**:
```
Authorization: Bearer <JWT_TOKEN>
```

**Response (200)**:
```json
{
  "connected": true,
  "account_id": "424851482219",
  "role_arn": "arn:aws:iam::424851482219:role/YuktiFinOpsRole",
  "verified": true,
  "last_verified_at": "2025-01-31T10:30:00Z"
}
```

**Response (404)**:
```json
{
  "connected": false
}
```

## Database Tables

### yt_aws_connections
- `id` (serial, primary key)
- `tenant_id` (integer, unique)
- `account_id` (varchar, 12 chars)
- `role_arn` (varchar)
- `external_id` (varchar, auto-generated)
- `verified` (boolean)
- `last_verified_at` (timestamp)
- `created_at` (timestamp)
- `updated_at` (timestamp)

## UI Components

### Page
- **Path**: `/onboarding`
- **File**: `frontend/src/pages/Onboarding.tsx`

### Components Used
- OnboardingGuard (redirects if already connected)
- StepIndicator (shows current step)
- CopyButton (copy external ID to clipboard)
- ValidationMessage (shows errors)
- SuccessAlert (shows success message)

## Business Rules
1. External ID auto-generated: `yukti-{tenant_id}-{random_12_chars}`
2. Account ID must be exactly 12 digits
3. Role ARN must match format: `arn:aws:iam::{account_id}:role/{role_name}`
4. Backend verifies role assumption before saving
5. Connection saved only if verification succeeds
6. User cannot access dashboard without AWS connection
7. OnboardingGuard checks connection status on all protected routes

## Security Features
- ✅ External ID prevents confused deputy attack
- ✅ Backend-generated external ID (not user-provided)
- ✅ Real-time role verification (STS AssumeRole)
- ✅ Credentials validation (GetCallerIdentity)
- ✅ Tenant isolation (one connection per tenant)
- ✅ Audit logging (connection attempts tracked)

## Error Handling

### Error Types
1. **INVALID_ACCOUNT_ID**: Account ID not 12 digits
2. **INVALID_ARN**: Role ARN format incorrect
3. **ACCESS_DENIED**: Trust policy missing/incorrect
4. **INVALID_EXTERNAL_ID**: External ID mismatch
5. **ROLE_NOT_FOUND**: Role doesn't exist
6. **NETWORK_ERROR**: AWS API unreachable

### User-Friendly Messages
- ACCESS_DENIED → "Check trust policy includes Yukti Account ID"
- INVALID_EXTERNAL_ID → "Use exact external ID shown above"
- ROLE_NOT_FOUND → "Verify role ARN is correct"
- INVALID_ARN → "ARN must start with arn:aws:iam::"
- NETWORK_ERROR → "Check internet connection"

## Implementation Status
- ✅ Frontend: `frontend/src/pages/Onboarding.tsx`
- ✅ Frontend: `frontend/src/components/Auth/OnboardingGuard.tsx`
- ✅ Backend: `internal/api/handlers/onboarding.go`
- ✅ Backend: `internal/aws/role_verifier.go` (STS verification)
- ✅ Database: `yt_aws_connections` table
- ✅ Testing: Manual testing complete with real AWS account
- ✅ Deployment: Live in Docker container

## Test Data
- **Yukti Account ID**: 144403604430
- **Customer Account ID**: 424851482219
- **Role ARN**: arn:aws:iam::424851482219:role/YuktiFinOpsRole
- **External ID**: yukti-27-a8f3d9e2b1c4 (auto-generated)

## Future Enhancements
- Add CloudFormation template download (one-click setup)
- Add video tutorial (embedded YouTube)
- Add trust policy validator (check before submission)
- Add multi-account support (connect multiple AWS accounts)
- Add connection health check (periodic re-verification)
