# Feature: Login Page

## Priority: HIGH (IMPLEMENTED ✅)

## What It Does
User authentication with email/password, JWT token generation, and multi-tenant support.

## Visual Reference
```
┌─────────────────────────────────────────────┐
│                                             │
│              YUKTI FINOPS                   │
│         Cloud Cost Optimization             │
│                                             │
│  ┌───────────────────────────────────────┐ │
│  │ Email                                 │ │
│  │ [yourname123@example.com          ] │ │
│  └───────────────────────────────────────┘ │
│                                             │
│  ┌───────────────────────────────────────┐ │
│  │ Password                              │ │
│  │ [••••••••••••••••••              ] │ │
│  └───────────────────────────────────────┘ │
│                                             │
│  [ ] Remember me                            │
│                                             │
│  ┌───────────────────────────────────────┐ │
│  │         LOGIN                         │ │
│  └───────────────────────────────────────┘ │
│                                             │
│  Don't have an account? Sign up             │
│                                             │
└─────────────────────────────────────────────┘
```

## User Flow
1. User opens http://localhost:3000/login
2. Enters email and password
3. Clicks "Login" button
4. System validates credentials
5. System generates JWT token (24-hour expiry)
6. System stores token in localStorage
7. Redirects to /dashboard

## Data Requirements

### Input
- `email` (string, required, email format)
- `password` (string, required, min 8 chars)

### Output (Success)
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 25,
    "email": "chandrakantpatil1594@gmail.com",
    "tenant_id": 27,
    "role": "owner"
  }
}
```

### Output (Error)
```json
{
  "error": "Invalid credentials"
}
```

## API Endpoints

### POST /api/v1/auth/login
**Request**:
```json
{
  "email": "yourname123@example.com",
  "password": "Chandra!@#$143"
}
```

**Response (200)**:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 25,
    "email": "chandrakantpatil1594@gmail.com",
    "tenant_id": 27,
    "role": "owner"
  }
}
```

**Response (401)**:
```json
{
  "error": "Invalid credentials"
}
```

## Database Tables

### yt_users
- `id` (serial, primary key)
- `email` (varchar, unique)
- `password_hash` (varchar)
- `tenant_id` (integer, foreign key)
- `email_verified` (boolean)
- `created_at` (timestamp)

### yt_tenant_users
- `id` (serial, primary key)
- `user_id` (integer, foreign key)
- `tenant_id` (integer, foreign key)
- `role` (varchar: owner, admin, editor, viewer)
- `created_at` (timestamp)

## UI Components

### Page
- **Path**: `/login`
- **File**: `frontend/src/pages/Login.tsx`

### Components Used
- React Hook Form (form validation)
- Tailwind CSS (styling)
- Lucide Icons (eye icon for password toggle)

## Business Rules
1. Email must be verified before login
2. Password must be hashed with bcrypt
3. JWT token expires after 24 hours
4. Failed login attempts logged for security
5. Token stored in localStorage as `token`
6. Auto-redirect to /dashboard on success
7. Show error message on invalid credentials

## Security Features
- ✅ Password hashing (bcrypt)
- ✅ JWT token with expiration
- ✅ Email verification required
- ✅ HTTPS only in production
- ✅ Rate limiting (5 attempts per minute)
- ✅ Audit logging (IP, timestamp, success/failure)

## Implementation Status
- ✅ Frontend: `frontend/src/pages/Login.tsx`
- ✅ Backend: `internal/api/handlers/auth.go` (Login handler)
- ✅ Middleware: `internal/api/middleware/jwt_auth.go`
- ✅ Database: `yt_users`, `yt_tenant_users` tables
- ✅ Testing: Manual testing complete
- ✅ Deployment: Live in Docker container

## Test Credentials
- **Email**: yourname123@example.com
- **Password**: Chandra!@#$143
- **Tenant ID**: 18
