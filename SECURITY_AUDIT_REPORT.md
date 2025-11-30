# 🔒 Security Audit Report - Yukti Platform

**Date**: 2024
**Auditor**: Pro Developer Security Review
**Focus**: Data Leaks & Tenant Isolation

---

## Executive Summary

**Overall Security Rating**: ⚠️ **MEDIUM RISK** - Critical vulnerabilities found

**Critical Issues Found**: 5
**High Issues Found**: 3
**Medium Issues Found**: 4

**Immediate Action Required**: YES

---

## 🚨 CRITICAL VULNERABILITIES

### 1. **TENANT ISOLATION BYPASS - Query Parameter Injection**

**Location**: `internal/api/handlers/customers.go` - GetDashboard(), GetFindings()

**Vulnerability**:
```go
// VULNERABLE CODE
tenantID := r.URL.Query().Get("tenant_id")  // ❌ User-controlled input
err := h.db.QueryRow(`
    SELECT ... FROM yt_hidden_cost_findings 
    WHERE tenant_id = $1
`, tenantID).Scan(...)
```

**Attack Scenario**:
```bash
# Attacker can access ANY tenant's data by changing tenant_id
curl "http://api.yukti.io/api/customers/findings?tenant_id=victim-tenant-123"
```

**Impact**: **CRITICAL** - Complete data breach. Any user can access any tenant's data.

**Fix Required**:
```go
// SECURE CODE
tenantID, ok := middleware.GetTenantID(r.Context())  // ✅ From JWT token
if !ok {
    return Unauthorized()
}
// Now tenantID is from authenticated JWT, not user input
```

**Status**: ❌ **UNFIXED** - Affects 2 endpoints

---

### 2. **MISSING TENANT VALIDATION IN ONBOARDING**

**Location**: `internal/onboarding/service.go` - SaveAWSConnection(), SaveMetricsIntegration()

**Vulnerability**:
```go
// VULNERABLE CODE
func (s *Service) SaveAWSConnection(ctx context.Context, conn *AWSConnection) error {
    query := `INSERT INTO yt_aws_connections (tenant_id, ...) VALUES ($1, ...)`
    _, err := s.db.ExecContext(ctx, query, conn.TenantID, ...)  // ❌ No validation
}
```

**Attack Scenario**:
```bash
# Attacker can save AWS credentials to victim's tenant
POST /api/onboarding/aws
{
  "tenant_id": "victim-tenant-123",  // ❌ Attacker controls this
  "account_id": "attacker-account",
  "role_arn": "arn:aws:iam::attacker:role/Steal"
}
```

**Impact**: **CRITICAL** - Attacker can hijack victim's AWS integration.

**Fix Required**:
```go
// SECURE CODE
func (s *Service) SaveAWSConnection(ctx context.Context, userTenantID int, conn *AWSConnection) error {
    // Validate tenant_id matches authenticated user
    if conn.TenantID != userTenantID {
        return errors.New("tenant_id mismatch")
    }
    // ... rest of code
}
```

**Status**: ❌ **UNFIXED**

---

### 3. **TENANT ISOLATION MIDDLEWARE CHECKS WRONG TABLE**

**Location**: `internal/api/middleware/tenant_isolation.go`

**Vulnerability**:
```go
// VULNERABLE CODE
err := m.db.QueryRow("SELECT EXISTS(SELECT 1 FROM yt_customers WHERE tenant_id = $1)", tenantID).Scan(&exists)
```

**Problem**: 
- Checks `yt_customers` table (old onboarding table)
- Should check `yt_tenants` table (actual tenant table)
- `yt_customers.tenant_id` is a STRING, not INT
- Mismatch with JWT which uses INT tenant_id

**Impact**: **CRITICAL** - Tenant validation may fail or pass incorrectly.

**Fix Required**:
```go
// SECURE CODE
err := m.db.QueryRow("SELECT EXISTS(SELECT 1 FROM yt_tenants WHERE id = $1 AND status = 'active')", tenantID).Scan(&exists)
```

**Status**: ❌ **UNFIXED**

---

### 4. **RESOURCE DETAILS ENDPOINT - TENANT BYPASS POSSIBLE**

**Location**: `internal/api/handlers/resources.go` - GetResourceDetails()

**Vulnerability**:
```go
// PARTIALLY SECURE
tenantID, ok := middleware.GetTenantID(r.Context())  // ✅ Good
resourceID := r.URL.Query().Get("resource_id")      // ❌ User input

err := h.db.QueryRow(`
    SELECT ... FROM yt_tenant_resources r
    WHERE r.tenant_id = $1 AND r.resource_id = $2
`, tenantID, resourceID).Scan(...)
```

**Problem**: While tenant_id is validated, resource_id is user-controlled. If resource IDs are predictable (e.g., sequential), attacker can enumerate resources.

**Impact**: **HIGH** - Information disclosure, resource enumeration.

**Fix Required**: Add rate limiting, use UUIDs for resource IDs, add audit logging.

**Status**: ⚠️ **PARTIAL** - Needs hardening

---

### 5. **SIGNUP CREATES TENANT WITHOUT EMAIL VERIFICATION**

**Location**: `internal/api/handlers/auth.go` - Signup()

**Vulnerability**:
```go
// VULNERABLE CODE
// Create tenant first
err = h.db.QueryRow(`
    INSERT INTO yt_tenants (tenant_code, company_name, status, subscription_tier, created_at)
    VALUES ($1, $2, 'pending_verification', 'FREE', NOW())  // ❌ Tenant created immediately
    RETURNING id
`, tenantCode, req.CompanyName).Scan(&tenantID)

// Create user
_, err = h.db.Exec(`
    INSERT INTO yt_users (tenant_id, email, password_hash, role, is_active, email_verified, created_at)
    VALUES ($1, $2, $3, 'admin', true, false, NOW())  // ❌ is_active=true before verification
`, tenantID, req.Email, string(hashedPassword))
```

**Problem**:
- Tenant and user created before email verification
- Attacker can create unlimited tenants with fake emails
- Database pollution attack
- Resource exhaustion

**Impact**: **HIGH** - Spam, resource exhaustion, database bloat.

**Fix Required**:
```go
// SECURE CODE
// 1. Create tenant with status='pending_email_verification'
// 2. Create user with is_active=false
// 3. Only activate after email verification
// 4. Add rate limiting on signup (max 3 per IP per hour)
```

**Status**: ❌ **UNFIXED**

---

## 🔴 HIGH SEVERITY ISSUES

### 6. **LOGIN MISSING TENANT VALIDATION**

**Location**: `internal/api/handlers/auth.go` - Login()

**Issue**: Login only checks email/password, doesn't validate tenant status.

**Fix**: Add tenant status check before generating JWT.

---

### 7. **JWT CONTAINS TENANT_ID BUT NO CROSS-CHECK**

**Location**: `internal/api/middleware/jwt_auth.go`

**Issue**: JWT middleware validates user and tenant separately, but doesn't verify user belongs to tenant.

**Attack**: If JWT is compromised, attacker could modify tenant_id claim.

**Fix**:
```go
// Add cross-check
err = m.db.QueryRow(`
    SELECT EXISTS(SELECT 1 FROM yt_users WHERE id = $1 AND tenant_id = $2)
`, claims.UserID, claims.TenantID).Scan(&valid)
```

---

### 8. **MISSING RATE LIMITING ON SENSITIVE ENDPOINTS**

**Location**: All auth endpoints

**Issue**: No rate limiting on:
- `/api/v1/auth/signup` - Spam attack
- `/api/v1/auth/login` - Brute force attack
- `/api/v1/auth/verify` - OTP brute force

**Fix**: Add rate limiting middleware (max 5 attempts per 15 minutes per IP).

---

## 🟡 MEDIUM SEVERITY ISSUES

### 9. **OTP CODE RETURNED IN SIGNUP RESPONSE**

**Location**: `internal/api/handlers/auth.go` - Signup()

```go
// INSECURE
json.NewEncoder(w).Encode(map[string]interface{}{
    "success": true,
    "message": "Account created. Use the verification code below.",
    "otp_code": code,  // ❌ OTP in response (for development only!)
})
```

**Issue**: OTP should ONLY be sent via email, never in API response.

**Fix**: Remove `otp_code` from response in production.

---

### 10. **TENANT_CODE GENERATION PREDICTABLE**

**Location**: `internal/api/handlers/auth.go` - Signup()

```go
// WEAK
tenantCode := "tenant-" + req.Email[:strings.Index(req.Email, "@")] + "-" + strconv.FormatInt(time.Now().Unix(), 36)
```

**Issue**: Tenant codes are predictable (email prefix + timestamp).

**Fix**: Use cryptographically secure random generation.

---

### 11. **MISSING AUDIT LOGGING ON SENSITIVE OPERATIONS**

**Issue**: No audit logs for:
- Tenant data access
- AWS connection changes
- User role changes
- Failed login attempts

**Fix**: Add audit logging to `yt_audit_logs` table.

---

### 12. **PASSWORD VALIDATION INCONSISTENT**

**Location**: `internal/api/handlers/auth.go`

**Issue**: Signup uses `security.ValidatePasswordStrict()` but we don't see the implementation. Need to verify it enforces:
- Minimum 12 characters
- Uppercase, lowercase, number, special char
- No common passwords

---

## ✅ SECURITY STRENGTHS

1. ✅ **JWT Authentication** - Properly implemented with expiry
2. ✅ **Password Hashing** - Uses bcrypt with default cost
3. ✅ **SQL Injection Prevention** - Parameterized queries used
4. ✅ **HTTPS Ready** - No hardcoded HTTP URLs
5. ✅ **Role-Based Access Control** - Admin/Editor/Viewer roles
6. ✅ **Email Verification Flow** - Exists (but needs hardening)
7. ✅ **Refresh Tokens** - Implemented for session management

---

## 🔧 IMMEDIATE FIXES REQUIRED

### Priority 1 (Deploy Today):

1. **Fix GetDashboard() and GetFindings()** - Use JWT tenant_id, not query param
2. **Fix TenantIsolationMiddleware** - Check yt_tenants, not yt_customers
3. **Fix Onboarding** - Validate tenant_id matches authenticated user

### Priority 2 (Deploy This Week):

4. **Add rate limiting** - Signup, login, OTP verification
5. **Remove OTP from response** - Production security
6. **Add audit logging** - All sensitive operations
7. **Add tenant-user cross-check** - JWT middleware

### Priority 3 (Deploy Next Week):

8. **Harden signup flow** - Only activate after email verification
9. **Improve tenant_code generation** - Cryptographically secure
10. **Add resource enumeration protection** - Rate limiting, UUIDs
11. **Add failed login tracking** - Lock account after 5 failures

---

## 📋 TESTING CHECKLIST

Before deploying fixes, test:

- [ ] User A cannot access User B's data (different tenants)
- [ ] User A cannot modify User B's AWS connections
- [ ] Invalid tenant_id in query params is rejected
- [ ] JWT with mismatched user_id/tenant_id is rejected
- [ ] Signup rate limiting works (max 3 per IP per hour)
- [ ] Login rate limiting works (max 5 per IP per 15 min)
- [ ] OTP not returned in production signup response
- [ ] Audit logs created for sensitive operations
- [ ] Inactive tenants cannot access API
- [ ] Unverified emails cannot login

---

## 🎯 SECURITY RECOMMENDATIONS

### Short Term (1-2 weeks):
1. Implement all Priority 1 & 2 fixes
2. Add comprehensive audit logging
3. Set up security monitoring alerts
4. Conduct penetration testing

### Medium Term (1-2 months):
5. Implement Web Application Firewall (WAF)
6. Add DDoS protection (CloudFlare/AWS Shield)
7. Set up intrusion detection system
8. Regular security audits (monthly)

### Long Term (3-6 months):
9. Bug bounty program
10. SOC 2 compliance
11. Annual penetration testing
12. Security training for developers

---

## 📊 RISK MATRIX

| Vulnerability | Severity | Exploitability | Impact | Priority |
|---------------|----------|----------------|--------|----------|
| Tenant Isolation Bypass | Critical | Easy | Complete Data Breach | P0 |
| Onboarding Tenant Bypass | Critical | Easy | AWS Hijacking | P0 |
| Wrong Table Check | Critical | Medium | Auth Bypass | P0 |
| Resource Enumeration | High | Medium | Info Disclosure | P1 |
| Signup Without Verification | High | Easy | Spam/DoS | P1 |
| Missing Rate Limiting | High | Easy | Brute Force | P1 |
| OTP in Response | Medium | Easy | Account Takeover | P2 |
| Predictable Tenant Codes | Medium | Medium | Enumeration | P2 |

---

## 🚀 DEPLOYMENT PLAN

### Phase 1: Emergency Fixes (Deploy Immediately)
```bash
# Fix critical tenant isolation issues
git checkout -b security/critical-fixes
# Apply fixes 1, 2, 3
# Test thoroughly
# Deploy to production with zero downtime
```

### Phase 2: High Priority (Deploy Within 48 Hours)
```bash
# Add rate limiting and audit logging
git checkout -b security/high-priority
# Apply fixes 4, 5, 6, 7
# Test with load testing
# Deploy to production
```

### Phase 3: Hardening (Deploy Within 1 Week)
```bash
# Harden signup and improve security
git checkout -b security/hardening
# Apply fixes 8, 9, 10, 11
# Full security testing
# Deploy to production
```

---

## 📞 INCIDENT RESPONSE

If data breach suspected:
1. **Immediately** disable affected tenant accounts
2. Rotate all JWT secrets
3. Invalidate all refresh tokens
4. Notify affected customers within 72 hours
5. Conduct forensic analysis
6. File incident report

---

**End of Security Audit Report**

**Next Steps**: Review this report, prioritize fixes, and create GitHub issues for tracking.
