package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"yukti/internal/api/middleware"
	"yukti/internal/models"
)

type RecommendationHandler struct {
	db *sql.DB
}

func NewRecommendationHandler(db *sql.DB) *RecommendationHandler {
	return &RecommendationHandler{db: db}
}

func (h *RecommendationHandler) ListRecommendations(w http.ResponseWriter, r *http.Request) {
	// Use models package to prevent unused import
	_ = models.OptimizationRecommendation{}
	
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Unauthorized"})
		return
	}

	status := r.URL.Query().Get("status")
	if status == "" {
		status = "pending"
	}

	rows, err := h.db.Query(`
		SELECT id, recommendation_type, current_cost, optimized_cost, 
		       monthly_savings, confidence_score, status, created_at
		FROM yt_tenant_recommendations
		WHERE tenant_id = $1 AND status = $2
		ORDER BY monthly_savings DESC
		LIMIT 100`,
		tenantID, status,
	)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": err.Error()})
		return
	}
	defer rows.Close()

	recommendations := []map[string]interface{}{}
	totalSavings := 0.0

	for rows.Next() {
		var id int
		var recType string
		var currentCost, optimizedCost, savings, confidence float64
		var status, createdAt string

		rows.Scan(&id, &recType, &currentCost, &optimizedCost, &savings, &confidence, &status, &createdAt)
		
		recommendations = append(recommendations, map[string]interface{}{
			"id":                id,
			"type":              recType,
			"current_cost":      currentCost,
			"optimized_cost":    optimizedCost,
			"monthly_savings":   savings,
			"confidence_score":  confidence,
			"status":            status,
			"created_at":        createdAt,
		})
		totalSavings += savings
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"recommendations": recommendations,
			"total_savings":   totalSavings,
			"count":           len(recommendations),
		},
	})
}
