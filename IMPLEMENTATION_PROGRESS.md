# Implementation Progress - Yukti Auth Fix

## 🎯 GOAL
Fix authentication to work with email + password only (no tenant_code) + Add basic subscription protection

---

## ✅ STEP 1: Add Subscription Table (15 min)

### Create Migration
```sql
-- migrations/012_add_subscriptions.sql
CREATE TABLE IF NOT EXISTS yt_subscriptions (
    id SERIAL PRIMARY KEY,
    tenant_id INTEGER NOT NULL REFERENCES yt_tenants(id) ON DELETE CASCADE,
    plan VARCHAR(50) DEFAULT 'trial',
    current_period_end TIMESTAMP DEFAULT NOW() + INTERVAL '30 days',
    grace_until TIMESTAMP,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(tenant_id)
);

-- Create index
CREATE INDEX idx_subscriptions_tenant ON yt_subscriptions(tenant_id);

-- Auto-create subscription when tenant is created
CREATE OR REPLACE FUNCTION create_default_subscription()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO yt_subscriptions (tenant_id, plan, current_period_end)
    VALUES (NEW.id, 'trial', NOW() + INTERVAL '30 days');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_create_subscription
AFTER INSERT ON yt_tenants
FOR EACH ROW
EXECUTE FUNCTION create_default_subscription();
```

**Status**: ⏳ Pending

---

## ✅ STEP 2: Fix Auth Handler (1 hour)

### Modify Login to NOT require tenant_code

**File**: `internal/api/handlers/auth.go`

**Changes**:
1. Remove `tenant_code` from LoginRequest struct
2. Lookup user by email only (email is UNIQUE)
3. Get tenant_id from user record
4. Generate JWT with tenant_id

**Status**: ⏳ Pending

---

## ✅ STEP 3: Add Subscription Middleware (30 min)

### Create Middleware

**File**: `internal/api/middleware/subscription.go` (NEW)

```go
package middleware

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
)

type SubscriptionMiddleware struct {
	db *sql.DB
}

func NewSubscriptionMiddleware(db *sql.DB) *SubscriptionMiddleware {
	return &SubscriptionMiddleware{db: db}
}

func (m *SubscriptionMiddleware) CheckSubscription(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get tenant_id from JWT context
		tenantID, ok := GetTenantID(r.Context())
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
			return
		}

		// Check subscription status
		var isActive bool
		err := m.db.QueryRow(`
			SELECT 
				CASE 
					WHEN current_period_end > NOW() THEN true
					WHEN grace_until > NOW() THEN true
					ELSE false
				END as active
			FROM yt_subscriptions
			WHERE tenant_id = $1 AND is_active = true
		`, tenantID).Scan(&isActive)

		if err != nil || !isActive {
			w.WriteHeader(http.StatusPaymentRequired)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Subscription expired. Please contact support.",
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}
```

**Status**: ⏳ Pending

---

## ✅ STEP 4: Update Routes (15 min)

### Apply Subscription Middleware

**File**: `internal/api/routes/routes.go`

**Changes**:
1. Use `NewAuthHandler()` (not NewSimpleAuthHandler)
2. Add subscription middleware to protected routes
3. Remove auth_simple.go

**Status**: ⏳ Pending

---

## ✅ STEP 5: Fix Compilation Errors (30 min)

**Files to fix**:
- Remove duplicate declarations in middleware
- Fix billing.go type mismatches
- Remove auth_simple.go

**Status**: ⏳ Pending

---

## ✅ STEP 6: Test Complete Flow (1 hour)

### Test Cases:
1. ✅ Signup: email + password → Creates tenant + user + subscription
2. ✅ Login: email + password → Returns JWT
3. ✅ Dashboard: JWT → Shows data (subscription active)
4. ✅ Expired subscription: JWT → 402 Payment Required

**Status**: ⏳ Pending

---

## 📊 PROGRESS TRACKER

- [ ] Step 1: Add subscription table (15 min)
- [ ] Step 2: Fix auth handler (1 hour)
- [ ] Step 3: Add subscription middleware (30 min)
- [ ] Step 4: Update routes (15 min)
- [ ] Step 5: Fix compilation errors (30 min)
- [ ] Step 6: Test complete flow (1 hour)

**Total Estimated Time**: 3 hours 30 minutes
**Status**: Not Started
**Started At**: -
**Completed At**: -

---

## 🚀 EXECUTION PLAN

1. Create subscription migration
2. Run migration on database
3. Fix auth.go (remove tenant_code)
4. Create subscription middleware
5. Update routes.go
6. Remove auth_simple.go
7. Fix compilation errors
8. Rebuild Docker containers
9. Test signup → login → dashboard
10. Document test results

---

**Ready to start? Type 'yes' to begin execution.**
