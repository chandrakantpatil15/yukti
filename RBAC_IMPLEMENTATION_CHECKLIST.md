# RBAC Implementation Checklist

## ✅ Completed Tasks

### Database
- [x] Created `scripts/012_create_yt_users.sql` migration
- [x] User table with UUID primary key, tenant_id FK, email, password_hash, role, is_active
- [x] Unique constraint on (tenant_id, email)
- [x] Indexes for performance
- [x] Auto-update trigger for updated_at
- [x] Added user_id to yt_audit_logs

### Backend - Models & Security
- [x] Created `internal/models/user.go` with GORM model
- [x] Bcrypt password hashing (HashPassword, CheckPassword)
- [x] Helper methods: CreateUser, GetUserByEmailTenant, ListUsersByTenant
- [x] Updated `internal/security/jwt.go` with UserID, Email, Role in claims
- [x] Updated GenerateToken signature

### Backend - Handlers
- [x] Created `internal/api/handlers/auth.go`
  - [x] POST /api/v1/auth/signup (creates user + tenant, first user = admin)
  - [x] POST /api/v1/auth/login (returns JWT token)
  - [x] POST /api/v1/auth/logout (stateless)
  - [x] POST /api/v1/auth/api-keys (admin only)
- [x] Created `internal/api/handlers/filters.go`
  - [x] GET /api/v1/filters/resource-types
  - [x] GET /api/v1/filters/tags
  - [x] GET /api/v1/filters/services
  - [x] GET /api/v1/filters/accounts
  - [x] GET /api/v1/filters/regions

### Backend - Middleware
- [x] Created `internal/api/middleware/jwt_auth.go`
  - [x] Validates Bearer token
  - [x] Verifies user/tenant active
  - [x] Sets context: user_id, tenant_id, role, email
- [x] Created `internal/api/middleware/require_role.go`
  - [x] RequireRole() function for RBAC
  - [x] Returns 403 if unauthorized

### Backend - Routes & Integration
- [x] Updated `internal/api/routes/routes.go`
  - [x] Added auth routes (public: signup, login, logout)
  - [x] Added protected auth routes (api-keys: admin only)
  - [x] Added filter routes (JWT protected)
- [x] Updated `internal/database/database.go`
  - [x] Added User model to AutoMigrate
- [x] Created `internal/feature/feature_gate.go`
  - [x] Placeholder for subscription tier gating

### Frontend - Auth
- [x] Created `frontend/src/lib/auth.ts`
  - [x] Token management (localStorage - TODO: HttpOnly cookies)
  - [x] JWT decoding
  - [x] User context helpers
  - [x] Role checking
- [x] Created `frontend/src/pages/Auth/Login.tsx`
  - [x] React Hook Form + Zod validation
  - [x] Stores JWT on success
  - [x] Redirects to dashboard
- [x] Created `frontend/src/pages/Auth/Signup.tsx`
  - [x] React Hook Form + Zod validation
  - [x] Creates user and tenant
  - [x] Redirects to login

### Frontend - Routing & Protection
- [x] Created `frontend/src/components/Auth/ProtectedRoute.tsx`
  - [x] Checks authentication
  - [x] Checks role permissions
  - [x] Redirects to /login or /403
- [x] Updated `frontend/src/App.tsx`
  - [x] Integrated react-router-dom
  - [x] Added /login, /signup, /403 routes
  - [x] Protected all dashboard routes
  - [x] Admin routes require admin role

### Frontend - Dynamic Filters
- [x] Created `frontend/src/components/Filters/DynamicFilters.tsx`
  - [x] Fetches filter options from backend
  - [x] Renders resource types, services, regions, tags
  - [x] Debounced filter updates
  - [x] Fully data-driven (no hardcoded values)

### Testing
- [x] Created `internal/api/handlers/auth_test.go`
  - [x] Test structure with setupTestDB
  - [x] Stubs for signup/login tests
  - [x] User model tests (password hashing, retrieval)
- [x] Created `internal/api/handlers/filters_test.go`
  - [x] Test structure with seeded data
  - [x] Stubs for filter endpoint tests
- [x] Created `frontend/src/pages/Auth/__tests__/Login.test.tsx`
  - [x] MSW mock server setup
  - [x] Form validation tests
  - [x] Login flow tests (stubs)
- [x] Created `frontend/src/components/Filters/__tests__/DynamicFilters.test.tsx`
  - [x] MSW mocks for filter endpoints
  - [x] Component rendering tests
  - [x] Filter interaction tests

### Documentation
- [x] Created `api/openapi.yaml`
  - [x] Auth endpoints documented
  - [x] Filter endpoints documented
  - [x] Request/response schemas
  - [x] Security schemes (Bearer JWT)
- [x] Created `IMPLEMENTATION_SUMMARY_RBAC.md`
- [x] Created `RBAC_IMPLEMENTATION_CHECKLIST.md` (this file)

## 🔄 Next Steps (To Complete)

### Immediate
1. **Run database migration:**
   ```bash
   psql $DATABASE_URL -f scripts/012_create_yt_users.sql
   ```

2. **Set environment variable:**
   ```bash
   export JWT_SECRET=your-secret-key-change-in-production
   ```

3. **Install frontend test dependencies** (if missing):
   ```bash
   cd frontend
   npm install --save-dev msw @testing-library/user-event
   ```

4. **Test the implementation:**
   ```bash
   # Backend
   go test ./internal/api/handlers/... -v
   
   # Frontend
   cd frontend
   npm test
   ```

### Short-term Enhancements
- [ ] Implement full test cases (replace stubs with httptest)
- [ ] Add HttpOnly cookie support for token storage
- [ ] Update API service (`frontend/src/services/api.ts`) to use getAuthHeader()
- [ ] Add user management page (`frontend/src/pages/Admin/Users.tsx`)
- [ ] Implement actual feature gating with Stripe integration
- [ ] Add password reset functionality
- [ ] Add email verification

### Integration Tasks
- [ ] Update existing handlers to use JWT middleware instead of API key
- [ ] Add RBAC to existing endpoints (onboarding, whitelists, etc.)
- [ ] Update Dashboard to use DynamicFilters component
- [ ] Add user invitation flow (admin invites users)
- [ ] Add user profile page

## 📝 Notes

### Known Issues
1. **Token Storage**: Currently uses localStorage. Should migrate to HttpOnly cookies for better security.
2. **Feature Gating**: Placeholder implementation - defaults to enabled until Stripe integration.
3. **Test Stubs**: Test files contain structure but need full implementation with httptest/msw.
4. **Tenant ID Type**: Uses INTEGER to match existing schema (not UUID as originally requested).

### Ambiguities Resolved
- **Signup Flow**: Creates new tenant if company_name provided, or associates with existing tenant if email matches
- **First User Role**: First user for a tenant automatically gets 'admin' role
- **Token Expiration**: JWT tokens expire after 24 hours (configurable)

## 🧪 Testing Instructions

### Backend Tests
```bash
# Set test database URL
export TEST_DATABASE_URL="postgres://user:pass@localhost:5432/yukti_test?sslmode=disable"

# Run auth tests
go test ./internal/api/handlers/auth_test.go -v

# Run filter tests
go test ./internal/api/handlers/filters_test.go -v
```

### Frontend Tests
```bash
cd frontend
npm test -- --watchAll=false
```

### Manual Testing
1. Start backend: `make start` or `docker-compose up backend`
2. Start frontend: `cd frontend && npm start`
3. Navigate to http://localhost:3000/signup
4. Create account
5. Login at http://localhost:3000/login
6. Test protected routes
7. Test filter endpoints (check Network tab)

## 🔐 Security Considerations

- ✅ Passwords hashed with bcrypt (default cost)
- ✅ JWT tokens signed with HS256
- ✅ User and tenant status verified on each request
- ⚠️ Token in localStorage (TODO: HttpOnly cookies)
- ⚠️ JWT_SECRET should be strong random string in production
- ⚠️ Admin key still hardcoded (needs proper admin auth)

## 📊 Files Changed Summary

**New Files (15):**
- `scripts/012_create_yt_users.sql`
- `internal/models/user.go`
- `internal/api/handlers/auth.go`
- `internal/api/handlers/filters.go`
- `internal/api/middleware/jwt_auth.go`
- `internal/api/middleware/require_role.go`
- `internal/feature/feature_gate.go`
- `frontend/src/lib/auth.ts`
- `frontend/src/pages/Auth/Login.tsx`
- `frontend/src/pages/Auth/Signup.tsx`
- `frontend/src/components/Auth/ProtectedRoute.tsx`
- `frontend/src/components/Filters/DynamicFilters.tsx`
- `internal/api/handlers/auth_test.go`
- `internal/api/handlers/filters_test.go`
- `frontend/src/pages/Auth/__tests__/Login.test.tsx`
- `frontend/src/components/Filters/__tests__/DynamicFilters.test.tsx`
- `api/openapi.yaml`

**Modified Files (5):**
- `internal/security/jwt.go` - Added user/role to claims
- `internal/api/routes/routes.go` - Added auth and filter routes
- `internal/database/database.go` - Added User to AutoMigrate
- `frontend/src/App.tsx` - Added routing with react-router-dom
- `frontend/src/components/Filters/DynamicFilters.tsx` - Fixed import

---

**Status**: ✅ Core implementation complete. Ready for testing and integration.

