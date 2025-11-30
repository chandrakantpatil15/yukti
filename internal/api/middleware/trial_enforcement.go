package middleware

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// TrialEnforcementMiddleware enforces trial expiration and subscription requirements
type TrialEnforcementMiddleware struct {
	db *sql.DB
}

// NewTrialEnforcementMiddleware creates a new trial enforcement middleware
func NewTrialEnforcementMiddleware(db *sql.DB) *TrialEnforcementMiddleware {
	return &TrialEnforcementMiddleware{db: db}
}

// RequireActiveSubscription checks if tenant has active subscription or valid trial
func (m *TrialEnforcementMiddleware) RequireActiveSubscription(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantID, ok := GetTenantID(r.Context())
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Unauthorized",
			})
			return
		}

		// Check tenant subscription status
		var tier string
		var trialEndsAt *time.Time
		var status string
		err := m.db.QueryRow(`
			SELECT subscription_tier, trial_ends_at, status
			FROM yt_tenants WHERE id = $1
		`, tenantID).Scan(&tier, &trialEndsAt, &status)
		if err != nil {
			log.Printf("[ERROR] Failed to get tenant info: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Internal server error",
			})
			return
		}

		if status != "active" {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Tenant account is suspended",
			})
			return
		}

		// Check if trial expired
		if trialEndsAt != nil && time.Now().After(*trialEndsAt) {
			// Trial expired - check for active subscription
			var hasActiveSub bool
			m.db.QueryRow(`
				SELECT EXISTS(
					SELECT 1 FROM yt_billing_subscriptions
					WHERE tenant_id = $1 AND status IN ('active', 'trialing')
				)
			`, tenantID).Scan(&hasActiveSub)

			if !hasActiveSub && tier == "FREE" {
				// Trial expired and no subscription
				w.WriteHeader(http.StatusPaymentRequired) // 402
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"error":   "Trial expired. Please subscribe to continue.",
					"code":    "TRIAL_EXPIRED",
					"upgrade_url": "/billing/upgrade",
				})
				return
			}
		}

		// Allow request to proceed
		next.ServeHTTP(w, r)
	})
}

// CheckTrialStatus returns trial status information
func (m *TrialEnforcementMiddleware) CheckTrialStatus(tenantID int) (bool, *time.Time, error) {
	var tier string
	var trialEndsAt *time.Time
	err := m.db.QueryRow(`
		SELECT subscription_tier, trial_ends_at
		FROM yt_tenants WHERE id = $1
	`, tenantID).Scan(&tier, &trialEndsAt)
	if err != nil {
		return false, nil, err
	}

	if trialEndsAt == nil {
		return false, nil, nil // No trial
	}

	isExpired := time.Now().After(*trialEndsAt)
	return isExpired, trialEndsAt, nil
}

