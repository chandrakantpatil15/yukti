# Feature: Admin Portal

## Priority: HIGH (IMPLEMENTED ✅)

## What It Does
Platform admin dashboard for managing tenants, users, impersonation, and analytics.

## Visual Reference
```
┌─────────────────────────────────────────────────────────────┐
│  YUKTI ADMIN PORTAL                          [Logout]        │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  Platform Overview                                           │
│                                                               │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │   Tenants   │  │    Users    │  │  Resources  │         │
│  │     127     │  │     543     │  │   45,892    │         │
│  │   ↑ 12%    │  │   ↑ 8%     │  │   ↑ 15%    │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
│                                                               │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │  Findings   │  │   Savings   │  │ Active Scans│         │
│  │   12,450    │  │  $1.2M/mo   │  │      23     │         │
│  │   ↑ 20%    │  │   ↑ 18%    │  │   → 0%     │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
│                                                               │
│  Quick Actions                                               │
│  ┌─────────────────────────────────────────────────────┐    │
│  │  📊 Tenant Management                               │    │
│  │  View, suspend, activate, delete tenants            │    │
│  │  [Go to Tenants]                                    │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                               │
│  ┌─────────────────────────────────────────────────────┐    │
│  │  👥 User Management                                 │    │
│  │  View, suspend, activate users                      │    │
│  │  [Go to Users]                                      │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                               │
│  ┌─────────────────────────────────────────────────────┐    │
│  │  📈 Analytics                                       │    │
│  │  Platform growth, revenue, usage metrics            │    │
│  │  [Go to Analytics]                                  │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                               │
└─────────────────────────────────────────────────────────────┘
```

## User Flow

### Admin Login
1. Admin opens http://localhost:3000/admin/login
2. Enters admin email and password
3. System validates admin credentials
4. System generates admin JWT token (24-hour expiry)
5. Redirects to /admin/dashboard

### Tenant Management
1. Admin clicks "Tenant Management"
2. Sees list of all tenants with stats
3. Can search/filter tenants
4. Can view tenant details (users, resources, findings)
5. Can suspend/activate/delete tenant
6. Can impersonate tenant owner

### User Management
1. Admin clicks "User Management"
2. Sees list of all users across all tenants
3. Can search/filter users
4. Can view user details (tenant associations, role)
5. Can suspend/activate user
6. Can reset user password

### Impersonation
1. Admin clicks "Impersonate" on tenant/user
2. Modal opens asking for reason
3. Admin enters reason (required)
4. System creates impersonation session
5. System generates user JWT token
6. Yellow banner appears: "Impersonating user@example.com"
7. Admin can access user's dashboard/resources
8. Admin clicks "End Impersonation" to return

## Data Requirements

### Admin Login Input
```json
{
  "email": "admin@yukti.com",
  "password": "AdminPassword123!"
}
```

### Admin Login Output
```json
{
  "admin_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "admin": {
    "id": 1,
    "email": "admin@yukti.com",
    "role": "super_admin"
  }
}
```

### Platform Stats Output
```json
{
  "total_tenants": 127,
  "total_users": 543,
  "total_resources": 45892,
  "total_findings": 12450,
  "total_savings": 1200000.00,
  "active_scans": 23,
  "growth": {
    "tenants": 12,
    "users": 8,
    "resources": 15,
    "findings": 20
  }
}
```

### Tenant List Output
```json
{
  "tenants": [
    {
      "id": 1,
      "tenant_id": 18,
      "company_name": "Acme Corp",
      "status": "active",
      "user_count": 5,
      "resource_count": 847,
      "finding_count": 7,
      "monthly_savings": 425.60,
      "created_at": "2025-01-01T00:00:00Z"
    }
  ]
}
```

### Impersonation Input
```json
{
  "user_id": 25,
  "reason": "Customer support request - investigating dashboard issue"
}
```

### Impersonation Output
```json
{
  "impersonation_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "session_id": 123,
  "user": {
    "id": 25,
    "email": "chandrakantpatil1594@gmail.com",
    "tenant_id": 27
  },
  "expires_at": "2025-01-31T11:30:00Z"
}
```

## API Endpoints

### POST /api/admin/login
**Request**:
```json
{
  "email": "admin@yukti.com",
  "password": "AdminPassword123!"
}
```

**Response (200)**:
```json
{
  "admin_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "admin": {
    "id": 1,
    "email": "admin@yukti.com",
    "role": "super_admin"
  }
}
```

### GET /api/admin/stats
**Headers**:
```
Authorization: Bearer <ADMIN_TOKEN>
```

**Response (200)**:
```json
{
  "total_tenants": 127,
  "total_users": 543,
  "total_resources": 45892,
  "total_findings": 12450,
  "total_savings": 1200000.00,
  "active_scans": 23
}
```

### GET /api/admin/tenants
**Headers**:
```
Authorization: Bearer <ADMIN_TOKEN>
```

**Query Params**:
- `status` (optional: active, suspended, deleted)
- `search` (optional: company name)

**Response (200)**:
```json
{
  "tenants": [
    {
      "id": 1,
      "tenant_id": 18,
      "company_name": "Acme Corp",
      "status": "active",
      "user_count": 5,
      "resource_count": 847,
      "finding_count": 7,
      "monthly_savings": 425.60
    }
  ]
}
```

### GET /api/admin/tenants/:id
**Headers**:
```
Authorization: Bearer <ADMIN_TOKEN>
```

**Response (200)**:
```json
{
  "id": 1,
  "tenant_id": 18,
  "company_name": "Acme Corp",
  "status": "active",
  "users": [...],
  "resources": [...],
  "findings": [...],
  "aws_connection": {...}
}
```

### POST /api/admin/tenants/:id/suspend
**Headers**:
```
Authorization: Bearer <ADMIN_TOKEN>
```

**Response (200)**:
```json
{
  "message": "Tenant suspended successfully"
}
```

### POST /api/admin/tenants/:id/activate
**Headers**:
```
Authorization: Bearer <ADMIN_TOKEN>
```

**Response (200)**:
```json
{
  "message": "Tenant activated successfully"
}
```

### DELETE /api/admin/tenants/:id
**Headers**:
```
Authorization: Bearer <ADMIN_TOKEN>
```

**Response (200)**:
```json
{
  "message": "Tenant deleted successfully"
}
```

### GET /api/admin/users
**Headers**:
```
Authorization: Bearer <ADMIN_TOKEN>
```

**Response (200)**:
```json
{
  "users": [
    {
      "id": 25,
      "email": "chandrakantpatil1594@gmail.com",
      "tenant_count": 1,
      "status": "active",
      "created_at": "2025-01-01T00:00:00Z"
    }
  ]
}
```

### POST /api/admin/users/:id/suspend
**Headers**:
```
Authorization: Bearer <ADMIN_TOKEN>
```

**Response (200)**:
```json
{
  "message": "User suspended successfully"
}
```

### POST /api/admin/impersonate
**Headers**:
```
Authorization: Bearer <ADMIN_TOKEN>
```

**Request**:
```json
{
  "user_id": 25,
  "reason": "Customer support request"
}
```

**Response (200)**:
```json
{
  "impersonation_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "session_id": 123,
  "expires_at": "2025-01-31T11:30:00Z"
}
```

### POST /api/admin/end-impersonation
**Headers**:
```
Authorization: Bearer <ADMIN_TOKEN>
```

**Response (200)**:
```json
{
  "message": "Impersonation session ended"
}
```

### GET /api/admin/analytics
**Headers**:
```
Authorization: Bearer <ADMIN_TOKEN>
```

**Response (200)**:
```json
{
  "growth_metrics": {
    "new_tenants_30d": 12,
    "new_users_30d": 45,
    "total_resources": 45892,
    "total_findings": 12450,
    "avg_savings_per_tenant": 9448.82
  }
}
```

## Database Tables

### yt_admin_users
- `id` (serial, primary key)
- `email` (varchar, unique)
- `password_hash` (varchar)
- `role` (varchar: super_admin, support, analyst)
- `last_login` (timestamp)
- `created_at` (timestamp)

### yt_impersonation_sessions
- `id` (serial, primary key)
- `admin_id` (integer, foreign key)
- `user_id` (integer, foreign key)
- `reason` (text, required)
- `started_at` (timestamp)
- `ended_at` (timestamp)
- `is_active` (boolean)

### yt_admin_audit_logs
- `id` (serial, primary key)
- `admin_id` (integer, foreign key)
- `action` (varchar: suspend_tenant, activate_user, impersonate, etc.)
- `target_type` (varchar: tenant, user)
- `target_id` (integer)
- `details` (jsonb)
- `ip_address` (varchar)
- `created_at` (timestamp)

## UI Components

### Pages
- **Path**: `/admin/login` - `frontend/src/pages/Admin/AdminLogin.tsx`
- **Path**: `/admin/dashboard` - `frontend/src/pages/Admin/AdminDashboard.tsx`
- **Path**: `/admin/tenants` - `frontend/src/pages/Admin/AdminTenants.tsx`
- **Path**: `/admin/users` - `frontend/src/pages/Admin/AdminUsers.tsx`
- **Path**: `/admin/analytics` - `frontend/src/pages/Admin/AdminAnalytics.tsx`

### Components
- `ImpersonationModal.tsx` - Impersonation form
- `ImpersonationBanner.tsx` - Yellow banner during impersonation
- `adminApi.ts` - Admin API client

## Business Rules
1. Only super_admin can delete tenants
2. Only super_admin can suspend users
3. Support role can view but not modify
4. Analyst role can only view analytics
5. Impersonation requires reason (min 10 chars)
6. Impersonation session expires after 1 hour
7. All admin actions logged to audit table
8. Admin token expires after 24 hours

## Security Features
- ✅ Separate admin authentication (not mixed with user auth)
- ✅ Role-based access control (super_admin, support, analyst)
- ✅ Audit logging (all actions tracked with IP, timestamp)
- ✅ Impersonation reason required (compliance)
- ✅ Impersonation session tracking (start/end times)
- ✅ Admin token stored separately (admin_token vs token)

## Implementation Status
- ✅ Frontend: All 5 admin pages created
- ✅ Backend: `internal/api/handlers/admin_auth.go`
- ✅ Backend: `internal/api/handlers/admin_tenants.go`
- ✅ Backend: `internal/api/handlers/admin_impersonation.go`
- ✅ Backend: `internal/api/handlers/admin_analytics.go`
- ✅ Backend: `internal/services/impersonation_service.go`
- ✅ Database: All 3 admin tables created
- ✅ Testing: Manual testing complete
- ✅ Deployment: Live in Docker container

## Test Credentials
- **Email**: admin@yukti.com
- **Password**: AdminPassword123!
- **Role**: super_admin

## Future Enhancements
- Add admin activity dashboard (recent actions)
- Add tenant revenue tracking (MRR, churn)
- Add user activity logs (last login, actions)
- Add bulk operations (suspend multiple tenants)
- Add email notifications (tenant suspended, user activated)
- Add 2FA for admin login
