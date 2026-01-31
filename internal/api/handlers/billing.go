package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
	"yukti/internal/models"
	"yukti/internal/services"

	"github.com/gorilla/mux"
)

type BillingHandler struct {
	billingService *services.BillingService
}

func NewBillingHandler(db *sql.DB) (*BillingHandler, error) {
	billingService := services.NewBillingService(db)
	return &BillingHandler{billingService: billingService}, nil
}

// ListBillings - GET /api/admin/billing
func (h *BillingHandler) ListBillings(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	if page == 0 {
		page = 1
	}
	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit == 0 {
		limit = 20
	}
	status := q.Get("status")
	tenantID := q.Get("tenant_id")

	billings, total, err := h.billingService.ListBillings(page, limit, status, tenantID)
	if err != nil {
		http.Error(w, "Failed to fetch billings", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"billings": billings,
		"total":    total,
		"page":     page,
		"limit":    limit,
	})
}

// GetBilling - GET /api/admin/billing/:id
func (h *BillingHandler) GetBilling(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid billing ID", http.StatusBadRequest)
		return
	}

	billing, err := h.billingService.GetBilling(id)
	if err != nil {
		http.Error(w, "Billing not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(billing)
}

// CreateBilling - POST /api/admin/billing
func (h *BillingHandler) CreateBilling(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TenantID   int     `json:"tenant_id"`
		Plan       string  `json:"plan"`
		Amount     float64 `json:"amount"`
		Currency   string  `json:"currency"`
		DueDate    string  `json:"due_date"`
		InvoiceURL string  `json:"invoice_url"`
		Notes      string  `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Parse due date
	dueDate, err := time.Parse("2006-01-02", req.DueDate)
	if err != nil {
		http.Error(w, "Invalid due date format", http.StatusBadRequest)
		return
	}

	createReq := &models.CreateBillingRequest{
		TenantID: req.TenantID,
		Plan:     req.Plan,
		Amount:   req.Amount,
		Currency: req.Currency,
		DueDate:  dueDate,
		Notes:    req.Notes,
	}

	billing, err := h.billingService.CreateBilling(createReq)
	if err != nil {
		http.Error(w, "Failed to create billing", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(billing)
}

// UpdateBilling - PUT /api/admin/billing/:id
func (h *BillingHandler) UpdateBilling(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid billing ID", http.StatusBadRequest)
		return
	}

	var req models.UpdateBillingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	billing, err := h.billingService.UpdateBilling(id, &req)
	if err != nil {
		http.Error(w, "Failed to update billing", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(billing)
}

// DeleteBilling - DELETE /api/admin/billing/:id
func (h *BillingHandler) DeleteBilling(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid billing ID", http.StatusBadRequest)
		return
	}

	if err := h.billingService.DeleteBilling(id); err != nil {
		http.Error(w, "Failed to delete billing", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Billing deleted successfully"})
}

// MarkAsPaid - POST /api/admin/billing/:id/mark-paid
func (h *BillingHandler) MarkAsPaid(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid billing ID", http.StatusBadRequest)
		return
	}

	billing, err := h.billingService.MarkAsPaid(id)
	if err != nil {
		http.Error(w, "Failed to mark as paid", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(billing)
}

// GetBillingStats - GET /api/admin/billing/stats
func (h *BillingHandler) GetBillingStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.billingService.GetBillingStats()
	if err != nil {
		http.Error(w, "Failed to fetch stats", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// GetBillingInfo - GET /api/v1/billing/info (tenant-scoped)
func (h *BillingHandler) GetBillingInfo(w http.ResponseWriter, r *http.Request) {
	// Try header first, then query param
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		tenantID = r.URL.Query().Get("tenant_id")
	}

	if tenantID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "tenant_id required",
		})
		return
	}

	// Use existing ListBillings to fetch invoices for tenant
	billings, _, err := h.billingService.ListBillings(1, 50, "", tenantID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Failed to fetch billing info",
		})
		return
	}

	// Minimal response compatible with frontend expected shape
	resp := map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"subscription_tier": "FREE",
			"trial_ends_at":     nil,
			"subscription":      nil,
			"invoices":          billings,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ExportBillings - GET /api/admin/billing/export
func (h *BillingHandler) ExportBillings(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	status := q.Get("status")
	tenantID := q.Get("tenant_id")

	billings, _, err := h.billingService.ListBillings(1, 10000, status, tenantID)
	if err != nil {
		http.Error(w, "Failed to export billings", http.StatusInternalServerError)
		return
	}

	// CSV header
	csv := "ID,Tenant,Plan,Amount,Currency,Status,Due Date,Paid Date,Invoice URL\n"
	for _, b := range billings {
		paidDate := ""
		if b.PaidDate != nil {
			paidDate = b.PaidDate.Format("2006-01-02")
		}
		csv += strconv.Itoa(b.ID) + "," + strconv.Itoa(b.TenantID) + "," + b.Plan + "," +
			strconv.FormatFloat(b.Amount, 'f', 2, 64) + "," + b.Currency + "," +
			b.Status + "," + b.DueDate.Format("2006-01-02") + "," + paidDate + "," +
			b.InvoiceURL + "\n"
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=billings.csv")
	w.Write([]byte(csv))
}
