package handlers

import (
	"encoding/json"
	"net/http"

	"yukti/internal/hiddencosts"
	"yukti/internal/models"
)

type HiddenCostsHandler struct {
	detector *hiddencosts.Detector
}

func NewHiddenCostsHandler(detector *hiddencosts.Detector) *HiddenCostsHandler {
	return &HiddenCostsHandler{detector: detector}
}

func (h *HiddenCostsHandler) ListFindings(w http.ResponseWriter, r *http.Request) {
	// Use models package to prevent unused import
	_ = models.Resource{}
	
	tenantID := r.Context().Value("tenant_id").(string)
	
	// Mock resources for demo
	resources := []hiddencosts.Resource{
		{
			ARN:    "arn:aws:rds:us-east-1:123456789012:db:prod-db",
			Type:   "rds",
			Region: "us-east-1",
			Metadata: map[string]interface{}{
				"multi_az": true,
			},
		},
		{
			ARN:    "arn:aws:ec2:us-east-1:123456789012:natgateway/nat-abc123",
			Type:   "nat_gateway",
			Region: "us-east-1",
			Metadata: map[string]interface{}{
				"data_processed_gb": 1500.0,
			},
		},
	}

	findings, err := h.detector.RunDetection(r.Context(), tenantID, resources)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	totalSavings := 0.0
	categories := make(map[string]int)
	for _, f := range findings {
		totalSavings += f.EstimatedSavings
		categories[string(f.Category)]++
	}

	response := map[string]interface{}{
		"findings":                 findings,
		"total_findings":           len(findings),
		"total_estimated_savings":  totalSavings,
		"categories":               categories,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
