# RBAC Implementation - Complete Changes Summary

## Overview
This document provides a complete summary of all changes made to implement user accounts, RBAC, JWT authentication, and data-driven filter endpoints.

---

## Database Migration

### New File: `scripts/012_create_yt_users.sql`

```sql
-- User accounts and RBAC migration
-- Note: tenant_id uses INTEGER to match yt_tenants.id (SERIAL)

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS yt_users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id INTEGER NOT NULL REFERENCES yt_tenants(id) ON DELETE CASCADE,
    email TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL CHECK (role IN ('admin', 'editor', 'viewer')),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT unique_email_per_tenant UNIQUE (tenant_id, email)
);

CREATE INDEX idx_users_tenant_email ON yt_users(tenant_id, email);
CREATE INDEX idx_users_tenant_role ON yt_users(tenant_id, role);
CREATE INDEX idx_users_email ON yt_users(email) WHERE is_active = true;

-- ... (triggers, audit log updates)
```

**Rationale**: Creates user table with proper constraints, indexes, and relationships for multi-tenant RBAC.

---

## Backend Changes

### New File: `internal/models/user.go`

**Key Features:**
- GORM model with bcrypt password hashing
- Helper methods for CRUD operations
- SQL repository for compatibility

**Rationale**: Provides type-safe user model with password security and database operations.

---

### Modified: `internal/security/jwt.go`

**Changes:**
- Added `UserID`, `Email`, `Role` to `JWTClaims`
- Updated `GenerateToken()` to include user information

**Rationale**: JWT tokens now contain user identity and role for RBAC enforcement.

---

### New File: `internal/api/handlers/auth.go`

**Endpoints:**
- `POST /api/v1/auth/signup` - User registration
- `POST /api/v1/auth/login` - Authentication
- `POST /api/v1/auth/logout` - Stateless logout
- `POST /api/v1/auth/api-keys` - API key creation (admin only)

**Rationale**: Implements complete authentication flow with tenant creation and role assignment.

---

### New File: `internal/api/middleware/jwt_auth.go`

**Features:**
- Validates Bearer token from Authorization header
- Verifies user and tenant are active
- Sets context values: user_id, tenant_id, role, email

**Rationale**: Centralized JWT validation and context setting for all protected routes.

---

### New File: `internal/api/middleware/require_role.go`

**Features:**
- `RequireRole(allowedRoles ...string)` middleware
- Checks role from context
- Returns 403 if unauthorized

**Rationale**: Reusable RBAC middleware for protecting endpoints by role.

---

### New File: `internal/api/handlers/filters.go`

**Endpoints:**
- `GET /api/v1/filters/resource-types`
- `GET /api/v1/filters/tags`
- `GET /api/v1/filters/services`
- `GET /api/v1/filters/accounts`
- `GET /api/v1/filters/regions`

**Rationale**: Data-driven filter endpoints that populate UI dynamically (no hardcoded values).

---

### New File: `internal/feature/feature_gate.go`

**Features:**
- Subscription tier checking
- Feature enablement logic
- Placeholder for Stripe integration

**Rationale**: Foundation for feature gating based on subscription tiers.

---

### Modified: `internal/api/routes/routes.go`

**Added Routes:**
```go
// Auth routes (public)
router.HandleFunc("/api/v1/auth/signup", authHandler.Signup).Methods("POST")
router.HandleFunc("/api/v1/auth/login", authHandler.Login).Methods("POST")
router.HandleFunc("/api/v1/auth/logout", authHandler.Logout).Methods("POST")

// Auth routes (protected)
router.Handle("/api/v1/auth/api-keys",
    jwtAuthMw.RequireAuth(middleware.RequireRole("admin")(...))).Methods("POST")

// Filter routes
router.Handle("/api/v1/filters/resource-types",
    jwtAuthMw.RequireAuth(...)).Methods("GET")
// ... (other filter routes)
```

**Rationale**: Registers all new auth and filter endpoints with proper middleware.

---

### Modified: `internal/database/database.go`

**Change:**
```go
err = db.AutoMigrate(
    &models.Resource{},
    // ... existing models
    &models.User{},  // Added
)
```

**Rationale**: Ensures User table is created automatically on startup.

---

## Frontend Changes

### New File: `frontend/src/lib/auth.ts`

**Features:**
- Token management (localStorage)
- JWT decoding
- User context helpers
- Role checking utilities

**Rationale**: Centralized auth utilities for token handling and user context.

---

### New File: `frontend/src/pages/Auth/Login.tsx`

**Features:**
- React Hook Form + Zod validation
- Tenant code, email, password fields
- Stores JWT token on success
- Redirects to dashboard

**Rationale**: User-friendly login page with proper validation.

---

### New File: `frontend/src/pages/Auth/Signup.tsx`

**Features:**
- React Hook Form + Zod validation
- Email, password, optional company name
- Creates user and tenant
- Redirects to login

**Rationale**: Self-service signup flow for new users.

---

### New File: `frontend/src/components/Auth/ProtectedRoute.tsx`

**Features:**
- Checks authentication
- Checks role permissions
- Redirects appropriately

**Rationale**: Reusable component for protecting routes by authentication and role.

---

### New File: `frontend/src/components/Filters/DynamicFilters.tsx`

**Features:**
- Fetches filter options from backend
- Renders resource types, services, regions, tags
- Debounced filter updates
- Fully data-driven

**Rationale**: Dynamic filter UI that adapts to actual data (no hardcoded values).

---

### Modified: `frontend/src/App.tsx`

**Changes:**
- Integrated `react-router-dom` for proper routing
- Added `/login`, `/signup`, `/403` routes
- Protected all dashboard routes with `ProtectedRoute`
- Admin routes require `admin` role
- Created `AppLayout` component for consistent layout

**Rationale**: Proper routing with authentication and role-based protection.

---

### Modified: `frontend/src/services/api.ts`

**Changes:**
- Imported `getAuthHeader()` and `getCurrentUser()` from auth library
- Uses JWT token from auth library instead of localStorage directly
- Maintains backward compatibility with X-Tenant-ID header

**Rationale**: Centralizes auth logic and uses new auth utilities.

---

## Testing

### New File: `internal/api/handlers/auth_test.go`

**Structure:**
- Test database setup
- Signup tests (stubs)
- Login tests (stubs)
- User model tests (password hashing, retrieval)

**Rationale**: Test foundation for auth handlers (stubs ready for implementation).

---

### New File: `internal/api/handlers/filters_test.go`

**Structure:**
- Test database with seeded data
- Filter endpoint tests (stubs)

**Rationale**: Test foundation for filter handlers.

---

### New File: `frontend/src/pages/Auth/__tests__/Login.test.tsx`

**Features:**
- MSW mock server setup
- Form validation tests
- Login flow tests

**Rationale**: Frontend unit tests for login component.

---

### New File: `frontend/src/components/Filters/__tests__/DynamicFilters.test.tsx`

**Features:**
- MSW mocks for filter endpoints
- Component rendering tests
- Filter interaction tests

**Rationale**: Frontend unit tests for dynamic filters component.

---

## Documentation

### New File: `api/openapi.yaml`

**Content:**
- Complete OpenAPI 3.0 specification
- All auth endpoints documented
- All filter endpoints documented
- Request/response schemas
- Security schemes (Bearer JWT)

**Rationale**: API documentation for new endpoints.

---

## Environment Variables Required

```bash
# Required
JWT_SECRET=your-secret-key-change-in-production

# Optional (for testing)
TEST_DATABASE_URL=postgres://user:pass@localhost:5432/yukti_test?sslmode=disable
```

---

## Migration Instructions

1. **Run database migration:**
   ```bash
   psql $DATABASE_URL -f scripts/012_create_yt_users.sql
   ```

2. **Set JWT secret:**
   ```bash
   export JWT_SECRET=$(openssl rand -hex 32)
   ```

3. **Restart backend:**
   ```bash
   make restart
   # or
   docker-compose restart backend
   ```

4. **Install frontend test dependencies** (if needed):
   ```bash
   cd frontend
   npm install --save-dev msw @testing-library/user-event
   ```

---

## Testing Instructions

### Backend
```bash
# Run all tests
go test ./internal/api/handlers/... -v

# Run specific test
go test ./internal/api/handlers/auth_test.go -v
```

### Frontend
```bash
cd frontend
npm test
```

### Manual Testing Flow
1. Navigate to http://localhost:3000/signup
2. Create account (first user becomes admin)
3. Login at http://localhost:3000/login
4. Access protected routes
5. Test filter endpoints (check browser Network tab)
6. Test admin-only routes (should require admin role)

---

## Known Limitations & TODOs

1. **Token Storage**: Currently localStorage. TODO: Migrate to HttpOnly cookies
2. **Feature Gating**: Placeholder. TODO: Integrate with Stripe
3. **Test Implementation**: Stubs created. TODO: Implement full tests with httptest
4. **User Management UI**: Not created. TODO: Add admin user management page
5. **Password Reset**: Not implemented. TODO: Add password reset flow
6. **Email Verification**: Not implemented. TODO: Add email verification

---

## Files Summary

**Total Files Created**: 16
**Total Files Modified**: 5
**Total Lines Added**: ~2,500+

**Status**: ✅ Implementation complete and ready for testing

