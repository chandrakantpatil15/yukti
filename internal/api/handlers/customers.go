package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"yukti/internal/api/middleware"

	"github.com/google/uuid"
)

type CustomerHandler struct {
	db *sql.DB
}

func NewCustomerHandler(db *sql.DB) *CustomerHandler {
	return &CustomerHandler{db: db}
}

func (h *CustomerHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	// SECURITY FIX: Get tenant_id from JWT token, not query parameter
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		log.Printf("[WARN] GetDashboard called without valid JWT from IP: %s", r.RemoteAddr)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Unauthorized"})
		return
	}
	log.Printf("[INFO] GetDashboard called for tenant: %d from IP: %s", tenantID, r.RemoteAddr)

	// Get findings
	var totalSavings float64
	var findingsCount int
	log.Printf("[DEBUG] Fetching findings for tenant: %d", tenantID)
	// Convert tenant_id to string for yt_hidden_cost_findings.tenant_id (VARCHAR)
	tenantIDStr := "tenant-" + strconv.Itoa(tenantID)
	err := h.db.QueryRow(`
		SELECT COALESCE(SUM(estimated_savings), 0), COUNT(*) 
		FROM yt_hidden_cost_findings 
		WHERE tenant_id = $1
	`, tenantIDStr).Scan(&totalSavings, &findingsCount)
	if err != nil {
		log.Printf("[ERROR] Failed to get findings for tenant %s: %v", tenantID, err)
	}

	// Get budget
	var budgetAmount, currentSpend float64
	log.Printf("[DEBUG] Fetching budget for tenant: %d", tenantID)
	err = h.db.QueryRow(`
		SELECT COALESCE(amount, 0), COALESCE(current_spend, 0) 
		FROM yt_budgets 
		WHERE tenant_id = $1 AND status = 'active' 
		LIMIT 1
	`, tenantIDStr).Scan(&budgetAmount, &currentSpend)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("[ERROR] Failed to get budget for tenant %s: %v", tenantID, err)
	}

	// Get RI recommendations
	var riSavings float64
	log.Printf("[DEBUG] Fetching RI recommendations for tenant: %d", tenantID)
	err = h.db.QueryRow(`
		SELECT COALESCE(SUM(monthly_savings), 0) 
		FROM yt_ri_recommendations 
		WHERE tenant_id = $1
	`, tenantIDStr).Scan(&riSavings)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("[ERROR] Failed to get RI recommendations for tenant %s: %v", tenantID, err)
	}

	log.Printf("[INFO] Dashboard data for tenant %d: savings=%.2f, findings=%d, budget=%.2f", tenantID, totalSavings, findingsCount, budgetAmount)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"total_savings":  totalSavings,
			"findings_count": findingsCount,
			"budget_amount":  budgetAmount,
			"current_spend":  currentSpend,
			"ri_savings":     riSavings,
		},
	})
}

// GetFindings returns a paginated list of findings for a tenant
// Query parameters:
//   - tenant_id: (required) The ID of the tenant to get findings for
//   - category: (optional) Filter findings by category
//   - severity: (optional) Filter findings by severity
//   - page: (optional) Page number, starting from 1 (default: 1)
//   - per_page: (optional) Number of items per page, max 500 (default: 50)
func (h *CustomerHandler) GetFindings(w http.ResponseWriter, r *http.Request) {
	// SECURITY FIX: Get tenant_id from JWT token, not query parameter
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		log.Printf("[WARN] GetFindings called without valid JWT from IP: %s", r.RemoteAddr)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Unauthorized"})
		return
	}
	category := r.URL.Query().Get("category")
	severity := r.URL.Query().Get("severity")

	// Parse pagination parameters
	pageParam := r.URL.Query().Get("page")
	perPageParam := r.URL.Query().Get("per_page")
	page := 1
	perPage := 50

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
		} else if v < 1 || v > 500 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{"error": "per_page must be between 1 and 500"})
			return
		} else {
			perPage = v
		}
	}
	log.Printf("[INFO] GetFindings called for tenant: %d, category: %s, severity: %s from IP: %s", tenantID, category, severity, r.RemoteAddr)

	// Convert tenant_id to string for yt_hidden_cost_findings.tenant_id (VARCHAR)
	tenantIDStr := "tenant-" + strconv.Itoa(tenantID)

	query := `
		SELECT id, detector_name, category, severity, title, description, 
		       resource_arn, estimated_savings, confidence, created_at
		FROM yt_hidden_cost_findings
		WHERE tenant_id = $1
	`
	args := []interface{}{tenantIDStr}

	if category != "" {
		query += " AND category = $2"
		args = append(args, category)
	}
	if severity != "" {
		if category != "" {
			query += " AND severity = $3"
		} else {
			query += " AND severity = $2"
		}
		args = append(args, severity)
	}

	query += " ORDER BY estimated_savings DESC"

	// Count total for pagination
	countQuery := `SELECT COUNT(*) FROM yt_hidden_cost_findings WHERE tenant_id = $1`
	countArgs := []interface{}{tenantIDStr}
	if category != "" {
		countQuery += " AND category = $2"
		countArgs = append(countArgs, category)
	}
	if severity != "" {
		if category != "" {
			countQuery += " AND severity = $3"
		} else {
			countQuery += " AND severity = $2"
		}
		countArgs = append(countArgs, severity)
	}
	var total int
	_ = h.db.QueryRow(countQuery, countArgs...).Scan(&total)

	// Apply pagination
	offset := (page - 1) * perPage
	query += " LIMIT $" + strconv.Itoa(len(args)+1) + " OFFSET $" + strconv.Itoa(len(args)+2)
	args = append(args, perPage, offset)

	log.Printf("[DEBUG] Executing findings query with %d args", len(args))
	rows, err := h.db.Query(query, args...)
	if err != nil {
		log.Printf("[ERROR] Failed to query findings for tenant %s: %v", tenantID, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	defer rows.Close()

	findings := []map[string]interface{}{}
	for rows.Next() {
		var id, detectorName, category, severity, title, description, resourceARN string
		var estimatedSavings, confidence float64
		var createdAt sql.NullTime

		if err := rows.Scan(&id, &detectorName, &category, &severity, &title, &description, &resourceARN, &estimatedSavings, &confidence, &createdAt); err != nil {
			log.Printf("[WARN] Failed to scan finding row: %v", err)
			continue
		}

		findings = append(findings, map[string]interface{}{
			"id":                id,
			"detector_name":     detectorName,
			"category":          category,
			"severity":          severity,
			"title":             title,
			"description":       description,
			"resource_arn":      resourceARN,
			"estimated_savings": estimatedSavings,
			"confidence":        confidence,
			"created_at":        createdAt.Time.Format("2006-01-02 15:04:05"),
		})
	}

	log.Printf("[INFO] Returning %d findings for tenant %d (page=%d, per_page=%d, total=%d)", len(findings), tenantID, page, perPage, total)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"findings": findings,
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

func (h *CustomerHandler) CreateCustomer(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] CreateCustomer called from IP: %s", r.RemoteAddr)
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		CompanyName string `json:"company_name"`
		Email       string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[ERROR] Failed to decode create customer request: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Invalid JSON format",
		})
		return
	}

	// Validate input
	if req.CompanyName == "" || req.Email == "" {
		log.Printf("[WARN] CreateCustomer validation failed: missing company_name or email")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "company_name and email are required",
		})
		return
	}

	id := uuid.New().String()
	tenantID := "tenant-" + uuid.New().String()[:8]
	log.Printf("[INFO] Creating new customer: %s (tenant: %s, email: %s)", req.CompanyName, tenantID, req.Email)

	_, err := h.db.Exec(`
		INSERT INTO yt_customers (id, tenant_id, company_name, email, onboarding_status, onboarding_step, created_at)
		VALUES ($1, $2, $3, $4, 'pending', 'company_info', NOW())
	`, id, tenantID, req.CompanyName, req.Email)

	if err != nil {
		log.Printf("[ERROR] Failed to create customer %s: %v", req.CompanyName, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	log.Printf("[INFO] Successfully created customer: %s with tenant_id: %s", req.CompanyName, tenantID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"tenant_id": tenantID,
		"id":        id,
	})
}
