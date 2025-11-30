package feature

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"yukti/internal/api/middleware"
)

// Feature represents a feature that can be gated by subscription tier
type Feature string

const (
	FeatureHiddenCostDetection Feature = "hidden_cost_detection"
	FeatureIaCGeneration        Feature = "iac_generation"
	FeatureMLForecasting        Feature = "ml_forecasting"
	FeatureMultiAccount         Feature = "multi_account"
	FeatureAPIKeys              Feature = "api_keys"
	FeatureWhitelisting         Feature = "whitelisting"
	FeatureBudgetTracking       Feature = "budget_tracking"
)

// SubscriptionTier represents subscription tier levels
type SubscriptionTier string

const (
	TierFree         SubscriptionTier = "FREE"
	TierProfessional SubscriptionTier = "PROFESSIONAL"
	TierEnterprise   SubscriptionTier = "ENTERPRISE"
	TierFinancial    SubscriptionTier = "FINANCIAL"
)

// FeatureGate checks if a feature is enabled for a tenant's subscription tier
type FeatureGate struct {
	db *sql.DB
}

// NewFeatureGate creates a new feature gate
func NewFeatureGate(db *sql.DB) *FeatureGate {
	return &FeatureGate{db: db}
}

// IsEnabled checks if a feature is enabled for a tenant
func (fg *FeatureGate) IsEnabled(tenantID int, feature Feature) (bool, error) {
	// Get tenant subscription tier and trial status
	var tier string
	var trialEndsAt *time.Time
	var hasActiveSub bool
	err := fg.db.QueryRow(`
		SELECT t.subscription_tier, t.trial_ends_at,
		       EXISTS(
		           SELECT 1 FROM yt_billing_subscriptions s
		           WHERE s.tenant_id = t.id AND s.status IN ('active', 'trialing')
		       ) as has_active_sub
		FROM yt_tenants t WHERE t.id = $1
	`, tenantID).Scan(&tier, &trialEndsAt, &hasActiveSub)
	if err != nil {
		return false, fmt.Errorf("failed to get tenant tier: %w", err)
	}

	// Check if trial expired and no active subscription
	if trialEndsAt != nil && time.Now().After(*trialEndsAt) && !hasActiveSub && tier == "FREE" {
		// Trial expired - only allow basic features
		switch feature {
		case FeatureBudgetTracking:
			return true, nil
		default:
			return false, nil
		}
	}

	// Feature gating by subscription tier
	switch SubscriptionTier(tier) {
	case TierFree:
		// Free tier: limited features
		switch feature {
		case FeatureHiddenCostDetection, FeatureIaCGeneration, FeatureMLForecasting, FeatureAPIKeys:
			return false, nil
		default:
			return true, nil
		}
	case TierProfessional:
		// Professional tier: most features enabled
		switch feature {
		case FeatureAPIKeys:
			return false, nil // API keys only for Enterprise+
		default:
			return true, nil
		}
	case TierEnterprise, TierFinancial:
		// Enterprise/Financial: all features enabled
		return true, nil
	default:
		// Unknown tier: default to enabled
		return true, nil
	}
}

// CheckFeature is a convenience function that returns an error if feature is disabled
func (fg *FeatureGate) CheckFeature(tenantID int, feature Feature) error {
	enabled, err := fg.IsEnabled(tenantID, feature)
	if err != nil {
		return err
	}
	if !enabled {
		// Get tier for error message
		var tier string
		fg.db.QueryRow(`SELECT subscription_tier FROM yt_tenants WHERE id = $1`, tenantID).Scan(&tier)
		return fmt.Errorf("feature %s is not available for %s tier. Upgrade to access this feature", feature, tier)
	}
	return nil
}

// RequireFeatureMiddleware creates middleware that checks feature access
func RequireFeatureMiddleware(db *sql.DB, feature Feature) func(http.Handler) http.Handler {
	gate := NewFeatureGate(db)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tenantID, ok := middleware.GetTenantID(r.Context())
			if !ok {
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"error":   "Unauthorized",
				})
				return
			}

			err := gate.CheckFeature(tenantID, feature)
			if err != nil {
				w.WriteHeader(http.StatusPaymentRequired) // 402
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"error":   err.Error(),
					"code":    "FEATURE_NOT_AVAILABLE",
					"upgrade_url": "/billing/upgrade",
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetTierFeatures returns list of enabled features for a tier
func GetTierFeatures(tier SubscriptionTier) []Feature {
	switch tier {
	case TierFree:
		return []Feature{
			FeatureBudgetTracking,
		}
	case TierProfessional:
		return []Feature{
			FeatureHiddenCostDetection,
			FeatureIaCGeneration,
			FeatureMLForecasting,
			FeatureMultiAccount,
			FeatureWhitelisting,
			FeatureBudgetTracking,
		}
	case TierEnterprise, TierFinancial:
		return []Feature{
			FeatureHiddenCostDetection,
			FeatureIaCGeneration,
			FeatureMLForecasting,
			FeatureMultiAccount,
			FeatureAPIKeys,
			FeatureWhitelisting,
			FeatureBudgetTracking,
		}
	default:
		return []Feature{}
	}
}

