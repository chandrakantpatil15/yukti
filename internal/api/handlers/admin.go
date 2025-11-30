package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type AdminHandler struct {
	db *sql.DB
}

func NewAdminHandler(db *sql.DB) *AdminHandler {
	return &AdminHandler{db: db}
}

// GetCustomers returns a paginated list of all customers with their total savings and findings count
// Query parameters:
//   - page: (optional) Page number, starting from 1 (default: 1)
//   - per_page: (optional) Number of items per page, max 100 (default: 20)
//   - search: (optional) Search company name or email
func (h *AdminHandler) GetCustomers(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] Admin GetCustomers called from IP: %s", r.RemoteAddr)
	w.Header().Set("Content-Type", "application/json")

	// Parse pagination parameters
	pageParam := r.URL.Query().Get("page")
	perPageParam := r.URL.Query().Get("per_page")
	searchTerm := strings.TrimSpace(r.URL.Query().Get("search"))

	page := 1
	perPage := 20

	if pageParam != "" {
		if v, err := strconv.Atoi(pageParam); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{"error": "invalid page parameter"})
			return
		} else if v < 1 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{"error": "page must be greater than 0"})
			return
		} else {
			page = v
		}
	}

	if perPageParam != "" {
		if v, err := strconv.Atoi(perPageParam); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{"error": "invalid per_page parameter"})
			return
		} else if v < 1 || v > 100 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{"error": "per_page must be between 1 and 100"})
			return
		} else {
			perPage = v
		}
	}

	// Count total for pagination
	countQuery := `
		SELECT COUNT(DISTINCT c.id)
		FROM yt_customers c
	`
	var countArgs []interface{}
	if searchTerm != "" {
		countQuery += ` WHERE c.company_name ILIKE $1 OR c.email ILIKE $1`
		countArgs = append(countArgs, "%"+searchTerm+"%")
	}

	var total int
	if err := h.db.QueryRow(countQuery, countArgs...).Scan(&total); err != nil {
		log.Printf("[ERROR] Failed to count customers: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Failed to fetch customers"})
		return
	}

	// Build main query
	query := `
		SELECT c.id, c.tenant_id, c.company_name, c.email, c.onboarding_status,
		       c.created_at, c.completed_at,
		       COALESCE(SUM(f.estimated_savings), 0) as total_savings,
		       COUNT(f.id) as findings_count
		FROM yt_customers c
		LEFT JOIN yt_hidden_cost_findings f ON c.tenant_id = f.tenant_id
	`
	var args []interface{}

	if searchTerm != "" {
		query += ` WHERE c.company_name ILIKE $1 OR c.email ILIKE $1`
		args = append(args, "%"+searchTerm+"%")
	}

	query += `
		GROUP BY c.id, c.tenant_id, c.company_name, c.email, c.onboarding_status, c.created_at, c.completed_at
		ORDER BY c.created_at DESC
		LIMIT $` + strconv.Itoa(len(args)+1) + ` OFFSET $` + strconv.Itoa(len(args)+2)

	args = append(args, perPage, (page-1)*perPage)

	rows, err := h.db.Query(query, args...)
	if err != nil {
		log.Printf("[ERROR] Failed to query customers: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Failed to fetch customers",
		})
		return
	}
	defer rows.Close()

	customers := []map[string]interface{}{}
	for rows.Next() {
		var id, tenantID, companyName, email, status string
		var createdAt, completedAt sql.NullTime
		var totalSavings float64
		var findingsCount int

		if err := rows.Scan(&id, &tenantID, &companyName, &email, &status, &createdAt, &completedAt, &totalSavings, &findingsCount); err != nil {
			log.Printf("[WARN] Failed to scan customer row: %v", err)
			continue
		}

		customers = append(customers, map[string]interface{}{
			"id":             id,
			"tenant_id":      tenantID,
			"company_name":   companyName,
			"email":          email,
			"status":         status,
			"created_at":     createdAt.Time,
			"completed_at":   completedAt.Time,
			"total_savings":  totalSavings,
			"findings_count": findingsCount,
		})
	}

	if err := rows.Err(); err != nil {
		log.Printf("[ERROR] Error iterating customer rows: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Error reading customer data",
		})
		return
	}

	log.Printf("[INFO] Successfully fetched %d customers (page=%d, per_page=%d, total=%d)", len(customers), page, perPage, total)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"customers": customers,
		"meta": map[string]interface{}{
			"page":     page,
			"per_page": perPage,
			"total":    total,
			"total_pages": func() int {
				if perPage == 0 {
					return 0
				}
				tp := total / perPage
				if total%perPage != 0 {
					tp++
				}
				return tp
			}(),
		},
	})
}

func (h *AdminHandler) ImpersonateTenant(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] Admin impersonation request from IP: %s", r.RemoteAddr)
	var req struct {
		TenantID string `json:"tenant_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[ERROR] Failed to decode impersonation request: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Invalid request"})
		return
	}

	// Log impersonation for security audit
	adminUser := r.Header.Get("X-Admin-User")
	// Extract IP without port
	ipAddr := r.RemoteAddr
	if idx := strings.LastIndex(ipAddr, ":"); idx != -1 {
		ipAddr = ipAddr[:idx]
	}
	log.Printf("[WARN] Admin %s impersonating tenant %s from IP %s", adminUser, req.TenantID, ipAddr)
	_, err := h.db.Exec(`
		INSERT INTO yt_audit_logs (user_id, action, resource_type, ip_address)
		VALUES ($1, $2, $3, $4)
	`, adminUser, "impersonate_tenant", "customer", ipAddr)
	if err != nil {
		log.Printf("[ERROR] Failed to log impersonation to audit table: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"tenant_id": req.TenantID,
	})
}

func (h *AdminHandler) SyncPricing(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("[INFO] Admin pricing sync requested from IP: %s", r.RemoteAddr)

	// TODO: Trigger background job to:
	// 1. Fetch latest AWS pricing data
	// 2. Update yt_aws_pricing table
	// 3. Emit metrics

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Pricing sync queued for processing",
	})
}

func (h *AdminHandler) SyncInventory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("[INFO] Admin inventory sync requested from IP: %s", r.RemoteAddr)

	var req struct {
		TenantID int `json:"tenant_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.TenantID == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "tenant_id required",
		})
		return
	}

	// TODO: Trigger background job to:
	// 1. Sync AWS resources for tenant
	// 2. Update yt_tenant_resources
	// 3. Run cost detectors

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"message":   "Inventory sync queued for processing",
		"tenant_id": req.TenantID,
	})
}

func (h *AdminHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] Admin GetMetrics called from IP: %s", r.RemoteAddr)
	var totalCustomers, activeTrials int
	var totalSavings, mrr float64

	if err := h.db.QueryRow("SELECT COUNT(*) FROM yt_customers").Scan(&totalCustomers); err != nil {
		log.Printf("[ERROR] Failed to get total customers: %v", err)
	}
	if err := h.db.QueryRow("SELECT COUNT(*) FROM yt_customers WHERE onboarding_status = 'in_progress'").Scan(&activeTrials); err != nil {
		log.Printf("[ERROR] Failed to get active trials: %v", err)
	}
	if err := h.db.QueryRow("SELECT COALESCE(SUM(estimated_savings), 0) FROM yt_hidden_cost_findings").Scan(&totalSavings); err != nil {
		log.Printf("[ERROR] Failed to get total savings: %v", err)
	}

	// Calculate MRR (simplified: $99 per completed customer)
	var completedCustomers int
	if err := h.db.QueryRow("SELECT COUNT(*) FROM yt_customers WHERE onboarding_status = 'completed'").Scan(&completedCustomers); err != nil {
		log.Printf("[ERROR] Failed to get completed customers: %v", err)
	}
	mrr = float64(completedCustomers) * 99.0
	log.Printf("[INFO] Metrics: customers=%d, trials=%d, savings=%.2f, mrr=%.2f", totalCustomers, activeTrials, totalSavings, mrr)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"metrics": map[string]interface{}{
			"total_customers": totalCustomers,
			"total_savings":   totalSavings,
			"active_trials":   activeTrials,
			"mrr":             mrr,
		},
	})
}
