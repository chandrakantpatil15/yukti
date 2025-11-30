package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"yukti/internal/api/middleware"
)

type AdminTenantsHandler struct {
	db *sql.DB
}

func NewAdminTenantsHandler(db *sql.DB) *AdminTenantsHandler {
	return &AdminTenantsHandler{db: db}
}

type AdminTenantInfo struct {
	ID                string `json:"id"`
	TenantID          string `json:"tenant_id"`
	CompanyName       string `json:"company_name"`
	Email             string `json:"email"`
	OnboardingStatus  string `json:"onboarding_status"`
	UserCount         int    `json:"user_count"`
	ResourceCount     int    `json:"resource_count"`
	FindingsCount     int    `json:"findings_count"`
	MonthlySavings    string `json:"monthly_savings"`
	CreatedAt         string `json:"created_at"`
}

type ListTenantsResponse struct {
	Success bool              `json:"success"`
	Tenants []AdminTenantInfo `json:"tenants"`
	Total   int               `json:"total"`
}

// ListTenants returns all tenants with stats
func (h *AdminTenantsHandler) ListTenants(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(`
		SELECT 
			c.id,
			c.tenant_id,
			c.company_name,
			c.email,
			c.onboarding_status,
			c.created_at,
			COALESCE((SELECT COUNT(*) FROM yt_tenant_users tu WHERE tu.tenant_id = c.id), 0) as user_count,
			COALESCE((SELECT COUNT(*) FROM yt_tenant_resources tr WHERE tr.tenant_id = c.id), 0) as resource_count,
			COALESCE((SELECT COUNT(*) FROM yt_hidden_cost_findings f WHERE f.tenant_id = c.id), 0) as findings_count,
			COALESCE((SELECT SUM(monthly_savings) FROM yt_hidden_cost_findings f WHERE f.tenant_id = c.id), 0) as monthly_savings
		FROM yt_customers c
		ORDER BY c.created_at DESC
	`)

	if err != nil {
		log.Printf("[ERROR] Failed to list tenants: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ListTenantsResponse{Success: false})
		return
	}
	defer rows.Close()

	tenants := []AdminTenantInfo{}
	for rows.Next() {
		var t AdminTenantInfo
		var savings float64
		rows.Scan(&t.ID, &t.TenantID, &t.CompanyName, &t.Email, &t.OnboardingStatus, 
			&t.CreatedAt, &t.UserCount, &t.ResourceCount, &t.FindingsCount, &savings)
		t.MonthlySavings = formatCurrency(savings)
		tenants = append(tenants, t)
	}

	json.NewEncoder(w).Encode(ListTenantsResponse{
		Success: true,
		Tenants: tenants,
		Total:   len(tenants),
	})
}

type TenantDetailResponse struct {
	Success bool        `json:"success"`
	Tenant  interface{} `json:"tenant,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// GetTenantDetails returns detailed tenant information
func (h *AdminTenantsHandler) GetTenantDetails(w http.ResponseWriter, r *http.Request) {
	tenantID := strings.TrimPrefix(r.URL.Path, "/api/admin/tenants/")

	var t AdminTenantInfo
	var savings float64
	err := h.db.QueryRow(`
		SELECT 
			c.id,
			c.tenant_id,
			c.company_name,
			c.email,
			c.onboarding_status,
			c.created_at,
			COALESCE((SELECT COUNT(*) FROM yt_tenant_users tu WHERE tu.tenant_id = c.id), 0),
			COALESCE((SELECT COUNT(*) FROM yt_tenant_resources tr WHERE tr.tenant_id = c.id), 0),
			COALESCE((SELECT COUNT(*) FROM yt_hidden_cost_findings f WHERE f.tenant_id = c.id), 0),
			COALESCE((SELECT SUM(monthly_savings) FROM yt_hidden_cost_findings f WHERE f.tenant_id = c.id), 0)
		FROM yt_customers c
		WHERE c.id = $1
	`, tenantID).Scan(&t.ID, &t.TenantID, &t.CompanyName, &t.Email, &t.OnboardingStatus,
		&t.CreatedAt, &t.UserCount, &t.ResourceCount, &t.FindingsCount, &savings)

	if err == sql.ErrNoRows {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(TenantDetailResponse{Success: false, Error: "Tenant not found"})
		return
	}

	if err != nil {
		log.Printf("[ERROR] Failed to get tenant details: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(TenantDetailResponse{Success: false, Error: "Failed to get tenant"})
		return
	}

	t.MonthlySavings = formatCurrency(savings)

	// Get users
	userRows, _ := h.db.Query(`
		SELECT u.email, tu.role, tu.is_active, tu.joined_at
		FROM yt_tenant_users tu
		JOIN yt_users u ON tu.user_id = u.id
		WHERE tu.tenant_id = $1
		ORDER BY tu.joined_at ASC
	`, tenantID)
	defer userRows.Close()

	users := []map[string]interface{}{}
	for userRows.Next() {
		var email, role, joinedAt string
		var isActive bool
		userRows.Scan(&email, &role, &isActive, &joinedAt)
		users = append(users, map[string]interface{}{
			"email":     email,
			"role":      role,
			"is_active": isActive,
			"joined_at": joinedAt,
		})
	}

	json.NewEncoder(w).Encode(TenantDetailResponse{
		Success: true,
		Tenant: map[string]interface{}{
			"id":                t.ID,
			"tenant_id":         t.TenantID,
			"company_name":      t.CompanyName,
			"email":             t.Email,
			"onboarding_status": t.OnboardingStatus,
			"user_count":        t.UserCount,
			"resource_count":    t.ResourceCount,
			"findings_count":    t.FindingsCount,
			"monthly_savings":   t.MonthlySavings,
			"created_at":        t.CreatedAt,
			"users":             users,
		},
	})
}

type TenantActionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// SuspendTenant suspends a tenant account
func (h *AdminTenantsHandler) SuspendTenant(w http.ResponseWriter, r *http.Request) {
	tenantID := strings.TrimPrefix(r.URL.Path, "/api/admin/tenants/")
	tenantID = strings.TrimSuffix(tenantID, "/suspend")

	adminID, _ := middleware.GetAdminUserID(r.Context())

	_, err := h.db.Exec(`
		UPDATE yt_customers SET onboarding_status = 'suspended' WHERE id = $1
	`, tenantID)

	if err != nil {
		log.Printf("[ERROR] Failed to suspend tenant: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(TenantActionResponse{Success: false, Error: "Failed to suspend tenant"})
		return
	}

	// Log admin action
	h.db.Exec(`
		INSERT INTO yt_admin_audit_logs (admin_user_id, action, resource_type, resource_id, tenant_id, details)
		VALUES ($1, 'suspend_tenant', 'tenant', $2, $2, '{}')
	`, adminID, tenantID)

	log.Printf("[ADMIN] Tenant suspended: %s by admin %s", tenantID, adminID)
	json.NewEncoder(w).Encode(TenantActionResponse{Success: true, Message: "Tenant suspended"})
}

// ActivateTenant activates a suspended tenant
func (h *AdminTenantsHandler) ActivateTenant(w http.ResponseWriter, r *http.Request) {
	tenantID := strings.TrimPrefix(r.URL.Path, "/api/admin/tenants/")
	tenantID = strings.TrimSuffix(tenantID, "/activate")

	adminID, _ := middleware.GetAdminUserID(r.Context())

	_, err := h.db.Exec(`
		UPDATE yt_customers SET onboarding_status = 'completed' WHERE id = $1
	`, tenantID)

	if err != nil {
		log.Printf("[ERROR] Failed to activate tenant: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(TenantActionResponse{Success: false, Error: "Failed to activate tenant"})
		return
	}

	// Log admin action
	h.db.Exec(`
		INSERT INTO yt_admin_audit_logs (admin_user_id, action, resource_type, resource_id, tenant_id, details)
		VALUES ($1, 'activate_tenant', 'tenant', $2, $2, '{}')
	`, adminID, tenantID)

	log.Printf("[ADMIN] Tenant activated: %s by admin %s", tenantID, adminID)
	json.NewEncoder(w).Encode(TenantActionResponse{Success: true, Message: "Tenant activated"})
}

// DeleteTenant deletes a tenant (soft delete)
func (h *AdminTenantsHandler) DeleteTenant(w http.ResponseWriter, r *http.Request) {
	tenantID := strings.TrimPrefix(r.URL.Path, "/api/admin/tenants/")

	adminID, _ := middleware.GetAdminUserID(r.Context())

	// Soft delete - mark as deleted
	_, err := h.db.Exec(`
		UPDATE yt_customers SET onboarding_status = 'deleted' WHERE id = $1
	`, tenantID)

	if err != nil {
		log.Printf("[ERROR] Failed to delete tenant: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(TenantActionResponse{Success: false, Error: "Failed to delete tenant"})
		return
	}

	// Log admin action
	h.db.Exec(`
		INSERT INTO yt_admin_audit_logs (admin_user_id, action, resource_type, resource_id, tenant_id, details)
		VALUES ($1, 'delete_tenant', 'tenant', $2, $2, '{}')
	`, adminID, tenantID)

	log.Printf("[ADMIN] Tenant deleted: %s by admin %s", tenantID, adminID)
	json.NewEncoder(w).Encode(TenantActionResponse{Success: true, Message: "Tenant deleted"})
}

func formatCurrency(amount float64) string {
	if amount == 0 {
		return "$0.00"
	}
	return "$" + formatFloat(amount)
}

func formatFloat(f float64) string {
	return strings.TrimRight(strings.TrimRight("0.00", "0"), ".")
}


