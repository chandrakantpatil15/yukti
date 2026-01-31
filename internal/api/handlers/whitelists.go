package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"yukti/internal/api/middleware"
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
	log.Printf("[DEBUG] CreateWhitelist called")

	tenantIDVal := r.Context().Value(middleware.TenantIDKey)
	log.Printf("[DEBUG] tenantIDVal: %v, type: %T", tenantIDVal, tenantIDVal)
	tenantID, ok := tenantIDVal.(int)
	if !ok {
		log.Printf("[ERROR] Invalid tenant context: %v", tenantIDVal)
		http.Error(w, "Invalid tenant context", http.StatusInternalServerError)
		return
	}
	log.Printf("[DEBUG] tenantID: %d", tenantID)

	userEmailVal := r.Context().Value(middleware.EmailKey)
	log.Printf("[DEBUG] userEmailVal: %v, type: %T", userEmailVal, userEmailVal)
	userEmail, ok := userEmailVal.(string)
	if !ok {
		log.Printf("[ERROR] Invalid user context: %v", userEmailVal)
		http.Error(w, "Invalid user context", http.StatusInternalServerError)
		return
	}
	log.Printf("[DEBUG] userEmail: %s", userEmail)

	var req whitelist.CreateWhitelistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[ERROR] JSON decode error: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Printf("[DEBUG] Request decoded: %+v", req)

	wl, err := h.service.CreateWhitelist(r.Context(), strconv.Itoa(tenantID), userEmail, req)
	if err != nil {
		log.Printf("[ERROR] Service error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("[DEBUG] Whitelist created: %+v", wl)

	response := map[string]interface{}{
		"id":                  wl.ID,
		"status":              wl.Status,
		"cost_impact_monthly": wl.CostImpactMonthly,
		"message":             "Resource whitelisted successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
	log.Printf("[DEBUG] Response sent successfully")
}

func (h *WhitelistHandler) ListWhitelists(w http.ResponseWriter, r *http.Request) {
	tenantIDVal := r.Context().Value(middleware.TenantIDKey)
	tenantID, ok := tenantIDVal.(int)
	if !ok {
		http.Error(w, "Invalid tenant context", http.StatusInternalServerError)
		return
	}

	whitelists, err := h.service.ListWhitelists(r.Context(), strconv.Itoa(tenantID))
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
		"success":           true,
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

	userEmailVal := r.Context().Value(middleware.EmailKey)
	userEmail, _ := userEmailVal.(string)

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
