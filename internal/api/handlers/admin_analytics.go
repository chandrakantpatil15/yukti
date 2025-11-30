package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

type AdminAnalyticsHandler struct {
	db *sql.DB
}

func NewAdminAnalyticsHandler(db *sql.DB) *AdminAnalyticsHandler {
	return &AdminAnalyticsHandler{db: db}
}

// GetPlatformStats returns overall platform statistics
func (h *AdminAnalyticsHandler) GetPlatformStats(w http.ResponseWriter, r *http.Request) {
	db := h.db

	var stats struct {
		TotalTenants   int     `json:"total_tenants"`
		ActiveTenants  int     `json:"active_tenants"`
		TotalUsers     int     `json:"total_users"`
		TotalResources int     `json:"total_resources"`
		TotalFindings  int     `json:"total_findings"`
		TotalSavings   float64 `json:"total_savings"`
	}

	// Get tenant counts
	db.QueryRow("SELECT COUNT(*) FROM yt_customers").Scan(&stats.TotalTenants)
	db.QueryRow("SELECT COUNT(*) FROM yt_customers WHERE status = 'active'").Scan(&stats.ActiveTenants)

	// Get user count
	db.QueryRow("SELECT COUNT(*) FROM yt_users").Scan(&stats.TotalUsers)

	// Get resource count
	db.QueryRow("SELECT COUNT(*) FROM yt_tenant_resources").Scan(&stats.TotalResources)

	// Get findings count and total savings
	db.QueryRow(`
		SELECT COUNT(*), COALESCE(SUM(monthly_savings), 0) 
		FROM yt_hidden_cost_findings 
		WHERE status = 'open'
	`).Scan(&stats.TotalFindings, &stats.TotalSavings)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// GetAnalytics returns detailed platform analytics
func (h *AdminAnalyticsHandler) GetAnalytics(w http.ResponseWriter, r *http.Request) {
	db := h.db

	var analytics struct {
		NewTenants30d        int     `json:"new_tenants_30d"`
		NewUsers30d          int     `json:"new_users_30d"`
		ActiveScans7d        int     `json:"active_scans_7d"`
		TotalResources       int     `json:"total_resources"`
		TotalFindings        int     `json:"total_findings"`
		AvgSavingsPerTenant  float64 `json:"avg_savings_per_tenant"`
	}

	// New tenants in last 30 days
	db.QueryRow(`
		SELECT COUNT(*) FROM yt_customers 
		WHERE created_at >= NOW() - INTERVAL '30 days'
	`).Scan(&analytics.NewTenants30d)

	// New users in last 30 days
	db.QueryRow(`
		SELECT COUNT(*) FROM yt_users 
		WHERE created_at >= NOW() - INTERVAL '30 days'
	`).Scan(&analytics.NewUsers30d)

	// Active scans in last 7 days (placeholder - would need scan tracking table)
	analytics.ActiveScans7d = 0

	// Total resources
	db.QueryRow("SELECT COUNT(*) FROM yt_tenant_resources").Scan(&analytics.TotalResources)

	// Total findings
	db.QueryRow("SELECT COUNT(*) FROM yt_hidden_cost_findings WHERE status = 'open'").Scan(&analytics.TotalFindings)

	// Average savings per tenant
	db.QueryRow(`
		SELECT COALESCE(AVG(tenant_savings), 0) FROM (
			SELECT tenant_id, SUM(monthly_savings) as tenant_savings
			FROM yt_hidden_cost_findings
			WHERE status = 'open'
			GROUP BY tenant_id
		) as tenant_totals
	`).Scan(&analytics.AvgSavingsPerTenant)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(analytics)
}
