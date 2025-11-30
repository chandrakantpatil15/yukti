package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"yukti/internal/whitelist"

	"github.com/gorilla/mux"
)

type WhitelistHandler struct {
	service *whitelist.Service
}

func NewWhitelistHandler(service *whitelist.Service) *WhitelistHandler {
	return &WhitelistHandler{service: service}
}

func (h *WhitelistHandler) CreateWhitelist(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(string)
	userEmail := r.Context().Value("user_email").(string)

	var req whitelist.CreateWhitelistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	wl, err := h.service.CreateWhitelist(r.Context(), tenantID, userEmail, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"id":                  wl.ID,
		"status":              wl.Status,
		"cost_impact_monthly": wl.CostImpactMonthly,
		"message":             "Resource whitelisted successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *WhitelistHandler) ListWhitelists(w http.ResponseWriter, r *http.Request) {
	tenantIDVal := r.Context().Value("tenant_id")
	tenantID, _ := tenantIDVal.(string)
	if tenantID == "" {
		// fallback to query param for robustness
		tenantID = r.URL.Query().Get("tenant_id")
	}
	if tenantID == "" {
		http.Error(w, "tenant_id required", http.StatusBadRequest)
		return
	}

	whitelists, err := h.service.ListWhitelists(r.Context(), tenantID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	totalImpact := 0.0
	summary := map[string]int{"active": 0, "expired": 0, "pending_approval": 0}
	for _, wl := range whitelists {
		totalImpact += wl.CostImpactMonthly
		summary[string(wl.Status)]++
	}

	response := map[string]interface{}{
		"whitelists":        whitelists,
		"total_cost_impact": totalImpact,
		"summary":           summary,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *WhitelistHandler) RevokeWhitelist(w http.ResponseWriter, r *http.Request) {
	// Prefer path param: /api/whitelists/{id}
	vars := mux.Vars(r)
	whitelistID := strings.TrimSpace(vars["id"])
	if whitelistID == "" {
		// Fallback to query param for backward compatibility
		whitelistID = strings.TrimSpace(r.URL.Query().Get("id"))
	}

	if whitelistID == "" {
		http.Error(w, "id required", http.StatusBadRequest)
		return
	}

	userEmail, _ := r.Context().Value("user_email").(string)

	var req struct {
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	if err := h.service.RevokeWhitelist(r.Context(), whitelistID, userEmail, req.Reason); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Whitelist revoked successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
