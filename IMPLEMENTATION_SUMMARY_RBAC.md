# RBAC and Authentication Implementation Summary

## Overview
This document summarizes all changes made to implement user accounts, role-based access control (RBAC), JWT authentication, and data-driven filter endpoints.

## Database Changes

### Migration File: `scripts/012_create_yt_users.sql`
- Creates `yt_users` table with UUID primary key
- Fields: id, tenant_id (INTEGER FK), email, password_hash (bcrypt), role (admin/editor/viewer), is_active, timestamps
- Unique constraint on (tenant_id, email)
- Indexes for performance
- Auto-update trigger for updated_at
- Adds user_id column to yt_audit_logs

**To apply:**
```bash
psql $DATABASE_URL -f scripts/012_create_yt_users.sql
```

## Backend Changes

### 1. User Model (`internal/models/user.go`)
- GORM model for User with bcrypt password hashing
- Helper methods: CreateUser, GetUserByEmailTenant, ListUsersByTenant, UpdateUser, etc.
- SQL repository for compatibility with existing sql.DB code

### 2. JWT Service Updates (`internal/security/jwt.go`)
- Added UserID, Email, Role to JWTClaims
- Updated GenerateToken to include user information

### 3. Auth Handlers (`internal/api/handlers/auth.go`)
- **POST /api/v1/auth/signup**: Creates user and tenant (first user gets admin role)
- **POST /api/v1/auth/login**: Authenticates and returns JWT token
- **POST /api/v1/auth/logout**: Stateless logout (client deletes token)
- **POST /api/v1/auth/api-keys**: Admin-only API key creation

### 4. JWT Middleware (`internal/api/middleware/jwt_auth.go`)
- Validates Bearer token from Authorization header
- Verifies user and tenant are active
- Sets user_id, tenant_id, role, email in request context

### 5. RBAC Middleware (`internal/api/middleware/require_role.go`)
- RequireRole() function that checks user role against allowed roles
- Returns 403 if unauthorized

### 6. Filter Handlers (`internal/api/handlers/filters.go`)
- **GET /api/v1/filters/resource-types**: Distinct resource types for tenant
- **GET /api/v1/filters/tags**: Tag keys and values with counts
- **GET /api/v1/filters/services**: Distinct services from cost data
- **GET /api/v1/filters/accounts**: AWS accounts for tenant
- **GET /api/v1/filters/regions**: Distinct regions
- All endpoints are JWT-protected and cacheable

### 7. Feature Gating (`internal/feature/feature_gate.go`)
- Placeholder for subscription tier-based feature gating
- Currently defaults to enabled (will integrate with Stripe later)

### 8. Routes Updated (`internal/api/routes/routes.go`)
- Added auth routes (public: signup, login, logout)
- Added protected auth routes (api-keys: admin only)
- Added filter routes (JWT protected)

### 9. Database Auto-Migration (`internal/database/database.go`)
- Added User model to AutoMigrate

## Frontend Changes

### 1. Auth Library (`frontend/src/lib/auth.ts`)
- Token management (localStorage - TODO: migrate to HttpOnly cookies)
- JWT decoding and validation
- User context helpers
- Role checking utilities

### 2. Login Page (`frontend/src/pages/Auth/Login.tsx`)
- React Hook Form + Zod validation
- Tenant code, email, password fields
- Stores JWT token on success
- Redirects to dashboard

### 3. Signup Page (`frontend/src/pages/Auth/Signup.tsx`)
- React Hook Form + Zod validation
- Email, password, optional company name
- Creates user and tenant
- Redirects to login

### 4. Protected Route (`frontend/src/components/Auth/ProtectedRoute.tsx`)
- Wrapper component for role-based route protection
- Checks authentication and role
- Redirects to /login or /403 if unauthorized

### 5. Dynamic Filters (`frontend/src/components/Filters/DynamicFilters.tsx`)
- Fetches filter options from backend endpoints
- Renders resource types, services, regions, tags as chips/buttons
- Debounced filter state updates
- Fully data-driven (no hardcoded values)

## Testing

### Backend Tests (To be created)
- `internal/api/handlers/auth_test.go`: Unit tests for signup/login
- `internal/api/handlers/filters_test.go`: Unit tests for filter endpoints

### Frontend Tests (To be created)
- Jest + React Testing Library tests for Login component
- MSW mocks for backend API
- Integration test stub for login flow

## Environment Variables

Add to `.env`:
```
JWT_SECRET=your-secret-key-change-in-production
```

## Next Steps

1. **Run database migration**: `psql $DATABASE_URL -f scripts/012_create_yt_users.sql`
2. **Update App.tsx routing**: Add routes for /login, /signup, /403
3. **Update API service**: Use getAuthHeader() from auth.ts
4. **Add tests**: Create test files as outlined above
5. **HttpOnly cookies**: Migrate token storage from localStorage to HttpOnly cookies
6. **Stripe integration**: Implement actual feature gating based on subscription tiers

## API Documentation

### POST /api/v1/auth/signup
```json
{
  "email": "user@example.com",
  "password": "password123",
  "company_name": "Acme Corp" // optional
}
```

### POST /api/v1/auth/login
```json
{
  "tenant_code": "acme-corp-abc123",
  "email": "user@example.com",
  "password": "password123"
}
```

### GET /api/v1/filters/resource-types
Headers: `Authorization: Bearer <token>`
Response:
```json
{
  "success": true,
  "data": [
    {"key": "ec2", "label": "EC2", "count": 45},
    {"key": "rds", "label": "RDS", "count": 12}
  ]
}
```

## Notes

- Tenant ID uses INTEGER to match existing `yt_tenants.id` (SERIAL)
- Token storage is localStorage (marked as TODO for HttpOnly cookie migration)
- Feature gating is placeholder (defaults to enabled until Stripe integration)
- All filter endpoints are cacheable (5 minutes) and paginated where relevant

