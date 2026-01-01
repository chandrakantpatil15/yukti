# Yukti Platform - Complete Repository Analysis

## 📊 DATABASE SCHEMA ANALYSIS

### Core Tables Structure

#### 1. **yt_tenants** (Main Tenant Table)
```
- id: integer (PRIMARY KEY, auto-increment)
- tenant_code: varchar(50) UNIQUE (e.g., "acme-corp-a1b2c3d4")
- company_name: varchar(200)
- subscription_tier: varchar(20) DEFAULT 'FREE'
- status: varchar(20) DEFAULT 'active'
- created_at: timestamp
- trial_ends_at: timestamp
```
**Purpose**: Main tenant/organization table. Each company gets ONE tenant.

#### 2. **yt_users** (User Accounts)
```
- id: uuid (PRIMARY KEY)
- tenant_id: bigint (FK to yt_tenants.id)
- email: text UNIQUE
- password_hash: text
- role: text (admin, editor, viewer)
- is_active: boolean DEFAULT true
- email_verified: boolean DEFAULT false
- created_at: timestamp
- updated_at: timestamp
```
**Purpose**: User accounts. One user can belong to ONE tenant (tenant_id).

#### 3. **yt_tenant_users** (Multi-Tenant User Mapping - RBAC)
```
- id: uuid (PRIMARY KEY)
- user_id: uuid (FK to yt_users.id)
- tenant_id: varchar (FK to yt_customers.id) ⚠️ MISMATCH!
- role: text (owner, admin, editor, viewer)
- is_active: boolean
- invited_by: uuid
- joined_at: timestamp
```
**Purpose**: Maps users to multiple tenants with roles (for RBAC/team features).
**⚠️ CRITICAL ISSUE**: tenant_id references yt_customers (varchar) NOT yt_tenants (int)!

#### 4. **yt_customers** (Legacy/Onboarding Table)
```
- id: varchar(50) (PRIMARY KEY)
- tenant_id: varchar(50)
- company_name: varchar(255)
- email: varchar(255)
- onboarding_status: varchar(50) DEFAULT 'pending'
- onboarding_step: varchar(50)
- created_at: timestamp
```
**Purpose**: Appears to be for onboarding flow tracking. NOT the main tenant table.

### Other Important Tables

- **yt_aws_connections**: AWS account configurations per tenant
- **yt_tenant_resources**: Discovered AWS resources (EC2, RDS, S3)
- **yt_hidden_cost_findings**: Cost optimization findings
- **yt_otp_codes**: Email verification OTPs
- **yt_refresh_tokens**: JWT refresh tokens
- **yt_admin_users**: Platform admin accounts
- **yt_user_invitations**: Team invitation tokens

---

## 🔄 CURRENT USER JOURNEY (As Designed)

### **Signup Flow**
```
1. User visits /signup
2. Enters: email, password
3. Backend (auth.go):
   a. Creates NEW tenant in yt_tenants (generates tenant_code)
   b. Creates user in yt_users (tenant_id = new tenant)
   c. Assigns role: "admin" (first user) or "viewer"
4. Returns: user_id, tenant_id
5. Frontend: Redirects to /login
```

### **Login Flow**
```
1. User visits /login
2. Enters: tenant_code, email, password ⚠️ REQUIRES TENANT CODE!
3. Backend (auth.go):
   a. Looks up tenant by tenant_code
   b. Finds user by email + tenant_id
   c. Verifies password
   d. Generates JWT with tenant_id
4. Returns: token, user info
5. Frontend: Stores token, redirects to /dashboard
```

### **Dashboard Access Flow**
```
1. User navigates to /dashboard
2. Frontend sends JWT in Authorization header
3. Backend JWT middleware:
   a. Validates token
   b. Extracts tenant_id from JWT
   c. Checks user is active
   d. Checks tenant is active
4. Dashboard handler:
   a. Uses tenant_id from JWT (NOT from query params)
   b. Fetches tenant-specific data
5. Returns: dashboard metrics
```

---

## ⚠️ CRITICAL ISSUES IDENTIFIED

### Issue #1: **Dual Handler Conflict**
- **auth.go**: Original complex handler (uses GORM, models, tenant_code)
- **auth_simple.go**: New simple handler (uses sql.DB, wrong schema)
- **routes.go**: Currently uses `NewSimpleAuthHandler()` ❌

**Problem**: Simple handler uses wrong database schema (yt_customers instead of yt_tenants)

### Issue #2: **Frontend Expects NO tenant_code**
- **Login.tsx**: Removed tenant_code field (only email + password)
- **auth.go**: REQUIRES tenant_code in login request
- **Mismatch**: Frontend can't login because backend expects tenant_code!

### Issue #3: **Schema Inconsistency**
- **yt_users.tenant_id**: bigint (references yt_tenants.id)
- **yt_tenant_users.tenant_id**: varchar (references yt_customers.id)
- **Problem**: RBAC table references wrong tenant table!

### Issue #4: **Multiple Compilation Errors**
- auth_refresh.go removed (had tokenService dependency)
- billing.go has type mismatches
- Middleware has duplicate declarations
- JWTClaims missing JTI field

---

## ✅ CORRECT IMPLEMENTATION PLAN

### **Step 1: Use Original auth.go Handler**
```go
// routes.go
authHandler := handlers.NewAuthHandler(db)  // NOT NewSimpleAuthHandler
```

### **Step 2: Fix Login to NOT Require tenant_code**
Two options:

**Option A**: Modify auth.go to lookup tenant by email
```go
// Get user by email (across all tenants)
var user User
err := h.db.QueryRow(`
    SELECT u.id, u.tenant_id, u.email, u.password_hash, u.role, t.tenant_code
    FROM yt_users u
    JOIN yt_tenants t ON u.tenant_id = t.id
    WHERE u.email = $1 AND u.is_active = true
`, email).Scan(&user.ID, &user.TenantID, &user.Email, &user.PasswordHash, &user.Role, &tenantCode)
```

**Option B**: Frontend sends tenant_code (add back to Login.tsx)
```tsx
// Login.tsx
<Input name="tenant_code" placeholder="Company Code" />
```

### **Step 3: Fix Database Schema Mismatch**
```sql
-- Option 1: Make yt_tenant_users.tenant_id reference yt_tenants
ALTER TABLE yt_tenant_users 
  ALTER COLUMN tenant_id TYPE integer USING tenant_id::integer;

-- Option 2: Keep as-is and don't use yt_tenant_users for now
-- (Defer RBAC to post-MVP)
```

### **Step 4: Remove auth_simple.go**
```bash
rm internal/api/handlers/auth_simple.go
```

### **Step 5: Fix Compilation Errors**
- Remove duplicate contextKey declarations
- Add JTI field to JWTClaims (or remove blacklist check)
- Fix billing.go type mismatches

---

## 🎯 RECOMMENDED USER JOURNEY (Simplified)

### **Signup** (No Changes Needed)
```
POST /api/v1/auth/signup
Body: { "email": "user@example.com", "password": "Pass1234!" }
Response: { "success": true, "user_id": "uuid", "tenant_id": 123 }
```

### **Login** (OPTION A - No tenant_code)
```
POST /api/v1/auth/login
Body: { "email": "user@example.com", "password": "Pass1234!" }
Backend: Lookup user by email, get tenant_id automatically
Response: { "success": true, "token": "jwt...", "user": {...} }
```

### **Login** (OPTION B - With tenant_code)
```
POST /api/v1/auth/login
Body: { 
  "tenant_code": "acme-corp-a1b2c3d4",
  "email": "user@example.com", 
  "password": "Pass1234!" 
}
Response: { "success": true, "token": "jwt...", "user": {...} }
```

### **Dashboard Access**
```
GET /api/v1/dashboard
Headers: { "Authorization": "Bearer jwt..." }
Backend: Extract tenant_id from JWT
Response: { "metrics": {...}, "findings": [...] }
```

---

## 📝 NEXT STEPS (In Order)

1. ✅ **Decision**: Choose Option A (no tenant_code) or Option B (with tenant_code)
2. ✅ **Remove**: Delete auth_simple.go
3. ✅ **Fix**: Update routes.go to use NewAuthHandler()
4. ✅ **Modify**: Update auth.go Login() based on chosen option
5. ✅ **Fix**: Resolve all compilation errors
6. ✅ **Test**: Signup → Login → Dashboard flow
7. ✅ **Deploy**: Rebuild Docker containers

---

## 🔍 CURRENT STATE SUMMARY

**What Works:**
- ✅ Database schema (yt_tenants, yt_users)
- ✅ JWT authentication middleware
- ✅ Signup flow (creates tenant + user)
- ✅ Dashboard API (uses JWT tenant_id)

**What's Broken:**
- ❌ Login flow (frontend/backend mismatch on tenant_code)
- ❌ Compilation errors (auth_simple.go, billing.go, middleware)
- ❌ Routes using wrong handler (NewSimpleAuthHandler)

**What's Confusing:**
- ⚠️ Two tenant tables (yt_tenants vs yt_customers)
- ⚠️ RBAC table references wrong tenant table
- ⚠️ Two auth handlers (auth.go vs auth_simple.go)

---

## 💡 RECOMMENDATION

**Use OPTION A** (No tenant_code in login):
- ✅ Better UX (users don't need to remember tenant code)
- ✅ Matches current frontend (Login.tsx has no tenant_code field)
- ✅ Standard for SaaS apps (email is unique identifier)
- ⚠️ Requires: Email must be unique across ALL tenants

**Implementation:**
1. Modify auth.go Login() to lookup user by email only
2. Get tenant_id from user record
3. Generate JWT with tenant_id
4. Frontend stays as-is (no changes needed)

---

**Status**: Ready for implementation once decision is made.
**Estimated Time**: 30 minutes to fix + 10 minutes to test
**Risk Level**: Low (clear path forward)
